package wfexec

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/stretchr/testify/require"
)

// snapshotGraph: start -> gate -> end, where gate suspends first (prompt or
// delay, depending on mode) and completes once resumed.
func snapshotGraph(mode string, until time.Time) (*Graph, *sesTestStep) {
	g := NewGraph()

	start := &sesTestStep{name: "start"}
	start.SetID(1)

	gate := &sesTestStep{name: "gate", exec: func(_ context.Context, r *ExecRequest) (ExecResponse, error) {
		switch mode {
		case SnapshotPrompted:
			if !r.Input.Has("answer") {
				p, _ := expr.NewVars(map[string]interface{}{"question": "ok?"})
				return Prompt(7, "approve", p), nil
			}
			return expr.NewVars(map[string]interface{}{"answer": expr.Must(expr.Select(r.Input, "answer")).Get()})

		default:
			if !r.Input.Has("resumed") {
				return Delay(until), nil
			}
			return expr.NewVars(map[string]interface{}{"woke": true})
		}
	}}
	gate.SetID(2)

	end := &sesTestStep{name: "end"}
	end.SetID(3)

	g.AddStep(start, gate)
	g.AddStep(gate, end)
	g.AddStep(end)
	return g, start
}

// testTypes stands in for the automation type registry.
func testTypes(typ string) expr.Type {
	switch typ {
	case "Integer":
		return &expr.Integer{}
	case "String":
		return &expr.String{}
	case "Boolean":
		return &expr.Boolean{}
	}
	return nil
}

func suspend(t *testing.T, g *Graph, start Step, want SessionStatus) *Session {
	t.Helper()
	ctx := auth.SetIdentityToContext(context.Background(), auth.Authenticated(9, 100))
	ses := NewSession(ctx, g, SetWorkerIntervalSuspended(time.Millisecond))

	scope, _ := expr.NewVars(map[string]interface{}{"order": "A-1"})
	require.NoError(t, ses.Exec(ctx, start, scope))

	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	require.NoError(t, ses.WaitUntil(wctx, want))
	return ses
}

func TestSnapshot_PromptSurvivesRestart(t *testing.T) {
	req := require.New(t)
	g, start := snapshotGraph(SnapshotPrompted, time.Time{})
	ses1 := suspend(t, g, start, SessionPrompted)

	snaps := ses1.Snapshots()
	req.Len(snaps, 1)
	req.Equal(SnapshotPrompted, snaps[0].Kind)
	req.Equal(uint64(2), snaps[0].StepID)
	req.Equal(uint64(1), snaps[0].ParentStepID)
	req.Equal("approve", snaps[0].PromptRef)
	req.Equal(uint64(7), snaps[0].PromptOwnerID)
	req.Equal(uint64(9), snaps[0].OwnerID)
	req.True(snaps[0].Resumable)

	// "restart": the old session is gone, only serialized data remains
	raw, err := json.Marshal(snaps)
	req.NoError(err)
	ses1.Stop()

	var loaded []SuspendedSnapshot
	req.NoError(json.Unmarshal(raw, &loaded))
	for i := range loaded {
		req.NoError(loaded[i].ResolveTypes(testTypes))
	}

	ctx := context.Background()
	ses2 := NewSession(ctx, g, SetSessionID(ses1.ID()), SetWorkerIntervalSuspended(time.Millisecond))
	req.Equal(ses1.ID(), ses2.ID())
	req.NoError(ses2.Restore(loaded))
	req.Equal(SessionPrompted, ses2.Status())
	req.Len(ses2.AllPendingPrompts(), 1)

	// a different user may not answer
	answer, _ := expr.NewVars(map[string]interface{}{"answer": "yes"})
	_, err = ses2.Resume(auth.SetIdentityToContext(ctx, auth.Authenticated(8)), loaded[0].StateID, answer)
	req.Error(err)

	_, err = ses2.Resume(auth.SetIdentityToContext(ctx, auth.Authenticated(7)), loaded[0].StateID, answer)
	req.NoError(err)

	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req.NoError(ses2.WaitUntil(wctx, SessionCompleted, SessionFailed))
	req.NoError(ses2.Error())
	req.Equal("yes", expr.Must(expr.Select(ses2.Result(), "answer")).Get())
	// scope from before the restart is intact
	req.Equal("A-1", expr.Must(expr.Select(ses2.Result(), "order")).Get())
}

func TestSnapshot_DelayResumesAfterRestart(t *testing.T) {
	req := require.New(t)
	g, start := snapshotGraph(SnapshotDelayed, time.Now().Add(time.Hour))
	ses1 := suspend(t, g, start, SessionDelayed)

	snaps := ses1.Snapshots()
	req.Len(snaps, 1)
	req.Equal(SnapshotDelayed, snaps[0].Kind)
	req.NotNil(snaps[0].ResumeAt)
	ses1.Stop()

	// the wake-up time passed while the process was down
	past := time.Now().Add(-time.Minute)
	snaps[0].ResumeAt = &past

	ctx := context.Background()
	ses2 := NewSession(ctx, g, SetSessionID(ses1.ID()), SetWorkerIntervalSuspended(time.Millisecond))
	req.NoError(ses2.Restore(snaps))

	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req.NoError(ses2.WaitUntil(wctx, SessionCompleted, SessionFailed))
	req.NoError(ses2.Error())
	req.Equal(true, expr.Must(expr.Select(ses2.Result(), "woke")).Get())
}

func TestSnapshot_RestoreRejectsBadSnapshots(t *testing.T) {
	g, _ := snapshotGraph(SnapshotPrompted, time.Time{})
	ses := NewSession(context.Background(), g)
	defer ses.Stop()

	cases := map[string]SuspendedSnapshot{
		"step removed from workflow": {Kind: SnapshotPrompted, StateID: 1, StepID: 999, Resumable: true},
		"inside a loop":              {Kind: SnapshotPrompted, StateID: 1, StepID: 2, Resumable: false},
		"unknown kind":               {Kind: "bogus", StateID: 1, StepID: 2, Resumable: true},
	}
	for name, snap := range cases {
		require.Error(t, ses.Restore([]SuspendedSnapshot{snap}), name)
	}
	require.Equal(t, SessionActive, ses.Status(), "failed restore must leave the session untouched")
}
