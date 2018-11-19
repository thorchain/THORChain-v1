package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CLP can mint new coins
type CLP struct {
	Creator        sdk.AccAddress `json:"creator"`
	Ticker         string         `json:"ticker"`
	Name           string         `json:"name"`
	Decimals       uint8          `json:"decimals"`
	ReserveRatio   int            `json:"reserveRatio"`
	InitialSupply  int64          `json:"initialSupply"`
	CurrentSupply  int64          `json:"currentSupply"`
	AccountAddress sdk.AccAddress `json:"account_address"`
}

func NewCLP(sender sdk.AccAddress, ticker string, name string, decimals uint8, reserveRatio int, initialSupply int64,
	accountAddress sdk.AccAddress) CLP {
	newClp := CLP{
		Creator:        sender,
		Ticker:         ticker,
		Name:           name,
		Decimals:       decimals,
		ReserveRatio:   reserveRatio,
		InitialSupply:  initialSupply,
		CurrentSupply:  initialSupply,
		AccountAddress: accountAddress,
	}
	return newClp
}

func NewCLPAddress(ticker string) sdk.AccAddress {
	return sdk.AccAddress([]byte(fmt.Sprintf("t0clpaddr%v", ticker)))
}

// String provides a human-readable representation of a coin
func (clp CLP) String() string {
	return fmt.Sprintf("%v%v%v%v%v%v%v%v", clp.Creator, clp.Ticker, clp.Name, clp.Decimals, clp.ReserveRatio,
		clp.InitialSupply, clp.CurrentSupply, clp.AccountAddress)
}
