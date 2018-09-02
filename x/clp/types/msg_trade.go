package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create type
type MsgTrade struct {
	Sender     sdk.AccAddress
	FromTicker string
	ToTicker   string
	FromAmount int
}

// new create message
func NewMsgMsgTrade(sender sdk.AccAddress, fromTicker string, toTicker string, fromAmount int) MsgTrade {
	return MsgTrade{
		Sender:     sender,
		FromTicker: fromTicker,
		ToTicker:   toTicker,
		FromAmount: fromAmount,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgTrade{}

//Get MsgCreate Type
func (msg MsgTrade) Type() string { return "clp" }

//Get Create Signers
func (msg MsgTrade) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgTrade) String() string {
	return fmt.Sprintf("MsgTrade{Sender: %v, FromTicker: %v, ToTicker: %v,  FromAmount: %v}", msg.Sender, msg.FromTicker, msg.ToTicker, msg.FromAmount)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgTrade) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgTrade) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
