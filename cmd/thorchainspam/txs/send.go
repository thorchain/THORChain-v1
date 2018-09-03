package txs

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/thorchain/THORChain/app"
	"github.com/thorchain/THORChain/cmd/thorchainspam/helpers"
	"github.com/thorchain/THORChain/cmd/thorchainspam/stats"

	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"
)

// Returns the command to ensure k accounts exist
func GetTxsSend(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		// parse spam prefix and password
		spamPrefix := viper.GetString(FlagSpamPrefix)
		spamPassword := viper.GetString(FlagSpamPassword)
		if spamPassword == "" {
			return fmt.Errorf("--spam-password is required")
		}

		stats := stats.NewStats()

		// create context and all spammer objects
		spammers, err := createSpammers(spamPrefix, spamPassword, &stats)
		if err != nil {
			return err
		}

		// ensure at least 1 spammer is present
		if len(spammers) == 0 {
			return fmt.Errorf("no spam accounts found, please create them with `thorchainspam account ensure`")
		}

		fmt.Printf("Found %v spam accounts\n", len(spammers))

		//Use all cores
		runtime.GOMAXPROCS(runtime.NumCPU())

		var wg sync.WaitGroup

		for i := 0; i < len(spammers); i++ {
			wg.Add(1)

			fmt.Printf("Spammer %v: Starting up...\n", i)
			nextSpammer := spammers[(i+1)%len(spammers)]
			go spammers[i].start(&nextSpammer, &stats)
			fmt.Printf("Spammer %v: Started...\n", i)

		}

		go printStats(&stats)
		wg.Add(1)

		wg.Wait()

		fmt.Printf("Done.")
		stats.Print()

		return nil
	}
}

func printStats(stats *stats.Stats) {
	time.Sleep(5 * time.Millisecond)
	stats.Print()
}

func createSpammers(spamPrefix string, spamPassword string, stats *stats.Stats) ([]Spammer, error) {
	kb, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	infos, err := kb.List()
	if err != nil {
		return nil, err
	}

	var spammers []Spammer
	for i, info := range infos {
		accountName := info.GetName()
		if strings.HasPrefix(accountName, spamPrefix) {
			newSpammer := SpawnSpammer(accountName, spamPassword, i, kb, info, stats)
			spammers = append(spammers, newSpammer)
			fmt.Printf("Spammer %v: Spawned...\n", i)
		}
	}

	return spammers, nil

}

//Spawn new spammer
func SpawnSpammer(localAccountName string, spamPassword string, index int, kb cryptokeys.Keybase, spammerInfo cryptokeys.Info, stats *stats.Stats) Spammer {
	fmt.Printf("Spammer %v: Spawning...\n", index)

	cdc := app.MakeCodec()
	fmt.Printf("Spammer %v: Made codec...\n", index)

	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
	fmt.Printf("Spammer %v: Made context...\n", index)

	from, err := helpers.GetFromAddress(kb, localAccountName)
	if err != nil {
		fmt.Println(err)
		return Spammer{}
	}

	ctx, err = helpers.SetupContext(ctx, from)
	if err != nil {
		fmt.Println(err)
		return Spammer{}
	}

	// get account balance from sender
	fromAcc, err := getAcc(ctx, spammerInfo)
	if err != nil {
		fmt.Printf("Iteration %v: Account not found, skipping\n", index)
		stats.AddAccountNotFound()
		return Spammer{}
	}

	// calculate random share of coins to be sent
	randomCoins := getRandomCoinsUpTo(fromAcc.GetCoins(), 1000)

	if !randomCoins.IsPositive() {
		fmt.Printf("Iteration %v: No coins to send, skipping\n", index)
		stats.AddNoCoinsToSend()
		return Spammer{}
	}

	fmt.Printf("Spammer %v: Making sequence...\n", index)

	sequence, err3 := ctx.NextSequence(from)
	if err3 != nil {
		fmt.Printf("Spammer %v: Sequence Error...\n", index)
	}
	priv, err := kb.ExportPrivateKeyObject(localAccountName, spamPassword)
	if err != nil {
		fmt.Println(err)
	}

	return Spammer{localAccountName, spamPassword, from, cdc, index, sequence, ctx, priv, randomCoins}

}

//All the things needed for a single spammer thread
type Spammer struct {
	accountName     string
	password        string
	accountAddress  sdk.AccAddress
	codec           *wire.Codec
	index           int
	currentSequence int64
	ctx             context.CoreContext
	priv            tmcrypto.PrivKey
	randomCoins     sdk.Coins
}

func (sp *Spammer) start(nextSpammer *Spammer, stats *stats.Stats) {
	for {
		// fmt.Printf("Spammer %v: Sending to self with sequence %v...\n", sp.index, sp.currentSequence)
		sp.ctx = sp.ctx.WithSequence(sp.currentSequence)

		clpMsg := rand.Float32() < 0.5
		var msg sdk.Msg
		if clpMsg {
			msg = clpTypes.NewMsgTrade(sp.accountAddress, "RUNE", "ETH", 1)
		} else {
			msg = client.BuildMsg(sp.accountAddress, nextSpammer.accountAddress, sp.randomCoins)

		}

		_, err := helpers.PrivProcessMsg(sp.ctx, sp.priv, sp.codec, msg)
		if err != nil {
			fmt.Println(err)
			return
		}
		sp.currentSequence = sp.currentSequence + 1
	}
}

func getRandomCoinsUpTo(coins sdk.Coins, divideBy int64) sdk.Coins {
	res := make([]sdk.Coin, 0, len(coins))

	for _, coin := range coins {
		amount := coin.Amount.Int64()
		var randAmount int64
		if amount/divideBy > 0 {
			randAmount = rand.Int63n(amount / divideBy)
		} else {
			randAmount = 0
		}
		if randAmount == 0 && amount >= 1 {
			randAmount = 1
		}

		res = append(res, sdk.Coin{
			Denom:  coin.Denom,
			Amount: sdk.NewInt(randAmount),
		})
	}

	return res
}

func getAddr(info cryptokeys.Info) []byte {
	return sdk.AccAddress(info.GetPubKey().Address())
}

func getAcc(ctx context.CoreContext, info cryptokeys.Info) (auth.Account, error) {
	accAddr := getAddr(info)

	accBytes, err := ctx.QueryStore(auth.AddressStoreKey(accAddr), ctx.AccountStore)
	if err != nil {
		return nil, err
	}

	// Check if account was found
	if accBytes == nil {
		return nil, errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", accAddr)
	}

	// Decode account
	acc, err := ctx.Decoder(accBytes)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
