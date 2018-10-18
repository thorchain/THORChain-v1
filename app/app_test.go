package app

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
)

func setGenesis(app *ThorchainApp, accs ...*auth.BaseAccount) error {
	genaccs := make([]GenesisAccount, len(accs))
	for i, acc := range accs {
		genaccs[i] = NewGenesisAccount(acc)
	}

	genesisState := GenesisState{
		Accounts:  genaccs,
		StakeData: stake.DefaultGenesisState(),
	}

	stateBytes, err := wire.MarshalJSONIndent(app.cdc, genesisState)
	if err != nil {
		return err
	}

	// Initialize the chain
	vals := []abci.Validator{}
	app.InitChain(abci.RequestInitChain{Validators: vals, AppStateBytes: stateBytes})
	app.Commit()

	return nil
}

func TestThorchaindExport(t *testing.T) {
	db := db.NewMemDB()
	app := NewThorchainApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil)
	setGenesis(app)

	// Making a new app object with the db, so that initchain hasn't been called
	newApp := NewThorchainApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)), db, nil)
	_, _, err := newApp.ExportAppStateAndValidators()
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
