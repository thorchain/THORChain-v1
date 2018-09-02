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

//Run Formula for Tokens Issued
func CalculateTokensIssued(supply int64, baseCoinAmount int64, baseTokenBalance int64, reserveRatio int) int64 {
	//tokens issued = supply * ((1 + (connectedTokensPaid/balance))^connectorWeight - 1)
	floatReserveRatio := float64(reserveRatio) / float64(100)
	tokensIssued := float64(supply) * (math.Pow(float64(1+(float64(baseCoinAmount)/float64(baseTokenBalance))), floatReserveRatio) - 1)
	// Leaving Debug prints here for future convenience for if decimal token precision is ever implemented
	// fmt.Printf("supply: %v\n", supply)
	// fmt.Printf("baseCoinAmount: %v\n", baseCoinAmount)
	// fmt.Printf("baseTokenBalance: %v\n", baseTokenBalance)
	// fmt.Printf("reserveRatio: %v\n", reserveRatio)
	// fmt.Printf("floatReserveRatio: %v\n", floatReserveRatio)
	// fmt.Printf("float64(supply): %v\n", float64(supply))
	// fmt.Printf("float64(baseCoinAmount)/float64(baseTokenBalance): %v\n", float64(baseCoinAmount)/float64(baseTokenBalance))
	// fmt.Printf("1+(float64(baseCoinAmount)/float64(baseTokenBalance)): %v\n", 1+(float64(baseCoinAmount)/float64(baseTokenBalance)))
	// fmt.Printf("1+(float64(baseCoinAmount)/float64(baseTokenBalance)): %v\n", 1+(float64(baseCoinAmount)/float64(baseTokenBalance)))
	// fmt.Printf("float64(1+(float64(baseCoinAmount)/float64(baseTokenBalance))): %v\n", float64(1+(float64(baseCoinAmount)/float64(baseTokenBalance))))
	// fmt.Printf("math.Pow(float64(1+(float64(baseCoinAmount)/float64(baseTokenBalance))), floatReserveRatio): %v\n", math.Pow(float64(1+(float64(baseCoinAmount)/float64(baseTokenBalance))), floatReserveRatio))
	// fmt.Printf("tokensIssued: %v\n", tokensIssued)
	// fmt.Printf("math.Round(tokensIssued): %v\n", math.Round(tokensIssued))
	// fmt.Printf("\n\n")
	return int64(math.Round(tokensIssued))
}

// Trade with CLP.
func (k Keeper) trade(ctx sdk.Context, sender sdk.AccAddress, fromTicker string, toTicker string, fromAmount int64) (int64, sdk.Error) {
	if fromAmount <= 0 {
		return 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}
	clp := k.GetCLP(ctx, toTicker)
	if clp.Ticker == "" {
		return 0, ErrCLPNotExists(DefaultCodespace).TraceSDK("")
	}
	currentSenderCoins := k.bankKeeper.GetCoins(ctx, sender)
	currentSenderBaseCoinAmount := currentSenderCoins.AmountOf(k.baseCoinTicker).Int64()
	if currentSenderBaseCoinAmount < fromAmount {
		return 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}
	supply := clp.CurrentSupply
	currentCLPCoins := k.bankKeeper.GetCoins(ctx, clp.AccountAddress)
	clpBaseTokenBalance := currentCLPCoins.AmountOf(k.baseCoinTicker).Int64()
	if clpBaseTokenBalance <= 0 {
		return 0, ErrCLPEmpty(DefaultCodespace).TraceSDK("")
	}
	reserveRatio := clp.ReserveRatio
	newCLPCoinsAmount := CalculateTokensIssued(supply, fromAmount, clpBaseTokenBalance, reserveRatio)
	newCLPCoins := sdk.Coins{sdk.NewCoin(toTicker, newCLPCoinsAmount)}

	spentBaseCoins := sdk.Coins{sdk.NewCoin(k.baseCoinTicker, fromAmount)}
	k.bankKeeper.AddCoins(ctx, sender, newCLPCoins)
	k.bankKeeper.SubtractCoins(ctx, sender, spentBaseCoins)
	k.bankKeeper.AddCoins(ctx, clp.AccountAddress, spentBaseCoins)
	k.bankKeeper.SubtractCoins(ctx, clp.AccountAddress, newCLPCoins)

	return newCLPCoinsAmount, nil
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
