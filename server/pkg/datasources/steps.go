package datasources

import (
	"fmt"

	"github.com/cortezaproject/corteza/server/pkg/dal"
	"github.com/cortezaproject/corteza/server/pkg/filter"
	"github.com/cortezaproject/corteza/server/system/types"
)

// convStepJoin converts ReportStepJoin to dal.Join
func convStepJoin(step types.ReportStepJoin, defs FrameDefinitionSet) (out *dal.Join, err error) {
	// Validation
	if len(defs) > 1 {
		err = fmt.Errorf("cannot convert join step: expecting at most one definition, got %d", len(defs))
		return
	}

	// Get additional filtering
	var extf filter.Filter
	if len(defs) == 1 {
		extf = filterFromDef(defs[0])
	}

	f, err := dal.FilterFromExpr(step.Filter.Node()).
		MergeFilters(extf)
	if err != nil {
		return
	}

	// Make pipeline step
	out = &dal.Join{
		Ident:    step.Name,
		RelLeft:  step.LocalSource,
		RelRight: step.ForeignSource,

		Filter: f,

		On: dal.JoinPredicate{
			Left:  step.LocalColumn,
			Right: step.ForeignColumn,
		},
	}
	return
}

// convStepLink converts ReportStepLink to dal.Link
func convStepLink(step types.ReportStepLink, defs FrameDefinitionSet) (out *dal.Link, err error) {
	// Validation
	if len(defs) > 2 {
		err = fmt.Errorf("cannot convert join step: expecting at most two definitions, got %d", len(defs))
		return
	}

	// @todo additional filtering; will need to split the dal.Link filter into
	//       left and right for more control and clarity

	var extf filter.Filter
	if len(defs) > 0 {
		extf = filterFromDef(defs[0])
	}

	// Get additional filtering
	f, err := dal.FilterFromExpr(step.Filter.Node()).
		MergeFilters(extf)
	if err != nil {
		return
	}

	// @todo temporary constraint
	if len(defs) == 2 {
		d := defs[1]
		if d.Filter != nil && d.Filter.Node() != nil || len(d.Sort) > 0 {
			return nil, fmt.Errorf("temporary constraint: cannot apply sorting/filtering to foreign frame definition")
		}
	}

	// Make pipeline step
	out = &dal.Link{
		Ident:    step.Name,
		RelLeft:  step.LocalSource,
		RelRight: step.ForeignSource,

		On: dal.LinkPredicate{
			Left:  step.LocalColumn,
			Right: step.ForeignColumn,
		},
		Filter: f,
	}
	return
}
