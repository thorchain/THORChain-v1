package main

import (
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/app"
	"github.com/thorchain/THORChain/version"

	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/lcd"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	govcmd "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	clpcmd "github.com/thorchain/THORChain/x/clp/client/cli"
	clpRest "github.com/thorchain/THORChain/x/clp/client/rest"
	exchangecmd "github.com/thorchain/THORChain/x/exchange/client/cli"
	exchangeRest "github.com/thorchain/THORChain/x/exchange/client/rest"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "thorchaincli",
		Short: "THORChain light-client",
	}
	serveCommandCallback = func(ctx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb cryptokeys.Keybase) {
		clpRest.RegisterRoutes(ctx, r, cdc, kb, app.AppBaseCoinTicker)
		exchangeRest.RegisterRoutes(ctx, r, cdc, kb, "exchange")
	}
)

func main() {
	cobra.EnableCommandSorting = false
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc commands
	rpc.AddCommands(rootCmd)

	//Add state commands
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint state querying subcommands",
	}
	tendermintCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(tendermintCmd, cdc)

	//Add IBC commands
	ibcCmd := &cobra.Command{
		Use:   "ibc",
		Short: "Inter-Blockchain Communication subcommands",
	}
	ibcCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCTransferCmd(cdc),
			ibccmd.IBCRelayCmd(cdc),
		)...)

	advancedCmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced subcommands",
	}

	advancedCmd.AddCommand(
		tendermintCmd,
		ibcCmd,
		lcd.ServeCommandWithCallback(cdc, serveCommandCallback),
	)
	rootCmd.AddCommand(
		advancedCmd,
		client.LineBreak,
	)

	//Add stake commands
	stakeCmd := &cobra.Command{
		Use:   "stake",
		Short: "Stake and validation subcommands",
	}
	stakeCmd.AddCommand(
		client.GetCommands(
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegation("stake", cdc),
			stakecmd.GetCmdQueryUnbondingDelegations("stake", cdc),
			stakecmd.GetCmdQueryRedelegation("stake", cdc),
			stakecmd.GetCmdQueryRedelegations("stake", cdc),
			slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
		)...)
	stakeCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond("stake", cdc),
			stakecmd.GetCmdRedelegate("stake", cdc),
			slashingcmd.GetCmdUnrevoke(cdc),
		)...)
	rootCmd.AddCommand(
		stakeCmd,
	)

	//Add governance commands
	govCmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance and voting subcommands",
	}
	govCmd.AddCommand(
		client.GetCommands(
			govcmd.GetCmdQueryProposal("gov", cdc),
			govcmd.GetCmdQueryVote("gov", cdc),
			govcmd.GetCmdQueryVotes("gov", cdc),
			govcmd.GetCmdQueryProposals("gov", cdc),
		)...)
	govCmd.AddCommand(
		client.PostCommands(
			govcmd.GetCmdSubmitProposal(cdc),
			govcmd.GetCmdDeposit(cdc),
			govcmd.GetCmdVote(cdc),
		)...)
	rootCmd.AddCommand(
		govCmd,
	)

	//Add CLP commands
	clpCmd := &cobra.Command{
		Use:   "clp",
		Short: "CLP subcommands",
	}
	clpCmd.AddCommand(
		client.PostCommands(
			clpcmd.CreateTxCmd(cdc),
			clpcmd.TradeBaseTxCmd(cdc),
		)...)
	clpCmd.AddCommand(
		client.GetCommands(
			clpcmd.GetCmd(cdc),
			clpcmd.GetAllCmd(cdc),
		)...)
	rootCmd.AddCommand(
		clpCmd,
	)

	// Add exchange commands
	exchangeCmd := &cobra.Command{
		Use:   "exchange",
		Short: "Exchange subcommands",
	}
	exchangeCmd.AddCommand(
		client.PostCommands(
			exchangecmd.GetCmdLimitOrderCreate(cdc),
		)...)
	exchangeCmd.AddCommand(
		client.GetCommands(
			exchangecmd.GetCmdQueryOrderbook("exchange", cdc),
		)...)
	rootCmd.AddCommand(
		exchangeCmd,
	)

	//Add auth and bank commands
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
		)...)

	// add proxy, version and key info
	rootCmd.AddCommand(
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "GA", app.DefaultCLIHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}
