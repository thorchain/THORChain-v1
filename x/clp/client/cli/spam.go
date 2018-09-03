package clp

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/thorchain/THORChain/app"

	"github.com/cosmos/cosmos-sdk/client/context"
	cosmosClientKeys "github.com/cosmos/cosmos-sdk/client/keys"
	lcdhelpers "github.com/cosmos/cosmos-sdk/client/lcd/helpers"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

// create new clp transaction
func RunSpamCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "run_spammers <start_index>",
		Short: "Run 10 previously created spammers starting from account index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("starting")
			startIndex, _ := strconv.Atoi(args[0])
			numSpammers := 80

			runtime.GOMAXPROCS(runtime.NumCPU())

			fmt.Println("Getting keybase")
			kb, err := cosmosClientKeys.GetKeyBase() //XXX
			if err != nil {
				fmt.Println(err)
				return err
			}
			fmt.Println("Got keybase")
			fmt.Printf("Start Index is: %v\n", startIndex)

			var wg sync.WaitGroup
			wg.Add(numSpammers)

			var spammers []Spammer
			for i := startIndex; i < startIndex+numSpammers; i++ {
				fmt.Printf("Spammer %v: Spawning...\n", i)
				var buffer bytes.Buffer
				buffer.WriteString("spammer")
				buffer.WriteString(strconv.FormatInt(int64(i), 10))
				accountName := buffer.String()

				newSpammer := SpawnSpammer(accountName, i, kb)
				spammers = append(spammers, newSpammer)
				fmt.Printf("Spammer %v: Spawned...\n", i)
			}

			for i := startIndex; i < startIndex+numSpammers; i++ {
				fmt.Printf("Spammer %v: Starting up...\n", i)
				go spammers[i].start()
				fmt.Printf("Spammer %v: Started...\n", i)
			}

			wg.Wait()

			return nil
		},
	}
}

//yezs
func CreateSpamCmd(cdc *wire.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create_spammers <start_index>",
		Short: "Create 20 spammers starting from account index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("starting")
			password := "1234567890"
			startIndex, err := strconv.Atoi(args[0])

			richAccountName := "local_validator"

			kb, err := cosmosClientKeys.GetKeyBase() //XXX
			if err != nil {
				fmt.Printf("Keybase get fail...\n")
			}

			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			richFrom, err := GetFromAddress(kb, richAccountName)
			if err != nil {
				fmt.Println(err)
				return err
			}

			ctx, err = SetupContext(ctx, richFrom)
			if err != nil {
				fmt.Println(err)
				return err
			}

			richSequence, err := ctx.NextSequence(richFrom)
			if err != nil {
				fmt.Println(err)
			}

			for i := startIndex; i < startIndex+20; i++ {
				fmt.Printf("Spammer %v: Creating Account...\n", i)
				var buffer bytes.Buffer
				buffer.WriteString("spammer")
				buffer.WriteString(strconv.FormatInt(int64(i), 10))
				accountName := buffer.String()
				spammerAddress := makeKey(kb, accountName, password)
				fmt.Printf("Spammer %v: Feeding...\n", i)

				feedMsg := client.BuildMsg(richFrom, spammerAddress, sdk.Coins{sdk.NewCoin("RUNE", 1000000)})
				ctx = ctx.WithSequence(richSequence)

				_, err2 := ProcessMsg(ctx, richAccountName, password, cdc, feedMsg)
				if err2 != nil {
					fmt.Println(err)
					return err
				}
				richSequence++
				fmt.Printf("Spammer %v: Feeding sent...\n", i)
			}
			return nil
		},
	}
}

//well
type Spammer struct {
	accountName     string
	password        string
	accountAddress  sdk.AccAddress
	codec           *wire.Codec
	index           int
	currentSequence int64
	ctx             context.CoreContext
	priv            tmcrypto.PrivKey
}

var (
	_1Rune = sdk.Coins{sdk.NewCoin("RUNE", 1)}
)

func (sp *Spammer) start() {
	for {
		// fmt.Printf("Spammer %v: Sending to self with sequence %v...\n", sp.index, sp.currentSequence)
		sp.ctx = sp.ctx.WithSequence(sp.currentSequence)

		msg2 := client.BuildMsg(sp.accountAddress, sp.accountAddress, _1Rune)

		_, err := PrivProcessMsg(sp.ctx, sp.priv, sp.codec, msg2)
		if err != nil {
			fmt.Println(err)
			return
		}
		sp.currentSequence = sp.currentSequence + 1
	}
}

//blah
func SpawnSpammer(localAccountName string, index int, kb keys.Keybase) Spammer {
	fmt.Printf("Spammer %v: Spawning...\n", index)

	cdc := app.MakeCodec()
	fmt.Printf("Spammer %v: Made codec...\n", index)

	localPassword := "1234567890"

	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
	fmt.Printf("Spammer %v: Made context...\n", index)

	from, err := GetFromAddress(kb, localAccountName)
	if err != nil {
		fmt.Println(err)
		return Spammer{}
	}

	ctx, err = SetupContext(ctx, from)
	if err != nil {
		fmt.Println(err)
		return Spammer{}
	}
	fmt.Printf("Spammer %v: Making sequence...\n", index)

	sequence, err3 := ctx.NextSequence(from)
	if err3 != nil {
		fmt.Printf("Spammer %v: Sequence Error...\n", index)
	}
	priv, err := kb.ExportPrivateKeyObject(localAccountName, localPassword)
	if err != nil {
		fmt.Println(err)
	}

	return Spammer{localAccountName, localPassword, from, cdc, index, sequence, ctx, priv}

}

//zzz
func SetupContext(ctx context.CoreContext, from sdk.AccAddress) (context.CoreContext, error) {
	// add gas to context
	ctx = ctx.WithGas(10000)

	// add chain-id to context
	ctx = ctx.WithChainID("test-chain-local")

	//add account number and sequence
	ctx, err := lcdhelpers.EnsureAccountNumber(ctx, 0, from)
	if err != nil {
		fmt.Println(err)
		return ctx, err
	}
	return ctx, nil
}

//asd
func PrivProcessMsg(ctx context.CoreContext, priv tmcrypto.PrivKey, cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	//sign
	txBytes, err := PrivSignAndBuild(priv, []sdk.Msg{msg}, cdc, ctx)
	if err != nil {
		return nil, err
	}

	// send
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return nil, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		return nil, err
	}
	return output, err
}

//ugh
func ProcessMsg(ctx context.CoreContext, accountName string, password string, cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	//sign
	txBytes, err := ctx.SignAndBuild(accountName, password, []sdk.Msg{msg}, cdc)
	if err != nil {
		return nil, err
	}

	// send
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return nil, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		return nil, err
	}
	return output, err
}

func makeKey(kb keys.Keybase, name string, password string) sdk.AccAddress {
	algo := keys.SigningAlgo("secp256k1")
	info, _, _ := kb.CreateMnemonic(name, keys.English, password, algo)
	account := sdk.AccAddress(info.GetPubKey().Address().Bytes())
	// _, _ := sdk.Bech32ifyAccPub(info.GetPubKey())
	return account
}

//ew
func GetFromAddress(kb keys.Keybase, localAccountName string) (sdk.AccAddress, error) {
	info, err := kb.Get(localAccountName)
	if err != nil {
		fmt.Println(err)
		return sdk.AccAddress{}, err
	}

	from := sdk.AccAddress(info.GetPubKey().Address())
	return from, nil
}

//giev private key to sign
func Sign(priv tmcrypto.PrivKey, msg []byte) (sig tmcrypto.Signature, pub tmcrypto.PubKey, err error) {
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub = priv.PubKey()
	return sig, pub, nil
}

// sign and build the transaction from the msg
func PrivSignAndBuild(priv tmcrypto.PrivKey, msgs []sdk.Msg, cdc *wire.Codec, ctx context.CoreContext) ([]byte, error) {
	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return nil, errors.Errorf("chain ID required but not specified")
	}
	accnum := ctx.AccountNumber
	sequence := ctx.Sequence
	memo := ctx.Memo

	fee := sdk.Coin{}
	if ctx.Fee != "" {
		parsedFee, err := sdk.ParseCoin(ctx.Fee)
		if err != nil {
			return nil, err
		}
		fee = parsedFee
	}

	signMsg := auth.StdSignMsg{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
		Msgs:          msgs,
		Memo:          memo,
		Fee:           auth.NewStdFee(ctx.Gas, fee), // TODO run simulate to estimate gas?
	}

	// sign and build
	bz := signMsg.Bytes()

	sig, pubkey, err := Sign(priv, bz)
	if err != nil {
		return nil, err
	}
	sigs := []auth.StdSignature{{
		PubKey:        pubkey,
		Signature:     sig,
		AccountNumber: accnum,
		Sequence:      sequence,
	}}

	// marshal bytes
	tx := auth.NewStdTx(signMsg.Msgs, signMsg.Fee, sigs, memo)

	return cdc.MarshalBinary(tx)
}
