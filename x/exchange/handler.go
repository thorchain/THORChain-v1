package exchange

import (
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
	orderID, err := k.processLimitOrder(ctx, msg.Sender, msg.Kind, msg.Amount, msg.Price, msg.ExpiresAt)
	if err != nil {
		return err.Result()
	}
	resultLog := "json{\"orderFullyFilled\": true}json"
	if orderID > -1 {
		resultLog = fmt.Sprintf("json{\"orderFullyFilled\": false, \"orderID\": %v}json", orderID)
	}
	return sdk.Result{Log: resultLog}
}
