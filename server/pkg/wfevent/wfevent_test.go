package wfevent

import (
	"context"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/stretchr/testify/require"
)

func finding(severity string) *Event {
	return New(ResourceAnomalyFinding, OnCreate, map[string]map[string]interface{}{
		PropFinding: {"severity": severity, "moduleID": "42", "score": 3.5},
		PropRule:    {"detector": "zscore"},
	})
}

func TestEvent_ConstraintMatching(t *testing.T) {
	eq := func(name, value string) eventbus.ConstraintMatcher {
		return eventbus.MustMakeConstraint(name, "=", value)
	}

	ev := finding("high")
	require.True(t, ev.Match(eq("finding.severity", "high")))
	require.False(t, ev.Match(eq("finding.severity", "low")))
	require.True(t, ev.Match(eq("finding.moduleID", "42")))
	require.True(t, ev.Match(eq("rule.detector", "zscore")))
	require.True(t, ev.Match(eq("finding.score", "3.5")), "non-string values are compared as text")

	require.False(t, ev.Match(eq("finding.nope", "x")), "unknown field")
	require.False(t, ev.Match(eq("nope.severity", "high")), "unknown prop")
	require.False(t, ev.Match(eq("severity", "high")), "name without a prop")
}

func TestEvent_EncodeVars(t *testing.T) {
	vars, err := finding("high").EncodeVars()
	require.NoError(t, err)
	require.True(t, vars.Has("finding"))
	require.True(t, vars.Has("rule"))
}

func TestEmit_DeliversToMatchingSubscribersOnly(t *testing.T) {
	got := make(chan eventbus.Event, 4)
	handler := func(_ context.Context, ev eventbus.Event) error { got <- ev; return nil }

	high := eventbus.Service().Register(handler,
		eventbus.For(ResourceAnomalyFinding),
		eventbus.On(OnCreate),
		eventbus.Constraint(eventbus.MustMakeConstraint("finding.severity", "=", "high")),
	)
	defer eventbus.Service().Unregister(high)

	Emit(context.Background(), finding("low"))
	Emit(context.Background(), finding("high"))

	select {
	case ev := <-got:
		require.Equal(t, ResourceAnomalyFinding, ev.ResourceType())
		require.Equal(t, OnCreate, ev.EventType())
	case <-time.After(3 * time.Second):
		t.Fatal("subscriber never received the high-severity finding")
	}

	select {
	case <-got:
		t.Fatal("the low-severity finding must not reach a severity=high subscriber")
	case <-time.After(150 * time.Millisecond):
	}
}

func TestDefinitions(t *testing.T) {
	defs := Definitions()
	require.Len(t, defs, 4)
	require.Len(t, RequireRunAs(), 4)

	seen := map[string]bool{}
	for _, d := range defs {
		key := d.ResourceType + "/" + d.EventType
		require.False(t, seen[key], "duplicate %s", key)
		seen[key] = true
		require.NotEmpty(t, d.Properties)
		require.NotEmpty(t, d.Constraints)
	}
}

// The point of the event is that a workflow can read it: the encoded variables
// must be addressable with ordinary workflow expressions.
func TestEncodeVars_ReadableFromWorkflowExpressions(t *testing.T) {
	vars, err := New(ResourceAnomalyFinding, OnCreate, map[string]map[string]interface{}{
		PropFinding: {
			"severity":    "high",
			"score":       3.5,
			"explanation": map[string]interface{}{"detector": "zscore"},
		},
	}).EncodeVars()
	require.NoError(t, err)

	eval := func(src string) interface{} {
		t.Helper()
		e, err := expr.NewParser().Parse(src)
		require.NoError(t, err, src)
		out, err := e.Eval(context.Background(), vars)
		require.NoError(t, err, src)
		return out
	}

	require.Equal(t, "high", eval("finding.severity"))
	require.Equal(t, "zscore", eval("finding.explanation.detector"))
	require.Equal(t, true, eval("finding.score > 3"))
}
