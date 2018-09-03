package clp

import (
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/thorchain/THORChain/x/clp/types"
)

//Function to register a codec with this packages concretes/interfaces
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(types.MsgCreate{}, "clp/MsgCreate", nil)
	cdc.RegisterConcrete(types.MsgTrade{}, "clp/MsgTrade", nil)
}
