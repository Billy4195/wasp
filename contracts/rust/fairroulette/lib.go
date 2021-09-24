// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package fairroulette

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

func OnLoad() {
	exports := wasmlib.NewScExports()
	exports.AddFunc(FuncLockBets, funcLockBetsThunk)
	exports.AddFunc(FuncPayWinners, funcPayWinnersThunk)
	exports.AddFunc(FuncPlaceBet, funcPlaceBetThunk)
	exports.AddFunc(FuncPlayPeriod, funcPlayPeriodThunk)
	exports.AddView(ViewLastWinningNumber, viewLastWinningNumberThunk)

	for i, key := range keyMap {
		idxMap[i] = key.KeyID()
	}
}

type LockBetsContext struct {
	State MutableFairRouletteState
}

func funcLockBetsThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairroulette.funcLockBets")
	// only SC itself can invoke this function
	ctx.Require(ctx.Caller() == ctx.AccountID(), "no permission")

	f := &LockBetsContext{
		State: MutableFairRouletteState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	funcLockBets(ctx, f)
	ctx.Log("fairroulette.funcLockBets ok")
}

type PayWinnersContext struct {
	State MutableFairRouletteState
}

func funcPayWinnersThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairroulette.funcPayWinners")
	// only SC itself can invoke this function
	ctx.Require(ctx.Caller() == ctx.AccountID(), "no permission")

	f := &PayWinnersContext{
		State: MutableFairRouletteState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	funcPayWinners(ctx, f)
	ctx.Log("fairroulette.funcPayWinners ok")
}

type PlaceBetContext struct {
	Params ImmutablePlaceBetParams
	State  MutableFairRouletteState
}

func funcPlaceBetThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairroulette.funcPlaceBet")
	f := &PlaceBetContext{
		Params: ImmutablePlaceBetParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairRouletteState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Number().Exists(), "missing mandatory number")
	funcPlaceBet(ctx, f)
	ctx.Log("fairroulette.funcPlaceBet ok")
}

type PlayPeriodContext struct {
	Params ImmutablePlayPeriodParams
	State  MutableFairRouletteState
}

func funcPlayPeriodThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairroulette.funcPlayPeriod")
	// only SC creator can update the play period
	ctx.Require(ctx.Caller() == ctx.ContractCreator(), "no permission")

	f := &PlayPeriodContext{
		Params: ImmutablePlayPeriodParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairRouletteState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.PlayPeriod().Exists(), "missing mandatory playPeriod")
	funcPlayPeriod(ctx, f)
	ctx.Log("fairroulette.funcPlayPeriod ok")
}

type LastWinningNumberContext struct {
	Results MutableLastWinningNumberResults
	State   ImmutableFairRouletteState
}

func viewLastWinningNumberThunk(ctx wasmlib.ScViewContext) {
	ctx.Log("fairroulette.viewLastWinningNumber")
	f := &LastWinningNumberContext{
		Results: MutableLastWinningNumberResults{
			id: wasmlib.OBJ_ID_RESULTS,
		},
		State: ImmutableFairRouletteState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	viewLastWinningNumber(ctx, f)
	ctx.Log("fairroulette.viewLastWinningNumber ok")
}
