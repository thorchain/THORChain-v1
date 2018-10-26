package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/thorchain/THORChain/x/exchange"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
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
		RunE: func(_ *cobra.Command, _ []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// parse inputs

			// get the from address from the name flag
			sender, err := cliCtx.GetFromAddress()
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

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
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
		RunE: func(_ *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			kind, err := exchange.ParseKind(viper.GetString(flagKind))
			if err != nil {
				return err
			}

			amountDenom := viper.GetString(flagAmountDenom)
			priceDenom := viper.GetString(flagPriceDenom)

			res, err2 := cliCtx.QueryStore(exchange.MakeKeyOrderBook(kind, amountDenom, priceDenom), storeName)
			if err2 != nil {
				return err2
			}

			var orderbook exchange.OrderBook

			if len(res) > 0 {
				cdc.MustUnmarshalBinary(res, &orderbook)
			} else {
				orderbook = exchange.NewOrderBook(kind, amountDenom, priceDenom)
			}

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
