package clp

import "github.com/cosmos/cosmos-sdk/wire"

//Function to register a codec with this packages concretes/interfaces
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterConcrete(MsgCreate{}, "clp/MsgCreate", nil)
}
