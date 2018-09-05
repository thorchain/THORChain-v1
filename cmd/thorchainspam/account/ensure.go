package account

import (
	"fmt"
	"runtime"

	"github.com/tendermint/go-amino"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thorchain/THORChain/cmd/thorchainspam/helpers"

	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// Returns the command to ensure k accounts exist
func GetAccountEnsure(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

		chainId := viper.GetString(FlagChainID)
		if chainId == "" {
			return fmt.Errorf("--chain-id is required")
		}

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
		from, err := ctx.GetFromAddress()
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

		err = ensureFromAccHasEnoughCoins(ctx, from, totalCoinsNeeded)
		if err != nil {
			return err
		}

		sendCoins(numAccsToCreate, spamPrefix, numExistingAccs,
			kb, spamPassword, signPassword, from,
			coins, ctx, cdc, chainId)
		return nil
	}
}

func ensureFromAccHasEnoughCoins(ctx context.CoreContext, from sdk.AccAddress, coins sdk.Coins) error {
	fromAcc, err := ctx.QueryStore(auth.AddressStoreKey(from), ctx.AccountStore)
	if err != nil {
		return err
	}

	// Check if account was found
	if fromAcc == nil {
		return errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", from)
	}

	account, err := ctx.Decoder(fromAcc)
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
	coins sdk.Coins, ctx context.CoreContext, cdc *amino.Codec, chainId string) error {

	//Set to use max CPUs
	runtime.GOMAXPROCS(runtime.NumCPU())

	ctx, err := helpers.SetupContext(ctx, from, chainId, 0)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fromSequence, err := ctx.NextSequence(from)
	if err != nil {
		fmt.Println(err)
	}

	// how many msgs to send in 1 tx
	// for each required account, build the required amount of keys and transfer the coins
	for i := 0; i < numAccsToCreate; i++ {
		accountName := fmt.Sprintf("%v-%v", spamPrefix, i+numExistingAccs)
		to, err := createSpamAccountKey(kb, accountName, spamPassword)
		if err != nil {
			return err
		}

		msg := client.BuildMsg(from, to, coins)
		ctx = ctx.WithSequence(fromSequence)

		_, err2 := helpers.ProcessMsg(ctx, ctx.FromAddressName, signPassword, cdc, msg)
		if err2 != nil {
			fmt.Println(err)
			return err
		}
		fromSequence++
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

// sign and build the transaction from the msg
func ensureSignBuild(ctx context.CoreContext, name string, passphrase string, msgs []sdk.Msg, cdc *wire.Codec) (tyBytes []byte, err error) {
	// ctx = ctx.WithFromAddressName(name)

	err = context.EnsureAccountExists(ctx, name)
	if err != nil {
		return nil, err
	}

	ctx, err = context.EnsureAccountNumber(ctx)
	if err != nil {
		return nil, err
	}
	// default to next sequence number if none provided
	ctx, err = context.EnsureSequence(ctx)
	if err != nil {
		return nil, err
	}

	var txBytes []byte

	txBytes, err = ctx.SignAndBuild(name, passphrase, msgs, cdc)
	if err != nil {
		return nil, fmt.Errorf("Error signing transaction: %v", err)
	}

	return txBytes, err
}

// sign and build the transaction from the msg
func ensureSignBuildBroadcast(ctx context.CoreContext, name string, passphrase string, msgs []sdk.Msg, cdc *wire.Codec) (err error) {
	txBytes, err := ensureSignBuild(ctx, name, passphrase, msgs, cdc)
	if err != nil {
		return err
	}

	if ctx.Async {
		res, err := ctx.BroadcastTxAsync(txBytes)
		if err != nil {
			return err
		}
		if ctx.JSON {
			type toJSON struct {
				TxHash string
			}
			valueToJSON := toJSON{res.Hash.String()}
			JSON, err := cdc.MarshalJSON(valueToJSON)
			if err != nil {
				return err
			}
			fmt.Println(string(JSON))
		} else {
			fmt.Println("Async tx sent. tx hash: ", res.Hash.String())
		}
		return nil
	}
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}
	if ctx.JSON {
		// Since JSON is intended for automated scripts, always include response in JSON mode
		type toJSON struct {
			Height   int64
			TxHash   string
			Response string
		}
		valueToJSON := toJSON{res.Height, res.Hash.String(), fmt.Sprintf("%+v", res.DeliverTx)}
		JSON, err := cdc.MarshalJSON(valueToJSON)
		if err != nil {
			return err
		}
		fmt.Println(string(JSON))
		return nil
	}
	if ctx.PrintResponse {
		fmt.Printf("Committed at block %d. Hash: %s Response:%+v \n", res.Height, res.Hash.String(), res.DeliverTx)
	} else {
		fmt.Printf("Committed at block %d. Hash: %s \n", res.Height, res.Hash.String())
	}
	return nil
}
