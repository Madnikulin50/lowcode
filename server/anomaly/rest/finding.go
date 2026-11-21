package rest

import (
	"context"
	"time"

	anomalyService "github.com/madnikulin50/lowcode/server/anomaly/service"
	"github.com/madnikulin50/lowcode/server/anomaly/types"
	composeService "github.com/madnikulin50/lowcode/server/compose/service"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"

	"github.com/madnikulin50/lowcode/server/anomaly/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/store"
)

type (
	Finding struct{}

	findingSetPayload struct {
		Filter types.FindingFilter `json:"filter"`
		Set    types.FindingSet    `json:"set"`
	}
)

func (Finding) New() *Finding {
	return &Finding{}
}

func (ctrl Finding) Search(ctx context.Context, r *request.FindingSearch) (interface{}, error) {
	var (
		err error
		f   = types.FindingFilter{
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			RecordID:    r.RecordID,
			Status:      r.Status,
			Severity:    r.Severity,
		}
	)

	if r.ModuleID > 0 {
		if _, err = canManageRules(ctx, r.NamespaceID, r.ModuleID); err != nil {
			return nil, err
		}
	} else if !composeService.DefaultAccessControl.CanReadNamespace(ctx, &composeTypes.Namespace{ID: r.NamespaceID}) {
		return nil, errors.Unauthorized("not allowed to read this namespace")
	}

	if f.Paging, err = filter.NewPaging(r.Limit, r.PageCursor); err != nil {
		return nil, err
	}

	if f.Sorting, err = filter.NewSorting(r.Sort); err != nil {
		return nil, err
	}

	set, f, err := store.SearchAnomalyFindings(ctx, anomalyService.DefaultStore, f)
	if err != nil {
		return nil, err
	}

	// A namespace-wide search (no moduleID) can span modules the caller
	// can't read; drop those findings rather than exposing them.
	if r.ModuleID == 0 {
		visible := make(types.FindingSet, 0, len(set))
		checked := map[uint64]bool{}
		for _, finding := range set {
			if _, ok := checked[finding.ModuleID]; !ok {
				_, err := canManageRules(ctx, finding.NamespaceID, finding.ModuleID)
				checked[finding.ModuleID] = err == nil
			}
			if checked[finding.ModuleID] {
				visible = append(visible, finding)
			}
		}
		set = visible
	}

	return findingSetPayload{Filter: f, Set: set}, nil
}

func (ctrl Finding) UpdateStatus(ctx context.Context, r *request.FindingUpdateStatus) (interface{}, error) {
	finding, err := store.LookupAnomalyFindingByID(ctx, anomalyService.DefaultStore, r.FindingID)
	if err != nil {
		return nil, err
	}

	if _, err = canManageRules(ctx, finding.NamespaceID, finding.ModuleID); err != nil {
		return nil, err
	}

	switch r.Status {
	case types.StatusNew, types.StatusAcknowledged, types.StatusResolved, types.StatusFalsePositive:
		finding.Status = r.Status
	default:
		return nil, errors.InvalidData("invalid finding status: %s", r.Status)
	}

	updatedAt := time.Now()
	finding.UpdatedAt = &updatedAt

	return finding, store.UpdateAnomalyFinding(ctx, anomalyService.DefaultStore, finding)
}
