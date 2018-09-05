package main

import (
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/app"
	"github.com/thorchain/THORChain/cmd/thorchainspam/account"
	"github.com/thorchain/THORChain/cmd/thorchainspam/txs"
	"github.com/thorchain/THORChain/version"

	"github.com/tendermint/tendermint/libs/cli"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	cdc := app.MakeCodec()
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "thorchainspam",
		Short: "Spam commands to create artificial load on THORChain",
	}

	// --- account commands ---

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
	accountEnsureCmd.Flags().String(account.FlagAmount, "", "Amount of coins to send to each account")
	accountEnsureCmd.Flags().String(account.FlagSpamPrefix, "spam", "Prefix for the name of spam account keys")
	accountEnsureCmd.Flags().String(account.FlagSpamPassword, "", "Password for spam account keys")
	accountEnsureCmd.Flags().String(account.FlagSignPassword, "", "Password to sign the transactions with")
	accountEnsureCmd.Flags().String(account.FlagChainID, "", "Chain ID of tendermint node")
	accountEnsureCmd.Flags().String(account.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")

	accountCmd.AddCommand(accountEnsureCmd)

	rootCmd.AddCommand(accountCmd)

	// --- txs commands ---

	txsCmd := &cobra.Command{
		Use:   "txs",
		Short: "Transaction subcommands for creating transactions between spam accounts",
	}

	txsSendCmd := &cobra.Command{
		Use:   "send",
		Short: "Sends random transactions between spam accounts",
		RunE:  txs.GetTxsSend(cdc),
	}

	txsSendCmd.Flags().Float64(txs.FlagRateLimit, 200, "Prefix for the name of spam account keys")
	txsSendCmd.Flags().String(txs.FlagSpamPrefix, "spam", "Prefix for the name of spam account keys")
	txsSendCmd.Flags().String(txs.FlagSpamPassword, "", "Password for spam account keys")
	txsSendCmd.Flags().Int(txs.FlagTxConcurrency, 200, "Number of concurrent txs")
	txsSendCmd.Flags().String(txs.FlagChainID, "", "Chain ID of tendermint node")
	txsSendCmd.Flags().String(txs.FlagNode, "tcp://localhost:26657", "<host>:<port> to tendermint rpc interface for this chain")

	txsCmd.AddCommand(txsSendCmd)

	rootCmd.AddCommand(txsCmd)

	// add version
	rootCmd.AddCommand(version.VersionCmd)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		panic(err)
	}
}
