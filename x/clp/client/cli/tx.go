package clp

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/x/clp"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// create new clp transaction
func CreateTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create <ticker> <name> <reserve_ratio>",
		Short: "Create a token with CLP",
		Args:  cobra.ExactArgs(3),
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
			msg := clp.NewMsgCreate(from, ticker, name, reserveRatio)

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
			clp := new(clp.CLP)
			err2 := cdc.UnmarshalBinary(res, &clp)
			if err2 != nil {
				return err2
			}
			fmt.Printf("CLP details \nCreator: %s \nTicker: %v \nName: %v \nReserve Ratio: %v \n", clp.Creator, clp.Ticker, clp.Name, clp.ReserveRatio)
			return nil
		},
	}
}
