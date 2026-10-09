package riskengine

import "testing"

// auditBayesFixture is the example walked through in the architecture note:
// two root factors with their own, unrelated state vocabularies —
// "давность аудита" (recent/aging/overdue) and "текучесть персонала"
// (stable/elevated/exodus) — feeding one output node "auditRisk"
// (low/high), each state carrying a Severity for UI purposes only.
func auditBayesFixture() (*RiskModel, map[uint64]*RiskFactorDef) {
	factors := map[uint64]*RiskFactorDef{
		1: {
			ID: 1, Handle: "daysSinceAudit", Scale: ScaleStates,
			States: []RiskFactorState{
				{Code: "recent", Label: "Недавно", Severity: SeverityLow, Order: 0},
				{Code: "aging", Label: "Давно", Severity: SeverityMedium, Order: 1},
				{Code: "overdue", Label: "Просрочено", Severity: SeverityHigh, Order: 2},
			},
		},
		2: {
			ID: 2, Handle: "staffTurnover", Scale: ScaleStates,
			States: []RiskFactorState{
				{Code: "stable", Label: "Стабильно", Severity: SeverityLow, Order: 0},
				{Code: "elevated", Label: "Повышенная", Severity: SeverityMedium, Order: 1},
				{Code: "exodus", Label: "Массовый отток", Severity: SeverityCritical, Order: 2},
			},
		},
		3: {
			ID: 3, Handle: "auditRisk", Scale: ScaleStates,
			States: []RiskFactorState{
				{Code: "low", Label: "Низкий", Severity: SeverityLow, Order: 0},
				{Code: "high", Label: "Высокий", Severity: SeverityHigh, Order: 1},
			},
		},
	}

	// Row order: parent0 (daysSinceAudit) slowest, parent1 (staffTurnover)
	// fastest — recent×{stable,elevated,exodus}, aging×{...}, overdue×{...}.
	cpt := &CPT{
		ParentStates: [][]string{
			{"recent", "aging", "overdue"},
			{"stable", "elevated", "exodus"},
		},
		NodeStates: []string{"low", "high"},
		Rows: [][]float64{
			{0.95, 0.05}, {0.85, 0.15}, {0.60, 0.40}, // recent
			{0.80, 0.20}, {0.60, 0.40}, {0.35, 0.65}, // aging
			{0.55, 0.45}, {0.30, 0.70}, {0.05, 0.95}, // overdue
		},
	}

	model := &RiskModel{
		Strategy: StrategyBayes,
		Config: mustJSON(BayesConfig{
			OutputNode:   "risk",
			ScoreByState: map[string]float64{"low": 20, "high": 80},
		}),
		Nodes: []RiskNode{
			{ID: "audit", FactorID: 1, Role: RoleEvidence},
			{ID: "turnover", FactorID: 2, Role: RoleEvidence},
			{ID: "risk", FactorID: 3, Role: RoleOutput, CPT: cpt},
		},
		Edges: []RiskEdge{
			{From: "audit", To: "risk"},
			{From: "turnover", To: "risk"},
		},
	}
	return model, factors
}

func TestBayesPosterior_FullyObservedMatchesCPTRowExactly(t *testing.T) {
	model, factors := auditBayesFixture()
	net, err := buildBayesNet(model, factors, nil)
	if err != nil {
		t.Fatal(err)
	}
	post, err := net.posterior("risk", map[string]string{"audit": "overdue", "turnover": "exodus"})
	if err != nil {
		t.Fatal(err)
	}
	approxEqual(t, "P(low)", post["low"], 0.05, 1e-9)
	approxEqual(t, "P(high)", post["high"], 0.95, 1e-9)
}

func TestBayesPosterior_MarginalizesUnobservedParentViaUniformPrior(t *testing.T) {
	model, factors := auditBayesFixture()
	net, err := buildBayesNet(model, factors, nil)
	if err != nil {
		t.Fatal(err)
	}
	// turnover unobserved -> averaged uniformly over its 3 states for
	// audit=recent: (0.05+0.15+0.40)/3 = 0.2
	post, err := net.posterior("risk", map[string]string{"audit": "recent"})
	if err != nil {
		t.Fatal(err)
	}
	approxEqual(t, "P(high | audit=recent)", post["high"], 0.2, 1e-9)
}

func TestEvaluateBayes(t *testing.T) {
	model, factors := auditBayesFixture()
	evidence := Evidence{"daysSinceAudit": "overdue", "staffTurnover": "exodus"}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	// posterior = [0.05, 0.95] -> score = 0.05*20 + 0.95*80 = 77
	approxEqual(t, "InherentScore", a.InherentScore, 77, 0.01)
}

func TestExplainBayes_ReconcilesWithScoreAndIsMonotonic(t *testing.T) {
	model, factors := auditBayesFixture()
	evidence := Evidence{"daysSinceAudit": "overdue", "staffTurnover": "exodus"}
	a, err := Evaluate(model, factors, evidence)
	if err != nil {
		t.Fatal(err)
	}
	attr, err := Explain(model, factors, evidence, a)
	if err != nil {
		t.Fatal(err)
	}
	if attr.Method != MethodExact {
		t.Errorf("Method = %v, want %v", attr.Method, MethodExact)
	}
	var sum float64
	for _, c := range attr.Contributions {
		sum += c.Value
	}
	approxEqual(t, "baseline+Σcontrib", attr.Baseline+sum, a.InherentScore, 0.01)

	// Both factors are at their worst observed state, so both should push
	// the score up (positive contribution) relative to the marginalized
	// baseline.
	for _, c := range attr.Contributions {
		if c.Value <= 0 {
			t.Errorf("contribution for %s = %v, want > 0 (both factors are worst-case here)", c.Handle, c.Value)
		}
	}
}

func TestEvaluateBayes_UnknownStateCodeErrors(t *testing.T) {
	model, factors := auditBayesFixture()
	evidence := Evidence{"daysSinceAudit": "nonsense", "staffTurnover": "exodus"}
	if _, err := Evaluate(model, factors, evidence); err == nil {
		t.Error("expected error for unknown state code, got nil")
	}
}

func TestExactShapley_AdditiveGameReturnsInputsThemselves(t *testing.T) {
	values := []float64{3, -2, 7, 1}
	n := len(values)
	v := func(mask int) float64 {
		var s float64
		for i, val := range values {
			if mask&(1<<uint(i)) != 0 {
				s += val
			}
		}
		return s
	}
	phi := exactShapley(n, v)
	for i, want := range values {
		approxEqual(t, "phi", phi[i], want, 1e-9)
	}
}

func TestSampledShapley_ApproximatesExactOnAdditiveGame(t *testing.T) {
	values := []float64{3, -2, 7, 1, 5}
	n := len(values)
	v := func(mask int) float64 {
		var s float64
		for i, val := range values {
			if mask&(1<<uint(i)) != 0 {
				s += val
			}
		}
		return s
	}
	phi := sampledShapley(n, v, 300)
	for i, want := range values {
		approxEqual(t, "phi", phi[i], want, 1e-9) // additive game: even one permutation is exact
	}
}
