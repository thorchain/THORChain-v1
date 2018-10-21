package clp

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/clp", postClpHandlerFn(cdc, kb, cliCtx)).Methods("POST")
	r.HandleFunc("/clp_trade", postClpHandlerTradeFn(cdc, kb, cliCtx)).Methods("POST")
}

type clpCreateBody struct {
	BaseReq           baseReq `json:"base_req"`
	Ticker            string  `json:"ticker"`
	TokenName         string  `json:"token_name"`
	ReserveRatio      int     `json:"reserve_ratio"`
	InitialSupply     int64   `json:"initial_supply"`
	InitialRuneAmount int64   `json:"initial_rune_amount"`
}

type clpTradeBody struct {
	BaseReq    baseReq `json:"base_req"`
	FromTicker string  `json:"from_ticker"`
	ToTicker   string  `json:"to_ticker"`
	FromAmount int     `json:"from_amount"`
}

func postClpHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req clpCreateBody
		err := buildReq(w, r, cdc, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		info, err := kb.Get(req.BaseReq.Name)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// create the message
		msg := clpTypes.NewMsgCreate(sdk.AccAddress(info.GetPubKey().Address()), req.Ticker, req.TokenName,
			req.ReserveRatio, req.InitialSupply, req.InitialRuneAmount)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		signAndBuild(w, cliCtx, req.BaseReq, msg, cdc)
	}
}

func postClpHandlerTradeFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req clpTradeBody
		err := buildReq(w, r, cdc, &req)
		if err != nil {
			return
		}

		if !req.BaseReq.baseReqValidate(w) {
			return
		}

		info, err := kb.Get(req.BaseReq.Name)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// create the message
		msg := clpTypes.NewMsgTrade(sdk.AccAddress(info.GetPubKey().Address()), req.FromTicker, req.ToTicker,
			req.FromAmount)
		err = msg.ValidateBasic()
		if err != nil {
			writeErr(&w, http.StatusBadRequest, err.Error())
			return
		}

		signAndBuild(w, cliCtx, req.BaseReq, msg, cdc)
	}
}
