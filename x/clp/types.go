package clp

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

// Get MsgTest Type
func (msg MsgTest) Type() string { return "clp" }

//Get Test Signers
func (msg MsgTest) GetSigners() []sdk.AccAddress {
	fmt.Println("getting test signers")
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
