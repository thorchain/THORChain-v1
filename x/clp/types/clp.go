package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
