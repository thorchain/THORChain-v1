package exchange

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrderBook that is stored in orderbook
type OrderBook struct {
	Key         []byte       `json:"key"`
	Kind        OrderKind    `json:"kind"`
	AmountDenom string       `json:"amountDenom"`
	PriceDenom  string       `json:"priceDenom"`
	Orders      []LimitOrder `json:"orders"`
}

// NewOrderBook creates a new order book for the given key
func NewOrderBook(kind OrderKind, amountDenom string, priceDenom string) OrderBook {
	newOrderBook := OrderBook{
		Key:         MakeKeyOrderBook(kind, amountDenom, priceDenom),
		Kind:        kind,
		AmountDenom: amountDenom,
		PriceDenom:  priceDenom,
		Orders:      []LimitOrder{}, // TODO use more efficient data structure like linked list for better performance
	}

	return newOrderBook
}

// Key for getting a specific order book from the store
func MakeKeyOrderBook(kind OrderKind, amountDenom string, priceDenom string) []byte {
	return []byte(fmt.Sprintf("orderBook:%v:%v:%v", kind, amountDenom, priceDenom))
}

// String provides a human-readable representation of an order
func (ob *OrderBook) String() string {
	return fmt.Sprintf("OrderBook{Key: %v, Kind: %v, AmountDenom: %v, PriceDenom: %v, Orders: %v}",
		ob.Key, ob.Kind, ob.AmountDenom, ob.PriceDenom, ob.Orders)
}

// AddLimitOrder adds a limit order into the orderbook. It does so by ordering it by price, then by time:
// A buy orderbook will have most expensive orders first, a sell order book will have cheapest first.
func (ob *OrderBook) AddLimitOrder(limitOrder LimitOrder) sdk.Error {
	// Check if correct order book selected
	if ob.AmountDenom != limitOrder.Amount.Denom {
		return sdk.ErrInternal(
			fmt.Sprintf("Amount denom does not match between limit order %v and order book %v", limitOrder, ob))
	}
	if ob.PriceDenom != limitOrder.Price.Denom {
		return sdk.ErrInternal(
			fmt.Sprintf("Price denom does not match between limit order %v and order book %v", limitOrder, ob))
	}
	if ob.Kind != limitOrder.Kind {
		return sdk.ErrInternal(
			fmt.Sprintf("Kind does not match between limit order %v and order book %v", limitOrder, ob))
	}

	if len(ob.Orders) == 0 {
		ob.Orders = []LimitOrder{limitOrder}
		return nil
	}

	// New orders list
	newOrders := make([]LimitOrder, len(ob.Orders)+1)

	added := false

	// TODO use more efficient search to insert in the right place, e. g. like binary search
	for i, order := range ob.Orders {
		if added {
			newOrders[i+1] = order
			continue
		}

		if shouldInsertBefore(ob.Kind, limitOrder.Price, order.Price) {
			newOrders[i] = limitOrder
			added = true

			newOrders[i+1] = order
		} else {
			newOrders[i] = order
		}
	}

	if !added {
		newOrders[len(ob.Orders)] = limitOrder
	}

	ob.Orders = newOrders

	return nil
}

// Removes limit orders that have expired
func (ob *OrderBook) RemoveExpiredLimitOrders() {
	// New orders list
	newOrders := make([]LimitOrder, 0, len(ob.Orders))

	for _, order := range ob.Orders {
		// skip expired orders
		if order.ExpiresAt.Before(time.Now()) {
			continue
		}

		// add others
		newOrders = append(newOrders, order)
	}

	ob.Orders = newOrders
}

// Removes limit orders that are filled (amount is zero), useful for efficient cleanup after order matching
func (ob *OrderBook) RemoveFilledLimitOrders() {
	// New orders list
	newOrders := make([]LimitOrder, 0, len(ob.Orders))

	for _, order := range ob.Orders {
		// skip expired orders
		if order.Amount.IsZero() {
			continue
		}

		// add others
		newOrders = append(newOrders, order)
	}

	ob.Orders = newOrders
}

// In a buy orderbook, highest prices come first, in a sell orderbook, lowest come first.
func shouldInsertBefore(kind OrderKind, order1Price sdk.Coin, order2Price sdk.Coin) bool {
	if kind == BuyOrder {
		if !order2Price.IsGTE(order1Price) {
			return true
		}
		return false
	}

	if !order1Price.IsGTE(order2Price) {
		return true
	}
	return false
}
