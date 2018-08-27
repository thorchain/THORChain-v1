package clp

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "clp" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(context sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreate:
			return handleMsgCreate(keeper, context, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized CLP Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgCreateCLP This is the engine of your module
func handleMsgCreate(k Keeper, ctx sdk.Context, msg MsgCreate) sdk.Result {
	err := k.create(ctx, msg.Sender, msg.Ticker, msg.Name, msg.ReserveRatio)
	if err != nil {
		return err.Result()
	}
	fmt.Println("New CLP created!")
	return sdk.Result{}
}
