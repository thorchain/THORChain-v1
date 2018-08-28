package main

import (
	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/app"
	"github.com/thorchain/THORChain/cmd/thorchainspam/account"

	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	cdc := app.MakeCodec()
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "thorchainspam",
		Short: "Spam commands to create artificial load on THORChain",
	}

	accountCmd := &cobra.Command{
		Use:   "account",
		Short: "Account subcommands for ensuring accounts exist for further testing",
	}

	accountEnsureCmd := &cobra.Command{
		Use:   "ensure",
		Short: "Ensures k accounts exist by distributing the given amount of tokens over them",
		RunE:  account.GetAccountEnsure(cdc),
	}

	accountEnsureCmd.Flags().String(account.FlagFrom, "", "Name of private key with which to sign")
	accountEnsureCmd.Flags().Int(account.FlagK, 10, "Number of accounts to create")
	accountEnsureCmd.Flags().String(account.FlagAmount, "", "Maximum total amount of coins to distribute over the k accounts")
	accountEnsureCmd.Flags().String(account.FlagChainID, "", "Chain ID of tendermint node")
	accountEnsureCmd.Flags().String(account.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")

	accountCmd.AddCommand(accountEnsureCmd)

	rootCmd.AddCommand(accountCmd)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
