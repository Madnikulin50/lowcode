package agent_test

import (
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/invest/agent"
)

func TestDocumentRefPrefersDocumentID(t *testing.T) {
	req := agent.JobRequest{RecordID: "111", DocumentID: "222", ProjectID: "111"}
	req.Normalize()
	if req.DocumentRef() != "222" {
		t.Fatalf("DocumentRef=%q", req.DocumentRef())
	}
	req = agent.JobRequest{RecordID: "333"}
	req.Normalize()
	if req.DocumentRef() != "333" {
		t.Fatalf("fallback DocumentRef=%q", req.DocumentRef())
	}
}

func TestBuildApprovalSteps_AuthorThenPMOInvestor(t *testing.T) {
	members := []agent.Member{
		{User: "u-pmo", Role: "pmo"},
		{User: "u-inv", Role: "investor"},
		{User: "u-bank", Role: "bank"},
	}
	got := agent.BuildApprovalSteps("u-author", members, false)
	if len(got) != 3 {
		t.Fatalf("len=%d want 3: %+v", len(got), got)
	}
	if got[0].Approver != "u-author" || got[1].Approver != "u-pmo" || got[2].Approver != "u-inv" {
		t.Fatalf("order %+v", got)
	}
}

func TestBuildApprovalSteps_EmptyWithoutPeople(t *testing.T) {
	got := agent.BuildApprovalSteps("", nil, false)
	if len(got) != 0 {
		t.Fatalf("len=%d want 0", len(got))
	}
}

func TestBuildApprovalSteps_DedupAuthorIsPMO(t *testing.T) {
	members := []agent.Member{{User: "same", Role: "pmo"}, {User: "inv", Role: "investor"}}
	got := agent.BuildApprovalSteps("same", members, true)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(got), got)
	}
}

func TestSimulateImpact(t *testing.T) {
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	before, after, endAfter := agent.SimulateImpact(1000, 800, 200, end, 10)
	if before != 1000 || after != 1200 {
		t.Fatalf("before=%v after=%v", before, after)
	}
	want := end.Add(10 * 24 * time.Hour)
	if !endAfter.Equal(want) {
		t.Fatalf("endAfter=%v want %v", endAfter, want)
	}
	before, after, _ = agent.SimulateImpact(0, 500, 50, time.Time{}, 0)
	if before != 500 || after != 550 {
		t.Fatalf("fallback EAC before=%v after=%v", before, after)
	}
}

func TestApplyBudgetDelta(t *testing.T) {
	if agent.ApplyBudgetDelta(100, 20) != 120 {
		t.Fatal("add")
	}
	if agent.ApplyBudgetDelta(10, -50) != 0 {
		t.Fatal("floor at 0")
	}
}

func TestRiskScore(t *testing.T) {
	if agent.RiskScore("high", "critical") != 12 {
		t.Fatalf("got %v", agent.RiskScore("high", "critical"))
	}
}

func TestCollectReserveAndRFC(t *testing.T) {
	now := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	lines := []agent.BudgetLine{
		{Article: "СМР", Project: "1", Planned: 100, Actual: 90, Reserve: 0},
		{Article: "OK", Project: "1", Planned: 10, Actual: 1, Reserve: 5},
	}
	got := agent.CollectReserveAlerts(lines)
	if len(got) != 1 || got[0].Kind != "reserve_exhausted" {
		t.Fatalf("%+v", got)
	}
	rfcs := []agent.RFC{
		{Title: "Late", Status: "in_review", Project: "1", EndAfter: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Draft", Status: "draft", EndAfter: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	}
	rfcAlerts := agent.CollectRFCOverdue(rfcs, now)
	if len(rfcAlerts) != 1 {
		t.Fatalf("%+v", rfcAlerts)
	}
}

func TestCollectAlerts_StillFindsOverdueAndCPI(t *testing.T) {
	now := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	docs := []agent.Document{
		{Title: "Договор", Status: "in_review", DueDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), Project: "1"},
		{Title: "Смета", Status: "approved", DueDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)},
	}
	items := []agent.WBSItem{
		{ID: 9, Name: "Фундамент", Code: "1.1", ProjectID: "1", CPI: 0.7},
		{ID: 10, Name: "OK", Code: "1.2", CPI: 1.1},
	}
	got := agent.CollectAlerts(docs, items, now, 0.9)
	if len(got) != 2 {
		t.Fatalf("got %d alerts: %+v", len(got), got)
	}
}
