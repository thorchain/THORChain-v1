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
	clpKeeper := NewKeeper(clpKey, "RUNE", bankKeeper, DefaultCodespace)
	address := sdk.AccAddress([]byte("address1"))
	account := accountMapper.NewAccountWithAddress(ctx, address)
	accountMapper.SetAccount(ctx, account)
	return clpKeeper, cdc, bankKeeper, address
}

func TestCoolKeeperCreate(t *testing.T) {
	clpKey := sdk.NewKVStoreKey("clpTestKey")
	ctx := setupContext(clpKey)
	keeper, _, bankKeeper, address := setupKeepers(clpKey, ctx)

	baseTokenTicker := "RUNE"
	baseTokenName := "Rune"
	ticker := "eth"
	name := "ethereum"
	reserveRatio := 1
	ticker2 := "btc"
	name2 := "bitcoin"
	reserveRatio2 := 0
	ticker3 := "cos"
	name3 := "cosmos"
	reserveRatio3 := 200
	initialCoinSupply := int64(1000)
	initialBaseCoins := int64(5)
	clpAddress := types.NewCLPAddress(ticker)
	fiftyRune := sdk.NewCoin("RUNE", 50)

	validCLP := types.NewCLP(address, ticker, name, reserveRatio, initialCoinSupply, clpAddress)
	bankKeeper.SetCoins(ctx, address, sdk.Coins{fiftyRune})

	//Test happy path creation
	err1 := keeper.create(ctx, address, ticker, name, reserveRatio, initialCoinSupply, initialBaseCoins)
	require.Nil(t, err1)

	//Get created CLP and confirm values are correct
	newClp := keeper.GetCLP(ctx, ticker)
	require.Equal(t, newClp, &validCLP)

	//Get account coins and confirm debited and credited correctly
	addressCoins := bankKeeper.GetCoins(ctx, address)
	clpCoins := bankKeeper.GetCoins(ctx, clpAddress)
	addressRuneAmount := addressCoins.AmountOf("RUNE").Int64()
	clpRuneAmount := clpCoins.AmountOf("RUNE").Int64()
	require.Equal(t, addressRuneAmount, int64(45))
	require.Equal(t, clpRuneAmount, int64(5))

	//Test duplicate ticker
	err2 := keeper.create(ctx, address, ticker, name2, reserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err2)

	//Test bad ratios
	err4 := keeper.create(ctx, address, ticker2, name2, reserveRatio2, initialCoinSupply, initialBaseCoins)
	require.Error(t, err4)
	err5 := keeper.create(ctx, address, ticker3, name3, reserveRatio3, initialCoinSupply, initialBaseCoins)
	require.Error(t, err5)

	//Test cannot create CLP for base token
	err6 := keeper.create(ctx, address, baseTokenTicker, baseTokenName, reserveRatio, initialCoinSupply, initialBaseCoins)
	require.Error(t, err6)

	//Test cannot create CLP with bad initial supply
	err7 := keeper.create(ctx, address, "eth5", "ethereum4", reserveRatio, 0, initialBaseCoins)
	require.Error(t, err7)
	err8 := keeper.create(ctx, address, "eth5", "ethereum5", reserveRatio, -5, initialBaseCoins)
	require.Error(t, err8)

	//Test cannot create CLP with bad initial coins
	err9 := keeper.create(ctx, address, "eth6", "ethereum6", reserveRatio, initialCoinSupply, 0)
	require.Error(t, err9)
	err10 := keeper.create(ctx, address, "eth7", "ethereum7", reserveRatio, initialCoinSupply, -5)
	require.Error(t, err10)

	//Test cannot create CLP with more initial coins than owned
	err11 := keeper.create(ctx, address, "eth8", "ethereum8", reserveRatio, initialCoinSupply, 500)
	require.Error(t, err11)
}

func TestCoolKeeperTradeBase(t *testing.T) {
	clpKey := sdk.NewKVStoreKey("clpTestKey")
	ctx := setupContext(clpKey)
	keeper, _, bankKeeper, address := setupKeepers(clpKey, ctx)

	ticker := "eth"
	name := "ethereum"
	reserveRatio := 1
	tenRune := sdk.NewCoin("RUNE", 10)
	twentyRune := sdk.NewCoin("RUNE", 20)
	initialCoinSupply := int64(1000)
	initialBaseCoins := int64(5)
	fiftyRune := sdk.NewCoin("RUNE", 50)

	bankKeeper.SetCoins(ctx, address, sdk.Coins{fiftyRune})

	keeper.create(ctx, address, ticker, name, reserveRatio, initialCoinSupply, initialBaseCoins)

	//Test happy path trading
	bankKeeper.SetCoins(ctx, address, sdk.Coins{tenRune})
	_, err1 := keeper.tradeBase(ctx, address, ticker, 10)
	coins := bankKeeper.GetCoins(ctx, address)
	ethAmount := coins.AmountOf("eth").Int64()
	runeAmount := coins.AmountOf("RUNE").Int64()
	require.Nil(t, err1)
	require.Equal(t, ethAmount, int64(10))
	require.Equal(t, runeAmount, int64(0))

	//Test double trade
	bankKeeper.SetCoins(ctx, address, sdk.Coins{twentyRune})
	keeper.tradeBase(ctx, address, ticker, 10)
	keeper.tradeBase(ctx, address, ticker, 10)
	coins = bankKeeper.GetCoins(ctx, address)
	ethAmount = coins.AmountOf("eth").Int64()
	runeAmount = coins.AmountOf("RUNE").Int64()
	require.Equal(t, ethAmount, int64(20))
	require.Equal(t, runeAmount, int64(0))

	//Test invalid trade with nonexistent clp
	bankKeeper.SetCoins(ctx, address, sdk.Coins{tenRune})
	_, err2 := keeper.tradeBase(ctx, address, "btc", 10)
	require.Error(t, err2)
	coins = bankKeeper.GetCoins(ctx, address)
	btcAmount := coins.AmountOf("btc").Int64()
	runeAmount = coins.AmountOf("RUNE").Int64()
	require.Equal(t, btcAmount, int64(0))
	require.Equal(t, runeAmount, int64(10))

	// //Test invalid trade with too little rune
	bankKeeper.SetCoins(ctx, address, sdk.Coins{tenRune})
	_, err3 := keeper.tradeBase(ctx, address, ticker, int64(20))
	require.Error(t, err3)
	coins = bankKeeper.GetCoins(ctx, address)
	ethAmount = coins.AmountOf("eth").Int64()
	runeAmount = coins.AmountOf("RUNE").Int64()
	require.Equal(t, ethAmount, int64(0))
	require.Equal(t, runeAmount, int64(10))

	// //Test invalid trade with negative rune
	bankKeeper.SetCoins(ctx, address, sdk.Coins{tenRune})
	_, err4 := keeper.tradeBase(ctx, address, ticker, int64(-20))
	require.Error(t, err4)
	coins = bankKeeper.GetCoins(ctx, address)
	ethAmount = coins.AmountOf("eth").Int64()
	runeAmount = coins.AmountOf("RUNE").Int64()
	require.Equal(t, ethAmount, int64(0))
	require.Equal(t, runeAmount, int64(10))

}
