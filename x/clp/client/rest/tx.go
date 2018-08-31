package clp

import (
	"io/ioutil"
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
	r.HandleFunc("/clp", SendRequestHandlerFn(cdc, kb, ctx, buildCreateMsg)).Methods("POST")
	r.HandleFunc("/clp_trade_rune", SendRequestHandlerFn(cdc, kb, ctx, buildTradeBaseMsg)).Methods("POST")
}

type baseSendBody struct {
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	AccountNumber    int64  `json:"account_number"`
	Sequence         int64  `json:"sequence"`
	Gas              int64  `json:"gas"`
}

type clpSendBody struct {
	// fees and gas is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Ticker            string `json:"ticker"`
	TokenName         string `json:"token_name"`
	ReserveRatio      int    `json:"reserve_ratio"`
	InitialSupply     int64  `json:"initial_supply"`
	InitialRuneAmount int64  `json:"initial_rune_amount"`
	RuneAmount        int    `json:"rune_amount"`
}

func buildCreateMsg(w http.ResponseWriter, cdc *wire.Codec, from sdk.AccAddress, body []byte) (sdk.Msg, error) {
	var m clpSendBody
	err := cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	return clpTypes.NewMsgCreate(from, m.Ticker, m.TokenName, m.ReserveRatio, m.InitialSupply, m.InitialRuneAmount), nil
}

func buildTradeBaseMsg(w http.ResponseWriter, cdc *wire.Codec, from sdk.AccAddress, body []byte) (sdk.Msg, error) {
	var m clpSendBody
	err := cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	return clpTypes.NewMsgTradeBase(from, m.Ticker, m.RuneAmount), nil
}

func extractRequest(w http.ResponseWriter, r *http.Request, cdc *wire.Codec) (baseSendBody, []byte, error) {
	var m baseSendBody
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return baseSendBody{}, nil, err
	}
	err = cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return baseSendBody{}, nil, err
	}
	return m, body, nil
}

func setupContext(w http.ResponseWriter, ctx context.CoreContext, m baseSendBody, from sdk.AccAddress) (context.CoreContext, error) {
	// add gas to context
	ctx = ctx.WithGas(m.Gas)

	// add chain-id to context
	ctx = ctx.WithChainID(m.ChainID)

	//add account number and sequence
	ctx, err := lcdhelpers.EnsureAccountNumber(ctx, m.AccountNumber, from)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return ctx, err
	}
	ctx, err = lcdhelpers.EnsureSequence(ctx, m.Sequence, from)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return ctx, err
	}
	return ctx, nil
}

func getFromAddress(w http.ResponseWriter, kb keys.Keybase, localAccountName string) (sdk.AccAddress, error) {
	info, err := kb.Get(localAccountName)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return sdk.AccAddress{}, err
	}

	from := sdk.AccAddress(info.GetPubKey().Address())
	return from, nil
}

func processMsg(w http.ResponseWriter, ctx context.CoreContext, localAccountName string, password string, cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	//sign
	txBytes, err := ctx.SignAndBuild(localAccountName, password, []sdk.Msg{msg}, cdc)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	// send
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return nil, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return nil, err
	}
	return output, err
}

// SendRequestHandlerFn - http request handler to send coins to a address
func SendRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext, msgBuilder func(http.ResponseWriter, *wire.Codec, sdk.AccAddress, []byte) (sdk.Msg, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, body, err := extractRequest(w, r, cdc)
		if err != nil {
			return
		}

		from, err := getFromAddress(w, kb, m.LocalAccountName)
		if err != nil {
			return
		}

		ctx, err = setupContext(w, ctx, m, from)
		if err != nil {
			return
		}

		// build message
		msg, err := msgBuilder(w, cdc, from, body)
		if err != nil {
			return
		}

		output, err := processMsg(w, ctx, m.LocalAccountName, m.Password, cdc, msg)
		if err != nil {
			return
		}

		w.Write(output)
	}
}
