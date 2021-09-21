// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package coreaccounts

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type DepositCall struct {
	Func   *wasmlib.ScFunc
	Params MutableDepositParams
}

type HarvestCall struct {
	Func   *wasmlib.ScFunc
	Params MutableHarvestParams
}

type WithdrawCall struct {
	Func *wasmlib.ScFunc
}

type AccountsCall struct {
	Func    *wasmlib.ScView
	Results ImmutableAccountsResults
}

type BalanceCall struct {
	Func    *wasmlib.ScView
	Params  MutableBalanceParams
	Results ImmutableBalanceResults
}

type GetAccountNonceCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetAccountNonceParams
	Results ImmutableGetAccountNonceResults
}

type TotalAssetsCall struct {
	Func    *wasmlib.ScView
	Results ImmutableTotalAssetsResults
}

type Funcs struct{}

var ScFuncs Funcs

func (sc Funcs) Deposit(ctx wasmlib.ScFuncCallContext) *DepositCall {
	f := &DepositCall{Func: wasmlib.NewScFunc(HScName, HFuncDeposit)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) Harvest(ctx wasmlib.ScFuncCallContext) *HarvestCall {
	f := &HarvestCall{Func: wasmlib.NewScFunc(HScName, HFuncHarvest)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) Withdraw(ctx wasmlib.ScFuncCallContext) *WithdrawCall {
	return &WithdrawCall{Func: wasmlib.NewScFunc(HScName, HFuncWithdraw)}
}

func (sc Funcs) Accounts(ctx wasmlib.ScViewCallContext) *AccountsCall {
	f := &AccountsCall{Func: wasmlib.NewScView(HScName, HViewAccounts)}
	f.Func.SetPtrs(nil, &f.Results.id)
	return f
}

func (sc Funcs) Balance(ctx wasmlib.ScViewCallContext) *BalanceCall {
	f := &BalanceCall{Func: wasmlib.NewScView(HScName, HViewBalance)}
	f.Func.SetPtrs(&f.Params.id, &f.Results.id)
	return f
}

func (sc Funcs) GetAccountNonce(ctx wasmlib.ScViewCallContext) *GetAccountNonceCall {
	f := &GetAccountNonceCall{Func: wasmlib.NewScView(HScName, HViewGetAccountNonce)}
	f.Func.SetPtrs(&f.Params.id, &f.Results.id)
	return f
}

func (sc Funcs) TotalAssets(ctx wasmlib.ScViewCallContext) *TotalAssetsCall {
	f := &TotalAssetsCall{Func: wasmlib.NewScView(HScName, HViewTotalAssets)}
	f.Func.SetPtrs(nil, &f.Results.id)
	return f
}

func OnLoad() {
	exports := wasmlib.NewScExports()
	exports.AddFunc(FuncDeposit, nil)
	exports.AddFunc(FuncHarvest, nil)
	exports.AddFunc(FuncWithdraw, nil)
	exports.AddView(ViewAccounts, nil)
	exports.AddView(ViewBalance, nil)
	exports.AddView(ViewGetAccountNonce, nil)
	exports.AddView(ViewTotalAssets, nil)
}