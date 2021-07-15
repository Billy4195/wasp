// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

const (
	mgmtFuncClaimOwnership  = "claimOwnership"
	mgmtFuncWithdrawGasFees = "withdrawGasFees"
)

// TODO this SC could adjust gasPerIota based on some conditions
var evmChainMgmtInterface = coreutil.NewContractInterface("EVMChainManagement", "EVM chain management").WithFunctions(nil, []coreutil.ContractFunctionInterface{
	coreutil.Func(mgmtFuncClaimOwnership, func(ctx iscp.Sandbox) (dict.Dict, error) {
		a := assert.NewAssert(ctx.Log())
		_, err := ctx.Call(Interface.Hname(), iscp.Hn(FuncClaimOwnership), nil, nil)
		a.RequireNoError(err)
		return nil, nil
	}),
	coreutil.Func(mgmtFuncWithdrawGasFees, func(ctx iscp.Sandbox) (dict.Dict, error) {
		a := assert.NewAssert(ctx.Log())
		_, err := ctx.Call(Interface.Hname(), iscp.Hn(FuncWithdrawGasFees), nil, nil)
		a.RequireNoError(err)
		return nil, nil
	}),
})

func TestRequestGasFees(t *testing.T) {
	evmChain := initEVMChain(t, evmChainMgmtInterface)
	soloChain := evmChain.soloChain

	err := soloChain.DeployContract(nil, evmChainMgmtInterface.Name, evmChainMgmtInterface.ProgramHash)
	require.NoError(t, err)

	// deploy solidity `storage` contract (just to produce some fees to be claimed)
	evmChain.deployStorageContract(evmChain.faucetKey, 42)

	// change owner to evnchainmanagement SC
	managerAgentID := iscp.NewAgentID(soloChain.ChainID.AsAddress(), iscp.Hn(evmChainMgmtInterface.Name))
	_, err = soloChain.PostRequestSync(
		solo.NewCallParams(Interface.Name, FuncSetNextOwner, FieldNextEvmOwner, managerAgentID).
			WithIotas(1),
		soloChain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	// claim ownership
	_, err = soloChain.PostRequestSync(
		solo.NewCallParams(evmChainMgmtInterface.Name, mgmtFuncClaimOwnership).WithIotas(1),
		soloChain.OriginatorKeyPair,
	)
	require.NoError(t, err)

	// call requestGasFees manually, so that the manager SC request funds from the evm chain, check funds are received by the manager SC
	balance0, _ := soloChain.GetAccountBalance(managerAgentID).Get(ledgerstate.ColorIOTA)

	_, err = soloChain.PostRequestSync(
		solo.NewCallParams(evmChainMgmtInterface.Name, mgmtFuncWithdrawGasFees).WithIotas(1),
		soloChain.OriginatorKeyPair,
	)
	require.NoError(t, err)
	balance1, _ := soloChain.GetAccountBalance(managerAgentID).Get(ledgerstate.ColorIOTA)

	require.Greater(t, balance1, balance0)
}
