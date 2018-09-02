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

func (k Keeper) ensureExistentCLP(ctx sdk.Context, ticker string) sdk.Error {
	clp := k.GetCLP(ctx, ticker)
	if clp.Ticker == "" {
		return ErrCLPNotExists(DefaultCodespace).TraceSDK("")
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

//Run Formula for buying CLP coins
func CalculateCLPCoinsEmitted(clpCoinSupply int64, baseCoinsPaid int64, baseCoinBalance int64, reserveRatio int) int64 {
	//clpCoinsEmitted = clpCoinSupply * ((1 + (baseCoinsPaid/baseCoinBalance))^reserveRatio - 1)
	a := float64(clpCoinSupply)
	b := float64(baseCoinsPaid)
	c := float64(baseCoinBalance)
	d := float64(reserveRatio) / float64(100)
	return RunCLPFormula(a, b, c, d)
}

//Run Formula for selling CLP coins
func CalculateBaseCoinsEmitted(baseCoinBalance int64, clpCoinsPaid int64, clpCoinSupply int64, reserveRatio int) int64 {
	//baseCoinsEmitted = baseCoinBalance * ((1 + (clpCoinsPaid/clpCoinSupply))^(1/reserveRatio) - 1)
	a := float64(baseCoinBalance)
	b := float64(clpCoinsPaid)
	c := float64(clpCoinSupply)
	d := float64(1) / (float64(reserveRatio) / float64(100))
	return RunCLPFormula(a, b, c, d)
}

// Trade with CLP.
func (k Keeper) trade(ctx sdk.Context, sender sdk.AccAddress, fromTicker string, toTicker string, fromAmount int64) (int64, sdk.Error) {
	//Check different tickers
	if fromTicker == toTicker {
		return 0, ErrSameCoin(DefaultCodespace).TraceSDK("")
	}

	//Check sender coins ok
	currentSenderCoins := k.bankKeeper.GetCoins(ctx, sender)
	currentSenderFromCoinAmount := currentSenderCoins.AmountOf(fromTicker).Int64()
	if fromAmount <= 0 || currentSenderFromCoinAmount < fromAmount {
		return 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}

	var fromClp, toClp *types.CLP
	var fromClpCoins, toClpCoins sdk.Coins
	if fromTicker != k.baseCoinTicker {
		fromClp = k.GetCLP(ctx, fromTicker)
	}
	if toTicker != k.baseCoinTicker {
		toClp = k.GetCLP(ctx, toTicker)
	}

	//Check from clp if needed exists and coins ok
	if fromClp != nil {
		if fromClp.Ticker == "" {
			return 0, ErrCLPNotExists(DefaultCodespace).TraceSDK("")
		}
		fromClpCoins = k.bankKeeper.GetCoins(ctx, fromClp.AccountAddress)
		fromCLPBaseCoinBalance := fromClpCoins.AmountOf(k.baseCoinTicker).Int64()
		if fromCLPBaseCoinBalance <= 0 {
			return 0, ErrCLPEmpty(DefaultCodespace).TraceSDK("")
		}
	}

	//Check to clp exists if needed and coins ok
	if toClp != nil {
		if toClp.Ticker == "" {
			return 0, ErrCLPNotExists(DefaultCodespace).TraceSDK("")
		}
		toClpCoins = k.bankKeeper.GetCoins(ctx, toClp.AccountAddress)
		toCLPCoinBalance := toClpCoins.AmountOf(toTicker).Int64()
		if toCLPCoinBalance <= 0 {
			return 0, ErrCLPEmpty(DefaultCodespace).TraceSDK("")
		}
	}

	if fromTicker == k.baseCoinTicker && toTicker != k.baseCoinTicker {
		toClpCoinSupply := toClp.CurrentSupply
		toCLPBaseCoinBalance := toClpCoins.AmountOf(k.baseCoinTicker).Int64()
		toClpReserveRatio := toClp.ReserveRatio
		emittedCLPCoinsAmount := CalculateCLPCoinsEmitted(toClpCoinSupply, fromAmount, toCLPBaseCoinBalance, toClpReserveRatio)

		spentFromCoins := sdk.Coins{sdk.NewCoin(fromTicker, fromAmount)}
		emittedCLPCoins := sdk.Coins{sdk.NewCoin(toTicker, emittedCLPCoinsAmount)}

		k.bankKeeper.SubtractCoins(ctx, sender, spentFromCoins)
		k.bankKeeper.AddCoins(ctx, sender, emittedCLPCoins)
		k.bankKeeper.SubtractCoins(ctx, toClp.AccountAddress, emittedCLPCoins)
		k.bankKeeper.AddCoins(ctx, toClp.AccountAddress, spentFromCoins)
		return emittedCLPCoinsAmount, nil

	} else if toTicker == k.baseCoinTicker && fromTicker != k.baseCoinTicker {
		fromCLPBaseCoinBalance := fromClpCoins.AmountOf(k.baseCoinTicker).Int64()
		fromClpCoinSupply := fromClp.CurrentSupply
		fromClpReserveRatio := fromClp.ReserveRatio
		emittedBaseCoinsAmount := CalculateBaseCoinsEmitted(fromCLPBaseCoinBalance, fromAmount, fromClpCoinSupply, fromClpReserveRatio)

		spentFromCoins := sdk.Coins{sdk.NewCoin(fromTicker, fromAmount)}
		emittedBaseCoins := sdk.Coins{sdk.NewCoin(k.baseCoinTicker, emittedBaseCoinsAmount)}

		k.bankKeeper.SubtractCoins(ctx, sender, spentFromCoins)
		k.bankKeeper.AddCoins(ctx, sender, emittedBaseCoins)
		k.bankKeeper.SubtractCoins(ctx, fromClp.AccountAddress, emittedBaseCoins)
		k.bankKeeper.AddCoins(ctx, fromClp.AccountAddress, spentFromCoins)
		return emittedBaseCoinsAmount, nil
	}

	fromCLPBaseCoinBalance := fromClpCoins.AmountOf(k.baseCoinTicker).Int64()
	fromClpCoinSupply := fromClp.CurrentSupply
	fromClpReserveRatio := fromClp.ReserveRatio
	emittedBaseCoinsAmount := CalculateBaseCoinsEmitted(fromCLPBaseCoinBalance, fromAmount, fromClpCoinSupply, fromClpReserveRatio)

	toClpCoinSupply := toClp.CurrentSupply
	toCLPBaseCoinBalance := toClpCoins.AmountOf(k.baseCoinTicker).Int64()
	toClpReserveRatio := toClp.ReserveRatio
	emittedCLPCoinsAmount := CalculateCLPCoinsEmitted(toClpCoinSupply, emittedBaseCoinsAmount, toCLPBaseCoinBalance, toClpReserveRatio)

	spentFromCoins := sdk.Coins{sdk.NewCoin(fromTicker, fromAmount)}
	emittedBaseCoins := sdk.Coins{sdk.NewCoin(k.baseCoinTicker, emittedBaseCoinsAmount)}
	emittedCLPCoins := sdk.Coins{sdk.NewCoin(toTicker, emittedCLPCoinsAmount)}

	k.bankKeeper.SubtractCoins(ctx, sender, spentFromCoins)
	k.bankKeeper.AddCoins(ctx, fromClp.AccountAddress, spentFromCoins)
	k.bankKeeper.SubtractCoins(ctx, fromClp.AccountAddress, emittedBaseCoins)
	k.bankKeeper.AddCoins(ctx, toClp.AccountAddress, emittedBaseCoins)
	k.bankKeeper.SubtractCoins(ctx, toClp.AccountAddress, emittedCLPCoins)
	k.bankKeeper.AddCoins(ctx, sender, emittedCLPCoins)

	return emittedCLPCoinsAmount, nil
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
