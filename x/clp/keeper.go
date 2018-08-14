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

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, bankKeeper bank.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{key, bankKeeper, codespace}
}
