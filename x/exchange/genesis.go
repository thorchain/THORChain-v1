package exchange

import sdk "github.com/cosmos/cosmos-sdk/types"

// GenesisState - all exchange state that must be provided at genesis
type GenesisState struct {
	StartingOrderID int64 `json:"starting_orderID"`
}

func NewGenesisState(startingOrderID int64) GenesisState {
	return GenesisState{
		StartingOrderID: startingOrderID,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		StartingOrderID: 1,
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	err := k.setInitialOrderID(ctx, data.StartingOrderID)
	if err != nil {
		panic(err)
	}
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	startingOrderID, _ := k.getNewOrderID(ctx)

	return GenesisState{
		StartingOrderID: startingOrderID,
	}
}
