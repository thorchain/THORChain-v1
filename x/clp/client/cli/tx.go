package clp

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/x/clp"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

// create new clp transaction
func CreateTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create <ticker> <name> <decimals> <reserve_ratio> <initial_supply> <initial_rune_amount>",
		Short: "Create a token with CLP",
		Args:  cobra.ExactArgs(6),
		RunE: func(_ *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			ticker := args[0]
			name := args[1]
			decimalsInt, _ := strconv.Atoi(args[2])
			reserveRatio, _ := strconv.Atoi(args[3])

			if decimalsInt < 0 || decimalsInt > 255 {
				return clp.ErrInvalidDecimals(clp.DefaultCodespace)
			}

			decimals := uint8(decimalsInt)

			initialSupply, _ := strconv.Atoi(args[4])
			initialBaseCoinAmount, _ := strconv.Atoi(args[5])
			msg := clpTypes.NewMsgCreate(from, ticker, name, decimals, reserveRatio, int64(initialSupply), int64(initialBaseCoinAmount))

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
}

// create new clp transaction
func TradeBaseTxCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade <from_ticker> <to_ticker> <from_amount>",
		Short: "Trade from one token to another token via CLP",
		Args:  cobra.ExactArgs(3),
		RunE: func(_ *cobra.Command, args []string) error {
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			// create the message
			fromTicker := args[0]
			toTicker := args[1]
			fromAmount, _ := strconv.Atoi(args[2])
			msg := clpTypes.NewMsgTrade(from, fromTicker, toTicker, fromAmount)

			// Build and sign the transaction, then broadcast to a Tendermint
			// node.
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}
}

// get clp data
func GetCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get <ticker>",
		Short: "Get clp for given token",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			ticker := args[0]

			res, err := cliCtx.QueryStore(clp.MakeCLPStoreKey(ticker), "clp")
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
			fmt.Printf("CLP details \nCreator: %s \nTicker: %v \nName: %v \nDecimals: %v \nReserve Ratio: %v \nInitial Supply: %v \nAccount Address: %v \n", clp.Creator, clp.Ticker, clp.Name, clp.Decimals, clp.ReserveRatio, clp.InitialSupply, clp.AccountAddress.String())
			return nil
		},
	}
}

// get clp data
func GetAllCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "get_all",
		Short: "Get all clps",
		RunE: func(_ *cobra.Command, _ []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			clpSubspace := []byte("clp:")

			res, err := cliCtx.QuerySubspace(clpSubspace, "clp")
			if err != nil {
				return err
			}
			if res == nil {
				fmt.Printf("No CLPs \n")
				return nil
			}

			var clps []clpTypes.CLP
			for i := 0; i < len(res); i++ {
				clp := new(clpTypes.CLP)
				err2 := cdc.UnmarshalBinary(res[i].Value, &clp)
				if err2 != nil {
					return err2
				}
				clps = append(clps, *clp)
			}
			fmt.Printf("CLP details \n\n")

			for i := 0; i < len(clps); i++ {
				fmt.Printf("Creator: %s \nTicker: %v \nName: %v \nDecimals: %v \nReserve Ratio: %v \nInitial Supply: %v \nAccount Address: %v \n\n", clps[i].Creator, clps[i].Ticker, clps[i].Name, clps[i].Decimals, clps[i].ReserveRatio, clps[i].InitialSupply, clps[i].AccountAddress.String())
			}

			return nil
		},
	}
}
