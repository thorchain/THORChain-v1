package clp

import (
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
	k.SetCLP(ctx, clp)
	return nil
}

// Trade with CLP.
func (k Keeper) tradeBase(ctx sdk.Context, sender sdk.AccAddress, ticker string, baseCoinAmount int64) (int64, sdk.Error) {
	if baseCoinAmount <= 0 {
		return 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}
	err := k.ensureExistentCLP(ctx, ticker)
	if err != nil {
		return 0, err
	}
	currentCoins := k.bankKeeper.GetCoins(ctx, sender)
	currentBaseCoinAmount := currentCoins.AmountOf(k.baseCoinTicker).Int64()
	if currentBaseCoinAmount < baseCoinAmount {
		return 0, ErrNotEnoughCoins(DefaultCodespace).TraceSDK("")
	}
	newCoinsAmount := baseCoinAmount
	newCoins := sdk.Coins{sdk.NewCoin(ticker, newCoinsAmount)}
	spentBaseCoins := sdk.Coins{sdk.NewCoin(k.baseCoinTicker, baseCoinAmount)}
	finalCoins := currentCoins.Plus(newCoins).Minus(spentBaseCoins)
	k.bankKeeper.SetCoins(ctx, sender, finalCoins)
	return newCoinsAmount, nil
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
