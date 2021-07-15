package request

import (
	"net/http"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/testutil/testchain"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/webapi/model"
	"github.com/iotaledger/wasp/packages/webapi/routes"
	"github.com/iotaledger/wasp/packages/webapi/testutil"
)

func createMockedGetChain(t *testing.T) getChainFn {
	return func(chainID *iscp.ChainID) chain.ChainCore {
		return testchain.NewMockedChainCore(t, *chainID, testlogger.NewLogger(t))
	}
}

func getAccountBalanceMocked(ch chain.ChainCore, agentID *iscp.AgentID) (map[ledgerstate.Color]uint64, error) {
	ret := make(map[ledgerstate.Color]uint64)
	ret[ledgerstate.ColorIOTA] = 100
	return ret, nil
}

func dummyOffledgerRequest() *request.RequestOffLedger {
	contract := iscp.Hn("somecontract")
	entrypoint := iscp.Hn("someentrypoint")
	args := requestargs.New(dict.Dict{})
	req := request.NewRequestOffLedger(contract, entrypoint, args)
	keys, _ := testkey.GenKeyAddr()
	req.Sign(keys)
	return req
}

func TestNewRequestBase64(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain:          createMockedGetChain(t),
		getAccountBalance: getAccountBalanceMocked,
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": iscp.RandomChainID().Base58()},
		model.OffLedgerRequestBody{Request: model.NewBytes(dummyOffledgerRequest().Bytes())},
		nil,
		http.StatusAccepted,
	)
}

func TestNewRequestBinary(t *testing.T) {
	instance := &offLedgerReqAPI{
		getChain:          createMockedGetChain(t),
		getAccountBalance: getAccountBalanceMocked,
	}

	testutil.CallWebAPIRequestHandler(
		t,
		instance.handleNewRequest,
		http.MethodPost,
		routes.NewRequest(":chainID"),
		map[string]string{"chainID": iscp.RandomChainID().Base58()},
		dummyOffledgerRequest().Bytes(),
		nil,
		http.StatusAccepted,
	)
}
