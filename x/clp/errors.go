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
	CodeCLPNotExists        CodeType = 143
	CodeInvalidTickerName   CodeType = 144
	CodeCLPParsing          CodeType = 145
	CodeNotEnoughCoins      CodeType = 146
)

//Reserve ratio error
func ErrInvalidReserveRatio(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidReserveRatio, "reserve ratio must be higher than zero and less than or equal to 100")
}

//Existing CLP error
func ErrCLPExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeCLPExists, "clp already exists for this ticker symbol")
}

//Non Existing CLP error
func ErrCLPNotExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeCLPNotExists, "clp already exists for this ticker symbol")
}

//Base ticker name error
func ErrInvalidTickerName(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTickerName, "cannot create clp for base ticker symbol")
}

//Parse CLP Error
func ErrCLPParsing(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeCLPParsing, "cannot parse text into clp")
}

//Not enough coins Error
func ErrNotEnoughCoins(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNotEnoughCoins, "not enough coins to make trade")
}
