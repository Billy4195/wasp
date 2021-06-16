// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
)

type candidateBlock struct {
	block         state.Block
	local         bool
	votes         int
	approved      bool
	nextStateHash hashing.HashValue
	nextState     state.VirtualState
}

func newCandidateBlock(block state.Block, nextStateIfProvided state.VirtualState) *candidateBlock {
	var local bool
	var stateHash hashing.HashValue
	if nextStateIfProvided == nil {
		local = false
		stateHash = hashing.NilHash
	} else {
		local = true
		stateHash = nextStateIfProvided.Hash()
	}
	return &candidateBlock{
		block:         block,
		local:         local,
		votes:         1,
		approved:      false,
		nextStateHash: stateHash,
		nextState:     nextStateIfProvided,
	}
}

func (cT *candidateBlock) getBlock() state.Block {
	return cT.block
}

func (cT *candidateBlock) addVote() {
	cT.votes++
}

func (cT *candidateBlock) getVotes() int {
	return cT.votes
}

func (cT *candidateBlock) isLocal() bool {
	return cT.local
}

func (cT *candidateBlock) isApproved() bool {
	return cT.approved
}

func (cT *candidateBlock) approveIfRightOutput(output *ledgerstate.AliasOutput) {
	if cT.block.BlockIndex() == output.GetStateIndex() {
		outputID := output.ID()
		finalHash, err := hashing.HashValueFromBytes(output.GetStateData())
		if err != nil {
			return
		}
		if cT.isLocal() {
			if cT.nextStateHash == finalHash {
				cT.approved = true
				cT.block.SetApprovingOutputID(outputID)
			}
		} else {
			if cT.block.ApprovingOutputID() == outputID {
				cT.approved = true
				cT.nextStateHash = finalHash
			}
		}
	}
}

func (cT *candidateBlock) getNextStateHash() hashing.HashValue {
	return cT.nextStateHash
}

func (ct *candidateBlock) getNextState(currentState state.VirtualState) (state.VirtualState, error) {
	if ct.isLocal() {
		return ct.nextState, nil
	} else {
		err := currentState.ApplyBlock(ct.block)
		return currentState, err
	}
}

func (cT *candidateBlock) getApprovingOutputID() ledgerstate.OutputID {
	return cT.block.ApprovingOutputID()
}
