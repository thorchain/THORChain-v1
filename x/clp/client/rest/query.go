package clp

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thorchain/THORChain/x/clp"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const storeName = "clp"

// register REST routes
func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc(
		"/clp",
		QueryAccountRequestHandlerFn(cdc, authcmd.GetAccountDecoder(cdc), ctx),
	).Methods("GET")
}

// query accountREST Handler
func QueryAccountRequestHandlerFn(cdc *wire.Codec, decoder auth.AccountDecoder, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(clp.GetTestKey(), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// the query will return empty if there is no data for this account
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Write(res)

	}
}
