package agent_test

import (
	"testing"

	"github.com/madnikulin50/lowcode/agents/invest/agent"
)

func TestComputeCriticalPath_Linear(t *testing.T) {
	acts := []agent.Activity{
		{ID: "A", DurationDays: 5},
		{ID: "B", DurationDays: 3, Predecessors: []string{"A"}},
		{ID: "C", DurationDays: 2, Predecessors: []string{"B"}},
	}
	got := agent.ComputeCriticalPath(acts)
	by := map[string]agent.Activity{}
	for _, a := range got {
		by[a.ID] = a
	}
	if !by["A"].Critical || !by["B"].Critical || !by["C"].Critical {
		t.Fatalf("all should be critical: %+v", got)
	}
	if by["C"].EF != 10 {
		t.Fatalf("project end EF=%v want 10", by["C"].EF)
	}
}

func TestComputeCriticalPath_Float(t *testing.T) {
	// A(5) -> C(2)
	// B(1) -> C(2)   B has float
	acts := []agent.Activity{
		{ID: "A", DurationDays: 5},
		{ID: "B", DurationDays: 1},
		{ID: "C", DurationDays: 2, Predecessors: []string{"A", "B"}},
	}
	got := agent.ComputeCriticalPath(acts)
	by := map[string]agent.Activity{}
	for _, a := range got {
		by[a.ID] = a
	}
	if !by["A"].Critical || !by["C"].Critical {
		t.Fatalf("A and C should be critical: %+v", got)
	}
	if by["B"].Critical {
		t.Fatalf("B should have float: %+v", by["B"])
	}
	if by["B"].Float != 4 {
		t.Fatalf("B float=%v want 4", by["B"].Float)
	}
}
