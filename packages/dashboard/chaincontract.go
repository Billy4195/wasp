package dashboard

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/labstack/echo/v4"
)

//go:embed templates/chaincontract.tmpl
var tplChainContract string

func (d *Dashboard) initChainContract(e *echo.Echo, r renderer) {
	route := e.GET("/chain/:chainid/contract/:hname", d.handleChainContract)
	route.Name = "chainContract"
	r[route.Path] = d.makeTemplate(e, tplChainContract, tplWs)
}

func (d *Dashboard) handleChainContract(c echo.Context) error {
	chainID, err := iscp.ChainIDFromBase58(c.Param("chainid"))
	if err != nil {
		return err
	}

	hname, err := iscp.HnameFromString(c.Param("hname"))
	if err != nil {
		return err
	}

	result := &ChainContractTemplateParams{
		BaseTemplateParams: d.BaseParams(c, chainBreadcrumb(c.Echo(), *chainID), Tab{
			Path:  c.Path(),
			Title: fmt.Sprintf("Contract %d", hname),
			Href:  "#",
		}),
		ChainID: chainID,
		Hname:   hname,
	}

	chain := d.wasp.GetChain(chainID)
	if chain != nil {
		r, err := d.wasp.CallView(chain, root.Interface.Hname(), root.FuncFindContract, codec.MakeDict(map[string]interface{}{
			root.ParamHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		result.ContractRecord, err = root.DecodeContractRecord(r[root.VarData])
		if err != nil {
			return err
		}

		r, err = d.wasp.CallView(chain, eventlog.Interface.Hname(), eventlog.FuncGetRecords, codec.MakeDict(map[string]interface{}{
			eventlog.ParamContractHname: codec.EncodeHname(hname),
		}))
		if err != nil {
			return err
		}
		records := collections.NewArray16ReadOnly(r, eventlog.ParamRecords)
		result.Log = make([]*collections.TimestampedLogRecord, records.MustLen())
		for i := uint16(0); i < records.MustLen(); i++ {
			b := records.MustGetAt(i)
			result.Log[i], err = collections.ParseRawLogRecord(b)
			if err != nil {
				return err
			}
		}

		result.RootInfo, err = d.fetchRootInfo(chain)
		if err != nil {
			return err
		}
	}

	return c.Render(http.StatusOK, c.Path(), result)
}

type ChainContractTemplateParams struct {
	BaseTemplateParams

	ChainID *iscp.ChainID
	Hname   iscp.Hname

	ContractRecord *root.ContractRecord
	Log            []*collections.TimestampedLogRecord
	RootInfo       RootInfo
}
