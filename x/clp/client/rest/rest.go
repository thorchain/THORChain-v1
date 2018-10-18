package clp

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase, baseCoinTicker string) {
	registerQueryRoutes(cliCtx, r, cdc, baseCoinTicker)
	registerTxRoutes(cliCtx, r, cdc, kb)
}
