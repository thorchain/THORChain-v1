package clp

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	clpPackage "github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const storeName = "clp"

// register REST routes
func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, baseCoinTicker string) {
	r.HandleFunc(
		"/clp/{ticker}",
		queryClpRequestHandlerFn(cdc, cliCtx, authcmd.GetAccountDecoder(cdc), baseCoinTicker),
	).Methods("GET")
	r.HandleFunc(
		"/clps",
		queryClpsRequestHandlerFn(cdc, cliCtx, authcmd.GetAccountDecoder(cdc), baseCoinTicker),
	).Methods("GET")

}

func queryClpRequestHandlerFn(
	cdc *wire.Codec, cliCtx context.CLIContext, accountDecoder auth.AccountDecoder, baseCoinTicker string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithCodec(cdc)

		vars := mux.Vars(r)
		ticker := vars["ticker"]

		res, err := cliCtx.QueryStore(clpPackage.MakeCLPStoreKey(ticker), storeName)
		// the query will return empty if there is no data for this account
		if err != nil || len(res) == 0 {
			w.WriteHeader(http.StatusNotFound)
			err := errors.Errorf("clp for ticker [%d] does not exist", ticker)
			w.Write([]byte(err.Error()))

			return
		}

		// decode the value
		var clp clpTypes.CLP
		err2 := cdc.UnmarshalBinary(res, &clp)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("couldn't parse query result. Result: %s. Error: %s", res, err2.Error())))
			return
		}

		// print out whole account
		output, err := clpToJSONOutput(cdc, cliCtx, accountDecoder, baseCoinTicker, w, &clp)
		if err != nil {
			return
		}
		w.Write(output)
	}
}

func getCLPAccountOutput(
	clp *clpTypes.CLP, cliCtx context.CLIContext, decoder auth.AccountDecoder, cdc *wire.Codec, baseCoinTicker string,
) ([]byte, error) {
	clpAddr := clp.AccountAddress

	accountRes, err := cliCtx.QueryStore(auth.AddressStoreKey(clpAddr), "acc")
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

	jsonOutput := fmt.Sprintf("{\"%v\":%v,\"%v\":%v,\"price\":%v}", baseCoinTicker, baseCoinAmount, clp.Ticker,
		tickerCoinAmount, price)

	return []byte(jsonOutput), nil
}

func clpToJSONOutput(cdc *wire.Codec, cliCtx context.CLIContext, accountDecoder auth.AccountDecoder,
	baseCoinTicker string, w http.ResponseWriter, clp *clpTypes.CLP) ([]byte, error) {
	output, err3 := cdc.MarshalJSON(clp)
	if err3 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("couldn't marshall query result. Error: %s", err3.Error())))
		return []byte(""), err3
	}
	accountOutput, err := getCLPAccountOutput(clp, cliCtx, accountDecoder, cdc, baseCoinTicker)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return []byte(""), err
	}

	finalOutput := append([]byte("{\"clp\":"), output...)
	finalOutput = append(finalOutput, []byte(",\"account\": ")...)
	finalOutput = append(finalOutput, accountOutput...)
	finalOutput = append(finalOutput, []byte("}")...)

	return finalOutput, nil
}

// query all Handler
func queryClpsRequestHandlerFn(
	cdc *wire.Codec, cliCtx context.CLIContext, accountDecoder auth.AccountDecoder, baseCoinTicker string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx = cliCtx.WithCodec(cdc)

		clpSubspace := []byte("clp:")

		res, err := cliCtx.QuerySubspace(clpSubspace, "clp")
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
			output, err := clpToJSONOutput(cdc, cliCtx, accountDecoder, baseCoinTicker, w, &clps[i])
			if err != nil {
				return
			}
			outputs = append(outputs, output)
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
