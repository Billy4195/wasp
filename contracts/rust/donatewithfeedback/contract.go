// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

package donatewithfeedback

import "github.com/iotaledger/wasp/packages/vm/wasmlib"

type DonateCall struct {
	Func   *wasmlib.ScFunc
	Params MutableDonateParams
}

type WithdrawCall struct {
	Func   *wasmlib.ScFunc
	Params MutableWithdrawParams
}

type DonationCall struct {
	Func    *wasmlib.ScView
	Params  MutableDonationParams
	Results ImmutableDonationResults
}

type DonationInfoCall struct {
	Func    *wasmlib.ScView
	Results ImmutableDonationInfoResults
}

type Funcs struct{}

var ScFuncs Funcs

func (sc Funcs) Donate(ctx wasmlib.ScFuncCallContext) *DonateCall {
	f := &DonateCall{Func: wasmlib.NewScFunc(HScName, HFuncDonate)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) Withdraw(ctx wasmlib.ScFuncCallContext) *WithdrawCall {
	f := &WithdrawCall{Func: wasmlib.NewScFunc(HScName, HFuncWithdraw)}
	f.Func.SetPtrs(&f.Params.id, nil)
	return f
}

func (sc Funcs) Donation(ctx wasmlib.ScViewCallContext) *DonationCall {
	f := &DonationCall{Func: wasmlib.NewScView(HScName, HViewDonation)}
	f.Func.SetPtrs(&f.Params.id, &f.Results.id)
	return f
}

func (sc Funcs) DonationInfo(ctx wasmlib.ScViewCallContext) *DonationInfoCall {
	f := &DonationInfoCall{Func: wasmlib.NewScView(HScName, HViewDonationInfo)}
	f.Func.SetPtrs(nil, &f.Results.id)
	return f
}
