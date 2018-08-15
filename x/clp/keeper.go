package clp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// Keeper - handlers sets/gets of custom variables for your module
type Keeper struct {
	storeKey sdk.StoreKey // The (unexposed) key used to access the store from the Context.

	bankKeeper bank.Keeper

	codespace sdk.CodespaceType
}

// Key for storing the test message!
var testKey = []byte("TestKey")

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, bankKeeper bank.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{key, bankKeeper, codespace}
}

// GetTest - returns the current test
func (k Keeper) GetTest(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	valueBytes := store.Get(testKey)
	return string(valueBytes)
}

// Implements sdk.AccountMapper.
func (k Keeper) setTest(ctx sdk.Context, newTestValue string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(testKey, []byte(newTestValue))
}

// InitGenesis - store the genesis trend
func InitGenesis(ctx sdk.Context, k Keeper, data Genesis) error {
	k.setTest(ctx, data.Test)
	return nil
}

// WriteGenesis - output the genesis trend
func WriteGenesis(ctx sdk.Context, k Keeper) Genesis {
	test := k.GetTest(ctx)
	return Genesis{test}
}
