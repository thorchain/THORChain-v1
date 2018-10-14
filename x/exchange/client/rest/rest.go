package rest

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes registers exchange related REST handlers to a router
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase, storeName string) {
	// registerCreateLimitOrderRoute(ctx, r, cdc, kb)
	registerQueryOrderbookRoute(ctx, r, cdc, kb, storeName)
}
