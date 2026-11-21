package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
)

// Ready-made AI workflows. They are starting points: each shows one pattern
// end to end (an event, an ai* step, a decision), ending in a log step the
// author replaces with the real action - a record update, an email, an
// aiRunCalls for approved agent actions.
//
// Installing creates the workflow and its trigger *disabled*: a template must
// be reviewed before it starts reacting to production events. It runs as the
// installing user, which background events require (see
// validateWorkflowTriggers).

type (
	AITemplateInfo struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Trigger     string `json:"trigger"`
	}

	aiTemplate struct {
		info  AITemplateInfo
		build func(installer uint64) (*types.Workflow, *types.Trigger)
	}
)

var aiTemplates = map[string]aiTemplate{
	"anomaly-explain": {
		info: AITemplateInfo{
			Key:         "anomaly-explain",
			Name:        "AI: explain anomaly finding",
			Description: "When a high-severity anomaly is found, an AI step explains it in plain language and what to check first. Replace the final log step with a notification or a record update.",
			Trigger:     wfevent.ResourceAnomalyFinding + " / " + wfevent.OnCreate + " (finding.severity = high)",
		},
		build: buildAnomalyExplain,
	},
	"risk-escalation-review": {
		info: AITemplateInfo{
			Key:         "risk-escalation-review",
			Name:        "AI: review risk escalation",
			Description: "When a risk level rises, an AI step proposes what to do and asks you to approve or decline; the two branches are placeholders for the real follow-up (for example aiRunCalls to carry out approved agent actions).",
			Trigger:     wfevent.ResourceRiskAssessment + " / " + wfevent.OnEscalated,
		},
		build: buildRiskEscalationReview,
	},
}

func AITemplates() []AITemplateInfo {
	out := make([]AITemplateInfo, 0, len(aiTemplates))
	for _, t := range aiTemplates {
		out = append(out, t.info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// InstallAITemplate creates the template's workflow and trigger, both
// disabled, and returns the workflow.
func InstallAITemplate(ctx context.Context, key string) (*types.Workflow, error) {
	t, ok := aiTemplates[key]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", key)
	}
	if DefaultWorkflow == nil || DefaultTrigger == nil {
		return nil, fmt.Errorf("workflow service not available")
	}

	ident := auth.GetIdentityFromContext(ctx)
	if ident == nil {
		return nil, fmt.Errorf("a user is required to install a template")
	}

	wf, trigger := t.build(ident.Identity())

	// a re-install must not clash with an earlier copy
	if existing, _, err := DefaultWorkflow.Search(ctx, types.WorkflowFilter{Handle: wf.Handle}); err == nil && len(existing) > 0 {
		return nil, fmt.Errorf("workflow %q is already installed", wf.Handle)
	}

	created, err := DefaultWorkflow.Create(ctx, wf)
	if err != nil {
		return nil, err
	}

	trigger.WorkflowID = created.ID
	if _, err = DefaultTrigger.Create(ctx, trigger); err != nil {
		return created, fmt.Errorf("workflow created but its trigger was not: %w", err)
	}

	return created, nil
}

// str is a String-typed argument expression; function arguments must declare
// their type so the converter can check them against the function's signature.
func str(target, expression string) *types.Expr {
	return &types.Expr{Target: target, Expr: expression, Type: "String"}
}

// strVal is a constant String argument.
func strVal(target, value string) *types.Expr {
	return &types.Expr{Target: target, Value: value, Type: "String"}
}

func fn(id uint64, ref, name string, args, results []*types.Expr) *types.WorkflowStep {
	return &types.WorkflowStep{
		ID:        id,
		Kind:      types.WorkflowStepKindFunction,
		Ref:       ref,
		Arguments: args,
		Results:   results,
		Meta:      types.WorkflowStepMeta{Name: name},
	}
}

func buildAnomalyExplain(installer uint64) (*types.Workflow, *types.Trigger) {
	wf := &types.Workflow{
		Handle: "ai_anomaly_explain",
		Meta: &types.WorkflowMeta{
			Name:        "AI: explain anomaly finding",
			Description: "Explains a high-severity anomaly finding in plain language.",
		},
		RunAs: installer,
		Steps: types.WorkflowStepSet{
			fn(1, "aiAsk", "Explain finding",
				[]*types.Expr{
					str("prompt", `"A monitoring rule flagged this record as anomalous. In two or three sentences, explain in plain language why the value is unusual and what to check first. Do not invent facts that are not in the data.\n\nData: " + toPlainJSON(finding)`),
				},
				[]*types.Expr{{Target: "explanation", Source: "text"}}),
			fn(2, "logInfo", "Report",
				[]*types.Expr{
					str("message", `format("Anomaly on record %v (%v): %v", finding.recordID, finding.severity, explanation)`),
				}, nil),
		},
		Paths: types.WorkflowPathSet{{ParentID: 1, ChildID: 2}},
	}

	trigger := &types.Trigger{
		ResourceType: wfevent.ResourceAnomalyFinding,
		EventType:    wfevent.OnCreate,
		Constraints: types.TriggerConstraintSet{
			{Name: "finding.severity", Op: "=", Values: []string{"high"}},
		},
	}
	return wf, trigger
}

func buildRiskEscalationReview(installer uint64) (*types.Workflow, *types.Trigger) {
	wf := &types.Workflow{
		Handle: "ai_risk_escalation_review",
		Meta: &types.WorkflowMeta{
			Name:        "AI: review risk escalation",
			Description: "Proposes a response to a risk escalation and asks for approval.",
		},
		RunAs: installer,
		Steps: types.WorkflowStepSet{
			fn(1, "aiAsk", "Propose response",
				[]*types.Expr{
					str("prompt", `"The risk level of a monitored subject rose from '" + assessment.previousLevel + "' to '" + assessment.level + "'. In three sentences at most, recommend what the owner should do next, based only on this assessment: " + toPlainJSON(assessment)`),
				},
				[]*types.Expr{{Target: "recommendation", Source: "text"}}),
			{
				ID:   2,
				Kind: types.WorkflowStepKindPrompt,
				Ref:  "choice",
				Arguments: []*types.Expr{
					{Target: "owner", Value: installer, Type: "ID"},
					strVal("title", "Risk escalation"),
					str("message", `format("Risk rose to %v (was %v). Proposed response: %v", assessment.level, assessment.previousLevel, recommendation)`),
					strVal("confirmButtonLabel", "Approve"),
					strVal("rejectButtonLabel", "Decline"),
				},
				Results: []*types.Expr{{Target: "approved", Source: "value"}},
				Meta:    types.WorkflowStepMeta{Name: "Ask for approval"},
			},
			{ID: 3, Kind: types.WorkflowStepKindGateway, Ref: "excl", Meta: types.WorkflowStepMeta{Name: "Decision"}},
			fn(4, "logInfo", "Approved - replace with the real action",
				[]*types.Expr{str("message", `"Risk escalation approved: " + recommendation`)}, nil),
			fn(5, "logInfo", "Declined",
				[]*types.Expr{strVal("message", "Risk escalation declined")}, nil),
		},
		Paths: types.WorkflowPathSet{
			{ParentID: 1, ChildID: 2},
			{ParentID: 2, ChildID: 3},
			{ParentID: 3, ChildID: 4, Expr: `approved == true`},
			{ParentID: 3, ChildID: 5, Expr: `approved != true`},
		},
	}

	trigger := &types.Trigger{
		ResourceType: wfevent.ResourceRiskAssessment,
		EventType:    wfevent.OnEscalated,
	}
	return wf, trigger
}
