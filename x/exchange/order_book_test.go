package exchange

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestAddLimitOrderToEmptyBuyOrderBook(t *testing.T) {
	orderBook1 := NewOrderBook(BuyOrder, "ETH", "BTC")

	lo := LimitOrder{
		OrderID: 4,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	err := orderBook1.AddLimitOrder(lo)

	require.Nil(t, err)

	require.Len(t, orderBook1.Orders, 1)
	require.Equal(t, lo, orderBook1.Orders[0])
}

func TestAddLimitOrderToBuyOrderBook(t *testing.T) {
	orderBook1 := NewOrderBook(BuyOrder, "ETH", "BTC")

	lo1 := LimitOrder{
		OrderID: 1,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 80),
		Price:   sdk.NewCoin("BTC", 180),
	}
	lo2 := LimitOrder{
		OrderID: 2,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 20),
		Price:   sdk.NewCoin("BTC", 150),
	}
	lo3 := LimitOrder{
		OrderID: 3,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 200),
		Price:   sdk.NewCoin("BTC", 100),
	}

	orderBook1.Orders = []LimitOrder{lo1, lo2, lo3}

	lo4 := LimitOrder{
		OrderID: 4,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	err := orderBook1.AddLimitOrder(lo4)

	require.Nil(t, err)
	require.Len(t, orderBook1.Orders, 4)
	require.Equal(t, lo1, orderBook1.Orders[0])
	require.Equal(t, lo2, orderBook1.Orders[1])
	require.Equal(t, lo4, orderBook1.Orders[2])
	require.Equal(t, lo3, orderBook1.Orders[3])
}

func TestAddLimitOrderToEmptySellOrderBook(t *testing.T) {
	orderBook1 := NewOrderBook(SellOrder, "ETH", "BTC")

	lo := LimitOrder{
		OrderID: 4,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	err := orderBook1.AddLimitOrder(lo)

	require.Nil(t, err)
	require.Len(t, orderBook1.Orders, 1)
	require.Equal(t, lo, orderBook1.Orders[0])
}

func TestAddLimitOrderToSellOrderBook(t *testing.T) {
	orderBook1 := NewOrderBook(SellOrder, "ETH", "BTC")

	lo1 := LimitOrder{
		OrderID: 1,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 80),
		Price:   sdk.NewCoin("BTC", 90),
	}
	lo2 := LimitOrder{
		OrderID: 2,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 20),
		Price:   sdk.NewCoin("BTC", 150),
	}
	lo3 := LimitOrder{
		OrderID: 3,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 200),
		Price:   sdk.NewCoin("BTC", 151),
	}

	orderBook1.Orders = []LimitOrder{lo1, lo2, lo3}

	lo4 := LimitOrder{
		OrderID: 4,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	err := orderBook1.AddLimitOrder(lo4)

	require.Nil(t, err)
	require.Len(t, orderBook1.Orders, 4)
	require.Equal(t, lo1, orderBook1.Orders[0])
	require.Equal(t, lo2, orderBook1.Orders[1])
	require.Equal(t, lo4, orderBook1.Orders[2])
	require.Equal(t, lo3, orderBook1.Orders[3])
}

func TestRemoveExpiredLimitOrders(t *testing.T) {
	orderBook1 := NewOrderBook(BuyOrder, "ETH", "BTC")

	lo1 := LimitOrder{
		OrderID:   4,
		Kind:      BuyOrder,
		Amount:    sdk.NewCoin("ETH", 60),
		Price:     sdk.NewCoin("BTC", 150),
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	lo2 := LimitOrder{
		OrderID:   5,
		Kind:      BuyOrder,
		Amount:    sdk.NewCoin("ETH", 30),
		Price:     sdk.NewCoin("BTC", 130),
		ExpiresAt: time.Now().Add(time.Minute),
	}

	err := orderBook1.AddLimitOrder(lo1)
	require.Nil(t, err)
	err = orderBook1.AddLimitOrder(lo2)
	require.Nil(t, err)

	require.Len(t, orderBook1.Orders, 2)
	require.Equal(t, lo1, orderBook1.Orders[0])
	require.Equal(t, lo2, orderBook1.Orders[1])

	orderBook1.RemoveExpiredLimitOrders()
	require.Len(t, orderBook1.Orders, 1)
	require.Equal(t, lo2, orderBook1.Orders[0])
}

func TestRemoveFilledLimitOrders(t *testing.T) {
	orderBook1 := NewOrderBook(BuyOrder, "ETH", "BTC")

	lo1 := LimitOrder{
		OrderID:   4,
		Kind:      BuyOrder,
		Amount:    sdk.NewCoin("ETH", 0),
		Price:     sdk.NewCoin("BTC", 150),
		ExpiresAt: time.Now().Add(time.Minute),
	}
	lo2 := LimitOrder{
		OrderID:   5,
		Kind:      BuyOrder,
		Amount:    sdk.NewCoin("ETH", 30),
		Price:     sdk.NewCoin("BTC", 130),
		ExpiresAt: time.Now().Add(time.Minute),
	}

	err := orderBook1.AddLimitOrder(lo1)
	require.Nil(t, err)
	err = orderBook1.AddLimitOrder(lo2)
	require.Nil(t, err)

	require.Len(t, orderBook1.Orders, 2)
	require.Equal(t, lo1, orderBook1.Orders[0])
	require.Equal(t, lo2, orderBook1.Orders[1])

	orderBook1.RemoveFilledLimitOrders()
	require.Len(t, orderBook1.Orders, 1)
	require.Equal(t, lo2, orderBook1.Orders[0])
}
