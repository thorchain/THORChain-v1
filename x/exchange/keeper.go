package exchange

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// The exchange keeper contains one buy and one sell orderbook for each token pair.
// Each orderbook contains a list of orders, sorted by best price, then time
// (first in, first out).
type Keeper struct {
	storeKey sdk.StoreKey // The (unexposed) key used to access the store from the Context.

	bankKeeper bank.Keeper

	codespace sdk.CodespaceType

	cdc *wire.Codec
}

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, bankKeeper bank.Keeper, codespace sdk.CodespaceType) Keeper {
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)
	return Keeper{key, bankKeeper, codespace, cdc}
}

// getOrderBook returns the orderbook for the given token pair. If no order book exists for these tokens right now,
// a new (empty) orderbook will be returned.
func (k Keeper) getOrderBook(
	ctx sdk.Context, kind OrderKind, amountDenom string, priceDenom string) OrderBook {
	key := MakeKeyOrderBook(kind, amountDenom, priceDenom)

	store := ctx.KVStore(k.storeKey)
	valueBytes := store.Get(key)

	if valueBytes == nil {
		return NewOrderBook(kind, amountDenom, priceDenom)
	}

	orderBook := new(OrderBook)
	k.cdc.UnmarshalBinary(valueBytes, &orderBook)

	return *orderBook
}

// setOrderBook saves the order book to the store
func (k Keeper) setOrderBook(ctx sdk.Context, orderBook OrderBook) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinary(orderBook)
	if err != nil {
		panic(err)
	}
	store.Set(orderBook.Key, bz)
}

// processLimitOrder processes a limit order. After error checking, it tries to
// execute the order with existing trades – if that is not possible, a new entry
// for this limit order will be created in the corresponding order book.
// nolint gocyclo
func (k Keeper) processLimitOrder(
	ctx sdk.Context, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin,
	expiresAt time.Time) (int64, sdk.Error) {

	// error if already expired
	if expiresAt.Before(time.Now()) {
		return -1, ErrOrderExpired(k.codespace)
	}

	// error if kind not supported
	if kind != BuyOrder && kind != SellOrder {
		return -1, ErrInvalidKind(k.codespace)
	}

	// error if amount and price denom are the same
	if amount.Denom == price.Denom {
		return -1, ErrSameDenom(k.codespace)
	}

	// error if amount negative
	if !amount.IsPositive() {
		return -1, ErrAmountNotPositive(k.codespace)
	}

	// error if price negative
	if !price.IsPositive() {
		return -1, ErrPriceNotPositive(k.codespace)
	}

	// check if enough coins to place order
	totalPrice := sdk.Coin{price.Denom, amount.Amount.Mul(price.Amount)}
	if kind == BuyOrder && !k.bankKeeper.HasCoins(ctx, sender, sdk.Coins{totalPrice}) {
		return -1, sdk.ErrInsufficientCoins(fmt.Sprintf(
			"Must have at least %v to place this buy limit order", totalPrice))
	}
	if kind == SellOrder && !k.bankKeeper.HasCoins(ctx, sender, sdk.Coins{amount}) {
		return -1, sdk.ErrInsufficientCoins(fmt.Sprintf(
			"Must have at least %v to place this sell limit order", amount))
	}

	// fill order if possible
	amount, err := k.fillOrderIfPossible(ctx, sender, kind, amount, price, expiresAt)
	if err != nil {
		return -1, err
	}

	// store unfilled order
	return k.storeUnfilledLimitOrder(ctx, sender, kind, amount, price, expiresAt)
}

// k. tries to fill the order. Returns the amount that could not be filled.
func (k Keeper) fillOrderIfPossible(
	ctx sdk.Context, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin, expiresAt time.Time,
) (sdk.Coin, sdk.Error) {
	// get matching order book to fill the order
	matchingKind := SellOrder
	if kind == SellOrder {
		matchingKind = BuyOrder
	}
	orderBook := k.getOrderBook(ctx, matchingKind, amount.Denom, price.Denom)

	// remove expired
	orderBook.RemoveExpiredLimitOrders()

	// unfilled amount
	unfilledAmt := amount

	var err sdk.Error

	for i, storedOrder := range orderBook.Orders {
		// end loop if storedOrder cannot fill our order => since orders are sorted, this means there will be no more
		// match
		ok, fillAmount, fillPrice := storedOrder.DoesFill(kind, unfilledAmt, price)
		if !ok {
			break
		}

		fillTotalPrice := sdk.Coin{fillPrice.Denom, fillAmount.Amount.Mul(fillPrice.Amount)}

		// skip this stored order if sender does not have enough coins
		// TODO: better handling of this case => delete order, lock coins at order creation lock margin?
		if !k.hasOrderSenderEnoughCoins(ctx, storedOrder, fillAmount, fillTotalPrice) {
			continue
		}

		var buyer, seller sdk.AccAddress
		if kind == BuyOrder {
			buyer = sender
			seller = storedOrder.Sender
		} else {
			buyer = storedOrder.Sender
			seller = sender
		}

		err := k.exchangeCoins(ctx, buyer, seller, fillAmount, fillTotalPrice)

		if err != nil {
			break
		}

		// update unfilled amount
		unfilledAmt = unfilledAmt.Minus(fillAmount)

		// replace the amount of the stored order with the remaining part
		orderBook.Orders[i].Amount = storedOrder.Amount.Minus(fillAmount)
	}

	orderBook.RemoveFilledLimitOrders()

	k.setOrderBook(ctx, orderBook)

	return unfilledAmt, err
}

func (k Keeper) hasOrderSenderEnoughCoins(ctx sdk.Context, storedOrder LimitOrder, totalAmount, totalPrice sdk.Coin,
) bool {
	if storedOrder.Kind == BuyOrder {
		return k.bankKeeper.HasCoins(ctx, storedOrder.Sender, sdk.Coins{totalPrice})
	}

	return k.bankKeeper.HasCoins(ctx, storedOrder.Sender, sdk.Coins{totalAmount})
}

func (k Keeper) exchangeCoins(ctx sdk.Context, buyer, seller sdk.AccAddress, totalAmount, totalPrice sdk.Coin,
) sdk.Error {
	_, err := k.bankKeeper.SendCoins(ctx, seller, buyer, sdk.Coins{totalAmount})
	if err != nil {
		return err
	}

	_, err = k.bankKeeper.SendCoins(ctx, buyer, seller, sdk.Coins{totalPrice})
	if err != nil {
		return err
	}

	return nil
}

// storeUnfilledLimitOrder creates a new limit order, finds the corresponding order book, adds the limit order
// to the right place and saves the orderbook. Returns the id of the new order.
func (k Keeper) storeUnfilledLimitOrder(
	ctx sdk.Context, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin, expiresAt time.Time,
) (int64, sdk.Error) {
	if amount.IsZero() {
		return -1, nil
	}

	// get orderbook
	orderBook := k.getOrderBook(ctx, kind, amount.Denom, price.Denom)

	// create a new limit order and then add it to the orderbook
	newOrderID, err := k.getNewOrderID(ctx)
	if err != nil {
		return -1, err
	}

	limitOrder := NewLimitOrder(newOrderID, sender, kind, amount, price, expiresAt)

	err = orderBook.AddLimitOrder(limitOrder) // TODO
	if err != nil {
		return -1, err
	}

	k.setOrderBook(ctx, orderBook)

	return newOrderID, nil
}

// GetCLP - returns the clp
// func (k Keeper) GetCLP(ctx sdk.Context, ticker string) *CLP { // TODO
// 	store := ctx.KVStore(k.storeKey)
// 	valueBytes := store.Get(MakeCLPStoreKey(ticker))

// 	clp := new(CLP)
// 	k.cdc.UnmarshalBinary(valueBytes, &clp)

// 	return clp
// }

func (k Keeper) setInitialOrderID(ctx sdk.Context, orderID int64) sdk.Error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(KeyNextOrderID)
	if bz != nil {
		return ErrInvalidGenesis(k.codespace, "Initial OrderId already set")
	}
	bz = k.cdc.MustMarshalBinary(orderID)
	store.Set(KeyNextOrderID, bz)
	return nil
}

// Get the last used proposal ID
func (k Keeper) getLastOrderID(ctx sdk.Context) (orderID int64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(KeyNextOrderID)
	if bz == nil {
		return 0
	}
	k.cdc.MustUnmarshalBinary(bz, &orderID)
	orderID--
	return
}

func (k Keeper) getNewOrderID(ctx sdk.Context) (orderID int64, err sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(KeyNextOrderID)
	if bz == nil {
		return -1, ErrInvalidGenesis(k.codespace, "InitialOrderID never set")
	}
	k.cdc.MustUnmarshalBinary(bz, &orderID)
	bz = k.cdc.MustMarshalBinary(orderID + 1)
	store.Set(KeyNextOrderID, bz)
	return orderID, nil
}
