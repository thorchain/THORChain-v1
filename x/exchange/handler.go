package exchange

import (
	"encoding/json"
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "exchange" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateLimitOrder:
			return handleMsgCreateLimitOrder(keeper, ctx, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized exchange msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle eMsgCreateLimitOrder This is the engine of your module
func handleMsgCreateLimitOrder(k Keeper, ctx sdk.Context, msg MsgCreateLimitOrder) sdk.Result {
	processed, filled, err := k.processLimitOrder(ctx, msg.Sender, msg.Kind, msg.Amount, msg.Price, msg.ExpiresAt)

	if err != nil {
		return err.Result()
	}

	type toJSON struct {
		Processed ProcessedLimitOrder `json:"processed"`
		Filled    []FilledLimitOrder  `json:"filled"`
	}

	b, err2 := json.Marshal(toJSON{processed, filled})
	if err2 != nil {
		return sdk.ErrInternal(fmt.Sprintf("Error marshalling json: %v", err2)).Result()
	}

	resultLog := fmt.Sprintf("json%vjson", string(b))

	return sdk.Result{Log: resultLog}
}
