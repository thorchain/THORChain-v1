package clp

import (
	"testing"

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

func setupKeeper(clpKey *sdk.KVStoreKey) Keeper {
	cdc := wire.NewCodec()
	auth.RegisterBaseAccount(cdc)
	accountMapper := auth.NewAccountMapper(cdc, clpKey, auth.ProtoBaseAccount)
	bankKeeper := bank.NewKeeper(accountMapper)
	clpKeeper := NewKeeper(clpKey, bankKeeper, DefaultCodespace)
	return clpKeeper
}

func TestCoolKeeper(t *testing.T) {
	clpKey := sdk.NewKVStoreKey("clpTestKey")
	keeper := setupKeeper(clpKey)
	ctx := setupContext(clpKey)

	err := InitGenesis(ctx, keeper, Genesis{"first_test_value"})
	require.Nil(t, err)

	genesis := WriteGenesis(ctx, keeper)
	require.Nil(t, err)
	require.Equal(t, genesis, Genesis{"first_test_value"})

	res := keeper.GetTest(ctx)
	require.Equal(t, res, "first_test_value")

	keeper.setTest(ctx, "second_test_value")
	res = keeper.GetTest(ctx)
	require.Equal(t, res, "second_test_value")
}
