package clp

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/thorchain/THORChain/x/clp/types"
)

var (
	priv1     = ed25519.GenPrivKey()
	pub1      = priv1.PubKey()
	addr1     = sdk.AccAddress(pub1.Address())
	fiftyRune = sdk.NewInt64Coin("RUNE", 50)

	createMsg1 = types.MsgCreate{
		Sender:                addr1,
		Ticker:                "eth",
		Name:                  "ethereum",
		Decimals:              18,
		ReserveRatio:          100,
		InitialSupply:         int64(1000),
		InitialBaseCoinAmount: int64(5),
	}
	createMsg2 = types.MsgCreate{
		Sender:                addr1,
		Ticker:                "btc",
		Name:                  "bitcoin",
		Decimals:              8,
		ReserveRatio:          50,
		InitialSupply:         int64(1000),
		InitialBaseCoinAmount: int64(5),
	}
	createMsg3 = types.MsgCreate{
		Sender:                addr1,
		Ticker:                "eth",
		Name:                  "ethereum2",
		Decimals:              18,
		ReserveRatio:          100,
		InitialSupply:         int64(1000),
		InitialBaseCoinAmount: int64(5),
	}
	createMsg4 = types.MsgCreate{
		Sender:                addr1,
		Ticker:                "cos",
		Name:                  "cosmos",
		Decimals:              0,
		ReserveRatio:          200,
		InitialSupply:         int64(1000),
		InitialBaseCoinAmount: int64(5),
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	app := mock.NewApp()

	RegisterWire(app.Cdc)
	clpKey := sdk.NewKVStoreKey("clpAppTestKey")
	bankKeeper := bank.NewKeeper(app.AccountMapper)
	clpKeeper := NewKeeper(clpKey, "RUNE", bankKeeper, app.RegisterCodespace(DefaultCodespace))
	app.Router().AddRoute("clp", NewHandler(clpKeeper))

	app.SetInitChainer(getInitChainer(app, clpKeeper, bankKeeper))

	require.NoError(t, app.CompleteSetup([]*sdk.KVStoreKey{clpKey}))
	return app
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, clpKeeper Keeper, bankKeeper bank.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)
		return abci.ResponseInitChain{}
	}
}

func TestMsgCreate(t *testing.T) {
	app := getMockApp(t)

	// Construct genesis state
	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{fiftyRune},
	}
	accs := []auth.Account{acc1}

	// Initialize the chain (nil)
	mock.SetGenesis(app, accs)

	// A checkTx context (true)
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res1 := app.AccountMapper.GetAccount(ctxCheck, addr1)
	require.Equal(t, acc1, res1)

	// Create first clp and check it succeeds
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{createMsg1}, []int64{0}, []int64{0}, true, priv1)

	// Create second clp and check it succeeds
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{createMsg2}, []int64{0}, []int64{1}, true, priv1)

	// Create bad clp with duplicate ticker and check it fails
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{createMsg3}, []int64{0}, []int64{2}, false, priv1)

	// Create bad clp with bad ratio and check it fails
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{createMsg4}, []int64{0}, []int64{3}, false, priv1)

}
