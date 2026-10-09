package rulesgo

import (
	"context"
	"encoding/json"
	"testing"
)

func TestScoreMatrixProduct(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	eng.RegisterChain(&Chain{
		ID: "m", EntryNode: "n1",
		Nodes: []ChainNode{{ID: "n1", Type: "score.matrix", Config: json.RawMessage(`{}`)}},
	})
	res, err := eng.Run(context.Background(), "m", map[string]interface{}{
		"likelihood": 4, "impact": 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("failed: %s", res.Error)
	}
	if got := res.Output["score"]; got != float64(20) {
		t.Fatalf("score want 20 got %v", got)
	}
}

func TestScoreWeightedStoreRisk(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	cfg := `{
		"factors":[
			{"field":"shrinkPct","weight":0.35,"max":10},
			{"field":"incidents90d","weight":0.25,"max":20},
			{"field":"daysSinceAudit","weight":0.15,"max":365},
			{"field":"revenueImpact","weight":0.25,"max":5}
		]
	}`
	eng.RegisterChain(&Chain{
		ID: "w", EntryNode: "n1",
		Nodes: []ChainNode{{ID: "n1", Type: "score.weighted", Config: json.RawMessage(cfg)}},
	})
	res, err := eng.Run(context.Background(), "w", map[string]interface{}{
		"shrinkPct": 5.8, "incidents90d": 8, "daysSinceAudit": 210, "revenueImpact": 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	score, _ := res.Output["score"].(float64)
	if score < 60 || score > 70 {
		t.Fatalf("expected score ~64, got %v", score)
	}
}

func TestRiskBandResidualCritical(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	eng.RegisterChain(&Chain{
		ID: "b", EntryNode: "n1",
		Nodes: []ChainNode{{
			ID: "n1", Type: "risk.band",
			Config: json.RawMessage(`{"controlField":"controlEffectiveness","bands":[
				{"name":"low","max":25},{"name":"medium","max":50},
				{"name":"high","max":70},{"name":"critical","max":100}
			]}`),
		}},
	})
	res, err := eng.Run(context.Background(), "b", map[string]interface{}{
		"score": 80, "controlEffectiveness": 0.05,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Output["level"] != "critical" {
		t.Fatalf("want critical got %v (residual=%v)", res.Output["level"], res.Output["residualScore"])
	}
	if res.Output["is_critical"] != "true" {
		t.Fatalf("is_critical want true got %v", res.Output["is_critical"])
	}
}

func TestDemoStoreRiskChain(t *testing.T) {
	reg := DefaultRegistry(nil)
	eng := NewEngine(reg)
	chain := DemoStoreRiskChain()
	eng.RegisterChain(chain)

	for _, store := range DemoStores() {
		res, err := eng.Run(context.Background(), chain.ID, store.Input())
		if err != nil {
			t.Fatalf("%s: %v", store.Name, err)
		}
		if !res.Success {
			t.Fatalf("%s failed: %s", store.Name, res.Error)
		}
		level, _ := res.Output["level"].(string)
		score, _ := res.Output["score"].(float64)
		residual, _ := res.Output["residualScore"].(float64)
		t.Logf("%-28s score=%.1f residual=%.1f level=%s", store.Name, score, residual, level)
		if level == "" {
			t.Fatalf("%s: empty level", store.Name)
		}
	}
}
