package exchange

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Local code type
type CodeType = sdk.CodeType

//Exported code type numbers
const (
	DefaultCodespace sdk.CodespaceType = 15

	CodeInvalidKind        CodeType = 1
	CodeInvalidGenesis     CodeType = 2
	CodeOrderExpired       CodeType = 3
	CodeSameDenom          CodeType = 4
	CodeAmountNotPositive  CodeType = 5
	CodePriceNotPositive   CodeType = 6
	CodeOrderBookDirection CodeType = 7
)

// Invalid order kind error
func ErrInvalidKind(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidKind, "kind must be 'buy' for buy orders or 'sell' for sell orders")
}

// Invalid genesis error
func ErrInvalidGenesis(codespace sdk.CodespaceType, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGenesis, msg)
}

// Order expired error
func ErrOrderExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOrderExpired, "order must not be expired")
}

// Amount and price have same denom error
func ErrSameDenom(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSameDenom, "denom of amount and price must not be the same")
}

// Amount not positive error
func ErrAmountNotPositive(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAmountNotPositive, "amount must be positive")
}

// Price not positive error
func ErrPriceNotPositive(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePriceNotPositive, "price must be positive")
}

// Orderbook dircetion error
func ErrOrderBookDirection(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOrderBookDirection,
		"orderbook direction is not supported, please swap amount and price denoms")
}
