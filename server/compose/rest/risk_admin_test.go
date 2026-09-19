package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// doRisk sends a request to a RiskAdmin handler, injecting chi URL params
// the same way the router would, and returns the unwrapped response body.
func doRisk(t *testing.T, handler http.HandlerFunc, method, body string, params map[string]string) map[string]interface{} {
	t.Helper()
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, "/", bytes.NewBufferString(body))
	} else {
		r = httptest.NewRequest(method, "/", nil)
	}
	r = withChiParams(r, params)
	w := httptest.NewRecorder()
	handler(w, r)
	return unwrapResponse(t, w.Body.Bytes())
}

// TestRiskAdmin_EndToEnd exercises the full REST surface added in
// server/compose/rest/risk_admin.go against the same "Москва · Авиапарк"
// numbers already verified against server/pkg/riskengine's own tests and
// the architecture note: factor + model CRUD, Test, ExplainNarrative,
// SuggestFactors, and the binding -> Assess -> ListAssessments ->
// PortfolioSummary pipeline (phase 4's manual-recalculation path).
func TestRiskAdmin_EndToEnd(t *testing.T) {
	req := require.New(t)
	admin := RiskAdmin{}

	// --- factor library ---
	factorSpecs := []struct {
		handle string
		max    float64
	}{
		{"shrinkPct", 10}, {"incidents90d", 20}, {"daysSinceAudit", 365}, {"revenueImpact", 5},
	}
	factorIDs := map[string]string{}
	for _, f := range factorSpecs {
		payload, _ := json.Marshal(map[string]interface{}{
			"namespaceID": "1", "handle": f.handle, "label": f.handle, "scale": "numeric", "max": f.max,
			"source": "field", "fieldName": f.handle,
		})
		resp := doRisk(t, admin.CreateFactor, http.MethodPost, string(payload), nil)
		req.Equal(true, resp["created"])
		factor := resp["factor"].(map[string]interface{})
		factorIDs[f.handle] = factor["factorID"].(string)
	}

	// --- model: same weighted graph as demo_store_risk.go ---
	modelPayload, _ := json.Marshal(map[string]interface{}{
		"namespaceID": "1",
		"name":        "Операционный риск магазина",
		"strategy":    "weighted",
		"status":      "published",
		"config":      map[string]interface{}{},
		"nodes": []map[string]interface{}{
			{"id": "shrink", "factorID": factorIDs["shrinkPct"], "role": "evidence"},
			{"id": "incidents", "factorID": factorIDs["incidents90d"], "role": "evidence"},
			{"id": "audit", "factorID": factorIDs["daysSinceAudit"], "role": "evidence"},
			{"id": "revenue", "factorID": factorIDs["revenueImpact"], "role": "evidence"},
			{"id": "out", "factorID": "0", "role": "output"},
		},
		"edges": []map[string]interface{}{
			{"from": "shrink", "to": "out", "weight": 0.35},
			{"from": "incidents", "to": "out", "weight": 0.25},
			{"from": "audit", "to": "out", "weight": 0.15},
			{"from": "revenue", "to": "out", "weight": 0.25},
		},
	})
	createResp := doRisk(t, admin.CreateModel, http.MethodPost, string(modelPayload), nil)
	req.Equal(true, createResp["created"])
	model := createResp["model"].(map[string]interface{})
	modelID := model["modelID"].(string)

	evidence := map[string]interface{}{
		"shrinkPct": 5.8, "incidents90d": 8.0, "daysSinceAudit": 210.0, "revenueImpact": 5.0,
	}

	// --- GetModel with resolved factors ---
	getReq := httptest.NewRequest(http.MethodGet, "/?withFactors=1", nil)
	getReq = withChiParams(getReq, map[string]string{"modelID": modelID})
	w := httptest.NewRecorder()
	admin.GetModel(w, getReq)
	getResp := unwrapResponse(t, w.Body.Bytes())
	req.NotNil(getResp["factors"])

	// --- Test (simulate, no persistence) ---
	testPayload, _ := json.Marshal(map[string]interface{}{"evidence": evidence, "withAttribution": true})
	testResp := doRisk(t, admin.Test, http.MethodPost, string(testPayload), map[string]string{"modelID": modelID})
	assessment := testResp["assessment"].(map[string]interface{})
	req.InDelta(63.93, assessment["inherentScore"].(float64), 0.01)

	// --- ExplainNarrative ---
	explainResp := doRisk(t, admin.ExplainNarrative, http.MethodPost, string(testPayload), map[string]string{"modelID": modelID})
	explained := explainResp["assessment"].(map[string]interface{})
	explanation, _ := explained["explanation"].(string)
	req.Contains(explanation, "Уровень риска")
	req.NotEmpty(explained["attribution"])

	// --- SuggestFactors ---
	suggestPayload, _ := json.Marshal(map[string]string{"description": "риск магазина: усадка и текучесть персонала"})
	suggestResp := doRisk(t, admin.SuggestFactors, http.MethodPost, string(suggestPayload), nil)
	suggested := suggestResp["factors"].([]interface{})
	req.NotEmpty(suggested)

	// --- binding + Assess (phase 4 manual recalculation) ---
	bindingPayload, _ := json.Marshal(map[string]interface{}{
		"namespaceID": "1", "modelID": modelID, "moduleID": "42", "autoRecalc": true,
	})
	bindingResp := doRisk(t, admin.CreateBinding, http.MethodPost, string(bindingPayload), nil)
	binding := bindingResp["binding"].(map[string]interface{})
	bindingID := binding["bindingID"].(string)

	assessPayload, _ := json.Marshal(map[string]interface{}{
		"subjectRecordID":     "101",
		"evidence":            evidence,
		"controlFactorHandle": "controlEffectiveness",
	})
	// controlEffectiveness isn't in evidence here, so control defaults to 0 — still a valid assessment.
	assessResp := doRisk(t, admin.Assess, http.MethodPost, string(assessPayload), map[string]string{"bindingID": bindingID})
	req.Equal(true, assessResp["created"])
	persisted := assessResp["assessment"].(map[string]interface{})
	req.Equal("101", persisted["subjectRecordID"])

	listReq := httptest.NewRequest(http.MethodGet, "/?latestPerSubject=1", nil)
	listReq = withChiParams(listReq, map[string]string{"bindingID": bindingID})
	w2 := httptest.NewRecorder()
	admin.ListAssessments(w2, listReq)
	listResp := unwrapResponse(t, w2.Body.Bytes())
	assessments := listResp["assessments"].([]interface{})
	req.Len(assessments, 1)

	// A second subject under the same binding, scored differently — this is
	// exactly what a record page's Risk block asks for: "the latest
	// assessment for *this* record", not "one row per every subject" (see
	// risk_admin.go's ListAssessments doc comment for the bug this pins
	// down: subjectRecordID used to be silently ignored whenever
	// latestPerSubject was also set).
	assessPayload2, _ := json.Marshal(map[string]interface{}{
		"subjectRecordID": "102",
		"evidence":        map[string]interface{}{"shrinkPct": 1.0, "incidents90d": 1.0, "daysSinceAudit": 5.0, "revenueImpact": 1.0},
	})
	doRisk(t, admin.Assess, http.MethodPost, string(assessPayload2), map[string]string{"bindingID": bindingID})

	perSubjectReq := httptest.NewRequest(http.MethodGet, "/?subjectRecordID=101&latestPerSubject=1", nil)
	perSubjectReq = withChiParams(perSubjectReq, map[string]string{"bindingID": bindingID})
	w2b := httptest.NewRecorder()
	admin.ListAssessments(w2b, perSubjectReq)
	perSubjectResp := unwrapResponse(t, w2b.Body.Bytes())
	perSubjectAssessments := perSubjectResp["assessments"].([]interface{})
	req.Len(perSubjectAssessments, 1, "subjectRecordID + latestPerSubject must return only that subject, not one row per subject")
	req.Equal("101", perSubjectAssessments[0].(map[string]interface{})["subjectRecordID"])

	// --- PortfolioSummary ---
	sumReq := httptest.NewRequest(http.MethodGet, "/", nil)
	sumReq = withChiParams(sumReq, map[string]string{"bindingID": bindingID})
	w3 := httptest.NewRecorder()
	admin.PortfolioSummary(w3, sumReq)
	sumResp := unwrapResponse(t, w3.Body.Bytes())
	req.EqualValues(2, sumResp["subjects"])
	factorsAgg := sumResp["factors"].([]interface{})
	req.NotEmpty(factorsAgg)

	// --- treatment ---
	treatmentPayload, _ := json.Marshal(map[string]interface{}{"description": "Провести внеплановый аудит"})
	treatResp := doRisk(t, admin.CreateTreatment, http.MethodPost, string(treatmentPayload), map[string]string{"bindingID": bindingID})
	req.Equal(true, treatResp["created"])
	treatment := treatResp["treatment"].(map[string]interface{})
	req.Equal("open", treatment["status"])
}
