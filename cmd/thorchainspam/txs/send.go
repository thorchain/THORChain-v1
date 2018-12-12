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
	"github.com/thorchain/THORChain/x/exchange"
)

var (
	defaultBlockTime time.Duration = 1000
)

type spamCtx struct {
	chainID         string
	stats           *stats.Stats
	limitOrderPairs [][]string
}

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

		limitOrderPairs, err := parseLimitOrderPairs(viper.GetString(FlagLimitOrderPairs))
		if err != nil {
			return err
		}

		spamCtx := spamCtx{chainID, &stats, limitOrderPairs}

		// create context and all spammer objects
		spammers, err := createSpammers(&spamCtx, cdc, spamPrefix, spamPassword)
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

		go doEvery(1*time.Second, spamCtx.stats.Print)

		i := 0

		for {
			<-limiter
			nextSpammer := spammers[(i+1)%len(spammers)]
			go spammers[i].send(&spamCtx, &nextSpammer)
			if i == len(spammers)-1 {
				i = 0
			} else {
				i++
			}
		}
	}
}

func parseLimitOrderPairs(input string) (result [][]string, err error) {
	if input == "" {
		err = fmt.Errorf("--limit-order-pairs is required")
		return
	}

	pairs := strings.Split(input, ",")

	for _, pair := range pairs {
		denoms := strings.Split(pair, ":")

		if len(denoms) != 2 {
			err = fmt.Errorf("--limit-order-pairs contained invalid pair '%v', needs to be of format 'TKNA:TKNB'", pair)
			return
		}

		result = append(result, denoms)
	}

	return
}

func doEvery(d time.Duration, f func()) {
	for range time.Tick(d) {
		go f()
	}
}

func createSpammers(spamCtx *spamCtx, cdc *wire.Codec, spamPrefix string, spamPassword string) ([]Spammer, error) {
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
			go SpawnSpammer(spamCtx, cdc, spamPassword, j, kb, info, &spammers, &wg, semaphore)
		}
	}

	wg.Wait()

	return spammers, nil
}

//Spawn new spammer
func SpawnSpammer(spamCtx *spamCtx, cdc *wire.Codec, spamPassword string, index int, kb cryptokeys.Keybase,
	info cryptokeys.Info, spammers *[]Spammer, wg *sync.WaitGroup,
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
		Gas:     20000,
		ChainID: spamCtx.chainID,
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
		info.GetName(), spamPassword, address, cdc, index, account.GetSequence(), cliCtx, txCtx, priv,
		account.GetCoins(), 0, queryFree,
	}

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

func (sp *Spammer) send(spamCtx *spamCtx, nextSpammer *Spammer) {
	<-sp.queryFree

	log.Log.Debugf("Spammer %v: Will transaction with sequence %v...\n", sp.index, sp.nextSequence)

	randType := rand.Float32()
	var msg sdk.Msg
	var coinsUsed sdk.Coins
	var ok bool

	switch {
	case randType < 0.5:
		msg, coinsUsed, ok = sp.makeRandomTxMsg(nextSpammer)
	case randType < 0.99:
		msg, coinsUsed, ok = sp.makeRandomClpMsg()
	default:
		msg, coinsUsed, ok = sp.makeRandomLimitOrderMsg(spamCtx)
	}

	if !ok {
		sp.updateSequenceAndCoins()
		sp.queryFree <- true
		return
	}

	sp.txCtx = sp.txCtx.WithSequence(sp.nextSequence)

	_, err := helpers.PrivBuildSignAndBroadcastMsg(sp.cdc, sp.cliCtx, sp.txCtx, sp.priv, msg)

	sp.nextSequence = sp.nextSequence + 1
	sp.sequenceCheck = sp.sequenceCheck + 1

	if err != nil {
		log.Log.Warningf("Spammer %v: Received error trying to send: %v\n", sp.index, err)
		spamCtx.stats.AddError()
		sp.updateSequenceAndCoins()
		sp.queryFree <- true
		return
	}
	log.Log.Debugf("Spammer %v: Sending successful\n", sp.index)
	spamCtx.stats.AddSuccess()

	sp.currentCoins = sp.currentCoins.Minus(coinsUsed)
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

func (sp *Spammer) makeRandomClpMsg() (sdk.Msg, sdk.Coins, bool) {
	clpFrom := "RUNE"
	clpTo := "XMR"

	if rand.Float32() < 0.5 {
		clpFrom = "XMR"
		clpTo = "RUNE"
	}

	randomCoins := getRandomCoinsUpTo(sp.currentCoins, 100000)
	clpAmount := randomCoins.AmountOf(clpFrom)

	if !clpAmount.GT(sdk.NewInt(0)) {
		clpFrom = "RUNE"
		clpTo = "XMR"
		clpAmount = randomCoins.AmountOf(clpFrom)
	}

	if !clpAmount.GT(sdk.NewInt(0)) {
		log.Log.Debugf("Spammer %v: Will not trade on CLP, has %v %v and %v %v\n", sp.index, clpAmount, clpFrom, randomCoins.AmountOf(clpTo), clpTo)
		return nil, sdk.Coins{}, false
	}

	msg := clpTypes.NewMsgTrade(sp.accountAddress, clpFrom, clpTo, int(clpAmount.Int64()))

	log.Log.Debugf("Spammer %v: Will trade on CLP: %v %v -> %v\n", sp.index, clpAmount, clpFrom, clpTo)

	return msg, sdk.Coins{sdk.NewCoin(clpFrom, clpAmount)}, true
}

func (sp *Spammer) makeRandomTxMsg(nextSpammer *Spammer) (sdk.Msg, sdk.Coins, bool) {
	randomCoins := getRandomCoinsUpTo(sp.currentCoins, 100000)
	msg := client.BuildMsg(sp.accountAddress, nextSpammer.accountAddress, randomCoins)

	log.Log.Debugf("Spammer %v: Will send: %v to %v\n", sp.index, randomCoins.String(), nextSpammer.accountAddress.String())

	return msg, randomCoins, true
}

func (sp *Spammer) makeRandomLimitOrderMsg(spamCtx *spamCtx) (sdk.Msg, sdk.Coins, bool) {
	amtDenom, priceDenom := getRandomLimitOrderDenoms(spamCtx)

	randomCoins := getRandomCoinsUpTo(sp.currentCoins, 1000)

	amtIfSell := sdk.NewCoin(amtDenom, randomCoins.AmountOf(amtDenom))
	totalPriceIfBuy := sdk.NewCoin(priceDenom, randomCoins.AmountOf(priceDenom))

	buy := true

	if rand.Float32() < 0.5 {
		buy = false
	}

	// check if amt/totalPrice > 0
	possibleKinds := 2
	if !totalPriceIfBuy.Amount.GT(sdk.NewInt(0)) {
		buy = false
		possibleKinds--
	}
	if !amtIfSell.Amount.GT(sdk.NewInt(0)) {
		buy = true
		possibleKinds--
	}
	if possibleKinds == 0 {
		log.Log.Debugf("Spammer %v: Will not create limit order, amtIfSell %v and totalPriceIfBuy %v\n", sp.index,
			amtIfSell, totalPriceIfBuy)
		return nil, sdk.Coins{}, false
	}

	var kind exchange.OrderKind
	var amt sdk.Coin
	var price sdk.Coin
	var coinsUsed sdk.Coin

	if buy {
		kind = exchange.BuyOrder
		price = sdk.NewInt64Coin(priceDenom, rand.Int63n(totalPriceIfBuy.Amount.Int64())+1) // TODO use market prices
		amt = sdk.NewCoin(amtDenom, totalPriceIfBuy.Amount.Div(price.Amount))
		coinsUsed = totalPriceIfBuy
	} else {
		kind = exchange.SellOrder
		price = sdk.NewInt64Coin(priceDenom, rand.Int63n(amtIfSell.Amount.Int64())+1)
		amt = amtIfSell
		coinsUsed = amtIfSell
	}

	msg := exchange.NewMsgCreateLimitOrder(sp.accountAddress, kind, amt, price, time.Now().Add(24*time.Hour))

	log.Log.Debugf("Spammer %v: Will create limit order, buy? %v with amt %v and price %v\\n", sp.index, buy, amt,
		price)

	return msg, sdk.Coins{coinsUsed}, true
}

func getRandomLimitOrderDenoms(spamCtx *spamCtx) (string, string) {
	randomPair := spamCtx.limitOrderPairs[rand.Intn(len(spamCtx.limitOrderPairs))]
	return randomPair[0], randomPair[1]
}

func getTotalPrice(amt sdk.Coin, price sdk.Coin) sdk.Coin {
	return sdk.Coin{price.Denom, amt.Amount.Mul(price.Amount)}
}
