package exchange

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrderKind is the kind an order can have, namely buy or sell
type OrderKind byte

const (
	// BuyOrder is the kind for buy orders
	BuyOrder OrderKind = 0x01
	// SellOrder is the kind for sell orders
	SellOrder OrderKind = 0x02
)

// LimitOrder that is stored in orderbook
type LimitOrder struct {
	OrderID   int64          `json:"order_id"`
	Sender    sdk.AccAddress `json:"sender"`
	Kind      OrderKind      `json:"kind"`
	Amount    sdk.Coin       `json:"amount"`
	Price     sdk.Coin       `json:"price"`
	ExpiresAt time.Time      `json:"expires_at"`
}

// ProcessedLimitOrder is return after order matching as a log entry to signal whether the order is fully filled or
// there is an open amount still sitting in the orderbook
type ProcessedLimitOrder struct {
	OrderID    int64    `json:"order_id"`
	OpenAmount sdk.Coin `json:"open_amt"`
}

// FilledLimitOrder is return after order matching as a log entry to signal what orders have been filled with
// which amount
type FilledLimitOrder struct {
	OrderID      int64    `json:"order_id"`
	FilledAmount sdk.Coin `json:"filled_amt"`
	FilledPrice  sdk.Coin `json:"filled_price"`
}

// NewLimitOrder creates a new limit order
func NewLimitOrder(orderID int64, sender sdk.AccAddress, kind OrderKind, amount sdk.Coin, price sdk.Coin, expiresAt time.Time) LimitOrder {
	newLimitOrder := LimitOrder{
		OrderID:   orderID,
		Sender:    sender,
		Kind:      kind,
		Amount:    amount,
		Price:     price,
		ExpiresAt: expiresAt,
	}

	return newLimitOrder
}

// String provides a human-readable representation of an order
func (lo *LimitOrder) String() string {
	return fmt.Sprintf("LimitOrder{Sender: %v, Kind: %v, Amount: %v, Price: %v, ExpiresAt: %v}",
		lo.Sender, lo.Kind, lo.Amount, lo.Price, lo.ExpiresAt)
}

// DoesFill checks if the stored order does fill the given parameters. If it does, it returns true, the amount that can
// be filled and the price at which it will be filled.
func (lo *LimitOrder) DoesFill(kind OrderKind, amount, price sdk.Coin) (bool, sdk.Coin, sdk.Coin) {
	// Check if amomunts and prices match
	if lo.Amount.Denom != amount.Denom {
		panic(fmt.Sprintf("Amount denom does not match between stored order %v and order to fill %v", lo,
			amount.Denom))
	}
	if lo.Price.Denom != price.Denom {
		panic(fmt.Sprintf("Price denom does not match between stored order %v and order to fill %v", lo,
			price.Denom))
	}
	if lo.Kind == kind {
		panic(fmt.Sprintf("Kind does not match between stored order %v and order to fill %v", lo, kind))
	}

	fillAmount := lo.Amount

	if !amount.IsGTE(lo.Amount) {
		// only part of the stored order's amount will be filled
		fillAmount = amount
	}

	if kind == BuyOrder {
		if price.IsGTE(lo.Price) {
			return true, fillAmount, lo.Price
		}
		return false, fillAmount, lo.Price
	}

	if lo.Price.IsGTE(price) {
		return true, fillAmount, lo.Price
	}
	return false, fillAmount, lo.Price
}
