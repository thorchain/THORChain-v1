package clp

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/x/clp"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// run the test transaction
func TestTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "test [string]",
		Short: "Test me?",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			msg := clp.NewMsgTest(from, args[0])

			// get account name
			name := ctx.FromAddressName

			// build and sign the transaction, then broadcast to Tendermint
			err = ctx.EnsureSignBuildBroadcast(name, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

// get test data
func GetTestCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get_test",
		Short: "Get test",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.QueryStore(clp.GetTestKey(), "clp")
			if err != nil {
				return err
			}
			fmt.Println("Test value is: %v", string(res))
			return nil
		},
	}
}
