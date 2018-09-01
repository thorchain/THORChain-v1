package clp

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	lcdhelpers "github.com/cosmos/cosmos-sdk/client/lcd/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/clp", lcdhelpers.RequestHandlerFn(cdc, kb, ctx, buildCreateMsg)).Methods("POST")
	r.HandleFunc("/clp_trade_rune", lcdhelpers.RequestHandlerFn(cdc, kb, ctx, buildTradeBaseMsg)).Methods("POST")
}

type clpCreateBody struct {
	Ticker            string `json:"ticker"`
	TokenName         string `json:"token_name"`
	ReserveRatio      int    `json:"reserve_ratio"`
	InitialSupply     int64  `json:"initial_supply"`
	InitialRuneAmount int64  `json:"initial_rune_amount"`
}

type clpTradeBody struct {
	Ticker     string `json:"ticker"`
	RuneAmount int    `json:"rune_amount"`
}

func buildCreateMsg(w http.ResponseWriter, cdc *wire.Codec, from sdk.AccAddress, body []byte, routeVars map[string]string) (sdk.Msg, error) {
	var m clpCreateBody
	err := cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	return clpTypes.NewMsgCreate(from, m.Ticker, m.TokenName, m.ReserveRatio, m.InitialSupply, m.InitialRuneAmount), nil
}

func buildTradeBaseMsg(w http.ResponseWriter, cdc *wire.Codec, from sdk.AccAddress, body []byte, routeVars map[string]string) (sdk.Msg, error) {
	var m clpTradeBody
	err := cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	return clpTypes.NewMsgTradeBase(from, m.Ticker, m.RuneAmount), nil
}
