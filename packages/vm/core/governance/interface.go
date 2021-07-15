// in the blocklog core contract the VM keeps indices of blocks and requests in an optimized way
// for fast checking and timestamp access.
package governance

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

const (
	Name        = coreutil.CoreContractGovernance
	description = "Governance contract"
)

var Interface = &coreutil.ContractInterface{
	Name:        Name,
	Description: description,
	ProgramHash: hashing.HashStrings(Name),
}

func init() {
	Interface.WithFunctions(initialize, []coreutil.ContractFunctionInterface{
		coreutil.Func(FuncRotateStateController, rotateStateController),
		coreutil.Func(FuncAddAllowedStateControllerAddress, addAllowedStateControllerAddress),
		coreutil.Func(FuncRemoveAllowedStateControllerAddress, removeAllowedStateControllerAddress),
		coreutil.ViewFunc(FuncGetAllowedStateControllerAddresses, getAllowedStateControllerAddresses),
	})
}

const (
	// functions
	FuncRotateStateController               = coreutil.CoreEPRotateStateController
	FuncAddAllowedStateControllerAddress    = "addAllowedStateControllerAddress"
	FuncRemoveAllowedStateControllerAddress = "removeAllowedStateControllerAddress"
	FuncGetAllowedStateControllerAddresses  = "getAllowedStateControllerAddresses"

	// state variables
	StateVarAllowedStateControllerAddresses = "a"
	StateVarRotateToAddress                 = "r"

	// params
	ParamStateControllerAddress          = coreutil.ParamStateControllerAddress
	ParamAllowedStateControllerAddresses = "a"
)
