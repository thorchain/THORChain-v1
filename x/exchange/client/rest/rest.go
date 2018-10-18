package rest

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes registers exchange related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase, storeName string) {
	// registerCreateLimitOrderRoute(cliCtx, r, cdc, kb)
	registerQueryOrderbookRoute(cliCtx, r, cdc, kb, storeName)
}
