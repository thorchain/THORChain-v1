package clp

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// genesis state - specify genesis test
type Genesis struct {
	Test string `json:"test"`
}

// Test type
type MsgTest struct {
	Sender sdk.AccAddress
	Test   string
}

// new test message
func NewMsgTest(sender sdk.AccAddress, test string) MsgTest {
	return MsgTest{
		Sender: sender,
		Test:   test,
	}
}

// enforce the msg type at compile time
var _ sdk.Msg = MsgTest{}

//Get MsgTest Type
func (msg MsgTest) Type() string { return "clp" }

//Get Test Signers
func (msg MsgTest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func (msg MsgTest) String() string {
	return fmt.Sprintf("MsgTest{Sender: %v, Test: %v}", msg.Sender, msg.Test)
}

// Validate Basic is used to quickly disqualify obviously invalid messages quickly
func (msg MsgTest) ValidateBasic() sdk.Error {
	if len(msg.Sender) == 0 {
		return sdk.ErrUnknownAddress(msg.Sender.String()).TraceSDK("")
	}
	if strings.Contains(msg.Test, "bad") {
		return sdk.ErrUnauthorized("").TraceSDK("bad test")
	}
	return nil
}

// Get the bytes for the message signer to sign on
func (msg MsgTest) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// Create type
type MsgCreate struct {
	Sender       sdk.AccAddress
	Ticker       string
	Name         string
	ReserveRatio int
}

// new test message
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

// CLP can mint new coins
type CLP struct {
	Creator      sdk.AccAddress `json:"creator"`
	Ticker       string         `json:"ticker"`
	Name         string         `json:"name"`
	ReserveRatio int            `json:"reserveRatio"`
}

func NewCLP(sender sdk.AccAddress, ticker string, name string, reserveRatio int) CLP {
	return CLP{
		Creator:      sender,
		Ticker:       ticker,
		Name:         name,
		ReserveRatio: reserveRatio,
	}
}

// String provides a human-readable representation of a coin
func (clp CLP) String() string {
	return fmt.Sprintf("%v%v%v%v", clp.Creator, clp.Ticker, clp.Name, clp.ReserveRatio)
}
