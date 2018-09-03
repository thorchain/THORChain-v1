package clp

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/thorchain/THORChain/x/clp/types"
)

// Keeper - handlers sets/gets of custom variables for your module
type Keeper struct {
	storeKey       sdk.StoreKey // The (unexposed) key used to access the store from the Context.
	baseCoinTicker string       // The base coin ticker for all clps.

	bankKeeper bank.Keeper

	codespace sdk.CodespaceType

	cdc *wire.Codec
}

// NewKeeper - Returns the Keeper
func NewKeeper(key sdk.StoreKey, baseCoinTicker string, bankKeeper bank.Keeper, codespace sdk.CodespaceType) Keeper {
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)
	return Keeper{key, baseCoinTicker, bankKeeper, codespace, cdc}
}

// InitGenesis - store the genesis trend
func InitGenesis(ctx sdk.Context, k Keeper, data types.Genesis) error {
	return nil
}

// WriteGenesis - output the genesis trend
func WriteGenesis(ctx sdk.Context, k Keeper) types.Genesis {
	return types.Genesis{}
}

// GetCLP - returns the clp
func (k Keeper) GetCLP(ctx sdk.Context, ticker string) *types.CLP {
	store := ctx.KVStore(k.storeKey)
	valueBytes := store.Get(MakeCLPStoreKey(ticker))

	clp := new(types.CLP)
	k.cdc.UnmarshalBinary(valueBytes, &clp)

	return clp
}

func (k Keeper) ensureNonexistentCLP(ctx sdk.Context, ticker string) sdk.Error {
	clp := k.GetCLP(ctx, ticker)
	if clp.Ticker != "" {
		return ErrCLPExists(DefaultCodespace).TraceSDK("")
	}
	return nil
}

// Create CLP.
func (k Keeper) create(ctx sdk.Context, sender sdk.AccAddress, ticker string, name string, reserveRatio int, initialSupply int64, initialBaseCoinAmount int64) sdk.Error {
	if initialSupply <= 0 {
		return ErrInvalidInitialSupply(DefaultCodespace).TraceSDK("")
	}
	if initialBaseCoinAmount <= 0 {
		return ErrInvalidInitialBaseCoins(DefaultCodespace).TraceSDK("")
	}
	if ticker == k.baseCoinTicker {
		return ErrInvalidTickerName(DefaultCodespace).TraceSDK("")
	}
	err := k.ensureNonexistentCLP(ctx, ticker)
	if err != nil {
		return err
	}
	if reserveRatio <= 0 || reserveRatio > 100 {
		return ErrInvalidReserveRatio(DefaultCodespace).TraceSDK("")
	}
	initialBaseCoins := sdk.NewCoin(k.baseCoinTicker, initialBaseCoinAmount)
	//Debit initial coins from sender
	_, _, err2 := k.bankKeeper.SubtractCoins(ctx, sender, sdk.Coins{initialBaseCoins})
	if err2 != nil {
		return err2
	}
	clpAddress := types.NewCLPAddress(ticker)
	clp := types.NewCLP(sender, ticker, name, reserveRatio, initialSupply, clpAddress)
	_, _, err3 := k.bankKeeper.AddCoins(ctx, clpAddress, sdk.Coins{initialBaseCoins})
	if err3 != nil {
		return err3
	}
	initialCLPCoins := sdk.Coins{sdk.NewCoin(ticker, initialSupply)}
	k.bankKeeper.AddCoins(ctx, clpAddress, initialCLPCoins)
	k.SetCLP(ctx, clp)
	return nil
}

//Run Formula for CLP command
func RunCLPFormula(a float64, b float64, c float64, d float64) int64 {
	//y = a * ((1 + (b/c))^d - 1)
	y := a * (math.Pow(float64(1+(b/c)), d) - 1)
	return int64(math.Round(y))
}

//Run Formula for trading CLP coins
func CalculateCoinsEmitted(clp *types.CLP, clpCoins sdk.Coins, coinsPaid int64, baseCoinTicker string, buy bool) int64 {
	//baseCoinsEmitted = baseCoinBalance * ((1 + (clpCoinsPaid/clpCoinSupply))^(1/reserveRatio) - 1)
	//clpCoinsEmitted = clpCoinSupply * ((1 + (baseCoinsPaid/baseCoinBalance))^reserveRatio - 1)
	baseCoinBalance := float64(clpCoins.AmountOf(baseCoinTicker).Int64())
	clpCoinSupply := float64(clp.CurrentSupply)
	reserveRatio := float64(clp.ReserveRatio)
	floatCoinsPaid := float64(coinsPaid)
	if buy {
		return RunCLPFormula(clpCoinSupply, floatCoinsPaid, baseCoinBalance, reserveRatio/float64(100))

	}
	return RunCLPFormula(baseCoinBalance, floatCoinsPaid, clpCoinSupply, float64(1)/(float64(reserveRatio)/float64(100)))
}

//Process a single CLP trade
func ProcessCLPTrade(ctx sdk.Context, sender sdk.AccAddress, clpTicker string, fromAmount int64, k Keeper, buy bool) (int64, sdk.Error) {
	clp := k.GetCLP(ctx, clpTicker)
	clpCoins := k.bankKeeper.GetCoins(ctx, clp.AccountAddress)
	var fromTicker, toTicker string
	if buy {
		fromTicker = k.baseCoinTicker
		toTicker = clp.Ticker
	} else {
		fromTicker = clp.Ticker
		toTicker = k.baseCoinTicker
	}

	//Check clp exists and coins ok
	if clp.Ticker == "" {
		return 0, ErrCLPNotExists(DefaultCodespace).TraceSDK("")
	}
	clpBaseCoinBalance := clpCoins.AmountOf(k.baseCoinTicker).Int64()
	clpClpCoinBalance := clpCoins.AmountOf(clp.Ticker).Int64()
	if (buy && clpClpCoinBalance <= 0) || (!buy && clpBaseCoinBalance <= 0) {
		return 0, ErrCLPEmpty(DefaultCodespace).TraceSDK("")
	}

	emittedCoinsAmount := CalculateCoinsEmitted(clp, clpCoins, fromAmount, k.baseCoinTicker, buy)

	spentFromCoins := sdk.Coins{sdk.NewCoin(fromTicker, fromAmount)}
	emittedCoins := sdk.Coins{sdk.NewCoin(toTicker, emittedCoinsAmount)}

	k.bankKeeper.SendCoins(ctx, sender, clp.AccountAddress, spentFromCoins)
	k.bankKeeper.SendCoins(ctx, clp.AccountAddress, sender, emittedCoins)
	return emittedCoinsAmount, nil
}

// Trade with CLP.
func (k Keeper) trade(ctx sdk.Context, sender sdk.AccAddress, fromTicker string, toTicker string, fromAmount int64) (int64, int64, sdk.Error) {
	//Check different tickers
	if fromTicker == toTicker {
		return 0, 0, ErrSameCoin(DefaultCodespace).TraceSDK("")
	}

	//Check sender coins ok
	currentSenderCoins := k.bankKeeper.GetCoins(ctx, sender)
	currentSenderFromCoinAmount := currentSenderCoins.AmountOf(fromTicker).Int64()
	if fromAmount <= 0 || currentSenderFromCoinAmount < fromAmount {
		return 0, 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}

	if fromTicker == k.baseCoinTicker && toTicker != k.baseCoinTicker {
		emittedCLPCoinsAmount, nil := ProcessCLPTrade(ctx, sender, toTicker, fromAmount, k, true)
		return emittedCLPCoinsAmount, fromAmount, nil

	} else if toTicker == k.baseCoinTicker && fromTicker != k.baseCoinTicker {
		emittedBaseCoinsAmount, nil := ProcessCLPTrade(ctx, sender, fromTicker, fromAmount, k, false)
		return emittedBaseCoinsAmount, emittedBaseCoinsAmount, nil
	}
	emittedBaseCoinsAmount, nil := ProcessCLPTrade(ctx, sender, fromTicker, fromAmount, k, false)
	emittedCLPCoinsAmount, nil := ProcessCLPTrade(ctx, sender, toTicker, emittedBaseCoinsAmount, k, true)
	return emittedCLPCoinsAmount, emittedBaseCoinsAmount, nil
}

// Implements sdk.AccountMapper.
func (k Keeper) SetCLP(ctx sdk.Context, clp types.CLP) {
	ticker := clp.Ticker
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.MarshalBinary(clp)
	if err != nil {
		panic(err)
	}
	store.Set(MakeCLPStoreKey(ticker), bz)
}

// Turn a clp ticker to key used to get it from the clp store
func MakeCLPStoreKey(ticker string) []byte {
	return append([]byte("clp:"), []byte(ticker)...)
}
