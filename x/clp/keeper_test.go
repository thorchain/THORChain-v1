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

func TestCoolKeeperTest(t *testing.T) {
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

func TestCoolKeeperCreate(t *testing.T) {
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)

	clpKey := sdk.NewKVStoreKey("clpTestKey")
	keeper := setupKeeper(clpKey)
	ctx := setupContext(clpKey)

	err := InitGenesis(ctx, keeper, Genesis{"first_test_value"})
	require.Nil(t, err)

	genesis := WriteGenesis(ctx, keeper)
	require.Nil(t, err)
	require.Equal(t, genesis, Genesis{"first_test_value"})

	ticker := "eth"
	name := "ethereum"
	reserveRatio := 1
	ticker2 := "btc"
	name2 := "bitcoin"
	reserveRatio2 := 0
	ticker3 := "cos"
	name3 := "cosmos"
	reserveRatio3 := 200
	validCLP := NewCLP(addr1, ticker, name, reserveRatio)
	validCLPBytes, err := cdc.MarshalBinary(validCLP)
	validCLPString := string(validCLPBytes)

	//Test happy path creation
	err1 := keeper.create(ctx, addr1, ticker, name, reserveRatio)
	require.Nil(t, err1)
	//Get created CLP and confirm values are correct
	newClp := keeper.GetCLP(ctx, ticker)
	require.Equal(t, newClp, validCLPString)

	//Test duplicate ticker
	err2 := keeper.create(ctx, addr1, ticker, name2, reserveRatio)
	require.Error(t, err2)

	//Test bad ratios
	err4 := keeper.create(ctx, addr1, ticker2, name2, reserveRatio2)
	require.Error(t, err4)
	err5 := keeper.create(ctx, addr1, ticker3, name3, reserveRatio3)
	require.Error(t, err5)

}
