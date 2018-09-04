package clp

import (
	"testing"

	"github.com/tendermint/go-amino"
	"github.com/thorchain/THORChain/x/clp/types"

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
	clpKey              = sdk.NewKVStoreKey("clpTestKey")
	_1600Rune           = sdk.NewCoin("RUNE", 1600)
	_1000Rune           = sdk.NewCoin("RUNE", 1000)
	runeTicker          = "RUNE"
	runeTokenName       = "Rune"
	ethTicker           = "ETH"
	ethTokenName        = "ethereum"
	ethClpAddress       = types.NewCLPAddress(ethTicker)
	btcTicker           = "BTC"
	btcTokenName        = "bitcoin"
	tokTicker           = "TOK"
	tokTokenName        = "Token"
	blankTicker         = ""
	invalidTicker       = "INVALID"
	reserveRatio        = 100
	zeroReserveRatio    = 0
	over100ReserveRatio = 300
	initialCoinSupply   = int64(500)
	initialBaseCoins    = int64(500)
)

func setupContext(clpKey *sdk.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	multiStore := store.NewCommitMultiStore(db)
	multiStore.MountStoreWithDB(clpKey, sdk.StoreTypeIAVL, db)
	multiStore.LoadLatestVersion()
	ctx := sdk.NewContext(multiStore, abci.Header{}, false, nil)
	return ctx
}

func setupKeepers(clpKey *sdk.KVStoreKey, ctx sdk.Context) (Keeper, *amino.Codec, bank.Keeper, sdk.AccAddress) {
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)
	accountMapper := auth.NewAccountMapper(cdc, clpKey, auth.ProtoBaseAccount)
	bankKeeper := bank.NewKeeper(accountMapper)
	clpKeeper := NewKeeper(clpKey, runeTicker, bankKeeper, DefaultCodespace)
	address := sdk.AccAddress([]byte("address1"))
	account := accountMapper.NewAccountWithAddress(ctx, address)
	accountMapper.SetAccount(ctx, account)
	return clpKeeper, cdc, bankKeeper, address
}

func setupTradingTest() (sdk.Context, Keeper, bank.Keeper, sdk.AccAddress) {
	ctx := setupContext(clpKey)
	keeper, _, bankKeeper, address := setupKeepers(clpKey, ctx)

	bankKeeper.SetCoins(ctx, address, sdk.Coins{_1600Rune})

	keeper.create(ctx, address, ethTicker, ethTokenName, 100, int64(500), int64(500))
	keeper.create(ctx, address, btcTicker, btcTokenName, 100, int64(500), int64(500))
	keeper.create(ctx, address, tokTicker, tokTokenName, 100, 1000000, 100)

	return ctx, keeper, bankKeeper, address
}

func TestCoolKeeperCreate(t *testing.T) {
	ctx := setupContext(clpKey)
	keeper, _, bankKeeper, senderAddress := setupKeepers(clpKey, ctx)

	validCLP := types.NewCLP(senderAddress, ethTicker, ethTokenName, 100, int64(500), ethClpAddress)
	bankKeeper.SetCoins(ctx, senderAddress, sdk.Coins{_1000Rune})

	//Test happy path creation
	err1 := keeper.create(ctx, senderAddress, ethTicker, ethTokenName, 100, int64(500), int64(500))
	require.Nil(t, err1)

	//Get created CLP and confirm values are correct
	newClp := keeper.GetCLP(ctx, ethTicker)
	require.Equal(t, newClp, &validCLP)

	//Get account coins and confirm debited and credited correctly
	addressCoins := bankKeeper.GetCoins(ctx, senderAddress)
	clpCoins := bankKeeper.GetCoins(ctx, ethClpAddress)
	addressRuneAmount := addressCoins.AmountOf(runeTicker).Int64()
	clpRuneAmount := clpCoins.AmountOf(runeTicker).Int64()
	addressEthAmount := addressCoins.AmountOf(ethTicker).Int64()
	clpEthAmount := clpCoins.AmountOf(ethTicker).Int64()
	require.Equal(t, addressRuneAmount, int64(500))
	require.Equal(t, clpRuneAmount, int64(500))
	require.Equal(t, addressEthAmount, int64(0))
	require.Equal(t, clpEthAmount, int64(500))

	//Test duplicate ticker
	err2 := keeper.create(ctx, senderAddress, ethTicker, ethTokenName, reserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err2)

	//Test bad ratios
	err4 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, zeroReserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err4)
	err5 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, over100ReserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err5)

	//Test cannot create CLP for base token
	err6 := keeper.create(ctx, senderAddress, runeTicker, runeTokenName, reserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err6)

	//Test cannot create CLP with bad initial supply
	err7 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, reserveRatio, 0, initialBaseCoins)
	require.Error(t, err7)
	err8 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, reserveRatio, -5, initialBaseCoins)
	require.Error(t, err8)

	//Test cannot create CLP with bad initial coins
	err9 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, reserveRatio, initialCoinSupply, 0)
	require.Error(t, err9)
	err10 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, reserveRatio, initialCoinSupply, -5)
	require.Error(t, err10)

	//Test cannot create CLP with more initial coins than owned
	err11 := keeper.create(ctx, senderAddress, btcTicker, btcTokenName, reserveRatio, initialCoinSupply, 5000)
	require.Error(t, err11)
}

func TestCoolKeeperTradeBasic(t *testing.T) {
	ctx, keeper, bankKeeper, senderAddress := setupTradingTest()
	clp := keeper.GetCLP(ctx, ethTicker)

	//Test happy path trading
	_, _, err1 := keeper.trade(ctx, senderAddress, runeTicker, ethTicker, 10)
	clp = keeper.GetCLP(ctx, ethTicker)
	senderCoins := bankKeeper.GetCoins(ctx, senderAddress)
	senderEthAmount := senderCoins.AmountOf(ethTicker).Int64()
	senderRuneAmount := senderCoins.AmountOf(runeTicker).Int64()
	clpCoins := bankKeeper.GetCoins(ctx, clp.AccountAddress)
	clpEthAmount := clpCoins.AmountOf(ethTicker).Int64()
	clpRuneAmount := clpCoins.AmountOf(runeTicker).Int64()
	require.Nil(t, err1)
	require.Equal(t, senderEthAmount, int64(10))
	require.Equal(t, senderRuneAmount, int64(490))
	require.Equal(t, clpEthAmount, int64(490))
	require.Equal(t, clp.CurrentSupply, int64(500))
	require.Equal(t, clpRuneAmount, int64(510))

	//Test double trade
	keeper.trade(ctx, senderAddress, runeTicker, ethTicker, 10)
	keeper.trade(ctx, senderAddress, runeTicker, ethTicker, 10)
	clp = keeper.GetCLP(ctx, ethTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderEthAmount = senderCoins.AmountOf(ethTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	clpCoins = bankKeeper.GetCoins(ctx, clp.AccountAddress)
	clpEthAmount = clpCoins.AmountOf(ethTicker).Int64()
	clpRuneAmount = clpCoins.AmountOf(runeTicker).Int64()
	require.Equal(t, senderEthAmount, int64(30))
	require.Equal(t, senderRuneAmount, int64(470))
	require.Equal(t, clpEthAmount, int64(470))
	require.Equal(t, clp.CurrentSupply, int64(500))
	require.Equal(t, clpRuneAmount, int64(530))

	//Test Trade from token back to rune twice
	keeper.trade(ctx, senderAddress, ethTicker, runeTicker, 10)
	keeper.trade(ctx, senderAddress, ethTicker, runeTicker, 10)
	clp = keeper.GetCLP(ctx, ethTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderEthAmount = senderCoins.AmountOf(ethTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	clpCoins = bankKeeper.GetCoins(ctx, clp.AccountAddress)
	clpEthAmount = clpCoins.AmountOf(ethTicker).Int64()
	clpRuneAmount = clpCoins.AmountOf(runeTicker).Int64()
	require.Equal(t, senderEthAmount, int64(10))
	require.Equal(t, senderRuneAmount, int64(491))
	require.Equal(t, clpEthAmount, int64(490))
	require.Equal(t, clp.CurrentSupply, int64(500))
	require.Equal(t, clpRuneAmount, int64(509))
}

func TestCoolKeeperTradeBasicSad(t *testing.T) {
	ctx, keeper, bankKeeper, senderAddress := setupTradingTest()

	//Test invalid trades then confirm balances are still in check
	//Invalid trade rune to rune
	_, _, err := keeper.trade(ctx, senderAddress, runeTicker, runeTicker, 10)
	require.Error(t, err)
	//Invalid trade same token
	_, _, err = keeper.trade(ctx, senderAddress, ethTicker, ethTicker, 10)
	require.Error(t, err)
	//Invalid trade to nonexistent clp token
	_, _, err = keeper.trade(ctx, senderAddress, runeTicker, invalidTicker, 10)
	require.Error(t, err)
	//Invalid trade to empty clp token
	_, _, err = keeper.trade(ctx, senderAddress, runeTicker, blankTicker, 10)
	require.Error(t, err)
	//Invalid trade from nonexistent clp token
	_, _, err = keeper.trade(ctx, senderAddress, invalidTicker, ethTicker, 10)
	require.Error(t, err)
	//Invalid trade from empty token
	_, _, err = keeper.trade(ctx, senderAddress, blankTicker, ethTicker, 10)
	require.Error(t, err)
	//Invalid trade with too little rune
	_, _, err = keeper.trade(ctx, senderAddress, runeTicker, ethTicker, int64(5000))
	require.Error(t, err)
	//Invalid trade with negative rune
	_, _, err = keeper.trade(ctx, senderAddress, runeTicker, ethTicker, int64(-20))
	require.Error(t, err)

	//Check balances still the same after invalid trades
	ethClp := keeper.GetCLP(ctx, ethTicker)
	senderCoins := bankKeeper.GetCoins(ctx, senderAddress)
	senderEthAmount := senderCoins.AmountOf(ethTicker).Int64()
	senderInvalidAmount := senderCoins.AmountOf(invalidTicker).Int64()
	senderRuneAmount := senderCoins.AmountOf(runeTicker).Int64()
	clpCoins := bankKeeper.GetCoins(ctx, ethClp.AccountAddress)
	clpEthAmount := clpCoins.AmountOf(ethTicker).Int64()
	clpRuneAmount := clpCoins.AmountOf(runeTicker).Int64()
	require.Equal(t, senderInvalidAmount, int64(0))
	require.Equal(t, senderRuneAmount, int64(500))
	require.Equal(t, senderEthAmount, int64(0))
	require.Equal(t, clpEthAmount, int64(500))
	require.Equal(t, ethClp.CurrentSupply, int64(500))
	require.Equal(t, clpRuneAmount, int64(500))
}

func TestCoolKeeperTradeBridged(t *testing.T) {
	ctx, keeper, bankKeeper, senderAddress := setupTradingTest()
	btcClp := keeper.GetCLP(ctx, btcTicker)

	//Get some BTC:
	keeper.trade(ctx, senderAddress, runeTicker, btcTicker, 100)
	senderCoins := bankKeeper.GetCoins(ctx, senderAddress)
	senderBtcAmount := senderCoins.AmountOf(btcTicker).Int64()
	senderRuneAmount := senderCoins.AmountOf(runeTicker).Int64()
	btcClpCoins := bankKeeper.GetCoins(ctx, btcClp.AccountAddress)
	btcClpRuneAmount := btcClpCoins.AmountOf(runeTicker).Int64()
	btcClpBtcAmount := btcClpCoins.AmountOf(btcTicker).Int64()
	require.Equal(t, senderBtcAmount, int64(100))
	require.Equal(t, senderRuneAmount, int64(400))
	require.Equal(t, btcClpRuneAmount, int64(600))
	require.Equal(t, btcClpBtcAmount, int64(400))

	//Test happy path trading
	_, _, err1 := keeper.trade(ctx, senderAddress, btcTicker, ethTicker, 20)
	ethClp := keeper.GetCLP(ctx, ethTicker)
	btcClp = keeper.GetCLP(ctx, btcTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderBtcAmount = senderCoins.AmountOf(btcTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	senderEthAmount := senderCoins.AmountOf(ethTicker).Int64()
	btcClpCoins = bankKeeper.GetCoins(ctx, btcClp.AccountAddress)
	btcClpEthAmount := btcClpCoins.AmountOf(ethTicker).Int64()
	btcClpRuneAmount = btcClpCoins.AmountOf(runeTicker).Int64()
	btcClpBtcAmount = btcClpCoins.AmountOf(btcTicker).Int64()
	ethClpCoins := bankKeeper.GetCoins(ctx, ethClp.AccountAddress)
	ethClpEthAmount := ethClpCoins.AmountOf(ethTicker).Int64()
	ethClpRuneAmount := ethClpCoins.AmountOf(runeTicker).Int64()
	ethClpBtcAmount := ethClpCoins.AmountOf(btcTicker).Int64()
	require.Nil(t, err1)
	require.Equal(t, senderBtcAmount, int64(80))
	require.Equal(t, senderRuneAmount, int64(400))
	require.Equal(t, senderEthAmount, int64(24))
	require.Equal(t, btcClpBtcAmount, int64(420))
	require.Equal(t, btcClpRuneAmount, int64(576))
	require.Equal(t, btcClpEthAmount, int64(0))
	require.Equal(t, ethClpBtcAmount, int64(0))
	require.Equal(t, ethClpRuneAmount, int64(524))
	require.Equal(t, ethClpEthAmount, int64(476))

	//Test double trade
	keeper.trade(ctx, senderAddress, btcTicker, ethTicker, 10)
	keeper.trade(ctx, senderAddress, btcTicker, ethTicker, 10)
	ethClp = keeper.GetCLP(ctx, ethTicker)
	btcClp = keeper.GetCLP(ctx, btcTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderBtcAmount = senderCoins.AmountOf(btcTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	senderEthAmount = senderCoins.AmountOf(ethTicker).Int64()
	btcClpCoins = bankKeeper.GetCoins(ctx, btcClp.AccountAddress)
	btcClpEthAmount = btcClpCoins.AmountOf(ethTicker).Int64()
	btcClpRuneAmount = btcClpCoins.AmountOf(runeTicker).Int64()
	btcClpBtcAmount = btcClpCoins.AmountOf(btcTicker).Int64()
	ethClpCoins = bankKeeper.GetCoins(ctx, ethClp.AccountAddress)
	ethClpEthAmount = ethClpCoins.AmountOf(ethTicker).Int64()
	ethClpRuneAmount = ethClpCoins.AmountOf(runeTicker).Int64()
	ethClpBtcAmount = ethClpCoins.AmountOf(btcTicker).Int64()
	require.Nil(t, err1)
	require.Equal(t, senderBtcAmount, int64(60))
	require.Equal(t, senderRuneAmount, int64(400))
	require.Equal(t, senderEthAmount, int64(45))
	require.Equal(t, btcClpBtcAmount, int64(440))
	require.Equal(t, btcClpRuneAmount, int64(553))
	require.Equal(t, btcClpEthAmount, int64(0))
	require.Equal(t, ethClpBtcAmount, int64(0))
	require.Equal(t, ethClpRuneAmount, int64(547))
	require.Equal(t, ethClpEthAmount, int64(455))

	//Test Trade from token back to rune twice
	keeper.trade(ctx, senderAddress, ethTicker, btcTicker, 10)
	keeper.trade(ctx, senderAddress, ethTicker, btcTicker, 10)
	ethClp = keeper.GetCLP(ctx, ethTicker)
	btcClp = keeper.GetCLP(ctx, btcTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderBtcAmount = senderCoins.AmountOf(btcTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	senderEthAmount = senderCoins.AmountOf(ethTicker).Int64()
	btcClpCoins = bankKeeper.GetCoins(ctx, btcClp.AccountAddress)
	btcClpEthAmount = btcClpCoins.AmountOf(ethTicker).Int64()
	btcClpRuneAmount = btcClpCoins.AmountOf(runeTicker).Int64()
	btcClpBtcAmount = btcClpCoins.AmountOf(btcTicker).Int64()
	ethClpCoins = bankKeeper.GetCoins(ctx, ethClp.AccountAddress)
	ethClpEthAmount = ethClpCoins.AmountOf(ethTicker).Int64()
	ethClpRuneAmount = ethClpCoins.AmountOf(runeTicker).Int64()
	ethClpBtcAmount = ethClpCoins.AmountOf(btcTicker).Int64()
	require.Nil(t, err1)
	require.Equal(t, senderBtcAmount, int64(80))
	require.Equal(t, senderRuneAmount, int64(400))
	require.Equal(t, senderEthAmount, int64(25))
	require.Equal(t, btcClpBtcAmount, int64(420))
	require.Equal(t, btcClpRuneAmount, int64(575))
	require.Equal(t, btcClpEthAmount, int64(0))
	require.Equal(t, ethClpBtcAmount, int64(0))
	require.Equal(t, ethClpRuneAmount, int64(525))
	require.Equal(t, ethClpEthAmount, int64(475))
}

func TestCoolKeeperTradeScopingDoc(t *testing.T) {
	ctx, keeper, bankKeeper, senderAddress := setupTradingTest()
	tokClp := keeper.GetCLP(ctx, tokTicker)

	//Test Example from scoping doc
	keeper.trade(ctx, senderAddress, runeTicker, tokTicker, 90)
	senderCoins := bankKeeper.GetCoins(ctx, senderAddress)
	senderTokAmount := senderCoins.AmountOf(tokTicker).Int64()
	senderRuneAmount := senderCoins.AmountOf(runeTicker).Int64()
	clpCoins := bankKeeper.GetCoins(ctx, tokClp.AccountAddress)
	clpTokAmount := clpCoins.AmountOf(tokTicker).Int64()
	clpRuneAmount := clpCoins.AmountOf(runeTicker).Int64()
	require.Equal(t, senderTokAmount, int64(900000))
	require.Equal(t, senderRuneAmount, int64(410))
	require.Equal(t, clpTokAmount, int64(100000))
	require.Equal(t, tokClp.CurrentSupply, int64(1000000))
	require.Equal(t, clpRuneAmount, int64(190))

	//Test Second Trade on example from scoping doc
	keeper.trade(ctx, senderAddress, runeTicker, tokTicker, 5)
	tokClp = keeper.GetCLP(ctx, tokTicker)
	senderCoins = bankKeeper.GetCoins(ctx, senderAddress)
	senderTokAmount = senderCoins.AmountOf(tokTicker).Int64()
	senderRuneAmount = senderCoins.AmountOf(runeTicker).Int64()
	clpCoins = bankKeeper.GetCoins(ctx, tokClp.AccountAddress)
	clpTokAmount = clpCoins.AmountOf(tokTicker).Int64()
	clpRuneAmount = clpCoins.AmountOf(runeTicker).Int64()
	require.Equal(t, senderTokAmount, int64(926316))
	require.Equal(t, senderRuneAmount, int64(405))
	require.Equal(t, clpTokAmount, int64(73684))
	require.Equal(t, tokClp.CurrentSupply, int64(1000000))
	require.Equal(t, clpRuneAmount, int64(195))
}
