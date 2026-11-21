package rest

import (
	"context"
	"encoding/json"
	"time"

	anomalyService "github.com/madnikulin50/lowcode/server/anomaly/service"
	"github.com/madnikulin50/lowcode/server/anomaly/types"
	composeService "github.com/madnikulin50/lowcode/server/compose/service"
	composeTypes "github.com/madnikulin50/lowcode/server/compose/types"

	"github.com/madnikulin50/lowcode/server/anomaly/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/store"
	"go.uber.org/zap"
)

type (
	Rule struct{}

	ruleSetPayload struct {
		Filter types.RuleFilter `json:"filter"`
		Set    types.RuleSet    `json:"set"`
	}
)

func (Rule) New() *Rule {
	return &Rule{}
}

// canManageRules checks the caller is allowed to search/manage records on
// the module a rule (or rule request) refers to: anomaly rules deliberately
// have no RBAC operations of their own (see codegen/def/anomaly.go) - access
// is gated by the underlying Compose module's own permissions instead.
func canManageRules(ctx context.Context, namespaceID, moduleID uint64) (*composeTypes.Module, error) {
	m, err := composeService.DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil {
		return nil, err
	}

	if !composeService.DefaultAccessControl.CanSearchRecordsOnModule(ctx, m) {
		return nil, errors.Unauthorized("not allowed to manage anomaly rules on this module")
	}

	return m, nil
}

func (ctrl Rule) Search(ctx context.Context, r *request.RuleSearch) (interface{}, error) {
	var (
		err error
		f   = types.RuleFilter{
			NamespaceID: r.NamespaceID,
			ModuleID:    r.ModuleID,
			Query:       r.Query,
			Enabled:     r.Enabled,
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

	set, f, err := store.SearchAnomalyRules(ctx, anomalyService.DefaultStore, f)
	if err != nil {
		return nil, err
	}

	return ruleSetPayload{Filter: f, Set: set}, nil
}

// parseParams decodes a rule's detector-specific Params from the JSON
// object string the client sends (e.g. {"min":0,"max":100} for the range
// detector). An empty string means "no params" rather than an error.
func parseParams(raw string) (types.FindingExplanation, error) {
	if raw == "" {
		return nil, nil
	}

	var params types.FindingExplanation
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return nil, errors.InvalidData("invalid params: %v", err)
	}
	return params, nil
}

func (ctrl Rule) Create(ctx context.Context, r *request.RuleCreate) (interface{}, error) {
	if _, err := canManageRules(ctx, r.NamespaceID, r.ModuleID); err != nil {
		return nil, err
	}

	params, err := parseParams(r.Params)
	if err != nil {
		return nil, err
	}

	rule := &types.Rule{
		ID:          id.Next(),
		NamespaceID: r.NamespaceID,
		ModuleID:    r.ModuleID,
		Field:       r.Field,
		Detector:    r.Detector,
		Threshold:   r.Threshold,
		Params:      params,
		Enabled:     r.Enabled,
		CreatedAt:   time.Now(),
	}

	return rule, store.CreateAnomalyRule(ctx, anomalyService.DefaultStore, rule)
}

func (ctrl Rule) Update(ctx context.Context, r *request.RuleUpdate) (interface{}, error) {
	rule, err := store.LookupAnomalyRuleByID(ctx, anomalyService.DefaultStore, r.RuleID)
	if err != nil {
		return nil, err
	}

	if _, err = canManageRules(ctx, rule.NamespaceID, rule.ModuleID); err != nil {
		return nil, err
	}

	params, err := parseParams(r.Params)
	if err != nil {
		return nil, err
	}

	// A rule's baseline state is only meaningful for the exact
	// detector/threshold/params it was built under (see
	// anomaly/service/detector.go) - switching any of these out from under
	// it would have the new detector misinterpret the old one's state, so
	// reset it on any such change rather than leave it stale.
	resetBaseline := false

	if r.Detector != "" && r.Detector != rule.Detector {
		rule.Detector = r.Detector
		resetBaseline = true
	}
	if r.Threshold != 0 && r.Threshold != rule.Threshold {
		rule.Threshold = r.Threshold
		resetBaseline = true
	}
	if r.Params != "" {
		rule.Params = params
		resetBaseline = true
	}
	rule.Enabled = r.Enabled

	updatedAt := time.Now()
	rule.UpdatedAt = &updatedAt

	if err = store.UpdateAnomalyRule(ctx, anomalyService.DefaultStore, rule); err != nil {
		return nil, err
	}

	if resetBaseline {
		if err := resetRuleBaseline(ctx, rule.ID); err != nil {
			// The rule update itself already succeeded - a failed reset
			// just means the next scan re-warms from partly-stale state
			// instead of a clean one. Log it rather than fail the request.
			anomalyService.DefaultLogger.Error("failed to reset anomaly baseline after rule update",
				zap.Uint64("ruleID", rule.ID), zap.Error(err))
		}
	}

	return rule, nil
}

func resetRuleBaseline(ctx context.Context, ruleID uint64) error {
	existing, err := store.LookupAnomalyBaselineByRuleID(ctx, anomalyService.DefaultStore, ruleID)
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	return store.DeleteAnomalyBaselineByID(ctx, anomalyService.DefaultStore, existing.ID)
}

func (ctrl Rule) Delete(ctx context.Context, r *request.RuleDelete) (interface{}, error) {
	rule, err := store.LookupAnomalyRuleByID(ctx, anomalyService.DefaultStore, r.RuleID)
	if err != nil {
		return nil, err
	}

	if _, err = canManageRules(ctx, rule.NamespaceID, rule.ModuleID); err != nil {
		return nil, err
	}

	return api.OK(), store.DeleteAnomalyRuleByID(ctx, anomalyService.DefaultStore, r.RuleID)
}
