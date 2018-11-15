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
func (k Keeper) getOrderBook(ctx sdk.Context, kind OrderKind, amountDenom string, priceDenom string) OrderBook {
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
	expiresAt time.Time) (ProcessedLimitOrder, []FilledLimitOrder, sdk.Error) {

	// error if already expired
	if expiresAt.Before(time.Now()) {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, ErrOrderExpired(k.codespace)
	}

	// error if kind not supported
	if kind != BuyOrder && kind != SellOrder {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, ErrInvalidKind(k.codespace)
	}

	// error if amount and price denom are the same
	if amount.Denom == price.Denom {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, ErrSameDenom(k.codespace)
	}

	// error if amount negative
	if !amount.IsPositive() {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, ErrAmountNotPositive(k.codespace)
	}

	// error if price negative
	if !price.IsPositive() {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, ErrPriceNotPositive(k.codespace)
	}

	// check if enough coins to place order
	totalPrice := getTotalPrice(amount, price)
	if kind == BuyOrder && !k.bankKeeper.HasCoins(ctx, sender, sdk.Coins{totalPrice}) {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, sdk.ErrInsufficientCoins(fmt.Sprintf(
			"Must have at least %v to place this buy limit order", totalPrice))
	}
	if kind == SellOrder && !k.bankKeeper.HasCoins(ctx, sender, sdk.Coins{amount}) {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, sdk.ErrInsufficientCoins(fmt.Sprintf(
			"Must have at least %v to place this sell limit order", amount))
	}

	// fill order if possible
	unfilledAmt, filledOrders, err := k.fillOrderIfPossible(ctx, sender, kind, amount, price)
	if err != nil {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, err
	}

	// store unfilled order
	processedOrder, err := k.storeUnfilledLimitOrder(ctx, sender, kind, unfilledAmt, price, expiresAt)
	if err != nil {
		return ProcessedLimitOrder{}, []FilledLimitOrder{}, err
	}

	return processedOrder, filledOrders, nil
}

// fillOrderIfPossible tries to fill the order. Returns the amount that could not be filled and a slice of limit orders that have been filled
func (k Keeper) fillOrderIfPossible(
	ctx sdk.Context, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin,
) (sdk.Coin, []FilledLimitOrder, sdk.Error) {
	// get matching order book to fill the order
	matchingKind := SellOrder
	if kind == SellOrder {
		matchingKind = BuyOrder
	}
	orderBook := k.getOrderBook(ctx, matchingKind, amount.Denom, price.Denom)

	// unfilled amount
	unfilledAmt := amount

	// slice of filled orderIds, prices and amounts
	filledOrders := make([]FilledLimitOrder, 0, 10)

	var err sdk.Error

	for i, storedOrder := range orderBook.Orders {
		// end loop if unfilled amt is 0
		if unfilledAmt.IsZero() {
			break
		}

		// end loop if storedOrder cannot fill our order => since orders are sorted, this means there will be no more
		// match
		ok, fillAmount, fillPrice := storedOrder.DoesFill(kind, unfilledAmt, price)
		if !ok {
			break
		}

		fillTotalPrice := getTotalPrice(fillAmount, fillPrice)

		var coinsFromSenderToStoredSender, coinsToUnlockForSender sdk.Coin

		if kind == BuyOrder {
			// send totalPrice from buyer to seller
			coinsFromSenderToStoredSender = fillTotalPrice

			// give buyer locked coins from seller
			coinsToUnlockForSender = fillAmount
		} else {
			// send amount from seller to buyer
			coinsFromSenderToStoredSender = fillAmount

			// give seller locked coins from buyer
			coinsToUnlockForSender = fillTotalPrice
		}

		err = k.sendAndUnlockCoins(
			ctx, sender, storedOrder.Sender, coinsFromSenderToStoredSender, coinsToUnlockForSender)
		if err != nil {
			break
		}

		filledOrders = append(filledOrders, FilledLimitOrder{storedOrder.OrderID, fillAmount, fillPrice})

		// update unfilled amount
		unfilledAmt = unfilledAmt.Minus(fillAmount)

		// replace the amount of the stored order with the remaining part
		orderBook.Orders[i].Amount = storedOrder.Amount.Minus(fillAmount)
	}

	orderBook.RemoveFilledLimitOrders()

	k.setOrderBook(ctx, orderBook)

	return unfilledAmt, filledOrders, err
}

func (k Keeper) hasOrderSenderEnoughCoins(ctx sdk.Context, storedOrder LimitOrder, totalAmount, totalPrice sdk.Coin,
) bool {
	if storedOrder.Kind == BuyOrder {
		return k.bankKeeper.HasCoins(ctx, storedOrder.Sender, sdk.Coins{totalPrice})
	}

	return k.bankKeeper.HasCoins(ctx, storedOrder.Sender, sdk.Coins{totalAmount})
}

func (k Keeper) sendAndUnlockCoins(ctx sdk.Context, a, b sdk.AccAddress, coinsFromAToB, coinsToUnlockForA sdk.Coin,
) sdk.Error {
	_, err := k.bankKeeper.SendCoins(ctx, a, b, sdk.Coins{coinsFromAToB})
	if err != nil {
		return err
	}

	_, _, err = k.bankKeeper.AddCoins(ctx, a, sdk.Coins{coinsToUnlockForA})
	return err
}

// storeUnfilledLimitOrder creates a new limit order, finds the corresponding order book, adds the limit order
// to the right place and saves the orderbook. Returns a ProcessedLimitOrder
func (k Keeper) storeUnfilledLimitOrder(
	ctx sdk.Context, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin, expiresAt time.Time,
) (ProcessedLimitOrder, sdk.Error) {
	// create a new limit order and then add it to the orderbook
	newOrderID, err := k.getNewOrderID(ctx)
	if err != nil {
		return ProcessedLimitOrder{}, err
	}

	if amount.IsZero() {
		return ProcessedLimitOrder{newOrderID, amount}, nil
	}

	// get orderbook
	orderBook := k.getOrderBook(ctx, kind, amount.Denom, price.Denom)

	limitOrder := NewLimitOrder(newOrderID, sender, kind, amount, price, expiresAt)

	err = orderBook.AddLimitOrder(limitOrder)
	if err != nil {
		return ProcessedLimitOrder{}, err
	}

	// lock sender's coins to fill order in the future
	var coinToLock sdk.Coin
	if kind == BuyOrder {
		coinToLock = getTotalPrice(amount, price)
	} else {
		coinToLock = amount
	}
	_, _, err = k.bankKeeper.SubtractCoins(ctx, sender, sdk.Coins{coinToLock})
	if err != nil {
		return ProcessedLimitOrder{}, err
	}

	k.setOrderBook(ctx, orderBook)

	return ProcessedLimitOrder{newOrderID, amount}, nil
}

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

func (k Keeper) refundExpiredLimitOrders(ctx sdk.Context) {
	// Iterate over all orderbooks
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, orderBookSubspace)

	now := time.Now()
	obsToUpdate := make([]OrderBook, 0)
	osToRefund := make([]LimitOrder, 0)

	for ; iter.Valid(); iter.Next() {
		ob := new(OrderBook)
		k.cdc.UnmarshalBinary(iter.Value(), &ob)

		containsExpired := false
		newOrders := make([]LimitOrder, 0, len(ob.Orders))

		// iterate over orders, only keeping not expired ones, and add expired ones to refund array
		for _, order := range ob.Orders {
			if order.ExpiresAt.Before(now) {
				containsExpired = true
				osToRefund = append(osToRefund, order)
				continue
			}

			newOrders = append(newOrders, order)
		}

		// if orderbook contains expired orders, add them to the slice we need to update
		if containsExpired {
			ob.Orders = newOrders
			obsToUpdate = append(obsToUpdate, *ob)
		}
	}

	iter.Close()

	// save the orderbooks that need to be updated
	for _, ob := range obsToUpdate {
		k.setOrderBook(ctx, ob)
	}

	// refund orders that were expired
	for _, order := range osToRefund {
		var coinsToUnlock sdk.Coin
		if order.Kind == BuyOrder {
			coinsToUnlock = getTotalPrice(order.Amount, order.Price)
		} else {
			coinsToUnlock = order.Amount
		}
		_, _, err := k.bankKeeper.AddCoins(ctx, order.Sender, sdk.Coins{coinsToUnlock})
		if err != nil {
			panic(err)
		}
	}
}

func getTotalPrice(amt sdk.Coin, price sdk.Coin) sdk.Coin {
	return sdk.Coin{price.Denom, amt.Amount.Mul(price.Amount)}
}
