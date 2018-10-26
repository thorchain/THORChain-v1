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
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/thorchain/THORChain/cmd/thorchainspam/helpers"
	"github.com/thorchain/THORChain/cmd/thorchainspam/log"
	"github.com/thorchain/THORChain/cmd/thorchainspam/stats"
	clpTypes "github.com/thorchain/THORChain/x/clp/types"
)

var (
	defaultBlockTime time.Duration = 1000
)

// Returns the command to send txs between accounts
func GetTxsSend(cdc *wire.Codec) func(cmd *cobra.Command, args []string) error {
	return func(_ *cobra.Command, _ []string) error {
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
		spammers, err := createSpammers(cdc, spamPrefix, spamPassword, &stats, chainID)
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

func createSpammers(cdc *wire.Codec, spamPrefix string, spamPassword string, stats *stats.Stats, chainID string,
) ([]Spammer, error) {
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
			go SpawnSpammer(cdc, spamPassword, j, kb, info, stats, chainID, &spammers, &wg, semaphore)
		}
	}

	wg.Wait()

	return spammers, nil
}

//Spawn new spammer
func SpawnSpammer(cdc *wire.Codec, spamPassword string, index int, kb cryptokeys.Keybase,
	info cryptokeys.Info, stats *stats.Stats, chainID string, spammers *[]Spammer, wg *sync.WaitGroup,
	semaphore <-chan bool) {
	defer wg.Done()

	log.Log.Debugf("Spammer %v: Spawning...\n", index)

	log.Log.Debugf("Spammer %v: Making contexts...\n", index)

	cliCtx := context.NewCLIContext().
		WithCodec(cdc).
		WithAccountDecoder(authcmd.GetAccountDecoder(cdc)).
		WithFromAddressName(info.GetName())

	txCtx := authctx.TxContext{
		Codec:   cdc,
		Gas:     10000,
		ChainID: chainID,
	}

	log.Log.Debugf("Spammer %v: Finding account...\n", index)

	address := sdk.AccAddress(info.GetPubKey().Address())
	account, err3 := cliCtx.GetAccount(address)
	if err3 != nil {
		log.Log.Errorf("Spammer %v: Account not found, skipping\n", index)
		<-semaphore
		return
	}

	txCtx = txCtx.WithAccountNumber(account.GetAccountNumber())
	txCtx = txCtx.WithSequence(account.GetSequence())

	// get private key
	priv, err := kb.ExportPrivateKeyObject(info.GetName(), spamPassword)
	if err != nil {
		panic(err)
	}

	queryFree := make(chan bool, 1)
	queryFree <- true

	newSpammer := Spammer{
		info.GetName(), spamPassword, address, cdc, index, account.GetSequence(), cliCtx, txCtx, priv, account.GetCoins(), 0, queryFree}

	*spammers = append(*spammers, newSpammer)
	log.Log.Infof("Spammer %v: Spawned...\n", index)
	<-semaphore
}

//All the things needed for a single spammer thread
type Spammer struct {
	accountName    string
	password       string
	accountAddress sdk.AccAddress
	cdc            *wire.Codec
	index          int
	nextSequence   int64
	cliCtx         context.CLIContext
	txCtx          authctx.TxContext
	priv           tmcrypto.PrivKey
	currentCoins   sdk.Coins
	sequenceCheck  int
	queryFree      chan bool
}

func (sp *Spammer) send(nextSpammer *Spammer, stats *stats.Stats) {
	<-sp.queryFree

	log.Log.Debugf("Spammer %v: Will transaction with sequence %v...\n", sp.index, sp.nextSequence)

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

	sp.txCtx = sp.txCtx.WithSequence(sp.nextSequence)

	_, err := helpers.PrivBuildSignAndBroadcastMsg(sp.cdc, sp.cliCtx, sp.txCtx, sp.priv, msg)

	sp.nextSequence = sp.nextSequence + 1
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
		sp.currentCoins = sp.currentCoins.Minus([]sdk.Coin{sdk.NewInt64Coin(clpFrom, clpAmount.Int64())})
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

	log.Log.Debugf("Spammer %v: Querying account for new sequence and coins...\n", sp.index)
	fromAcc, err := sp.cliCtx.GetAccount(sp.accountAddress)
	if err != nil {
		log.Log.Errorf("Spammer %v: Account not found, skipping\n", sp.index)
		return
	}

	sequence, err := sp.cliCtx.GetAccountSequence(sp.accountAddress)
	if err != nil {
		log.Log.Errorf("Spammer %v: Error getting sequence: %v\n", sp.index, err)
	}
	sp.nextSequence = sequence
	log.Log.Debugf("Spammer %v: Sequence updated to %v\n", sp.index, sp.nextSequence)
	sp.sequenceCheck = 0

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
