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

var (
	defaultBlockTime time.Duration = 8000
)

// Returns the command to send txs between accounts
func GetTxsSend(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		chainID := viper.GetString(FlagChainID)
		if chainID == "" {
			return fmt.Errorf("--chain-id is required")
		}

		rateLimit := viper.GetFloat64(FlagRateLimit)

		// parse spam prefix and password
		spamPrefix := viper.GetString(FlagSpamPrefix)
		spamPassword := viper.GetString(FlagSpamPassword)
		if spamPassword == "" {
			return fmt.Errorf("--spam-password is required")
		}

		stats := stats.NewStats()

		// create context and all spammer objects
		spammers, err := createSpammers(spamPrefix, spamPassword, &stats, chainID)
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

		// rate limiter to allow x events per second
		limiter := time.Tick(time.Duration(rateLimit) * time.Millisecond)

		go doEvery(1*time.Second, stats.Print)

		i := 0

		for {
			<-limiter
			nextSpammer := spammers[(i+1)%len(spammers)]
			go spammers[i].send(&nextSpammer, &stats)
			if i == len(spammers)-1 {
				i = 0
			} else {
				i++
			}
		}
	}
}

func doEvery(d time.Duration, f func()) {
	for _ = range time.Tick(d) {
		go f()
	}
}

func createSpammers(spamPrefix string, spamPassword string, stats *stats.Stats, chainID string) ([]Spammer, error) {
	kb, err := keys.GetKeyBase()
	if err != nil {
		return nil, err
	}

	infos, err := kb.List()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	semaphore := make(chan bool, 50)
	var spammers []Spammer
	var j = -1
	for _, info := range infos {
		accountName := info.GetName()
		if strings.HasPrefix(accountName, spamPrefix) {
			j++
			wg.Add(1)
			semaphore <- true
			go SpawnSpammer(accountName, spamPassword, j, kb, info, stats, chainID, &spammers, &wg, semaphore)
		}
	}

	wg.Wait()

	return spammers, nil
}

//Spawn new spammer
func SpawnSpammer(localAccountName string, spamPassword string, index int, kb cryptokeys.Keybase, spammerInfo cryptokeys.Info, stats *stats.Stats, chainId string, spammers *[]Spammer, wg *sync.WaitGroup, semaphore <-chan bool) {
	defer wg.Done()

	fmt.Printf("Spammer %v: Spawning...\n", index)

	cdc := app.MakeCodec()
	fmt.Printf("Spammer %v: Made codec...\n", index)

	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
	fmt.Printf("Spammer %v: Made context...\n", index)

	from, err := helpers.GetFromAddress(kb, localAccountName)
	if err != nil {
		fmt.Println(err)
		<-semaphore
		return
	}

	ctx, err = helpers.SetupContext(ctx, from, chainId, 0)
	if err != nil {
		fmt.Println(err)
		<-semaphore
		return
	}

	// get account balance from sender
	fromAcc, err := getAcc(ctx, spammerInfo)
	if err != nil {
		fmt.Printf("Iteration %v: Account not found, skipping\n", index)
		<-semaphore
		return
	}

	// calculate random share of coins to be sent
	randomCoins := getRandomCoinsUpTo(fromAcc.GetCoins(), 1000)

	fmt.Printf("Spammer %v: Finding sequence...\n", index)

	sequence, err3 := ctx.NextSequence(from)
	if err3 != nil {
		fmt.Printf("Spammer %v: Sequence Error...\n", index)
	}

	priv, err := kb.ExportPrivateKeyObject(localAccountName, spamPassword)
	if err != nil {
		panic(err)
	}

	queryFree := make(chan bool, 1)
	queryFree <- true

	newSpammer := Spammer{
		localAccountName, spamPassword, from, cdc, index, sequence, ctx, priv, randomCoins, 0, queryFree, "RUNE", "ETH"}

	*spammers = append(*spammers, newSpammer)
	fmt.Printf("Spammer %v: Spawned...\n", index)
	<-semaphore
}

//All the things needed for a single spammer thread
type Spammer struct {
	accountName     string
	password        string
	accountAddress  sdk.AccAddress
	cdc             *wire.Codec
	index           int
	currentSequence int64
	ctx             context.CoreContext
	priv            tmcrypto.PrivKey
	randomCoins     sdk.Coins
	sequenceCheck   int
	queryFree       chan bool
	clpFrom         string
	clpTo           string
}

func (sp *Spammer) send(nextSpammer *Spammer, stats *stats.Stats) {
	<-sp.queryFree

	// fmt.Printf("Spammer %v: Sending transaction with sequence %v...\n", sp.index, sp.currentSequence)
	sp.ctx = sp.ctx.WithSequence(sp.currentSequence)

	clpMsg := rand.Float32() < 0.5
	var msg sdk.Msg
	if clpMsg {
		msg = clpTypes.NewMsgTrade(sp.accountAddress, sp.clpFrom, sp.clpTo, 1)
	} else {
		msg = client.BuildMsg(sp.accountAddress, nextSpammer.accountAddress, sp.randomCoins)
	}

	_, err := helpers.PrivProcessMsg(sp.ctx, sp.priv, sp.cdc, msg)

	sp.currentSequence = sp.currentSequence + 1
	sp.sequenceCheck = sp.sequenceCheck + 1

	if err != nil {
		fmt.Println(err)
		stats.AddError()
		sp.updateContext()
		sp.queryFree <- true
		return
	}
	stats.AddSuccess()
	if sp.sequenceCheck >= 200 {
		sp.updateContext()
	}
	sp.queryFree <- true
}

func (sp *Spammer) flipCLPTickers() {
	tmp := sp.clpFrom
	sp.clpFrom = sp.clpTo
	sp.clpTo = tmp
}

func (sp *Spammer) updateContext() {
	fmt.Printf("Spammer %v: time to refresh sequence at %v, waiting for next block...\n", sp.index,
		time.Now().UTC().Format(time.RFC3339))
	time.Sleep(defaultBlockTime * time.Millisecond)
	fmt.Printf("Spammer %v: querying new sequence...\n", sp.index)

	nextSequence, err := sp.ctx.NextSequence(sp.accountAddress)
	if err != nil {
		fmt.Println(err)
	}
	sp.currentSequence = nextSequence
	fmt.Printf("Spammer %v: Sequence updated at %v to...%v\n", sp.index, time.Now().UTC().Format(time.RFC3339),
		sp.currentSequence)
	sp.sequenceCheck = 0
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
		if randAmount > 0 {
			res = append(res, sdk.Coin{
				Denom:  coin.Denom,
				Amount: sdk.NewInt(randAmount),
			})
		}
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
