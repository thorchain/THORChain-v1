package exchange

import (
	"testing"
	"time"

	"github.com/tendermint/go-amino"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

var (
	exchangeKey = sdk.NewKVStoreKey("orderTestKey")
)

func setupContext(exchangeKey *sdk.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	multiStore := store.NewCommitMultiStore(db)
	multiStore.MountStoreWithDB(exchangeKey, sdk.StoreTypeIAVL, db)
	multiStore.LoadLatestVersion()
	ctx := sdk.NewContext(multiStore, abci.Header{}, false, nil)
	return ctx
}

func setupKeepers(exchangeKey *sdk.KVStoreKey, ctx sdk.Context) (Keeper, *amino.Codec, bank.Keeper,
	sdk.AccAddress, sdk.AccAddress) {
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)
	accountMapper := auth.NewAccountMapper(cdc, exchangeKey, auth.ProtoBaseAccount)
	bankKeeper := bank.NewKeeper(accountMapper)
	exchangeKeeper := NewKeeper(exchangeKey, bankKeeper, DefaultCodespace)

	InitGenesis(ctx, exchangeKeeper, DefaultGenesisState())
	WriteGenesis(ctx, exchangeKeeper)

	buyerAddress := sdk.AccAddress([]byte("buyerAddress"))
	buyerAccount := accountMapper.NewAccountWithAddress(ctx, buyerAddress)
	accountMapper.SetAccount(ctx, buyerAccount)

	sellerAddress := sdk.AccAddress([]byte("sellerAddress"))
	sellerAccount := accountMapper.NewAccountWithAddress(ctx, sellerAddress)
	accountMapper.SetAccount(ctx, sellerAccount)

	return exchangeKeeper, cdc, bankKeeper, buyerAddress, sellerAddress
}

func setupCreateBuyLimitOrderTest() (sdk.Context, Keeper, bank.Keeper, sdk.AccAddress, sdk.AccAddress,
	LimitOrder, LimitOrder, LimitOrder, LimitOrder) {
	ctx := setupContext(exchangeKey)
	keeper, _, bankKeeper, buyer, seller := setupKeepers(exchangeKey, ctx)

	bankKeeper.SetCoins(ctx, buyer, sdk.Coins{sdk.NewCoin("RUNE", 2000)})
	bankKeeper.SetCoins(ctx, seller, sdk.Coins{sdk.NewCoin("ETH", 250)})

	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	id, _ := keeper.getNewOrderID(ctx)
	limitSellOrder1 := NewLimitOrder(id, seller, SellOrder, sdk.NewCoin("ETH", 120),
		sdk.NewCoin("RUNE", 6), time.Now().Add(time.Minute).UTC())
	id, _ = keeper.getNewOrderID(ctx)
	limitSellOrder2 := NewLimitOrder(id, seller, SellOrder, sdk.NewCoin("ETH", 100),
		sdk.NewCoin("RUNE", 7), time.Now().Add(time.Minute).UTC())
	sellOrderBook.Orders = []LimitOrder{limitSellOrder1, limitSellOrder2}
	keeper.setOrderBook(ctx, sellOrderBook)

	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	id, _ = keeper.getNewOrderID(ctx)
	limitBuyOrder1 := NewLimitOrder(id, buyer, BuyOrder, sdk.NewCoin("ETH", 50),
		sdk.NewCoin("RUNE", 4), time.Now().Add(time.Minute).UTC())
	id, _ = keeper.getNewOrderID(ctx)
	limitBuyOrder2 := NewLimitOrder(id, buyer, BuyOrder, sdk.NewCoin("ETH", 80),
		sdk.NewCoin("RUNE", 2), time.Now().Add(time.Minute).UTC())
	buyOrderBook.Orders = []LimitOrder{limitBuyOrder1, limitBuyOrder2}
	keeper.setOrderBook(ctx, buyOrderBook)

	return ctx, keeper, bankKeeper, buyer, seller, limitSellOrder1, limitSellOrder2, limitBuyOrder1, limitBuyOrder2
}

func TestKeeperCreateLimitOrderSad(t *testing.T) {
	ctx, keeper, bankKeeper, buyer, seller, _, _, limitOrder1, limitOrder2 := setupCreateBuyLimitOrderTest()

	// Test invalid limit orders then confirm balances and orderbook is still untouched
	// Invalid limit order that is expired
	_, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", 3), time.Now().Add(-time.Minute))
	require.EqualError(t, err, ErrOrderExpired(keeper.codespace).Error())

	// Invalid limit order with wrong kind
	_, err = keeper.processLimitOrder(
		ctx, buyer, 0x03, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", 3), time.Now().Add(time.Minute))
	require.EqualError(t, err, ErrInvalidKind(keeper.codespace).Error())

	// Invalid limit order token to same token
	_, err = keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("ETH", 3), time.Now().Add(time.Minute))
	require.EqualError(t, err, ErrSameDenom(keeper.codespace).Error())

	// Invalid limit order negative amount
	_, err = keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", -200), sdk.NewCoin("RUNE", 3), time.Now().Add(time.Minute))
	require.EqualError(t, err, ErrAmountNotPositive(keeper.codespace).Error())

	// Invalid limit order negative price
	_, err = keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", -3), time.Now().Add(time.Minute))
	require.EqualError(t, err, ErrPriceNotPositive(keeper.codespace).Error())

	// Invalid limit order not enough coins
	_, err = keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", 11), time.Now().Add(time.Minute))
	require.EqualError(t, err, sdk.ErrInsufficientCoins("Must have at least 2200RUNE to place this buy limit order").Error())

	// Check balances still the same after invalid trades
	buyerRuneAmount := bankKeeper.GetCoins(ctx, buyer).AmountOf("RUNE").Int64()
	require.Equal(t, int64(2000), buyerRuneAmount)
	sellerEthAmount := bankKeeper.GetCoins(ctx, seller).AmountOf("ETH").Int64()
	require.Equal(t, int64(250), sellerEthAmount)

	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Len(t, sellOrderBook.Orders, 2)
	require.Len(t, buyOrderBook.Orders, 2)
	require.Equal(t, limitOrder1, buyOrderBook.Orders[0])
	require.Equal(t, limitOrder2, buyOrderBook.Orders[1])
}

// Test if buy order cannot be filled => expect order to be placed on orderbook
func TestKeeperCreateBuyLimitOrderNotFilled(t *testing.T) {
	ctx, keeper, _, buyer, _, _, _, limitBuyOrder1, limitBuyOrder2 := setupCreateBuyLimitOrderTest()

	expiresAt := time.Now().Add(time.Minute).UTC()

	limitOrderID, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", 3), expiresAt)

	require.Nil(t, err)
	require.True(t, limitOrderID > 0)

	orderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Equal(t, limitBuyOrder1, orderBook.Orders[0])
	require.Equal(t, NewLimitOrder(limitBuyOrder2.OrderID+1, buyer, BuyOrder, sdk.NewCoin("ETH", 200),
		sdk.NewCoin("RUNE", 3), expiresAt), orderBook.Orders[1])
	require.Equal(t, limitBuyOrder2, orderBook.Orders[2])
}

// Test if buy order can be filled fully by 2 sell orders => expect sell orders to be updated
func TestKeeperCreateBuyLimitOrderFilled(t *testing.T) {
	ctx, keeper, bankKeeper, buyer, seller, _, limitSellOrder2, limitBuyOrder1, limitBuyOrder2 := setupCreateBuyLimitOrderTest()

	expiresAt := time.Now().Add(time.Minute).UTC()

	limitOrderID, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("RUNE", 8), expiresAt)

	require.Nil(t, err)
	require.Equal(t, int64(-1), limitOrderID)

	// buy orderbook untouched
	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Len(t, buyOrderBook.Orders, 2)
	require.Equal(t, limitBuyOrder1, buyOrderBook.Orders[0])
	require.Equal(t, limitBuyOrder2, buyOrderBook.Orders[1])

	// sell orderbook changed
	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	require.Len(t, sellOrderBook.Orders, 1)
	expectedLimitSellOrder2 := limitSellOrder2
	expectedLimitSellOrder2.Amount = sdk.NewCoin("ETH", 20)
	require.Equal(t, expectedLimitSellOrder2, sellOrderBook.Orders[0])

	// seller and buyer coins changed
	coinsSeller := bankKeeper.GetCoins(ctx, seller)
	require.Equal(t, "50ETH,1280RUNE", coinsSeller.String())
	coinsBuyer := bankKeeper.GetCoins(ctx, buyer)
	require.Equal(t, "200ETH,720RUNE", coinsBuyer.String())
}

// Test if buy order can be filled partially by 1 sell order => expect new partial buy order to be placed
func TestKeeperCreateBuyLimitOrderFilledPartially(t *testing.T) {
	ctx, keeper, bankKeeper, buyer, seller, _, limitSellOrder2, limitBuyOrder1, limitBuyOrder2 := setupCreateBuyLimitOrderTest()

	expiresAt := time.Now().Add(time.Minute).UTC()

	limitOrderID, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 210), sdk.NewCoin("RUNE", 6), expiresAt)

	require.Nil(t, err)
	require.True(t, limitOrderID > 0)

	// buy orderbook has additional buy order over partial filling
	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Len(t, buyOrderBook.Orders, 3)
	require.Equal(t, NewLimitOrder(limitOrderID, buyer, BuyOrder, sdk.NewCoin("ETH", 90),
		sdk.NewCoin("RUNE", 6), expiresAt), buyOrderBook.Orders[0])
	require.Equal(t, limitBuyOrder1, buyOrderBook.Orders[1])
	require.Equal(t, limitBuyOrder2, buyOrderBook.Orders[2])

	// sell orderbook changed
	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	require.Len(t, sellOrderBook.Orders, 1)
	expectedLimitSellOrder2 := limitSellOrder2
	require.Equal(t, expectedLimitSellOrder2, sellOrderBook.Orders[0])

	// seller and buyer coins changed
	coinsSeller := bankKeeper.GetCoins(ctx, seller)
	require.Equal(t, "130ETH,720RUNE", coinsSeller.String())
	coinsBuyer := bankKeeper.GetCoins(ctx, buyer)
	require.Equal(t, "120ETH,1280RUNE", coinsBuyer.String())
}

// Test if cheapest order in orderbook is expired
func TestKeeperCreateBuyLimitOrderFilledCheapestExpired(t *testing.T) {
	ctx, keeper, bankKeeper, buyer, seller, _, limitSellOrder2, limitBuyOrder1, limitBuyOrder2 :=
		setupCreateBuyLimitOrderTest()

	// let cheaper order be expired
	orderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	orderBook.Orders[0].ExpiresAt = time.Now().Add(-time.Minute).UTC()
	keeper.setOrderBook(ctx, orderBook)

	expiresAt := time.Now().Add(time.Minute).UTC()

	limitOrderID, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 70), sdk.NewCoin("RUNE", 8), expiresAt)

	require.Nil(t, err)
	require.Equal(t, int64(-1), limitOrderID)

	// buy orderbook untouched
	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Len(t, buyOrderBook.Orders, 2)
	require.Equal(t, limitBuyOrder1, buyOrderBook.Orders[0])
	require.Equal(t, limitBuyOrder2, buyOrderBook.Orders[1])

	// sell orderbook changed
	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	require.Len(t, sellOrderBook.Orders, 1)
	expectedLimitSellOrder2 := limitSellOrder2
	expectedLimitSellOrder2.Amount = sdk.NewCoin("ETH", 30)
	require.Equal(t, expectedLimitSellOrder2, sellOrderBook.Orders[0])

	// seller and buyer coins changed
	coinsSeller := bankKeeper.GetCoins(ctx, seller)
	require.Equal(t, "180ETH,490RUNE", coinsSeller.String())
	coinsBuyer := bankKeeper.GetCoins(ctx, buyer)
	require.Equal(t, "70ETH,1510RUNE", coinsBuyer.String())
}

// Test if order in orderbook not enough coins is skipped
func TestKeeperCreateBuyLimitOrderNotEnoughCoinsSkipped(t *testing.T) {
	ctx, keeper, bankKeeper, buyer, seller, limitSellOrder1, limitSellOrder2, limitBuyOrder1, limitBuyOrder2 :=
		setupCreateBuyLimitOrderTest()

	// let seller have not enough coins
	bankKeeper.SetCoins(ctx, seller, sdk.Coins{sdk.NewCoin("ETH", 80)})

	expiresAt := time.Now().Add(time.Minute).UTC()

	limitOrderID, err := keeper.processLimitOrder(
		ctx, buyer, BuyOrder, sdk.NewCoin("ETH", 110), sdk.NewCoin("RUNE", 8), expiresAt)

	require.Nil(t, err)
	require.True(t, limitOrderID > 0)

	// buy orderbook has additional buy order over partial filling
	buyOrderBook := keeper.getOrderBook(ctx, BuyOrder, "ETH", "RUNE")
	require.Len(t, buyOrderBook.Orders, 3)
	require.Equal(t, NewLimitOrder(limitOrderID, buyer, BuyOrder, sdk.NewCoin("ETH", 110),
		sdk.NewCoin("RUNE", 8), expiresAt), buyOrderBook.Orders[0])
	require.Equal(t, limitBuyOrder1, buyOrderBook.Orders[1])
	require.Equal(t, limitBuyOrder2, buyOrderBook.Orders[2])

	// sell orderbook untouched
	sellOrderBook := keeper.getOrderBook(ctx, SellOrder, "ETH", "RUNE")
	require.Len(t, sellOrderBook.Orders, 2)
	require.Equal(t, limitSellOrder1, sellOrderBook.Orders[0])
	require.Equal(t, limitSellOrder2, sellOrderBook.Orders[1])

	// seller and buyer coins unchanged
	coinsSeller := bankKeeper.GetCoins(ctx, seller)
	require.Equal(t, "80ETH", coinsSeller.String())
	coinsBuyer := bankKeeper.GetCoins(ctx, buyer)
	require.Equal(t, "2000RUNE", coinsBuyer.String())
}
