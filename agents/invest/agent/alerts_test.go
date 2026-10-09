package agent_test

import (
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/invest/agent"
)

func TestCollectAlerts(t *testing.T) {
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
