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
	"github.com/thorchain/THORChain/cmd/thorchainspam/log"
	"github.com/thorchain/THORChain/cmd/thorchainspam/stats"

	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"
)

var (
	defaultBlockTime time.Duration = 1000
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

		log.Log.Infof("Found %v spam accounts\n", len(spammers))

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
	for range time.Tick(d) {
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
func SpawnSpammer(localAccountName string, spamPassword string, index int, kb cryptokeys.Keybase, spammerInfo cryptokeys.Info, stats *stats.Stats, chainID string, spammers *[]Spammer, wg *sync.WaitGroup, semaphore <-chan bool) {
	defer wg.Done()

	log.Log.Debugf("Spammer %v: Spawning...\n", index)

	cdc := app.MakeCodec()
	log.Log.Debugf("Spammer %v: Made codec...\n", index)

	ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
	log.Log.Debugf("Spammer %v: Made context...\n", index)

	from, err := helpers.GetFromAddress(kb, localAccountName)
	if err != nil {
		log.Log.Errorf(err.Error())
		<-semaphore
		return
	}

	ctx, err = helpers.SetupContext(ctx, from, chainID, 0)
	if err != nil {
		log.Log.Errorf(err.Error())
		<-semaphore
		return
	}

	// get account balance from sender
	fromAcc, err := getAcc(ctx, spammerInfo)
	if err != nil {
		log.Log.Errorf("Iteration %v: Account not found, skipping\n", index)
		<-semaphore
		return
	}

	log.Log.Debugf("Spammer %v: Finding sequence...\n", index)

	sequence, err3 := ctx.NextSequence(from)
	if err3 != nil {
		log.Log.Errorf("Spammer %v: Sequence Error...\n", index)
	}

	priv, err := kb.ExportPrivateKeyObject(localAccountName, spamPassword)
	if err != nil {
		panic(err)
	}

	queryFree := make(chan bool, 1)
	queryFree <- true

	newSpammer := Spammer{
		localAccountName, spamPassword, from, cdc, index, sequence, ctx, priv, fromAcc.GetCoins(), 0, queryFree}

	*spammers = append(*spammers, newSpammer)
	log.Log.Infof("Spammer %v: Spawned...\n", index)
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
	currentCoins    sdk.Coins
	sequenceCheck   int
	queryFree       chan bool
}

func (sp *Spammer) send(nextSpammer *Spammer, stats *stats.Stats) {
	<-sp.queryFree

	log.Log.Debugf("Spammer %v: Will transaction with sequence %v...\n", sp.index, sp.currentSequence)

	sp.ctx = sp.ctx.WithSequence(sp.currentSequence)

	// calculate random share of coins to be sent
	randomCoins := getRandomCoinsUpTo(sp.currentCoins, 100000)
	clpFrom := "RUNE"
	clpTo := "XMR"
	var clpAmount sdk.Int

	clpMsg := rand.Float32() < 0.5

	var msg sdk.Msg
	if clpMsg {
		if rand.Float32() < 0.5 {
			clpFrom = "XMR"
			clpTo = "RUNE"
		}
		clpAmount = randomCoins.AmountOf(clpFrom)
		if !clpAmount.GT(sdk.NewInt(0)) {
			clpFrom = "RUNE"
			clpTo = "XMR"
			clpAmount = randomCoins.AmountOf(clpFrom)
		}
		if clpAmount.GT(sdk.NewInt(0)) {
			msg = clpTypes.NewMsgTrade(sp.accountAddress, clpFrom, clpTo, int(clpAmount.Int64()))

			log.Log.Debugf("Spammer %v: Will trade on CLP: %v %v -> %v\n", sp.index, clpAmount, clpFrom, clpTo)
		} else {
			log.Log.Debugf("Spammer %v: Will not trade on CLP, has %v %v and %v %v\n", sp.index, clpAmount, clpFrom, randomCoins.AmountOf(clpTo), clpTo)
			sp.updateSequenceAndCoins()
			sp.queryFree <- true
			return
		}
	} else {
		msg = client.BuildMsg(sp.accountAddress, nextSpammer.accountAddress, randomCoins)

		log.Log.Debugf("Spammer %v: Will send: %v to %v\n", sp.index, randomCoins.String(), nextSpammer.accountAddress.String())
	}

	_, err := helpers.PrivProcessMsg(sp.ctx, sp.priv, sp.cdc, msg)

	sp.currentSequence = sp.currentSequence + 1
	sp.sequenceCheck = sp.sequenceCheck + 1

	if err != nil {
		log.Log.Warningf("Spammer %v: Received error trying to send: %v\n", sp.index, err)
		stats.AddError()
		sp.updateSequenceAndCoins()
		sp.queryFree <- true
		return
	}
	log.Log.Debugf("Spammer %v: Sending successful\n", sp.index)
	stats.AddSuccess()

	if clpMsg {
		sp.currentCoins = sp.currentCoins.Minus([]sdk.Coin{sdk.NewCoin(clpFrom, clpAmount.Int64())})
	} else {
		sp.currentCoins = sp.currentCoins.Minus(randomCoins)
	}

	if sp.sequenceCheck >= 200 {
		sp.updateSequenceAndCoins()
	}
	sp.queryFree <- true
}

func (sp *Spammer) updateSequenceAndCoins() {
	log.Log.Debugf("Spammer %v: Time to refresh sequence and coins, waiting for next block...\n", sp.index)
	time.Sleep(defaultBlockTime * time.Millisecond)
	log.Log.Debugf("Spammer %v: Querying new sequence...\n", sp.index)

	nextSequence, err := sp.ctx.NextSequence(sp.accountAddress)
	if err != nil {
		log.Log.Errorf("Spammer %v: Error updating sequence: %v\n", sp.index, err)
	}
	sp.currentSequence = nextSequence
	log.Log.Debugf("Spammer %v: Sequence updated to %v\n", sp.index, sp.currentSequence)
	sp.sequenceCheck = 0

	log.Log.Debugf("Spammer %v: Querying coins...\n", sp.index)
	fromAcc, err := getAccFromAddr(sp.ctx, sp.accountAddress)
	if err != nil {
		log.Log.Errorf("Spammer %v: Account not found, skipping\n", sp.index)
		return
	}

	sp.currentCoins = fromAcc.GetCoins()
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

	return getAccFromAddr(ctx, accAddr)
}

func getAccFromAddr(ctx context.CoreContext, accAddr []byte) (auth.Account, error) {
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
