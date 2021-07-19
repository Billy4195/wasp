// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// +build !noevm

package jsonrpc

import (
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/tools/cluster"
	"github.com/iotaledger/wasp/tools/cluster/testutil"
	"github.com/stretchr/testify/require"
)

type clusterTestEnv struct {
	testEnv
	cluster *cluster.Cluster
	chain   *cluster.Chain
}

func newClusterTestEnv(t *testing.T) *clusterTestEnv {
	evmtest.InitGoEthLogger(t)

	clu := testutil.NewCluster(t)

	chain, err := clu.DeployDefaultChain()
	require.NoError(t, err)

	chainID := evm.DefaultChainID

	_, err = chain.DeployContract(
		evmchain.Contract.Name,
		evmchain.Contract.ProgramHash.String(),
		"EVM chain on top of ISCP",
		map[string]interface{}{
			evmchain.FieldChainID: codec.EncodeUint16(uint16(chainID)),
			evmchain.FieldGenesisAlloc: evmchain.EncodeGenesisAlloc(core.GenesisAlloc{
				evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
			}),
		},
	)
	require.NoError(t, err)

	signer, _, err := clu.NewKeyPairWithFunds()
	require.NoError(t, err)

	backend := NewWaspClientBackend(chain.Client(signer))
	evmChain := NewEVMChain(backend, chainID, evmchain.Contract.Name)

	accountManager := NewAccountManager(evmtest.Accounts)

	rpcsrv := NewServer(evmChain, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &clusterTestEnv{
		testEnv: testEnv{
			t:         t,
			server:    rpcsrv,
			client:    client,
			rawClient: rawClient,
			chainID:   chainID,
		},
		cluster: clu,
		chain:   chain,
	}
}

func TestRPCClusterGetLogs(t *testing.T) {
	newClusterTestEnv(t).testRPCGetLogs()
}
