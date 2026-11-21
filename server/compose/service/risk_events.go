package service

import (
	"context"
	"strconv"

	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
	"github.com/madnikulin50/lowcode/server/pkg/riskstore"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
)

var riskLevelRank = map[riskengine.Severity]int{
	riskengine.SeverityLow:      1,
	riskengine.SeverityMedium:   2,
	riskengine.SeverityHigh:     3,
	riskengine.SeverityCritical: 4,
}

// SaveRiskAssessment persists an assessment and tells workflows about it:
// always as onAssessed, and additionally as onEscalated when the level rose
// compared with the subject's previous assessment (a first assessment counts
// against "low"). This is the one place both the manual Assess endpoint and
// auto-recalc go through.
func SaveRiskAssessment(ctx context.Context, binding *riskengine.RiskSubjectBinding, a *riskengine.RiskAssessment) {
	previous := riskengine.Severity("")
	if history := riskstore.AssessmentsForBinding(binding.ID, a.SubjectRecordID); len(history) > 0 {
		previous = history[0].Level // newest first, before this one is saved
	}

	riskstore.SaveAssessment(a)
	publishRiskAssessment(ctx, binding, a, previous)
}

func publishRiskAssessment(ctx context.Context, binding *riskengine.RiskSubjectBinding, a *riskengine.RiskAssessment, previous riskengine.Severity) {
	props := map[string]map[string]interface{}{
		wfevent.PropAssessment: {
			"assessmentID":    strconv.FormatUint(a.ID, 10),
			"bindingID":       strconv.FormatUint(a.BindingID, 10),
			"subjectRecordID": strconv.FormatUint(a.SubjectRecordID, 10),
			"modelID":         strconv.FormatUint(a.ModelID, 10),
			"level":           string(a.Level),
			"previousLevel":   string(previous),
			"inherentScore":   a.InherentScore,
			"residualScore":   a.ResidualScore,
			"explanation":     a.Explanation,
		},
		wfevent.PropBinding: {
			"bindingID":   strconv.FormatUint(binding.ID, 10),
			"moduleID":    strconv.FormatUint(binding.ModuleID, 10),
			"namespaceID": strconv.FormatUint(binding.NamespaceID, 10),
			"modelID":     strconv.FormatUint(binding.ModelID, 10),
		},
	}

	wfevent.Emit(ctx, wfevent.New(wfevent.ResourceRiskAssessment, wfevent.OnAssessed, props))

	prevRank := riskLevelRank[previous]
	if prevRank == 0 {
		prevRank = riskLevelRank[riskengine.SeverityLow] // no history yet
	}
	if riskLevelRank[a.Level] > prevRank {
		wfevent.Emit(ctx, wfevent.New(wfevent.ResourceRiskAssessment, wfevent.OnEscalated, props))
	}
}
