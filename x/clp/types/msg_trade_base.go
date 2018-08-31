package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Create type
type MsgTradeBase struct {
	Sender         sdk.AccAddress
	Ticker         string
	BaseCoinAmount int
}

// new create message
func NewMsgTradeBase(sender sdk.AccAddress, ticker string, baseCoinAmount int) MsgTradeBase {
	return MsgTradeBase{
		Sender:         sender,
		Ticker:         ticker,
		BaseCoinAmount: baseCoinAmount,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgTradeBase{}

//Get MsgCreate Type
func (msg MsgTradeBase) Type() string { return "clp" }

//Get Create Signers
func (msg MsgTradeBase) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgTradeBase) String() string {
	return fmt.Sprintf("MsgTradeBase{Sender: %v, Ticker: %v,  BaseCoinAmount: %v}", msg.Sender, msg.Ticker, msg.BaseCoinAmount)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgTradeBase) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgTradeBase) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}
