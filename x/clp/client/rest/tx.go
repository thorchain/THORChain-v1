package clp

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"
	"github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
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
	Ticker            string `json:"ticker"`
	TokenName         string `json:"token_name"`
	ReserveRatio      int    `json:"reserve_ratio"`
	InitialSupply     int64  `json:"initial_supply"`
	InitialRuneAmount int64  `json:"initial_rune_amount"`
	RuneAmount        int    `json:"rune_amount"`
	LocalAccountName  string `json:"name"`
	Password          string `json:"password"`
	ChainID           string `json:"chain_id"`
	AccountNumber     int64  `json:"account_number"`
	Sequence          int64  `json:"sequence"`
	Gas               int64  `json:"gas"`
}

var msgCdc = wire.NewCodec()

func init() {
	bank.RegisterWire(msgCdc)
	clp.RegisterWire(msgCdc)
}

func buildCreateMsg(from sdk.AccAddress, m sendBody) sdk.Msg {
	return clpTypes.NewMsgCreate(from, m.Ticker, m.TokenName, m.ReserveRatio, m.InitialSupply, m.InitialRuneAmount)
}

func buildTradeBaseMsg(from sdk.AccAddress, m sendBody) sdk.Msg {
	return clpTypes.NewMsgTradeBase(from, m.Ticker, m.RuneAmount)
}

//Ensure Account Number is added to context for transaction
func EnsureAccountNumber(ctx context.CoreContext, m sendBody, from sdk.AccAddress) (context.CoreContext, error) {
	if m.AccountNumber != 0 {
		return ctx, nil
	}
	accnum, err := ctx.GetAccountNumber(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to account number: %d\n", accnum)
	ctx = ctx.WithAccountNumber(accnum)
	return ctx, nil
}

//Ensure Sequence is added to context for transaction
func EnsureSequence(ctx context.CoreContext, m sendBody, from sdk.AccAddress) (context.CoreContext, error) {
	if m.Sequence != 0 {
		return ctx, nil
	}
	seq, err := ctx.NextSequence(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to next sequence number: %d\n", seq)
	ctx = ctx.WithSequence(seq)
	return ctx, nil
}

// SendRequestHandlerFn - http request handler to send coins to a address
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext, msgBuilder func(sdk.AccAddress, sendBody) sdk.Msg) http.HandlerFunc {
	ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
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

		//add account number and sequence
		ctx, err = EnsureAccountNumber(ctx, m, from)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ctx, err = EnsureSequence(ctx, m, from)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		//sign
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
