package clp

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/thorchain/THORChain/x/clp/types"
)

// NewHandler returns a handler for "clp" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(context sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case types.MsgCreate:
			return handleMsgCreate(keeper, context, msg)
		case types.MsgTrade:
			return handleMsgTrade(keeper, context, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized CLP Msg type: %v", reflect.TypeOf(msg).Name())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgCreateCLP This is the engine of your module
func handleMsgCreate(k Keeper, ctx sdk.Context, msg types.MsgCreate) sdk.Result {
	err := k.create(ctx, msg.Sender, msg.Ticker, msg.Name, msg.ReserveRatio, msg.InitialSupply, msg.InitialBaseCoinAmount)
	if err != nil {
		return err.Result()
	}
	fmt.Println("New CLP created!")
	return sdk.Result{}
}

// Handle MsgCreateCLP This is the engine of your module
func handleMsgTrade(k Keeper, ctx sdk.Context, msg types.MsgTrade) sdk.Result {
	newCoinsAmount, err := k.trade(ctx, msg.Sender, msg.FromTicker, msg.ToTicker, int64(msg.FromAmount))
	if err != nil {
		return err.Result()
	}
	resultLog := fmt.Sprintf("in: %v%v, out: %v%v", msg.FromAmount, msg.FromTicker, newCoinsAmount, msg.ToTicker)
	return sdk.Result{Log: resultLog}
}
