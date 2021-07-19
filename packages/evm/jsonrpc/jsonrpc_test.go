// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package jsonrpc

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	t         *testing.T
	server    *rpc.Server
	client    *ethclient.Client
	rawClient *rpc.Client
	chainID   int
}

type soloTestEnv struct {
	testEnv
	solo *solo.Solo
}

func newSoloTestEnv(t *testing.T) *soloTestEnv {
	evmtest.InitGoEthLogger(t)

	chainID := evm.DefaultChainID

	s := solo.New(t, true, false).WithNativeContract(evmchain.Processor)
	chainOwner, _ := s.NewKeyPairWithFunds()
	chain := s.NewChain(chainOwner, "iscpchain")
	err := chain.DeployContract(chainOwner, "evmchain", evmchain.Contract.ProgramHash,
		evmchain.FieldChainID, codec.EncodeUint16(uint16(chainID)),
		evmchain.FieldGenesisAlloc, evmchain.EncodeGenesisAlloc(core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		}),
	)
	require.NoError(t, err)
	signer, _ := s.NewKeyPairWithFunds()
	backend := NewSoloBackend(s, chain, signer)
	evmChain := NewEVMChain(backend, chainID, evmchain.Contract.Name)

	accountManager := NewAccountManager(evmtest.Accounts)

	rpcsrv := NewServer(evmChain, accountManager)
	t.Cleanup(rpcsrv.Stop)

	rawClient := rpc.DialInProc(rpcsrv)
	client := ethclient.NewClient(rawClient)
	t.Cleanup(client.Close)

	return &soloTestEnv{
		testEnv: testEnv{
			t:         t,
			server:    rpcsrv,
			client:    client,
			rawClient: rawClient,
			chainID:   chainID,
		},
		solo: s,
	}
}

func generateKey(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	return key, addr
}

var requestFundsAmount = big.NewInt(1e18) // 1 ETH

func (e *testEnv) signer() types.Signer {
	return evm.Signer(big.NewInt(int64(e.chainID)))
}

func (e *testEnv) requestFunds(target common.Address) *types.Transaction {
	nonce, err := e.client.NonceAt(context.Background(), evmtest.FaucetAddress, nil)
	require.NoError(e.t, err)
	tx, err := types.SignTx(
		types.NewTransaction(nonce, target, requestFundsAmount, evm.TxGas, evm.GasPrice, nil),
		e.signer(),
		evmtest.FaucetKey,
	)
	require.NoError(e.t, err)
	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)
	return tx
}

func (e *testEnv) deployEVMContract(creator *ecdsa.PrivateKey, contractABI abi.ABI, contractBytecode []byte, args ...interface{}) (*types.Transaction, common.Address) {
	creatorAddress := crypto.PubkeyToAddress(creator.PublicKey)

	nonce := e.nonceAt(creatorAddress)

	constructorArguments, err := contractABI.Pack("", args...)
	require.NoError(e.t, err)

	data := concatenate(contractBytecode, constructorArguments)

	value := big.NewInt(0)

	gasLimit := e.estimateGas(ethereum.CallMsg{
		From:     creatorAddress,
		To:       nil, // contract creation
		GasPrice: evm.GasPrice,
		Value:    value,
		Data:     data,
	})

	tx, err := types.SignTx(
		types.NewContractCreation(nonce, value, gasLimit, evm.GasPrice, data),
		e.signer(),
		creator,
	)
	require.NoError(e.t, err)

	err = e.client.SendTransaction(context.Background(), tx)
	require.NoError(e.t, err)

	return tx, crypto.CreateAddress(creatorAddress, nonce)
}

func concatenate(a, b []byte) []byte {
	r := make([]byte, 0, len(a)+len(b))
	r = append(r, a...)
	r = append(r, b...)
	return r
}

func (e *testEnv) estimateGas(msg ethereum.CallMsg) uint64 {
	gas, err := e.client.EstimateGas(context.Background(), msg)
	require.NoError(e.t, err)
	return gas
}

func (e *testEnv) nonceAt(address common.Address) uint64 {
	nonce, err := e.client.NonceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return nonce
}

func (e *testEnv) blockNumber() uint64 {
	blockNumber, err := e.client.BlockNumber(context.Background())
	require.NoError(e.t, err)
	return blockNumber
}

func (e *testEnv) blockByNumber(number *big.Int) *types.Block {
	block, err := e.client.BlockByNumber(context.Background(), number)
	require.NoError(e.t, err)
	return block
}

func (e *testEnv) blockByHash(hash common.Hash) *types.Block {
	block, err := e.client.BlockByHash(context.Background(), hash)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.t, err)
	return block
}

func (e *testEnv) transactionByHash(hash common.Hash) *types.Transaction {
	tx, isPending, err := e.client.TransactionByHash(context.Background(), hash)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.t, err)
	require.False(e.t, isPending)
	return tx
}

func (e *testEnv) transactionByBlockHashAndIndex(blockHash common.Hash, index uint) *types.Transaction {
	tx, err := e.client.TransactionInBlock(context.Background(), blockHash, index)
	if errors.Is(err, ethereum.NotFound) {
		return nil
	}
	require.NoError(e.t, err)
	return tx
}

func (e *testEnv) uncleByBlockHashAndIndex(blockHash common.Hash, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.rawClient.Call(&uncle, "eth_getUncleByBlockHashAndIndex", blockHash, hexutil.Uint(index))
	require.NoError(e.t, err)
	return uncle
}

func (e *testEnv) transactionByBlockNumberAndIndex(blockNumber *big.Int, index uint) *RPCTransaction {
	var tx *RPCTransaction
	err := e.rawClient.Call(&tx, "eth_getTransactionByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	require.NoError(e.t, err)
	return tx
}

func (e *testEnv) uncleByBlockNumberAndIndex(blockNumber *big.Int, index uint) map[string]interface{} {
	var uncle map[string]interface{}
	err := e.rawClient.Call(&uncle, "eth_getUncleByBlockNumberAndIndex", (*hexutil.Big)(blockNumber), hexutil.Uint(index))
	require.NoError(e.t, err)
	return uncle
}

func (e *testEnv) blockTransactionCountByHash(hash common.Hash) uint {
	n, err := e.client.TransactionCount(context.Background(), hash)
	require.NoError(e.t, err)
	return n
}

func (e *testEnv) uncleCountByBlockHash(hash common.Hash) uint {
	var res hexutil.Uint
	err := e.rawClient.Call(&res, "eth_getUncleCountByBlockHash", hash)
	require.NoError(e.t, err)
	return uint(res)
}

func (e *testEnv) blockTransactionCountByNumber() uint {
	// the client only supports calling this method with "pending"
	n, err := e.client.PendingTransactionCount(context.Background())
	require.NoError(e.t, err)
	return n
}

func (e *testEnv) uncleCountByBlockNumber(blockNumber *big.Int) uint {
	var res hexutil.Uint
	err := e.rawClient.Call(&res, "eth_getUncleCountByBlockNumber", (*hexutil.Big)(blockNumber))
	require.NoError(e.t, err)
	return uint(res)
}

func (e *testEnv) balance(address common.Address) *big.Int {
	bal, err := e.client.BalanceAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return bal
}

func (e *testEnv) code(address common.Address) []byte {
	code, err := e.client.CodeAt(context.Background(), address, nil)
	require.NoError(e.t, err)
	return code
}

func (e *testEnv) storage(address common.Address, key common.Hash) []byte {
	data, err := e.client.StorageAt(context.Background(), address, key, nil)
	require.NoError(e.t, err)
	return data
}

func (e *testEnv) txReceipt(hash common.Hash) *types.Receipt {
	r, err := e.client.TransactionReceipt(context.Background(), hash)
	require.NoError(e.t, err)
	return r
}

func (e *testEnv) accounts() []common.Address {
	var res []common.Address
	err := e.rawClient.Call(&res, "eth_accounts")
	require.NoError(e.t, err)
	return res
}

func (e *testEnv) sign(address common.Address, data []byte) []byte {
	var res hexutil.Bytes
	err := e.rawClient.Call(&res, "eth_sign", address, hexutil.Bytes(data))
	require.NoError(e.t, err)
	return res
}

func (e *testEnv) signTransaction(args *SendTxArgs) []byte {
	var res hexutil.Bytes
	err := e.rawClient.Call(&res, "eth_signTransaction", args)
	require.NoError(e.t, err)
	return res
}

func (e *testEnv) sendTransaction(args *SendTxArgs) common.Hash {
	var res common.Hash
	err := e.rawClient.Call(&res, "eth_sendTransaction", args)
	require.NoError(e.t, err)
	return res
}

func (e *testEnv) getLogs(q ethereum.FilterQuery) []types.Log {
	logs, err := e.client.FilterLogs(context.Background(), q)
	require.NoError(e.t, err)
	return logs
}

func TestRPCGetBalance(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Zero(t, big.NewInt(0).Cmp(env.balance(receiverAddress)))
	env.requestFunds(receiverAddress)
	require.Zero(t, big.NewInt(1e18).Cmp(env.balance(receiverAddress)))
}

func TestRPCGetCode(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := generateKey(t)

	// account address
	{
		env.requestFunds(creatorAddress)
		require.Empty(t, env.code(creatorAddress))
	}
	// contract address
	{
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
		require.NotEmpty(t, env.code(contractAddress))
	}
}

func TestRPCGetStorage(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := generateKey(t)

	env.requestFunds(creatorAddress)

	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

	// first static variable in contract (uint32 n) has slot 0. See:
	// https://docs.soliditylang.org/en/v0.6.6/miscellaneous.html#layout-of-state-variables-in-storage
	slot := common.Hash{}
	ret := env.storage(contractAddress, slot)

	var v uint32
	err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
	require.NoError(t, err)
	require.Equal(t, uint32(42), v)
}

func TestRPCBlockNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.blockNumber())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockNumber())
}

func TestRPCGetTransactionCount(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.nonceAt(evmtest.FaucetAddress))
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.nonceAt(evmtest.FaucetAddress))
}

func TestRPCGetBlockByNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.EqualValues(t, 0, env.blockByNumber(big.NewInt(0)).Number().Uint64())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockByNumber(big.NewInt(1)).Number().Uint64())
}

func TestRPCGetBlockByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.blockByHash(common.Hash{}))
	require.EqualValues(t, 0, env.blockByHash(env.blockByNumber(big.NewInt(0)).Hash()).Number().Uint64())
	env.requestFunds(receiverAddress)
	require.EqualValues(t, 1, env.blockByHash(env.blockByNumber(big.NewInt(1)).Hash()).Number().Uint64())
}

func TestRPCGetTransactionByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByHash(common.Hash{}))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByHash(block1.Transactions()[0].Hash())
	require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetTransactionByBlockHashAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByBlockHashAndIndex(common.Hash{}, 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByBlockHashAndIndex(block1.Hash(), 0)
	require.Equal(t, block1.Transactions()[0].Hash(), tx.Hash())
}

func TestRPCGetUncleByBlockHashAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.uncleByBlockHashAndIndex(common.Hash{}, 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Nil(t, env.uncleByBlockHashAndIndex(block1.Hash(), 0))
}

func TestRPCGetTransactionByBlockNumberAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.transactionByBlockNumberAndIndex(big.NewInt(3), 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	tx := env.transactionByBlockNumberAndIndex(block1.Number(), 0)
	require.EqualValues(t, block1.Hash(), *tx.BlockHash)
	require.EqualValues(t, 0, *tx.TransactionIndex)
}

func TestRPCGetUncleByBlockNumberAndIndex(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	require.Nil(t, env.uncleByBlockNumberAndIndex(big.NewInt(3), 0))
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Nil(t, env.uncleByBlockNumberAndIndex(block1.Number(), 0))
}

func TestRPCGetTransactionCountByHash(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Positive(t, len(block1.Transactions()))
	require.EqualValues(t, len(block1.Transactions()), env.blockTransactionCountByHash(block1.Hash()))
	require.EqualValues(t, 0, env.blockTransactionCountByHash(common.Hash{}))
}

func TestRPCGetUncleCountByBlockHash(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Zero(t, len(block1.Uncles()))
	require.EqualValues(t, len(block1.Uncles()), env.uncleCountByBlockHash(block1.Hash()))
	require.EqualValues(t, 0, env.uncleCountByBlockHash(common.Hash{}))
}

func TestRPCGetTransactionCountByNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(nil)
	require.Positive(t, len(block1.Transactions()))
	require.EqualValues(t, len(block1.Transactions()), env.blockTransactionCountByNumber())
}

func TestRPCGetUncleCountByBlockNumber(t *testing.T) {
	env := newSoloTestEnv(t)
	_, receiverAddress := generateKey(t)
	env.requestFunds(receiverAddress)
	block1 := env.blockByNumber(big.NewInt(1))
	require.Zero(t, len(block1.Uncles()))
	require.EqualValues(t, len(block1.Uncles()), env.uncleCountByBlockNumber(big.NewInt(1)))
}

func TestRPCAccounts(t *testing.T) {
	env := newSoloTestEnv(t)
	accounts := env.accounts()
	require.Equal(t, len(evmtest.Accounts), len(accounts))
}

func TestRPCSign(t *testing.T) {
	env := newSoloTestEnv(t)
	signed := env.sign(evmtest.AccountAddress(0), []byte("hello"))
	require.NotEmpty(t, signed)
}

func TestRPCSignTransaction(t *testing.T) {
	env := newSoloTestEnv(t)

	from := evmtest.AccountAddress(0)
	to := evmtest.AccountAddress(1)
	gas := hexutil.Uint64(evm.TxGas)
	nonce := hexutil.Uint64(env.nonceAt(from))
	signed := env.signTransaction(&SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(requestFundsAmount),
		Nonce:    &nonce,
	})
	require.NotEmpty(t, signed)

	// test that the signed tx can be sent
	env.requestFunds(from)
	err := env.rawClient.Call(nil, "eth_sendRawTransaction", hexutil.Encode(signed))
	require.NoError(t, err)
}

func TestRPCSendTransaction(t *testing.T) {
	env := newSoloTestEnv(t)

	from := evmtest.AccountAddress(0)
	env.requestFunds(from)

	to := evmtest.AccountAddress(1)
	gas := hexutil.Uint64(evm.TxGas)
	nonce := hexutil.Uint64(env.nonceAt(from))
	txHash := env.sendTransaction(&SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(requestFundsAmount),
		Nonce:    &nonce,
	})
	require.NotEqualValues(t, common.Hash{}, txHash)
}

func TestRPCGetTxReceipt(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddr := generateKey(t)

	// regular transaction
	{
		tx := env.requestFunds(creatorAddr)
		receipt := env.txReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, common.Address{}, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(1), receipt.BlockNumber)
		require.EqualValues(t, env.blockByNumber(big.NewInt(1)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	}

	// contract creation
	{
		contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
		require.NoError(t, err)
		tx, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))
		receipt := env.txReceipt(tx.Hash())

		require.EqualValues(t, types.LegacyTxType, receipt.Type)
		require.EqualValues(t, types.ReceiptStatusSuccessful, receipt.Status)
		require.NotZero(t, receipt.CumulativeGasUsed)
		require.EqualValues(t, types.Bloom{}, receipt.Bloom)
		require.EqualValues(t, 0, len(receipt.Logs))

		require.EqualValues(t, tx.Hash(), receipt.TxHash)
		require.EqualValues(t, contractAddress, receipt.ContractAddress)
		require.NotZero(t, receipt.GasUsed)

		require.EqualValues(t, big.NewInt(2), receipt.BlockNumber)
		require.EqualValues(t, env.blockByNumber(big.NewInt(2)).Hash(), receipt.BlockHash)
		require.EqualValues(t, 0, receipt.TransactionIndex)
	}
}

func TestRPCCall(t *testing.T) {
	env := newSoloTestEnv(t)
	creator, creatorAddress := generateKey(t)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.StorageContractABI))
	require.NoError(t, err)
	_, contractAddress := env.deployEVMContract(creator, contractABI, evmtest.StorageContractBytecode, uint32(42))

	callArguments, err := contractABI.Pack("retrieve")
	require.NoError(t, err)

	ret, err := env.client.CallContract(context.Background(), ethereum.CallMsg{
		From: creatorAddress,
		To:   &contractAddress,
		Data: callArguments,
	}, nil)
	require.NoError(t, err)

	var v uint32
	err = contractABI.UnpackIntoInterface(&v, "retrieve", ret)
	require.NoError(t, err)
	require.Equal(t, uint32(42), v)
}

func TestRPCGetLogs(t *testing.T) {
	newSoloTestEnv(t).testRPCGetLogs()
}

func (e *testEnv) testRPCGetLogs() {
	creator, creatorAddress := evmtest.Accounts[0], evmtest.AccountAddress(0)
	contractABI, err := abi.JSON(strings.NewReader(evmtest.ERC20ContractABI))
	require.NoError(e.t, err)
	contractAddress := crypto.CreateAddress(creatorAddress, e.nonceAt(creatorAddress))

	filterQuery := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	require.Empty(e.t, e.getLogs(filterQuery))

	e.deployEVMContract(creator, contractABI, evmtest.ERC20ContractBytecode, "TestCoin", "TEST")

	require.Equal(e.t, 1, len(e.getLogs(filterQuery)))

	recipientAddress := evmtest.AccountAddress(1)
	nonce := hexutil.Uint64(e.nonceAt(creatorAddress))
	callArguments, err := contractABI.Pack("transfer", recipientAddress, big.NewInt(1337))
	value := big.NewInt(0)
	gas := hexutil.Uint64(e.estimateGas(ethereum.CallMsg{
		From:  creatorAddress,
		To:    &contractAddress,
		Value: value,
		Data:  callArguments,
	}))
	require.NoError(e.t, err)
	e.sendTransaction(&SendTxArgs{
		From:     creatorAddress,
		To:       &contractAddress,
		Gas:      &gas,
		GasPrice: (*hexutil.Big)(evm.GasPrice),
		Value:    (*hexutil.Big)(value),
		Nonce:    &nonce,
		Data:     (*hexutil.Bytes)(&callArguments),
	})

	require.Equal(e.t, 2, len(e.getLogs(filterQuery)))
}
