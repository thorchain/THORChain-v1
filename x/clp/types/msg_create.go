package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create type
type MsgCreate struct {
	Sender       sdk.AccAddress
	Ticker       string
	Name         string
	ReserveRatio int
}

// new create message
func NewMsgCreate(sender sdk.AccAddress, ticker string, name string, reserveRatio int) MsgCreate {
	return MsgCreate{
		Sender:       sender,
		Ticker:       ticker,
		Name:         name,
		ReserveRatio: reserveRatio,
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
	return fmt.Sprintf("MsgCreate{Sender: %v, Ticker: %v,  Name: %v,  ReserveRatio: %v}", msg.Sender, msg.Ticker, msg.Name, msg.ReserveRatio)
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
