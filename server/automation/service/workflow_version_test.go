package service

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWorkflowVersions(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
	)

	s, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))
	req.NoError(s.TruncateAutomationWorkflowVersions(ctx))

	wf := &types.Workflow{
		ID: nextID(), Handle: "approve", Enabled: true, RunAs: 7, OwnedBy: 8,
		Meta:  &types.WorkflowMeta{Name: "Approve"},
		Steps: types.WorkflowStepSet{{ID: 1, Kind: types.WorkflowStepKindExpressions}},
	}

	list := func() types.WorkflowVersionSet {
		set, _, err := store.SearchAutomationWorkflowVersions(ctx, s, types.WorkflowVersionFilter{WorkflowID: wf.ID})
		req.NoError(err)
		return set
	}

	req.NoError(saveWorkflowVersion(ctx, s, wf, ""))
	req.Len(list(), 1)

	t.Run("identical definition is not repeated", func(t *testing.T) {
		wf.RunAs = 9 // not part of the snapshot
		req.NoError(saveWorkflowVersion(ctx, s, wf, ""))
		req.Len(list(), 1)
	})

	t.Run("a change makes the next version", func(t *testing.T) {
		wf.Meta = &types.WorkflowMeta{Name: "Approve v2"}
		req.NoError(saveWorkflowVersion(ctx, s, wf, "renamed"))
		set := list()
		req.Len(set, 2)
		for _, v := range set {
			if v.Version == 2 {
				req.Equal("renamed", v.Note)
				req.Contains(v.Snapshot, "Approve v2")
			}
		}
	})

	t.Run("the oldest go over the cap", func(t *testing.T) {
		for i := 0; i < workflowVersionsKept+3; i++ {
			wf.KeepSessions = i + 1
			req.NoError(saveWorkflowVersion(ctx, s, wf, ""))
		}
		set := list()
		req.Len(set, workflowVersionsKept)
		min, max := set[0].Version, set[0].Version
		for _, v := range set {
			if v.Version < min {
				min = v.Version
			}
			if v.Version > max {
				max = v.Version
			}
		}
		req.Equal(max-workflowVersionsKept+1, min)
	})

	t.Run("restore leaves who and whether alone", func(t *testing.T) {
		cur := &types.Workflow{ID: wf.ID, Enabled: false, RunAs: 42, OwnedBy: 43, Meta: &types.WorkflowMeta{Name: "now"}}
		upd := restoredWorkflow(cur, workflowSnapshot{
			Handle: "old", Enabled: true, KeepSessions: 5,
			Meta:  &types.WorkflowMeta{Name: "then"},
			Steps: wf.Steps,
		})

		req.Equal("old", upd.Handle)
		req.Equal("then", upd.Meta.Name)
		req.Equal(5, upd.KeepSessions)
		req.False(upd.Enabled, "an old version must not switch the workflow on")
		req.Equal(uint64(42), upd.RunAs)
		req.Equal(uint64(43), upd.OwnedBy)
		req.Nil(upd.UpdatedAt, "no stale check on restore")
	})
}
