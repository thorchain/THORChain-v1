package exchange

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// slashing begin block functionality
func BeginBlocker(ctx sdk.Context, k Keeper) {
	k.refundExpiredLimitOrders(ctx)
}
