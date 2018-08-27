package clp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const storeName = "clp"

// register REST routes
func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc(
		"/clp/{ticker}",
		QueryAccountRequestHandlerFn(cdc, authcmd.GetAccountDecoder(cdc), ctx),
	).Methods("GET")
}

// query accountREST Handler
func QueryAccountRequestHandlerFn(cdc *wire.Codec, decoder auth.AccountDecoder, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ticker := vars["ticker"]

		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(clp.MakeCLPStoreKey(ticker), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// the query will return empty if there is no data for this account
		if len(res) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// decode the value
		clp := new(clpTypes.CLP)

		err2 := cdc.UnmarshalBinary(res, &clp)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't parse query result. Result: %s. Error: %s", res, err2.Error())))
		}

		// print out whole account
		output, err3 := cdc.MarshalJSON(clp)
		if err3 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't marshall query result. Error: %s", err3.Error())))
			return
		}

		w.Write(output)

	}
}
