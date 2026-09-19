package riskengine

import (
	"encoding/json"
	"math"
	"testing"
)

func approxEqual(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, want %v (±%v)", name, got, want, tol)
	}
}

// storeRiskFixture mirrors server/pkg/rulesgo/demo_store_risk.go's weighted
// config and the "Москва · Авиапарк" sample from DemoStores(), so the
// engine's output can be checked against the numbers already written up in
// the architecture note.
func storeRiskFixture() (*RiskModel, map[uint64]*RiskFactorDef, Evidence) {
	factors := map[uint64]*RiskFactorDef{
		1: {ID: 1, Handle: "shrinkPct", Scale: ScaleNumeric, Max: 10},
		2: {ID: 2, Handle: "incidents90d", Scale: ScaleNumeric, Max: 20},
		3: {ID: 3, Handle: "daysSinceAudit", Scale: ScaleNumeric, Max: 365},
		4: {ID: 4, Handle: "revenueImpact", Scale: ScaleNumeric, Max: 5},
	}
	model := &RiskModel{
		Strategy: StrategyWeighted,
		Config:   json.RawMessage(`{}`),
		Nodes: []RiskNode{
			{ID: "shrink", FactorID: 1, Role: RoleEvidence},
			{ID: "incidents", FactorID: 2, Role: RoleEvidence},
			{ID: "audit", FactorID: 3, Role: RoleEvidence},
			{ID: "revenue", FactorID: 4, Role: RoleEvidence},
			{ID: "out", Role: RoleOutput},
		},
		Edges: []RiskEdge{
			{From: "shrink", To: "out", Weight: 0.35},
			{From: "incidents", To: "out", Weight: 0.25},
			{From: "audit", To: "out", Weight: 0.15},
			{From: "revenue", To: "out", Weight: 0.25},
		},
	}
	evidence := Evidence{
		"shrinkPct":            5.8,
		"incidents90d":         8.0,
		"daysSinceAudit":       210.0,
		"revenueImpact":        5.0,
		"controlEffectiveness": 0.30,
	}
	return model, factors, evidence
}

func TestEvaluateWeighted_StoreRiskPilot(t *testing.T) {
	model, factors, evidence := storeRiskFixture()
	opts := EvalOptions{
		ControlFactorHandle: "controlEffectiveness",
		Bands: []Band{
			{Level: SeverityLow, Max: 20},
			{Level: SeverityMedium, Max: 40},
			{Level: SeverityHigh, Max: 60},
			{Level: SeverityCritical, Max: 100},
		},
	}
	a, err := EvaluateWithOptions(model, factors, evidence, opts)
	if err != nil {
		t.Fatal(err)
	}
	approxEqual(t, "InherentScore", a.InherentScore, 63.93, 0.01)
	approxEqual(t, "ResidualScore", a.ResidualScore, 44.75, 0.01)
	if a.Level != SeverityHigh {
		t.Errorf("Level = %v, want %v", a.Level, SeverityHigh)
	}
	if len(a.Observations) != 4 {
		t.Errorf("Observations = %d, want 4", len(a.Observations))
	}
}

func TestExplainWeighted_ReconcilesWithScore(t *testing.T) {
	model, factors, evidence := storeRiskFixture()
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	attr, err := Explain(model, factors, evidence, a)
	if err != nil {
		t.Fatal(err)
	}
	if attr.Method != MethodClosedForm {
		t.Errorf("Method = %v, want %v", attr.Method, MethodClosedForm)
	}

	want := map[string]float64{"shrink": 20.3, "incidents": 10.0, "audit": 8.63, "revenue": 25.0}
	var sum float64
	for _, c := range attr.Contributions {
		sum += c.Value
		if w, ok := want[c.NodeID]; ok {
			approxEqual(t, "contribution["+c.NodeID+"]", c.Value, w, 0.02)
		} else {
			t.Errorf("unexpected node %q in contributions", c.NodeID)
		}
	}
	// The defining property of an exact attribution: contributions plus
	// baseline reconstruct the score exactly (baseline is 0 here — no
	// RiskFactorDef.Baseline was set, so it defaults to the zero point).
	approxEqual(t, "baseline+Σcontrib", attr.Baseline+sum, a.InherentScore, 0.01)
}

func TestEvaluateMatrix(t *testing.T) {
	factors := map[uint64]*RiskFactorDef{
		1: {ID: 1, Handle: "likelihood", Scale: ScaleNumeric, Max: 5},
		2: {ID: 2, Handle: "impact", Scale: ScaleNumeric, Max: 5},
	}
	model := &RiskModel{
		Strategy: StrategyMatrix,
		Config:   json.RawMessage(`{"likelihoodNode":"l","impactNode":"i"}`),
		Nodes: []RiskNode{
			{ID: "l", FactorID: 1, Role: RoleEvidence},
			{ID: "i", FactorID: 2, Role: RoleEvidence},
		},
	}
	evidence := Evidence{"likelihood": 4.0, "impact": 5.0}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if a.InherentScore != 20 {
		t.Errorf("InherentScore = %v, want 20", a.InherentScore)
	}
}

func TestExplainMatrix_ExactTwoPlayerShapley(t *testing.T) {
	factors := map[uint64]*RiskFactorDef{
		1: {ID: 1, Handle: "likelihood", Scale: ScaleNumeric, Max: 5, Baseline: "1"},
		2: {ID: 2, Handle: "impact", Scale: ScaleNumeric, Max: 5, Baseline: "1"},
	}
	model := &RiskModel{
		Strategy: StrategyMatrix,
		Config:   json.RawMessage(`{"likelihoodNode":"l","impactNode":"i"}`),
		Nodes: []RiskNode{
			{ID: "l", FactorID: 1, Role: RoleEvidence},
			{ID: "i", FactorID: 2, Role: RoleEvidence},
		},
	}
	evidence := Evidence{"likelihood": 4.0, "impact": 5.0}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	attr, err := Explain(model, factors, evidence, a)
	if err != nil {
		t.Fatal(err)
	}
	if attr.Baseline != 1 { // 1x1 baseline cell
		t.Errorf("Baseline = %v, want 1", attr.Baseline)
	}
	var sum float64
	for _, c := range attr.Contributions {
		sum += c.Value
	}
	approxEqual(t, "baseline+Σcontrib", attr.Baseline+sum, a.InherentScore, 0.01)
}

func TestEvaluateBlend(t *testing.T) {
	model, factors, evidence := storeRiskFixture()
	weightedCfg, _ := json.Marshal(WeightedConfig{})

	blend := &RiskModel{
		Strategy: StrategyBlend,
		Nodes:    model.Nodes,
		Edges:    model.Edges,
		Config: mustJSON(BlendConfig{Components: []BlendComponent{
			{Strategy: StrategyWeighted, Weight: 1, Config: weightedCfg},
		}}),
	}
	a, err := Evaluate(blend, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	// A single-component blend must reproduce that component's own score.
	approxEqual(t, "blend score", a.InherentScore, 63.93, 0.01)
}

func mustJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
