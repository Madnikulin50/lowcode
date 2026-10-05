package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const persistTestWorkflowID = 11

// approvalGraph: start -> approve (suspends on a prompt for user 9) -> done
func approvalGraph() *wfexec.Graph {
	g := wfexec.NewGraph()

	start := wfexec.NewGenericStep(func(_ context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
		return expr.NewVars(map[string]interface{}{"started": true})
	})
	start.SetID(1)

	approve := wfexec.NewGenericStep(func(_ context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
		if !r.Input.Has("approved") {
			payload, _ := expr.NewVars(map[string]interface{}{"question": "ship it?"})
			return wfexec.Prompt(9, "approve", payload), nil
		}
		return expr.NewVars(map[string]interface{}{"approved": expr.Must(expr.Select(r.Input, "approved")).Get()})
	})
	approve.SetID(2)

	done := wfexec.NewGenericStep(func(_ context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
		return expr.NewVars(map[string]interface{}{"done": true})
	})
	done.SetID(3)

	g.AddStep(start, approve)
	g.AddStep(approve, done)
	g.AddStep(done)
	return g
}

func newPersistTestService(t *testing.T, s store.Storer) *session {
	t.Helper()
	return &session{
		store:      s,
		log:        zap.NewNop(),
		opt:        options.WorkflowOpt{},
		pool:       make(map[uint64]*types.Session),
		spawnQueue: make(chan *spawn),
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestSession_SuspendedApprovalSurvivesRestart(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
		g   = approvalGraph()
	)

	st, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), st))

	Registry().AddTypes(&expr.Any{}, &expr.Boolean{}, &expr.String{}, &expr.Vars{})
	prev := DefaultWorkflow
	DefaultWorkflow = &workflow{muxCache: &sync.RWMutex{}, cache: map[uint64]*wfCacheItem{persistTestWorkflowID: {g: g}}}
	defer func() { DefaultWorkflow = prev }()

	// --- first life of the server: start a workflow that waits for approval
	svc1 := newPersistTestService(t, st)
	wctx, stop := context.WithCancel(ctx)
	defer stop()
	svc1.Watch(wctx)

	input, _ := expr.NewVars(map[string]interface{}{"order": "A-1"})
	_, sessionID, err := svc1.Start(ctx, g, types.SessionStartParams{
		Invoker:    auth.Authenticated(9),
		WorkflowID: persistTestWorkflowID,
		Input:      input,
	})
	req.NoError(err)

	waitFor(t, "state to be persisted", func() bool { return len(svc1.states.ids(sessionID)) == 1 })

	rows, _, err := store.SearchAutomationStates(ctx, st, types.StateFilter{})
	req.NoError(err)
	req.Len(rows, 1)
	req.Equal(sessionID, rows[0].SessionID)
	req.Equal(wfexec.SnapshotPrompted, rows[0].Kind)
	stateID := rows[0].ID

	// identity objects are not stored; they are rebuilt on restore
	req.False(rows[0].Snapshot.Scope.Has("invoker"))
	req.False(rows[0].Snapshot.Scope.Has("runner"))

	// --- the server restarts: a new service over the same database
	stop()
	svc2 := newPersistTestService(t, st)
	req.NoError(svc2.restoreSuspended(ctx))

	req.Contains(svc2.pool, sessionID)
	req.Equal(types.SessionPrompted, svc2.pool[sessionID].Status)
	pending := svc2.PendingPrompts(auth.SetIdentityToContext(ctx, auth.Authenticated(9)))
	req.Len(pending, 1)
	req.Equal("approve", pending[0].Ref)
	req.Equal(stateID, pending[0].StateID)

	// somebody else cannot approve
	answer, _ := expr.NewVars(map[string]interface{}{"approved": true})
	req.Error(svc2.Resume(sessionID, stateID, auth.Authenticated(8), answer))

	req.NoError(svc2.Resume(sessionID, stateID, auth.Authenticated(9), answer))

	waitFor(t, "session to complete", func() bool {
		row, err := store.LookupAutomationSessionByID(ctx, st, sessionID)
		return err == nil && row.Status == types.SessionCompleted
	})

	// nothing left to restore
	left, _, err := store.SearchAutomationStates(ctx, st, types.StateFilter{})
	req.NoError(err)
	req.Empty(left)

	row, err := store.LookupAutomationSessionByID(ctx, st, sessionID)
	req.NoError(err)
	req.NotNil(row.CompletedAt)
	req.Empty(row.Error)
}

func TestSession_RestoreFailsCleanlyWhenWorkflowIsGone(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
		g   = approvalGraph()
	)

	st, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), st))

	Registry().AddTypes(&expr.Any{}, &expr.Boolean{}, &expr.String{}, &expr.Vars{})
	prev := DefaultWorkflow
	DefaultWorkflow = &workflow{muxCache: &sync.RWMutex{}, cache: map[uint64]*wfCacheItem{persistTestWorkflowID: {g: g}}}
	defer func() { DefaultWorkflow = prev }()

	svc1 := newPersistTestService(t, st)
	wctx, stop := context.WithCancel(ctx)
	defer stop()
	svc1.Watch(wctx)

	input, _ := expr.NewVars(map[string]interface{}{})
	_, sessionID, err := svc1.Start(ctx, g, types.SessionStartParams{Invoker: auth.Authenticated(9), WorkflowID: persistTestWorkflowID, Input: input})
	req.NoError(err)
	waitFor(t, "state to be persisted", func() bool { return len(svc1.states.ids(sessionID)) == 1 })
	stop()

	// while the server was down the workflow got deleted
	DefaultWorkflow = &workflow{muxCache: &sync.RWMutex{}, cache: map[uint64]*wfCacheItem{}}

	svc2 := newPersistTestService(t, st)
	req.NoError(svc2.restoreSuspended(ctx))
	req.NotContains(svc2.pool, sessionID)

	row, err := store.LookupAutomationSessionByID(ctx, st, sessionID)
	req.NoError(err)
	req.Equal(types.SessionFailed, row.Status)
	req.Contains(row.Error, "restart")

	left, _, err := store.SearchAutomationStates(ctx, st, types.StateFilter{})
	req.NoError(err)
	req.Empty(left)
}
