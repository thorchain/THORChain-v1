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
		case types.MsgTradeBase:
			return handleMsgTradeBase(keeper, context, msg)
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
func handleMsgTradeBase(k Keeper, ctx sdk.Context, msg types.MsgTradeBase) sdk.Result {
	newCoinsAmount, err := k.tradeBase(ctx, msg.Sender, msg.Ticker, int64(msg.BaseCoinAmount))
	if err != nil {
		return err.Result()
	}
	resultLog := fmt.Sprintf("in: %v%v, out: %v%v", msg.BaseCoinAmount, k.baseCoinTicker, newCoinsAmount, msg.Ticker)
	return sdk.Result{Log: resultLog}
}
