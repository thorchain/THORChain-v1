package clp

import "github.com/cosmos/cosmos-sdk/wire"

func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgTest{}, "clp/MsgTest", nil)
}
