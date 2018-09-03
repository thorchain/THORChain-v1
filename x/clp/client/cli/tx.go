package clp

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// create new clp transaction
func CreateTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create <ticker> <name> <reserve_ratio> <initial_supply> <initial_rune_amount>",
		Short: "Create a token with CLP",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			ticker := args[0]
			name := args[1]
			reserveRatio, err := strconv.Atoi(args[2])
			initialSupply, err := strconv.Atoi(args[3])
			initialBaseCoinAmount, err := strconv.Atoi(args[4])
			msg := clpTypes.NewMsgCreate(from, ticker, name, reserveRatio, int64(initialSupply), int64(initialBaseCoinAmount))

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
}

// create new clp transaction
func TradeBaseTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade <from_ticker> <to_ticker> <from_amount>",
		Short: "Trade from one token to another token via CLP",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			fromTicker := args[0]
			toTicker := args[1]
			fromAmount, err := strconv.Atoi(args[2])
			msg := clpTypes.NewMsgTrade(from, fromTicker, toTicker, fromAmount)

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
}

// get clp data
func GetCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get <ticker>",
		Short: "Get clp for given token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := context.NewCoreContextFromViper()
			ticker := args[0]

			res, err := ctx.QueryStore(clp.MakeCLPStoreKey(ticker), "clp")
			if err != nil {
				return err
			}
			if res == nil {
				fmt.Printf("No CLP for given ticker \n")
				return nil
			}
			cdc := wire.NewCodec()
			wire.RegisterCrypto(cdc)
			clp := new(clpTypes.CLP)
			err2 := cdc.UnmarshalBinary(res, &clp)
			if err2 != nil {
				return err2
			}
			fmt.Printf("CLP details \nCreator: %s \nTicker: %v \nName: %v \nReserve Ratio: %v \nInitial Supply: %v \nAccount Address: %v \n", clp.Creator, clp.Ticker, clp.Name, clp.ReserveRatio, clp.InitialSupply, clp.AccountAddress.String())
			return nil
		},
	}
}
