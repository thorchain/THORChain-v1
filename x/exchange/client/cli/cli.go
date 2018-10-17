package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/thorchain/THORChain/x/exchange"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

const (
	flagKind        = "kind"
	flagAmount      = "amount"
	flagPrice       = "price"
	flagExpiresAt   = "expires-at"
	flagAmountDenom = "amount-denom"
	flagPriceDenom  = "price-denom"
)

// get cmd to create new limit order
func GetCmdLimitOrderCreate(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-limit-order",
		Short: "Create a limit order",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// parse inputs

			// get the from address from the name flag
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			kind, err := exchange.ParseKind(viper.GetString(flagKind))
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoin(viper.GetString(flagAmount))
			if err != nil {
				return err
			}

			price, err := sdk.ParseCoin(viper.GetString(flagPrice))
			if err != nil {
				return err
			}

			expiresAt, err := time.Parse(time.RFC3339, viper.GetString(flagExpiresAt))
			if err != nil {
				return err
			}

			// create the msg
			msg := exchange.NewMsgCreateLimitOrder(sender, kind, amount, price, expiresAt)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// get account name
			addressName := ctx.FromAddressName

			// build and sign the transaction, then broadcast to Tendermint
			err = ctx.EnsureSignBuildBroadcast(addressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String(flagKind, "", "kind of the order ('sell' or 'buy')")
	cmd.Flags().String(flagAmount, "", "amount to be sold or bought, e. g. '8ETH'")
	cmd.Flags().String(flagPrice, "", "price limit per unit of amount (maximum buy price or minimum sell price), e. g. '25RUNE'")
	cmd.Flags().String(flagExpiresAt, "", "expiration of the order in RFC3339, e. g. '2018-10-31T11:45:05.000Z'")

	return cmd
}

// get command to query orderbook
func GetCmdQueryOrderbook(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-orderbook",
		Short: "Get sell or buy orderbook for given amount and price denoms",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			kind, err := exchange.ParseKind(viper.GetString(flagKind))
			if err != nil {
				return err
			}

			amountDenom := viper.GetString(flagAmountDenom)
			priceDenom := viper.GetString(flagPriceDenom)

			res, err2 := ctx.QueryStore(exchange.MakeKeyOrderBook(kind, amountDenom, priceDenom), storeName)
			if err2 != nil {
				return err2
			}

			var orderbook exchange.OrderBook

			if len(res) > 0 {
				cdc.MustUnmarshalBinary(res, &orderbook)
			} else {
				orderbook = exchange.NewOrderBook(kind, amountDenom, priceDenom)
			}

			orderbook.RemoveExpiredLimitOrders()

			output, err2 := wire.MarshalJSONIndent(cdc, orderbook)
			if err2 != nil {
				return err2
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().String(flagKind, "", "kind of the order ('sell' or 'buy')")
	cmd.Flags().String(flagAmountDenom, "", "denom of the amount to be sold or bought, e. g. 'ETH'")
	cmd.Flags().String(flagPriceDenom, "", "denom of the price limit for the sell or buy, e. g. 'RUNE'")

	return cmd
}
