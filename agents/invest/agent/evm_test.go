package agent_test

import (
	"math"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/invest/agent"
)

func almost(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestRecalcEVM_MidwayOnBudget(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)  // 60 days
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC) // 30 days in
	r := agent.RecalcEVM(1000, 50, 500, start, end, now)
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

func TestRecalcEVM_Overrun(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	now := end
	r := agent.RecalcEVM(100, 100, 200, start, end, now)
	if !almost(r.CPI, 0.5) {
		t.Fatalf("CPI=%v want 0.5", r.CPI)
	}
	if !almost(r.EAC, 200) {
		t.Fatalf("EAC=%v want 200", r.EAC)
	}
}

func TestAggregateProject(t *testing.T) {
	items := []agent.WBSItem{
		{ProjectID: "1", BudgetPlanned: 100, PV: 50, EV: 40, ActualCost: 60},
		{ProjectID: "1", BudgetPlanned: 100, PV: 50, EV: 50, ActualCost: 40},
		{ProjectID: "2", BudgetPlanned: 999, PV: 999, EV: 999, ActualCost: 999},
	}
	r := agent.AggregateProject(items, "1")
	if !almost(r.BAC, 200) || !almost(r.PV, 100) || !almost(r.EV, 90) || !almost(r.AC, 100) {
		t.Fatalf("agg %+v", r)
	}
}

func TestMergeFacts(t *testing.T) {
	items := []agent.WBSItem{{ID: 7, PercentComplete: 10, ActualCost: 1}}
	facts := []agent.ProgressFact{{WBSID: "7", Percent: 25, Cost: 40}, {WBSID: "7", Percent: 20, Cost: 10}}
	got := agent.MergeFacts(items, facts)
	if got[0].PercentComplete != 25 {
		t.Fatalf("percent=%v", got[0].PercentComplete)
	}
	if got[0].ActualCost != 50 {
		t.Fatalf("cost=%v", got[0].ActualCost)
	}
}
