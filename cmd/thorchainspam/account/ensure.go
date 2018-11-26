package account

import (
	"fmt"
	"os"
	"runtime"

	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thorchain/THORChain/cmd/thorchainspam/helpers"
)

// Returns the command to ensure k accounts exist
func GetAccountEnsure(cdc *wire.Codec) func(cnd *cobra.Command, args []string) error {
	return func(_ *cobra.Command, _ []string) error {
		cliCtx := context.NewCLIContext().
			WithCodec(cdc).
			WithLogger(os.Stdout).
			WithAccountDecoder(authcmd.GetAccountDecoder(cdc))

		// parse spam prefix and password
		spamPrefix := viper.GetString(FlagSpamPrefix)
		spamPassword := viper.GetString(FlagSpamPassword)
		if spamPassword == "" {
			return fmt.Errorf("--spam-password is required")
		}
		signPassword := viper.GetString(FlagSignPassword)
		if signPassword == "" {
			return fmt.Errorf("--sign-password is required")
		}

		// get list of all accounts starting with spamPrefix
		kb, err := keys.GetKeyBase()
		if err != nil {
			return err
		}

		infos, err := kb.List()
		if err != nil {
			return err
		}

		numExistingAccs := helpers.CountSpamAccounts(infos, spamPrefix)
		k := viper.GetInt(FlagK)
		numAccsToCreate := k - numExistingAccs

		if numAccsToCreate <= 0 {
			fmt.Printf("Found %v spam accounts, will not create additional ones\n", numExistingAccs)
			return nil
		}

		fmt.Printf("Found %v spam accounts, will create %v additional ones\n", numExistingAccs, numAccsToCreate)

		// get the from address
		from, err := cliCtx.GetFromAddress()
		if err != nil {
			return err
		}

		// parse coins trying to be sent
		amount := viper.GetString(FlagAmount)
		coins, err := sdk.ParseCoins(amount)
		if err != nil {
			return err
		}

		// ensure account has enough coins
		totalCoinsNeeded := multiplyCoins(coins, numAccsToCreate)

		err = ensureFromAccHasEnoughCoins(cliCtx, from, totalCoinsNeeded)
		if err != nil {
			return err
		}

		err = sendCoins(numAccsToCreate, spamPrefix, numExistingAccs, kb, spamPassword, signPassword, from, coins,
			cliCtx, cdc)
		if err != nil {
			return err
		}

		fmt.Printf("Done creating %v accounts\n", numAccsToCreate)

		return nil
	}
}

func ensureFromAccHasEnoughCoins(cliCtx context.CLIContext, from sdk.AccAddress, coins sdk.Coins) error {
	// Check if account exists
	err := cliCtx.EnsureAccountExistsFromAddr(from)
	if err != nil {
		return errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", from)
	}

	account, err := cliCtx.GetAccount(from)
	if err != nil {
		return err
	}

	if !account.GetCoins().IsGTE(coins) {
		return errors.Errorf("Account %s doesn't have enough coins to pay for all txs", from)
	}

	return nil
}

func sendCoins(numAccsToCreate int, spamPrefix string, numExistingAccs int,
	kb cryptokeys.Keybase, spamPassword string, signPassword string, from sdk.AccAddress,
	coins sdk.Coins, cliCtx context.CLIContext, cdc *amino.Codec) error {

	//Set to use max CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	fromAccount, err := cliCtx.GetAccount(from)
	if err != nil {
		return err
	}

	fromAccountNumber := fromAccount.GetAccountNumber()
	fromAccountSequence := fromAccount.GetSequence()

	txCtx := authctx.NewTxContextFromCLI().
		WithAccountNumber(fromAccountNumber).
		WithCodec(cdc).
		WithGas(20000)

	// how many msgs to send in 1 tx
	// for each required account, build the required amount of keys and transfer the coins
	for i := 0; i < numAccsToCreate; i++ {
		accountName := fmt.Sprintf("%v-%v", spamPrefix, i+numExistingAccs)
		to, err := createSpamAccountKey(kb, accountName, spamPassword)
		if err != nil {
			return err
		}

		msg := client.BuildMsg(from, to, coins)

		txCtx = txCtx.WithSequence(fromAccountSequence)
		fromAccountSequence++

		_, err = helpers.BuildSignAndBroadcastMsg(cdc, cliCtx, txCtx, cliCtx.FromAddressName, signPassword, msg)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func createSpamAccountKey(kb cryptokeys.Keybase, name string, pass string) (sdk.AccAddress, error) {
	algo := cryptokeys.SigningAlgo("secp256k1")
	info, _, err := kb.CreateMnemonic(name, cryptokeys.English, pass, algo)
	if err != nil {
		return nil, err
	}
	ko, err := keys.Bech32KeyOutput(info)
	if err != nil {
		return nil, err
	}
	return ko.Address, nil
}

func multiplyCoins(coins sdk.Coins, multiplyBy int) sdk.Coins {
	res := make([]sdk.Coin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, sdk.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.MulRaw(int64(multiplyBy)),
		})
	}
	return res
}
