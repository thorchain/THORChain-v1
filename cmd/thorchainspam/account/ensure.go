package account

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thorchain/THORChain/cmd/thorchainspam/constants"
	"github.com/thorchain/THORChain/cmd/thorchainspam/helpers"

	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// Returns the command to ensure k accounts exist
func GetAccountEnsure(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

		// get list of all accounts starting with "spam"
		kb, err := keys.GetKeyBase()
		if err != nil {
			return err
		}

		infos, err := kb.List()
		if err != nil {
			return err
		}

		numExistingAccs := helpers.CountSpamAccounts(infos)
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

		fromAcc, err := ctx.QueryStore(auth.AddressStoreKey(from), ctx.AccountStore)
		if err != nil {
			return err
		}

		// Check if account was found
		if fromAcc == nil {
			return errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", from)
		}

		// parse coins trying to be sent
		amount := viper.GetString(FlagAmount)
		totalCoins, err := sdk.ParseCoins(amount)
		if err != nil {
			return err
		}

		coins := divideCoins(totalCoins, k)

		// ensure account has enough coins
		account, err := ctx.Decoder(fromAcc)
		if err != nil {
			return err
		}
		if !account.GetCoins().IsGTE(coins) {
			return errors.Errorf("Address %s doesn't have enough coins to pay for this transaction.", from)
		}

		// for each required account, build the required amount of keys and transfer the coins
		for i := 0; i < numAccsToCreate; i++ {
			accountName := fmt.Sprintf("%v-%v", constants.SpamAccountPrefix, i)
			to, err := createSpamAccountKey(kb, accountName, constants.SpamAccountPassword)
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := client.BuildMsg(from, to, coins)

			err = ensureSignBuildBroadcast(ctx, ctx.FromAddressName, constants.SpamAccountPassword, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
		}

		return nil
	}
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

func divideCoins(coins sdk.Coins, divideBy int) sdk.Coins {
	res := make([]sdk.Coin, 0, len(coins))
	for _, coin := range coins {
		res = append(res, sdk.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.DivRaw(int64(divideBy)),
		})
	}
	return res
}

// sign and build the transaction from the msg
func ensureSignBuild(ctx context.CoreContext, name string, passphrase string, msgs []sdk.Msg, cdc *wire.Codec) (tyBytes []byte, err error) {
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
