package service

import (
	"context"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestRuleChainRunPersistence exercises the Phase 2 durable run-log path
// end to end against a real (in-memory sqlite) schema: it makes sure
// compose_rule_chain_run is auto-provisioned by store.Upgrade, that
// RuleChainDBPersist.SaveRun (the rulesgo.RunLogPersistence implementation
// EngineWithPersistence.RunWithLog calls) writes a row, and that it can be
// read back via SearchRuleChainRuns/LookupRuleChainRun.
func TestRuleChainRunPersistence(t *testing.T) {
	req := require.New(t)
	ctx := context.Background()

	s, err := sqlite.ConnectInMemoryWithDebug(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))

	prevStore := DefaultStore
	DefaultStore = s
	defer func() { DefaultStore = prevStore }()

	// register the owning chain first so SaveRun can resolve its namespace
	chain := &rulesgo.Chain{ID: "test_chain", Name: "Test Chain", NamespaceID: 42}
	persist := NewRuleChainPersistence()
	req.NoError(persist.SaveChain(ctx, chain))

	saver, ok := persist.(rulesgo.RunLogPersistence)
	req.True(ok, "RuleChainDBPersist must implement rulesgo.RunLogPersistence")

	started := time.Now().Add(-time.Second)
	finished := started.Add(250 * time.Millisecond)
	rec := rulesgo.ExecRecord{
		ChainID: "test_chain",
		Input:   map[string]interface{}{"x": 1},
		Result: &rulesgo.ChainResult{
			ChainID: "test_chain",
			Success: true,
			Output:  map[string]interface{}{"y": 2},
			Nodes: []rulesgo.NodeResult{
				{NodeID: "n1", Type: "condition", Output: map[string]interface{}{"ok": true}},
			},
		},
		StartedAt:   started,
		FinishedAt:  finished,
		Duration:    finished.Sub(started).String(),
		TriggerType: "manual-test",
	}

	req.NoError(saver.SaveRun(ctx, rec))

	runs, _, err := SearchRuleChainRuns(ctx, types.RuleChainRunFilter{ChainID: "test_chain"})
	req.NoError(err)
	req.Len(runs, 1)
	req.True(runs[0].Success)
	req.Equal("manual-test", runs[0].TriggerType)
	req.EqualValues(42, runs[0].NamespaceID)
	req.EqualValues(250, runs[0].DurationMs)
	req.NotNil(runs[0].FinishedAt)

	got, err := LookupRuleChainRun(ctx, runs[0].ID)
	req.NoError(err)
	req.Contains(string(got.Nodes), "n1")
	req.Contains(string(got.Output), "\"y\":2")

	// a failed run with no matching chainID filter should not show up
	otherRuns, _, err := SearchRuleChainRuns(ctx, types.RuleChainRunFilter{ChainID: "does_not_exist"})
	req.NoError(err)
	req.Len(otherRuns, 0)
}
