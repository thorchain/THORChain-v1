package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create type
type MsgCreate struct {
	Sender                sdk.AccAddress
	Ticker                string
	Name                  string
	Decimals              uint8
	ReserveRatio          int
	InitialSupply         int64
	InitialBaseCoinAmount int64
}

// new create message
func NewMsgCreate(sender sdk.AccAddress, ticker string, name string, decimals uint8, reserveRatio int,
	initialSupply int64, initialBaseCoinAmount int64) MsgCreate {
	return MsgCreate{
		Sender:                sender,
		Ticker:                ticker,
		Name:                  name,
		Decimals:              decimals,
		ReserveRatio:          reserveRatio,
		InitialSupply:         initialSupply,
		InitialBaseCoinAmount: initialBaseCoinAmount,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgCreate{}

//Get MsgCreate Type
func (msg MsgCreate) Type() string { return "clp" }

//Get Create Signers
func (msg MsgCreate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgCreate) String() string {
	return fmt.Sprintf("MsgCreate{Sender: %v, Ticker: %v, Name: %v, Decimals: %v, ReserveRatio: %v, InitialSupply: %v, InitialBaseCoinAmount: %v}", msg.Sender, msg.Ticker, msg.Name, msg.Decimals, msg.ReserveRatio, msg.InitialSupply, msg.InitialBaseCoinAmount)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgCreate) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgCreate) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
