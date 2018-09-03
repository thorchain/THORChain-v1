package helpers

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	lcdhelpers "github.com/cosmos/cosmos-sdk/client/lcd/helpers"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

//Setup ChainID, Gas, Account number for a context
func SetupContext(ctx context.CoreContext, from sdk.AccAddress, chainId string) (context.CoreContext, error) {
	// add gas to context
	ctx = ctx.WithGas(10000)

	// add chain-id to context
	ctx = ctx.WithChainID(chainId)

	//add account number and sequence
	ctx, err := lcdhelpers.EnsureAccountNumber(ctx, 0, from)
	if err != nil {
		fmt.Println(err)
		return ctx, err
	}
	return ctx, nil
}

//Process a Message with a private key, not using account from keybase
func PrivProcessMsg(ctx context.CoreContext, priv tmcrypto.PrivKey, cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	//sign
	txBytes, err := PrivSignAndBuild(priv, []sdk.Msg{msg}, cdc, ctx)
	if err != nil {
		return nil, err
	}

	// send
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return nil, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		return nil, err
	}
	return output, err
}

//Process a Message with a keybase accountName and password
func ProcessMsg(ctx context.CoreContext, accountName string, password string, cdc *wire.Codec, msg sdk.Msg) ([]byte, error) {
	//sign
	txBytes, err := ctx.SignAndBuild(accountName, password, []sdk.Msg{msg}, cdc)
	if err != nil {
		return nil, err
	}

	// send
	res, err := ctx.BroadcastTxAsync(txBytes)
	if err != nil {
		return nil, err
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		return nil, err
	}
	return output, err
}

//Get Address from account name in keybase
func GetFromAddress(kb keys.Keybase, localAccountName string) (sdk.AccAddress, error) {
	info, err := kb.Get(localAccountName)
	if err != nil {
		fmt.Println(err)
		return sdk.AccAddress{}, err
	}

	from := sdk.AccAddress(info.GetPubKey().Address())
	return from, nil
}

//Sign a transaction with a given private key
func SignPriv(priv tmcrypto.PrivKey, msg []byte) (sig tmcrypto.Signature, pub tmcrypto.PubKey, err error) {
	sig, err = priv.Sign(msg)
	if err != nil {
		return nil, nil, err
	}
	pub = priv.PubKey()
	return sig, pub, nil
}

//Sign and build the transaction from the msg with a private key
func PrivSignAndBuild(priv tmcrypto.PrivKey, msgs []sdk.Msg, cdc *wire.Codec, ctx context.CoreContext) ([]byte, error) {
	// build the Sign Messsage from the Standard Message
	chainID := ctx.ChainID
	if chainID == "" {
		return nil, errors.Errorf("chain ID required but not specified")
	}
	accnum := ctx.AccountNumber
	sequence := ctx.Sequence
	memo := ctx.Memo

	fee := sdk.Coin{}
	if ctx.Fee != "" {
		parsedFee, err := sdk.ParseCoin(ctx.Fee)
		if err != nil {
			return nil, err
		}
		fee = parsedFee
	}

	signMsg := auth.StdSignMsg{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
		Msgs:          msgs,
		Memo:          memo,
		Fee:           auth.NewStdFee(ctx.Gas, fee), // TODO run simulate to estimate gas?
	}

	// sign and build
	bz := signMsg.Bytes()

	sig, pubkey, err := SignPriv(priv, bz)
	if err != nil {
		return nil, err
	}
	sigs := []auth.StdSignature{{
		PubKey:        pubkey,
		Signature:     sig,
		AccountNumber: accnum,
		Sequence:      sequence,
	}}

	// marshal bytes
	tx := auth.NewStdTx(signMsg.Msgs, signMsg.Fee, sigs, memo)

	return cdc.MarshalBinary(tx)
}
