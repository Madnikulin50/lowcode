package service

import (
	"encoding/json"
	"fmt"

	"context"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
	"github.com/madnikulin50/lowcode/server/pkg/riskstore"
)

// chatRiskToolDefs backs the "risk" toolkit (RegisterComposeToolKits,
// toolkits.go) — the risk-analyst agent's actual hands. Both tools call the
// same riskengine functions server/compose/rest/risk_admin.go's
// SuggestFactors/ExplainNarrative endpoints do, so the agent and the plain
// REST caller always agree on behavior; only the transport differs.
func chatRiskToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "risk_suggest_factors",
			Description: "Suggest a draft list of risk factors for a free-text domain description. Returns a JSON array the caller must review before saving — never applied automatically.",
			Params: []chat.ParamDef{
				{Name: "description", Type: "string", Required: true, Description: "Free-text description of the domain to assess risk for, e.g. 'vendor onboarding risk'"},
			},
			Handler: chatRiskSuggestFactors,
		},
		{
			Name:        "risk_explain",
			Description: "Run a published risk model against evidence and return its score, level and a SHAP-style attribution breakdown as JSON. Use this instead of guessing at risk causes yourself.",
			Params: []chat.ParamDef{
				{Name: "modelID", Type: "string", Required: true, Description: "Risk model ID"},
				{Name: "evidence", Type: "string", Required: true, Description: "JSON object of factor handle -> value, e.g. {\"shrinkPct\":5.8}"},
			},
			Handler: chatRiskExplain,
		},
	}
}

func chatRiskSuggestFactors(_ context.Context, params map[string]string) string {
	desc := params["description"]
	if desc == "" {
		return "description is required"
	}
	b, err := json.Marshal(riskengine.SuggestFactors(desc))
	if err != nil {
		return fmt.Sprintf("marshal error: %v", err)
	}
	return string(b)
}

func chatRiskExplain(_ context.Context, params map[string]string) string {
	modelID := parseUint64(params["modelID"])
	if modelID == 0 {
		return "modelID is required"
	}
	model, ok := riskstore.GetModel(modelID)
	if !ok {
		return fmt.Sprintf("model %d not found", modelID)
	}
	var evidence riskengine.Evidence
	if err := json.Unmarshal([]byte(params["evidence"]), &evidence); err != nil {
		return fmt.Sprintf("invalid evidence JSON: %v", err)
	}

	factors := riskstore.FactorsForModel(model)
	assessment, err := riskengine.EvaluateWithOptions(model, factors, evidence, riskengine.EvalOptions{
		ControlFactorHandle: "controlEffectiveness",
	})
	if err != nil {
		return fmt.Sprintf("evaluation error: %v", err)
	}
	attr, err := riskengine.Explain(model, factors, evidence, assessment)
	if err != nil {
		return fmt.Sprintf("attribution error: %v", err)
	}
	assessment.Attribution = attr
	assessment.Explanation = riskengine.NarrateAttribution(assessment, attr, factors)

	b, err := json.Marshal(assessment)
	if err != nil {
		return fmt.Sprintf("marshal error: %v", err)
	}
	return string(b)
}
