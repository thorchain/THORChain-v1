package exchange

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDoesLimitOrderFillSad(t *testing.T) {
	lo := LimitOrder{
		OrderID: 42,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	require.PanicsWithValue(t, "Amount denom does not match between stored order LimitOrder{Sender: , Kind: 1, Amount: 60ETH, Price: 150BTC, ExpiresAt: 0001-01-01 00:00:00 +0000 UTC} and order to fill RUNE", func() {
		lo.DoesFill(BuyOrder, sdk.NewCoin("RUNE", 50), sdk.NewCoin("BTC", 140))
	})
	require.PanicsWithValue(t, "Price denom does not match between stored order LimitOrder{Sender: , Kind: 1, Amount: 60ETH, Price: 150BTC, ExpiresAt: 0001-01-01 00:00:00 +0000 UTC} and order to fill RUNE", func() {
		lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 50), sdk.NewCoin("RUNE", 140))
	})
	require.PanicsWithValue(t, "Kind does not match between stored order LimitOrder{Sender: , Kind: 1, Amount: 60ETH, Price: 150BTC, ExpiresAt: 0001-01-01 00:00:00 +0000 UTC} and order to fill 1", func() {
		lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 50), sdk.NewCoin("BTC", 140))
	})
}

func TestDoesLimitOrderFillSellOrder(t *testing.T) {
	lo := LimitOrder{
		OrderID: 42,
		Kind:    BuyOrder,
		Amount:  sdk.NewCoin("ETH", 60),
		Price:   sdk.NewCoin("BTC", 150),
	}

	ok, _, _ := lo.DoesFill(SellOrder, sdk.NewCoin("ETH", 50), sdk.NewCoin("BTC", 151))

	require.False(t, ok)

	ok, fillAmt, fillPrice := lo.DoesFill(SellOrder, sdk.NewCoin("ETH", 50), sdk.NewCoin("BTC", 140))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 50), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 150), fillPrice)

	ok, fillAmt, fillPrice = lo.DoesFill(SellOrder, sdk.NewCoin("ETH", 50), sdk.NewCoin("BTC", 150))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 50), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 150), fillPrice)

	ok, fillAmt, fillPrice = lo.DoesFill(SellOrder, sdk.NewCoin("ETH", 70), sdk.NewCoin("BTC", 100))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 60), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 150), fillPrice)
}

func TestDoesLimitOrderFillBuyOrder(t *testing.T) {
	lo := LimitOrder{
		OrderID: 42,
		Kind:    SellOrder,
		Amount:  sdk.NewCoin("ETH", 200),
		Price:   sdk.NewCoin("BTC", 11),
	}

	ok, _, _ := lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 180), sdk.NewCoin("BTC", 10))

	require.False(t, ok)

	ok, fillAmt, fillPrice := lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 180), sdk.NewCoin("BTC", 11))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 180), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 11), fillPrice)

	ok, fillAmt, fillPrice = lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 200), sdk.NewCoin("BTC", 13))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 200), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 11), fillPrice)

	ok, fillAmt, fillPrice = lo.DoesFill(BuyOrder, sdk.NewCoin("ETH", 220), sdk.NewCoin("BTC", 13))

	require.True(t, ok)
	require.Equal(t, sdk.NewCoin("ETH", 200), fillAmt)
	require.Equal(t, sdk.NewCoin("BTC", 11), fillPrice)
}
