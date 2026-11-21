package tests

import (
	"context"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/stretchr/testify/require"
)

func testAutomationStates(t *testing.T, s store.AutomationStates) {
	var (
		ctx = context.Background()

		makeNew = func(sessionID uint64) *types.State {
			scope, _ := expr.NewVars(map[string]interface{}{"order": "A-1", "n": 7})
			at := now().Add(time.Hour)
			return &types.State{
				ID:         id.Next(),
				SessionID:  sessionID,
				WorkflowID: 11,
				Kind:       wfexec.SnapshotDelayed,
				ResumeAt:   &at,
				CreatedAt:  *now(),
				Snapshot: &types.StateSnapshot{
					Kind:      wfexec.SnapshotDelayed,
					StepID:    2,
					Scope:     scope,
					ResumeAt:  &at,
					Resumable: true,
				},
			}
		}
	)

	t.Run("create and lookup keeps the snapshot", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationStates(ctx))

		st := makeNew(1)
		req.NoError(s.CreateAutomationState(ctx, st))

		got, err := s.LookupAutomationStateByID(ctx, st.ID)
		req.NoError(err)
		req.Equal(st.SessionID, got.SessionID)
		req.Equal(wfexec.SnapshotDelayed, got.Kind)
		req.NotNil(got.ResumeAt)
		req.NotNil(got.Snapshot)
		req.Equal(uint64(2), got.Snapshot.StepID)
		req.True(got.Snapshot.Resumable)

		// vars come back unresolved until matched against a registry
		req.NotNil(got.Snapshot.Scope)
		req.True(got.Snapshot.Scope.Has("order"))
	})

	t.Run("upsert replaces", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationStates(ctx))

		st := makeNew(1)
		req.NoError(s.UpsertAutomationState(ctx, st))
		st.Kind = wfexec.SnapshotPrompted
		req.NoError(s.UpsertAutomationState(ctx, st))

		set, _, err := s.SearchAutomationStates(ctx, types.StateFilter{})
		req.NoError(err)
		req.Len(set, 1)
		req.Equal(wfexec.SnapshotPrompted, set[0].Kind)
	})

	t.Run("search by session", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationStates(ctx))

		a1, a2, b := makeNew(100), makeNew(100), makeNew(200)
		req.NoError(s.CreateAutomationState(ctx, a1, a2, b))

		set, _, err := s.SearchAutomationStates(ctx, types.StateFilter{SessionID: []string{"100"}})
		req.NoError(err)
		req.Len(set, 2)

		all, _, err := s.SearchAutomationStates(ctx, types.StateFilter{})
		req.NoError(err)
		req.Len(all, 3)
	})

	t.Run("delete by ID", func(t *testing.T) {
		req := require.New(t)
		req.NoError(s.TruncateAutomationStates(ctx))

		st := makeNew(1)
		req.NoError(s.CreateAutomationState(ctx, st))
		req.NoError(s.DeleteAutomationStateByID(ctx, st.ID))

		set, _, err := s.SearchAutomationStates(ctx, types.StateFilter{})
		req.NoError(err)
		req.Empty(set)
	})
}
