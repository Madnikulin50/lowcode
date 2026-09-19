package service

import (
	"context"
	"log"
	"strconv"
	"sync"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
	"github.com/madnikulin50/lowcode/server/pkg/riskstore"
)

var startRiskRecordTriggersOnce sync.Once

// StartRiskRecordTriggers subscribes once to compose:record
// afterCreate/afterUpdate and reassesses risk for every RiskSubjectBinding
// whose ModuleID matches the event's module and whose AutoRecalc is set —
// the same eventbus mechanism StartRuleChainRecordTriggers (see
// rulechain_trigger.go) already uses for rule chains, so a record edit and a
// risk reassessment go through one consistent trigger surface rather than
// two competing ones.
func StartRiskRecordTriggers() {
	startRiskRecordTriggersOnce.Do(func() {
		eventbus.Service().Register(func(ctx context.Context, ev eventbus.Event) error {
			reassessMatchingBindings(ctx, ev)
			return nil
		},
			eventbus.For("compose:record"),
			eventbus.On("afterCreate", "afterUpdate"),
			eventbus.Weight(100),
		)
	})
}

func reassessMatchingBindings(ctx context.Context, ev eventbus.Event) {
	re, ok := ev.(composeRecordEvent) // declared in rulechain_trigger.go, same package
	if !ok {
		return
	}
	rec := re.Record()
	mod := re.Module()
	if rec == nil || mod == nil {
		return
	}

	for _, binding := range riskstore.BindingsForModule(mod.ID) {
		if !binding.AutoRecalc {
			continue
		}
		if ns := re.Namespace(); ns != nil && ns.ID != 0 && binding.NamespaceID != 0 && ns.ID != binding.NamespaceID {
			continue
		}
		if err := reassessOne(binding, rec); err != nil {
			log.Printf("[risk] auto-recalc binding %d record %d: %v", binding.ID, rec.ID, err)
		}
	}
}

// reassessOne mirrors the RiskAdmin.Assess REST handler
// (server/compose/rest/risk_admin.go) but sources Evidence from a Compose
// record's own field values via the binding's FieldMap instead of a request
// body — the "recalc on every record change" half of AutoRecalc, the manual
// "recalc on demand" half being that REST endpoint.
func reassessOne(binding *riskengine.RiskSubjectBinding, rec *types.Record) error {
	model, ok := riskstore.GetModel(binding.ModelID)
	if !ok {
		return nil // binding points at a model that no longer exists; nothing to do
	}
	factors := riskstore.FactorsForModel(model)
	evidence := evidenceFromRecord(rec, binding, factors)

	assessment, err := riskengine.EvaluateWithOptions(model, factors, evidence, riskengine.EvalOptions{
		ControlFactorHandle: "controlEffectiveness",
	})
	if err != nil {
		return err
	}
	if attr, err := riskengine.Explain(model, factors, evidence, assessment); err == nil {
		assessment.Attribution = attr
		assessment.Explanation = riskengine.NarrateAttribution(assessment, attr, factors)
	}
	assessment.BindingID = binding.ID
	assessment.SubjectRecordID = rec.ID

	riskstore.SaveAssessment(assessment)
	return nil
}

// evidenceFromRecord reads one Evidence value per factor referenced by
// model's nodes, via binding.FieldMap (RiskFactorDef.Handle -> record field
// name). A "states" factor's field value is used as-is (expected to already
// hold a state code); numeric/bool factors are parsed from the field's
// string value. Factors with no FieldMap entry (manual/expr/external
// sources) are left out of Evidence — Evaluate reports which ones were
// required but missing.
func evidenceFromRecord(rec *types.Record, binding *riskengine.RiskSubjectBinding, factors map[uint64]*riskengine.RiskFactorDef) riskengine.Evidence {
	evidence := riskengine.Evidence{}
	for _, fd := range factors {
		fieldName, ok := binding.FieldMap[fd.Handle]
		if !ok {
			continue
		}
		raw := rec.Values.Get(fieldName, 0)
		if raw == nil {
			continue
		}
		if fd.Scale == riskengine.ScaleStates {
			evidence[fd.Handle] = raw.Value
			continue
		}
		if f, err := strconv.ParseFloat(raw.Value, 64); err == nil {
			evidence[fd.Handle] = f
		}
	}
	return evidence
}
