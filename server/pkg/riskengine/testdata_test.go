package riskengine

import (
	"encoding/json"
	"os"
	"testing"
)

// TestTestdataFixtures_MatchWireFormat loads the example JSON payloads under
// testdata/ — the same shape the REST admin endpoints
// (server/compose/rest/risk_admin.go) accept and return — and evaluates them
// with the same evidence used in the hand-built fixtures above. It exists to
// pin the wire format (a struct tag typo here would fail this test) and to
// double as a runnable example for the frontend.
func loadFactors(t *testing.T, path string) map[uint64]*RiskFactorDef {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var list []*RiskFactorDef
	if err := json.Unmarshal(b, &list); err != nil {
		t.Fatal(err)
	}
	out := make(map[uint64]*RiskFactorDef, len(list))
	for _, f := range list {
		out[f.ID] = f
	}
	return out
}

func loadModel(t *testing.T, path string) *RiskModel {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m RiskModel
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return &m
}

func TestTestdataFixtures_StoreRiskWeighted(t *testing.T) {
	factors := loadFactors(t, "testdata/store_risk_factors.json")
	model := loadModel(t, "testdata/store_risk_model.json")

	if model.Strategy != StrategyWeighted {
		t.Fatalf("strategy = %v, want weighted", model.Strategy)
	}
	evidence := Evidence{
		"shrinkPct":      5.8,
		"incidents90d":   8.0,
		"daysSinceAudit": 210.0,
		"revenueImpact":  5.0,
	}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	approxEqual(t, "InherentScore", a.InherentScore, 63.93, 0.01)
}

func TestTestdataFixtures_AuditBayes(t *testing.T) {
	factors := loadFactors(t, "testdata/audit_bayes_factors.json")
	model := loadModel(t, "testdata/audit_bayes_model.json")

	if model.Strategy != StrategyBayes {
		t.Fatalf("strategy = %v, want bayes", model.Strategy)
	}
	evidence := Evidence{"daysSinceAudit": "overdue", "staffTurnover": "exodus"}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	approxEqual(t, "InherentScore", a.InherentScore, 77, 0.01)

	attr, err := Explain(model, factors, evidence, a)
	if err != nil {
		t.Fatal(err)
	}
	var sum float64
	for _, c := range attr.Contributions {
		sum += c.Value
	}
	approxEqual(t, "baseline+Σcontrib", attr.Baseline+sum, a.InherentScore, 0.01)
}
