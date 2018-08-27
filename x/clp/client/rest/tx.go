package clp

import (
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"
	"github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/clp", SendRequestHandlerFn(cdc, kb, ctx, buildCreateMsg)).Methods("POST")
	r.HandleFunc("/clp_trade_rune", SendRequestHandlerFn(cdc, kb, ctx, buildTradeBaseMsg)).Methods("POST")
}

type sendBody struct {
	// fees and gas is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Ticker           string `json:"ticker"`
	TokenName        string `json:"token_name"`
	ReserveRatio     int    `json:"reserve_ratio"`
	RuneAmount       int    `json:"rune_amount"`
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	AccountNumber    int64  `json:"account_number"`
	Sequence         int64  `json:"sequence"`
	Gas              int64  `json:"gas"`
}

var msgCdc = wire.NewCodec()

func init() {
	bank.RegisterWire(msgCdc)
	clp.RegisterWire(msgCdc)
}

func buildCreateMsg(from sdk.AccAddress, m sendBody) sdk.Msg {
	return clpTypes.NewMsgCreate(from, m.Ticker, m.TokenName, m.ReserveRatio)
}

func buildTradeBaseMsg(from sdk.AccAddress, m sendBody) sdk.Msg {
	return clpTypes.NewMsgTradeBase(from, m.Ticker, m.RuneAmount)
}

// SendRequestHandlerFn - http request handler to send coins to a address
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext, msgBuilder func(sdk.AccAddress, sendBody) sdk.Msg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m sendBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = msgCdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		from := sdk.AccAddress(info.GetPubKey().Address())

		// build message
		msg := msgBuilder(from, m)

		if err != nil { // XXX rechecking same error ?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// add gas to context
		ctx = ctx.WithGas(m.Gas)
		// add chain-id to context
		ctx = ctx.WithChainID(m.ChainID)

		// sign
		ctx = ctx.WithAccountNumber(m.AccountNumber)
		ctx = ctx.WithSequence(m.Sequence)
		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, []sdk.Msg{msg}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := wire.MarshalJSONIndent(cdc, res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
