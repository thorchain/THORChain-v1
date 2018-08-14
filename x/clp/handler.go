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
		case MsgTest:
			return handleMsgTest(keeper, context, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized test Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgQuiz This is the engine of your module
func handleMsgTest(k Keeper, ctx sdk.Context, msg MsgTest) sdk.Result {
	// k.setTrend(ctx, msg.Cool)
	fmt.Println("test message being handled!")
	return sdk.Result{}
}
