// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

type stateManager struct {
	committee committee.Committee

	// becomes true after initially loaded state is validated.
	// after that it is always true
	solidStateValid bool

	// pending batches of state updates are candidates to confirmation by the state transaction
	// which leads to the state transition
	// the map key is hash of the variable state which is a result of applying the
	// batch of state updates to the solid variable state
	pendingBatches map[hashing.HashValue]*pendingBatch

	// state transaction with +1 state index from the state index of solid variable state
	// it may be nil if does not exist or not fetched yet
	nextStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	// it may be nil at bootstrap when origin variable state is calculated
	solidVariableState state.VariableState

	// largest state index evidenced by other messages. If this index is more than 1 step ahead
	// of the solid variable state, it means the state of the smart contract in the current node
	// falls behind the state of the smart contract, i.e. it is not synced
	largestEvidencedStateIndex uint32

	// pseudo-random permutation of peer indices. Serves a sequence in which peers are queried for state updates
	// the permutation is calculated taking last solid variable state hash as a seed
	permutationOfPeers []uint16

	// next peer permutationOfPeers[permutationIndex] is a next peer will be asked for ths state update
	permutationIndex uint16

	// the timeout deadline for sync inquiries
	syncMessageDeadline time.Time

	// current batch being synced
	syncedBatch *syncedBatch

	// logger
	log *logger.Logger
}

type syncedBatch struct {
	msgCounter   uint16
	stateIndex   uint32
	stateUpdates []state.StateUpdate
	stateTxId    valuetransaction.ID
	ts           int64
}

type pendingBatch struct {
	// batch of state updates, not validated yet
	batch state.Batch
	// resulting variable state after applied the batch to the solidVariableState
	nextVariableState state.VariableState
	// state transaction request deadline. For committed batches only
	stateTransactionRequestDeadline time.Time
}

func New(committee committee.Committee, log *logger.Logger) committee.StateManager {
	ret := &stateManager{
		committee:          committee,
		pendingBatches:     make(map[hashing.HashValue]*pendingBatch),
		permutationOfPeers: util.GetPermutation(committee.Size(), nil),
		log:                log.Named("s"),
	}
	go ret.initLoadState()

	return ret
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error
	var batch state.Batch

	stateExist, err := state.StateExist(sm.committee.Address())
	if err != nil {
		sm.log.Error(err)
		sm.committee.Dismiss()
		return
	}

	if stateExist {
		// load last solid state and last state update batch
		sm.solidVariableState, batch, err = state.LoadVariableState(sm.committee.Address())
		if err != nil {
			sm.log.Error(err)
			sm.committee.Dismiss()
			return
		}
		h := sm.solidVariableState.Hash()
		txh := batch.StateTransactionId()
		sm.log.Debugw("solid state state has been loaded",
			"state index", sm.solidVariableState.StateIndex(),
			"state hash", h.String(),
			"approving tx", txh.String(),
		)
	} else {
		// origin state
		sm.solidVariableState = nil // por las dudas
		par := sm.committee.MetaData()
		batch = apilib.NewOriginBatch(apilib.NewOriginParams{
			Address:      par.Address,
			OwnerAddress: par.OwnerAddress,
			ProgramHash:  par.ProgramHash,
		})
		// committing a batch means linking it to the approving transaction
		// it doesn't change essence of the batch
		// here 'color' is the ID of the origin transaction
		batch.WithStateTransaction((valuetransaction.ID)(par.Color))

		sm.log.Infow("initial state wasn't found. Origin state update batch has been created",
			"state txid", batch.StateTransactionId().String())
	}
	// loaded solid variable state and the last batch of state updates
	// it needs to be validated by the state transaction, so it is added to the
	// pending batches
	if !sm.addPendingBatch(batch) {
		sm.log.Errorf("initial batch inconsistent")
		sm.committee.Dismiss()
		return
	} else {
		if sm.solidVariableState == nil {
			if !(len(sm.pendingBatches) == 1) {
				panic("assertion: len(sm.pendingBatches) == 1")
			}
		}
	}
	// open msg queue for the committee
	sm.committee.SetReadyStateManager()
}
