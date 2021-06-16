package transaction

import (
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
)

// NewChainOriginTransaction creates new origin transaction for the self-governed chain
// returns the transaction and newly minted chain ID
func NewChainOriginTransaction(
	keyPair *ed25519.KeyPair,
	stateAddress ledgerstate.Address,
	balance map[ledgerstate.Color]uint64,
	timestamp time.Time,
	allInputs ...ledgerstate.Output,
) (*ledgerstate.Transaction, chainid.ChainID, error) {
	walletAddr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	txb := utxoutil.NewBuilder(allInputs...).WithTimestamp(timestamp)

	stateHash := state.OriginStateHash()
	if len(balance) == 0 {
		balance = map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: ledgerstate.DustThresholdAliasOutputIOTA}
	}
	if err := txb.AddNewAliasMint(balance, stateAddress, stateHash.Bytes()); err != nil {
		return nil, chainid.ChainID{}, err
	}
	// adding reminder in compressing mode, i.e. all provided inputs will be consumed
	if err := txb.AddRemainderOutputIfNeeded(walletAddr, nil, true); err != nil {
		return nil, chainid.ChainID{}, err
	}
	tx, err := txb.BuildWithED25519(keyPair)
	if err != nil {
		return nil, chainid.ChainID{}, err
	}
	// determine aliasAddress of the newly minted chain
	chained, err := utxoutil.GetSingleChainedAliasOutput(tx)
	if err != nil {
		return nil, chainid.ChainID{}, err
	}
	chainID, err := chainid.ChainIDFromAddress(chained.Address())
	if err != nil {
		return nil, chainid.ChainID{}, err
	}
	return tx, *chainID, nil
}

// NewRootInitRequestTransaction is a first request to be sent to the uninitialized
// chain. At this moment it only is able to process this specific request
// the request contains minimum data needed to bootstrap the chain
// TransactionEssence must be signed by the same address which created origin transaction
func NewRootInitRequestTransaction(
	keyPair *ed25519.KeyPair,
	chainID chainid.ChainID,
	description string,
	timestamp time.Time,
	allInputs ...ledgerstate.Output,
) (*ledgerstate.Transaction, error) {
	txb := utxoutil.NewBuilder(allInputs...).WithTimestamp(timestamp)

	args := requestargs.New(nil)

	const paramChainID = "$$chainid$$"
	const paramDescription = "$$description$$"

	args.AddEncodeSimple(paramChainID, codec.EncodeChainID(chainID))
	args.AddEncodeSimple(paramDescription, codec.EncodeString(description))

	metadata := request.NewRequestMetadata().
		WithTarget(coretypes.Hn("root")).
		WithEntryPoint(coretypes.EntryPointInit).
		WithArgs(args).
		Bytes()
	err := txb.AddExtendedOutputConsume(chainID.AsAddress(), metadata, map[ledgerstate.Color]uint64{
		ledgerstate.ColorIOTA: 1,
	})
	if err != nil {
		return nil, err
	}
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	if err := txb.AddRemainderOutputIfNeeded(addr, nil, true); err != nil {
		return nil, err
	}
	tx, err := txb.BuildWithED25519(keyPair)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
