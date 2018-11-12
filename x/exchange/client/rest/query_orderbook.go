package rest

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/gorilla/mux"
	"github.com/thorchain/THORChain/x/exchange"
)

func registerQueryOrderbookRoute(ctx context.CLIContext, r *mux.Router, cdc *wire.Codec, _ keys.Keybase,
	storeName string) {
	r.HandleFunc("/exchange/query-order-book",
		handleQueryOrderbook(cdc, authcmd.GetAccountDecoder(cdc), ctx, storeName)).Methods("POST")
}

type queryOrderbookBody struct {
	Kind        string `json:"kind"`
	AmountDenom string `json:"amount_denom"`
	PriceDenom  string `json:"price_denom"`
}

func handleQueryOrderbook(cdc *wire.Codec, _ auth.AccountDecoder, ctx context.CLIContext,
	storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m queryOrderbookBody
		body, err := ioutil.ReadAll(r.Body)
		err = cdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		kind, err := exchange.ParseKind(m.Kind)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		if m.AmountDenom == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("amount_denom must not be empty"))
			return
		}

		if m.PriceDenom == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("price_denom must not be empty"))
			return
		}

		// ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

		res, err2 := ctx.QueryStore(exchange.MakeKeyOrderBook(kind, m.AmountDenom, m.PriceDenom), storeName)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err2.Error()))
			return
		}

		var orderbook exchange.OrderBook

		if len(res) > 0 {
			cdc.MustUnmarshalBinary(res, &orderbook)
		} else {
			orderbook = exchange.NewOrderBook(kind, m.AmountDenom, m.PriceDenom)
		}

		output, err2 := wire.MarshalJSONIndent(cdc, orderbook)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err2.Error()))
			return
		}

		w.Write(output)
	}
}
