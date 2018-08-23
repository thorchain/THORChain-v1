package clp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Local code type
type CodeType = sdk.CodeType

//Exported code type numbers
const (
	DefaultCodespace sdk.CodespaceType = 14

	CodeInvalidReserveRatio CodeType = 141
	CodeCLPExists           CodeType = 142
)

//Reserve ratio error
func ErrInvalidReserveRatio(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidReserveRatio, "reserve ratio must be higher than zero and less than or equal to 100")
}

//Existing CLP error
func ErrCLPExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeCLPExists, "clp already exists for this ticker symbol")
}
