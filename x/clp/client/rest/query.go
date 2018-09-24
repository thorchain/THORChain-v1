package clp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	clpPackage "github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const storeName = "clp"

// register REST routes
func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, baseCoinTicker string) {
	r.HandleFunc(
		"/clp/{ticker}",
		QueryAccountRequestHandlerFn(cdc, authcmd.GetAccountDecoder(cdc), ctx, baseCoinTicker),
	).Methods("GET")
	r.HandleFunc(
		"/clps",
		QueryAllRequestHandlerFn(cdc, authcmd.GetAccountDecoder(cdc), ctx, baseCoinTicker),
	).Methods("GET")

}

// query accountREST Handler
func QueryAccountRequestHandlerFn(cdc *wire.Codec, decoder auth.AccountDecoder, ctx context.CoreContext, baseCoinTicker string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ticker := vars["ticker"]

		ctx := context.NewCoreContextFromViper()

		res, err := ctx.QueryStore(clpPackage.MakeCLPStoreKey(ticker), storeName)
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

		accountOutput, err := getCLPAccountOutput(clp, ctx, decoder, cdc, baseCoinTicker)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		finalOutput := append([]byte("{\"clp\":"), output...)
		finalOutput = append(finalOutput, []byte(",\"account\": ")...)
		finalOutput = append(finalOutput, accountOutput...)
		finalOutput = append(finalOutput, []byte("}")...)
		w.Write(finalOutput)
	}
}

func getCLPAccountOutput(clp *clpTypes.CLP, ctx context.CoreContext, decoder auth.AccountDecoder, cdc *wire.Codec, baseCoinTicker string) ([]byte, error) {
	clpAddr := clp.AccountAddress

	accountRes, err := ctx.QueryStore(auth.AddressStoreKey(clpAddr), "acc")
	if err != nil {
		return nil, fmt.Errorf("couldn't query clp token balances. Error: %s", err.Error())
	}

	// decode the value
	clpAccount, err := decoder(accountRes)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse query result for clp token balances. Error: %s", err.Error())
	}

	tickerCoinAmount := clpAccount.GetCoins().AmountOf(clp.Ticker)
	baseCoinAmount := clpAccount.GetCoins().AmountOf(baseCoinTicker)
	price := clpPackage.CalculateCLPPrice(clp, clpAccount.GetCoins(), 1, baseCoinTicker)

	jsonOutput := fmt.Sprintf("{\"%v\":%v,\"%v\":%v,\"price\":%v}", baseCoinTicker, baseCoinAmount, clp.Ticker, tickerCoinAmount, price)

	return []byte(jsonOutput), nil
}

// query all Handler
func QueryAllRequestHandlerFn(cdc *wire.Codec, decoder auth.AccountDecoder, ctx context.CoreContext, baseCoinTicker string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.NewCoreContextFromViper()

		clpSubspace := []byte("clp:")

		res, err := ctx.QuerySubspace(cdc, clpSubspace, "clp")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if res == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var clps []clpTypes.CLP
		for i := 0; i < len(res); i++ {
			clp := new(clpTypes.CLP)
			err2 := cdc.UnmarshalBinary(res[i].Value, &clp)
			if err2 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't parse query result. Result: %s. Error: %s", res, err2.Error())))
				return
			}
			clps = append(clps, *clp)
		}
		fmt.Printf("CLP details \n\n")

		var outputs [][]byte
		for i := 0; i < len(clps); i++ {
			output, err3 := cdc.MarshalJSON(clps[i])
			if err3 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("couldn't marshall query result. Error: %s", err3.Error())))
				return
			}
			accountOutput, err := getCLPAccountOutput(&clps[i], ctx, decoder, cdc, baseCoinTicker)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			finalOutput := append([]byte("{\"clp\":"), output...)
			finalOutput = append(finalOutput, []byte(",\"account\": ")...)
			finalOutput = append(finalOutput, accountOutput...)
			finalOutput = append(finalOutput, []byte("}")...)

			outputs = append(outputs, finalOutput)
		}
		w.Write([]byte("["))

		for i := 0; i < len(outputs); i++ {
			w.Write(outputs[i])
			if i != len(outputs)-1 {
				w.Write([]byte(","))
			}
		}
		w.Write([]byte("]"))

	}
}
