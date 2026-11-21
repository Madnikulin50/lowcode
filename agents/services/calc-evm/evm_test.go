package calcevm

import (
	"math"
	"testing"
	"time"
)

func almost(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestRecalc_MidwayOnBudget(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	r := Recalc(1000, 50, 500, start, end, now)
	if !almost(r.PV, 500) {
		t.Fatalf("PV=%v want 500", r.PV)
	}
	if !almost(r.EV, 500) {
		t.Fatalf("EV=%v want 500", r.EV)
	}
	if !almost(r.SPI, 1) || !almost(r.CPI, 1) {
		t.Fatalf("SPI=%v CPI=%v", r.SPI, r.CPI)
	}
	if !almost(r.EAC, 1000) {
		t.Fatalf("EAC=%v want 1000", r.EAC)
	}
}

func TestRecalc_Overrun(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	r := Recalc(100, 100, 200, start, end, end)
	if !almost(r.CPI, 0.5) {
		t.Fatalf("CPI=%v want 0.5", r.CPI)
	}
	if !almost(r.EAC, 200) {
		t.Fatalf("EAC=%v want 200", r.EAC)
	}
}

func TestAggregate(t *testing.T) {
	items := []Item{
		{ProjectID: "1", BudgetPlanned: 100, PV: 50, EV: 40, ActualCost: 60},
		{ProjectID: "1", BudgetPlanned: 100, PV: 50, EV: 50, ActualCost: 40},
		{ProjectID: "2", BudgetPlanned: 999, PV: 999, EV: 999, ActualCost: 999},
	}
	r := Aggregate(items, "1")
	if !almost(r.BAC, 200) || !almost(r.PV, 100) || !almost(r.EV, 90) || !almost(r.AC, 100) {
		t.Fatalf("agg %+v", r)
	}
}

func TestMergeFacts(t *testing.T) {
	items := []Item{{ID: "7", PercentComplete: 10, ActualCost: 1}}
	facts := []Fact{{WBSID: "7", Percent: 25, Cost: 40}, {WBSID: "7", Percent: 20, Cost: 10}}
	got := MergeFacts(items, facts)
	if got[0].PercentComplete != 25 {
		t.Fatalf("percent=%v", got[0].PercentComplete)
	}
	if got[0].ActualCost != 50 {
		t.Fatalf("cost=%v", got[0].ActualCost)
	}
}

func TestRunAndDecode(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	in, err := InputFromParams("1", map[string]any{
		"now": now.Format(time.RFC3339),
		"items": []map[string]any{
			{"id": "7", "projectID": "1", "budgetPlanned": 1000, "percentComplete": 10, "actualCost": 1, "startPlanned": start, "endPlanned": end},
		},
		"facts": []map[string]any{{"wbsID": "7", "percent": 50, "cost": 500}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := Run(in)
	if out.WBS != 1 || !almost(out.SPI, 1) || !almost(out.CPI, 1) {
		t.Fatalf("%+v", out)
	}
	if out.Items[0].PercentComplete != 50 {
		t.Fatalf("merged percent %v", out.Items[0].PercentComplete)
	}
}

func TestCoerceComposeRecords(t *testing.T) {
	in, err := InputFromParams("", map[string]any{
		"projectID": "p1",
		"items": []map[string]any{
			{"recordID": "7", "project": "p1", "budget_planned": "1000", "percent_complete": "50", "actual_cost": "500", "start_planned": "2026-01-01", "end_planned": "2026-03-02"},
			{"recordID": "8", "project": "p2", "budget_planned": "999"},
		},
		"facts": []map[string]any{{"wbs": "7", "project": "p1", "percent": 50, "cost": 0}},
		"now":   "2026-01-31",
	})
	if err != nil {
		t.Fatal(err)
	}
	out := Run(in)
	if out.WBS != 1 || out.Items[0].ID != "7" {
		t.Fatalf("filtered %+v", out)
	}
	if len(out.Projects) != 1 || out.Projects[0].ProjectID != "p1" {
		t.Fatalf("projects %+v", out.Projects)
	}
}

func TestInputFromParamsRequiresItems(t *testing.T) {
	_, err := InputFromParams("", map[string]any{})
	if err == nil {
		t.Fatal("expected items required")
	}
}
