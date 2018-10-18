package exchange

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	priv1 = ed25519.GenPrivKey()
	priv2 = ed25519.GenPrivKey()
	pub1  = priv1.PubKey()
	pub2  = priv2.PubKey()
	addr1 = sdk.AccAddress(pub1.Address())
	addr2 = sdk.AccAddress(pub2.Address())

	sellLimOrder1 = MsgCreateLimitOrder{
		Sender:    addr1,
		Kind:      SellOrder,
		Amount:    sdk.NewInt64Coin("ETH", 70),
		Price:     sdk.NewInt64Coin("RUNE", 5),
		ExpiresAt: time.Now().Add(time.Minute),
	}
	buyLimOrder1 = MsgCreateLimitOrder{
		Sender:    addr2,
		Kind:      BuyOrder,
		Amount:    sdk.NewInt64Coin("ETH", 90),
		Price:     sdk.NewInt64Coin("RUNE", 4),
		ExpiresAt: time.Now().Add(time.Minute),
	}
	buyLimOrder2 = MsgCreateLimitOrder{
		Sender:    addr2,
		Kind:      BuyOrder,
		Amount:    sdk.NewInt64Coin("ETH", 50),
		Price:     sdk.NewInt64Coin("RUNE", 8),
		ExpiresAt: time.Now().Add(time.Minute),
	}
	buyLimOrder3 = MsgCreateLimitOrder{
		Sender:    addr2,
		Kind:      BuyOrder,
		Amount:    sdk.NewInt64Coin("ETH", 50000),
		Price:     sdk.NewInt64Coin("RUNE", 800),
		ExpiresAt: time.Now().Add(time.Minute),
	}
	buyLimOrder4 = MsgCreateLimitOrder{
		Sender:    addr2,
		Kind:      BuyOrder,
		Amount:    sdk.NewInt64Coin("ETH", 40),
		Price:     sdk.NewInt64Coin("RUNE", 8),
		ExpiresAt: time.Now().Add(-time.Minute),
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	app := mock.NewApp()

	RegisterWire(app.Cdc)
	exchangeKey := sdk.NewKVStoreKey("exchangeTestAppKey")
	bankKeeper := bank.NewKeeper(app.AccountMapper)
	exchangeKeeper := NewKeeper(exchangeKey, bankKeeper, app.RegisterCodespace(DefaultCodespace))
	app.Router().AddRoute("exchange", NewHandler(exchangeKeeper))

	app.SetInitChainer(getInitChainer(app, exchangeKeeper, bankKeeper))

	require.NoError(t, app.CompleteSetup([]*sdk.KVStoreKey{exchangeKey}))
	return app
}

// overwrite the mock init chainer
func getInitChainer(mapp *mock.App, exchangeKeeper Keeper, bankKeeper bank.Keeper) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		mapp.InitChainer(ctx, req)

		InitGenesis(ctx, exchangeKeeper, DefaultGenesisState())

		return abci.ResponseInitChain{}
	}
}

func TestMsgCreateLimitOrder(t *testing.T) {
	app := getMockApp(t)

	// Construct genesis state
	seller := &auth.BaseAccount{
		Address:       addr1,
		Coins:         sdk.Coins{sdk.NewInt64Coin("ETH", 250)},
		AccountNumber: 0,
	}
	buyer := &auth.BaseAccount{
		Address:       addr2,
		Coins:         sdk.Coins{sdk.NewInt64Coin("RUNE", 2000)},
		AccountNumber: 1,
	}
	accs := []auth.Account{seller, buyer}

	// Initialize the chain (nil)
	mock.SetGenesis(app, accs)

	// A checkTx context (true)
	ctxCheck := app.BaseApp.NewContext(true, abci.Header{})
	res1 := app.AccountMapper.GetAccount(ctxCheck, addr1)
	require.Equal(t, seller, res1)
	res2 := app.AccountMapper.GetAccount(ctxCheck, addr2)
	require.Equal(t, buyer, res2)

	// Create sell order and check it succeeds
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{sellLimOrder1}, []int64{0}, []int64{0}, true, priv1)

	// Create buy order and check it succeeds
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{buyLimOrder1}, []int64{1}, []int64{0}, true, priv2)

	// Create buy order and check it succeeds
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{buyLimOrder2}, []int64{1}, []int64{1}, true, priv2)

	// Create bad buy order (not enough coins) and check it is rejected
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{buyLimOrder3}, []int64{1}, []int64{2}, false, priv2)

	// Create bad buy order (expired) and check it is rejected
	mock.SignCheckDeliver(t, app.BaseApp, []sdk.Msg{buyLimOrder4}, []int64{1}, []int64{3}, false, priv2)
}
