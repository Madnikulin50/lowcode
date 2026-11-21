package service

import (
	"context"
	"fmt"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"sort"
	"sync"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
	"github.com/madnikulin50/lowcode/server/store"
	"go.uber.org/zap"
)

// Persistence of suspended sessions.
//
// A session waiting on a delay or a prompt (an approval, say) can wait for
// days; without this it is lost whenever the server restarts. While a
// session is suspended, each state it waits on is written to
// automation_states as a wfexec snapshot; the rows go away when the state is
// resumed or the session ends. On boot, restoreSuspended turns whatever rows
// are left back into live sessions.
//
// Limits (see wfexec/snapshot.go): only the states being waited on survive,
// not parallel branches that were mid-flight; states inside a loop cannot be
// restored; and a session whose workflow changed incompatibly while the
// server was down is failed instead of resumed.

// stateTracker remembers which state rows exist for a session, so a state is
// written once (it is immutable while suspended) and stale rows are found.
type stateTracker struct {
	mux sync.Mutex
	set map[uint64]map[uint64]struct{} // sessionID -> stateIDs
}

func (t *stateTracker) ids(sessionID uint64) (out []uint64) {
	t.mux.Lock()
	defer t.mux.Unlock()

	for id := range t.set[sessionID] {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return
}

func (t *stateTracker) has(sessionID, stateID uint64) bool {
	t.mux.Lock()
	defer t.mux.Unlock()
	_, ok := t.set[sessionID][stateID]
	return ok
}

func (t *stateTracker) add(sessionID, stateID uint64) {
	t.mux.Lock()
	defer t.mux.Unlock()

	if t.set == nil {
		t.set = make(map[uint64]map[uint64]struct{})
	}
	if t.set[sessionID] == nil {
		t.set[sessionID] = make(map[uint64]struct{})
	}
	t.set[sessionID][stateID] = struct{}{}
}

func (t *stateTracker) remove(sessionID, stateID uint64) {
	t.mux.Lock()
	defer t.mux.Unlock()

	delete(t.set[sessionID], stateID)
	if len(t.set[sessionID]) == 0 {
		delete(t.set, sessionID)
	}
}

// reconcileStates brings the stored state rows in line with what the live
// session is waiting on. Called from the state change handler (which holds
// svc.mux) after every status change.
func (svc *session) reconcileStates(ctx context.Context, ses *types.Session, s *wfexec.Session, status wfexec.SessionStatus) {
	log := svc.log.With(logger.Uint64("sessionID", ses.ID))

	switch status {
	case wfexec.SessionCompleted, wfexec.SessionFailed, wfexec.SessionCanceled:
		svc.dropStates(ctx, ses.ID, svc.states.ids(ses.ID)...)
		return
	}

	known := svc.states.ids(ses.ID)
	if len(known) == 0 && status != wfexec.SessionPrompted && status != wfexec.SessionDelayed {
		// the common case: an active session that never suspended
		return
	}

	live := make(map[uint64]struct{})
	for _, snap := range s.Snapshots() {
		live[snap.StateID] = struct{}{}

		if svc.states.has(ses.ID, snap.StateID) {
			continue
		}
		if !snap.Resumable {
			log.Warn("suspended state is inside a loop and will not survive a restart", logger.Uint64("stateID", snap.StateID))
			continue
		}

		if err := store.UpsertAutomationState(ctx, svc.store, stateRow(ses, snap)); err != nil {
			// not fatal: the session keeps running, it just is not durable
			log.Error("could not persist suspended state", logger.Uint64("stateID", snap.StateID), zap.Error(err))
			continue
		}
		svc.states.add(ses.ID, snap.StateID)
	}

	var gone []uint64
	for _, id := range known {
		if _, ok := live[id]; !ok {
			gone = append(gone, id)
		}
	}
	svc.dropStates(ctx, ses.ID, gone...)
}

func stateRow(ses *types.Session, snap wfexec.SuspendedSnapshot) *types.State {
	// invoker and runner are identity objects; they are rebuilt from the
	// session and the snapshot's owner on restore
	if snap.Scope != nil {
		if v, err := snap.Scope.Delete("invoker", "runner"); err == nil {
			if vars, ok := v.(*expr.Vars); ok {
				snap.Scope = vars
			}
		}
	}

	stored := types.StateSnapshot(snap)
	return &types.State{
		ID:         snap.StateID,
		SessionID:  ses.ID,
		WorkflowID: ses.WorkflowID,
		Kind:       snap.Kind,
		ResumeAt:   snap.ResumeAt,
		CreatedAt:  snap.CreatedAt,
		Snapshot:   &stored,
	}
}

func (svc *session) dropStates(ctx context.Context, sessionID uint64, stateIDs ...uint64) {
	for _, id := range stateIDs {
		if err := store.DeleteAutomationStateByID(ctx, svc.store, id); err != nil {
			svc.log.Error("could not remove persisted state",
				logger.Uint64("sessionID", sessionID), logger.Uint64("stateID", id), zap.Error(err))
			continue
		}
		svc.states.remove(sessionID, id)
	}
}

// restoreSuspended brings sessions that were suspended when the server last
// stopped back to life. It needs the workflows loaded (their graphs).
func (svc *session) restoreSuspended(ctx context.Context) error {
	rows, _, err := store.SearchAutomationStates(ctx, svc.store, types.StateFilter{})
	if err != nil {
		return fmt.Errorf("could not load persisted states: %w", err)
	}

	bySession := make(map[uint64]types.StateSet)
	var order []uint64
	for _, r := range rows {
		if _, ok := bySession[r.SessionID]; !ok {
			order = append(order, r.SessionID)
		}
		bySession[r.SessionID] = append(bySession[r.SessionID], r)
	}

	var restored, failed int
	for _, sid := range order {
		if err := svc.restoreSession(ctx, sid, bySession[sid]); err != nil {
			failed++
			svc.log.Warn("could not restore suspended session",
				logger.Uint64("sessionID", sid), zap.Error(err))
			svc.failRestored(ctx, sid, bySession[sid], err)
			continue
		}
		restored++
	}

	if restored+failed > 0 {
		svc.log.Info("restored suspended workflow sessions", zap.Int("restored", restored), zap.Int("failed", failed))
	}
	return nil
}

func (svc *session) restoreSession(ctx context.Context, sessionID uint64, rows types.StateSet) error {
	row, err := loadSession(ctx, svc.store, sessionID)
	if err != nil {
		return err
	}
	if row.CompletedAt != nil {
		// finished while its rows lingered - just clean up
		svc.dropStates(ctx, sessionID, rows.IDs()...)
		return nil
	}

	var g *wfexec.Graph
	if DefaultWorkflow != nil {
		g = DefaultWorkflow.graphOf(row.WorkflowID)
	}
	if g == nil {
		return fmt.Errorf("workflow %d is not available (deleted or disabled)", row.WorkflowID)
	}

	snaps := make([]wfexec.SuspendedSnapshot, 0, len(rows))
	for _, r := range rows {
		if r.Snapshot == nil {
			return fmt.Errorf("state %d has no snapshot", r.ID)
		}
		snap := wfexec.SuspendedSnapshot(*r.Snapshot)
		if err := snap.ResolveTypes(Registry().Type); err != nil {
			return err
		}
		snaps = append(snaps, snap)
	}

	// who the workflow runs as: the identity recorded on its suspended state
	var (
		invoker = auth.Authenticated(row.CreatedBy)
		runner  auth.Identifiable
	)
	for _, snap := range snaps {
		if snap.OwnerID != 0 {
			runner = auth.Authenticated(snap.OwnerID, snap.OwnerRoles...)
			break
		}
	}
	if runner == nil {
		runner = invoker
	}

	for i := range snaps {
		if snaps[i].Scope == nil {
			snaps[i].Scope = &expr.Vars{}
		}
		_ = snaps[i].Scope.AssignFieldValue("invoker", expr.Must(expr.NewAny(invoker)))
		_ = snaps[i].Scope.AssignFieldValue("runner", expr.Must(expr.NewAny(runner)))
	}

	execCtx := auth.SetIdentityToContext(context.Background(), runner)
	execCtx = context.WithValue(execCtx, workflowInvokerCtxKey{}, auth.Identifiable(invoker))
	// what the session spent before the restart is not remembered: it starts
	// the rest of its run with a fresh budget
	execCtx = aiagent.ContextWithBudget(execCtx, DefaultWorkflow.aiBudgetFor(row.WorkflowID))

	wfses := wfexec.NewSession(execCtx, g, append(
		svc.sessionOpts(ctx, row.WorkflowID, []uint64{row.WorkflowID}, runner),
		wfexec.SetSessionID(sessionID),
	)...)

	if err := wfses.Restore(snaps); err != nil {
		wfses.Stop()
		return err
	}

	ses := types.NewSession(wfses)
	ses.WorkflowID = row.WorkflowID
	ses.CreatedAt = row.CreatedAt
	ses.CreatedBy = row.CreatedBy
	ses.EventType = row.EventType
	ses.ResourceType = row.ResourceType
	ses.PurgeAt = row.PurgeAt
	ses.SuspendedAt = row.SuspendedAt
	ses.Input = row.Input
	ses.Stacktrace = row.Stacktrace
	ses.RuntimeStacktrace = row.Stacktrace
	if !svc.opt.StackTraceEnabled {
		ses.DisableStacktrace()
	}
	if svc.opt.StackTraceFull {
		ses.FullStacktrace()
	}

	status := wfexec.SessionDelayed
	ses.Status = types.SessionSuspended
	if wfses.Prompted() {
		status = wfexec.SessionPrompted
		ses.Status = types.SessionPrompted
	}

	svc.mux.Lock()
	svc.pool[sessionID] = ses
	svc.mux.Unlock()
	for _, r := range rows {
		svc.states.add(sessionID, r.ID)
	}

	// run it through the regular handler: it re-sends pending prompts to
	// their owners and refreshes the stored session
	return svc.stateChangeHandler(ctx)(status, nil, wfses)
}

// failRestored ends a session that could not be brought back, so it does not
// sit in the database as suspended forever.
func (svc *session) failRestored(ctx context.Context, sessionID uint64, rows types.StateSet, cause error) {
	svc.dropStates(ctx, sessionID, rows.IDs()...)

	row, err := loadSession(ctx, svc.store, sessionID)
	if err != nil {
		return
	}

	at := *now()
	row.Status = types.SessionFailed
	row.SuspendedAt = nil
	row.CompletedAt = &at
	row.Error = "interrupted by a server restart and could not be resumed: " + cause.Error()
	if err := store.UpsertAutomationSession(ctx, svc.store, row); err != nil {
		svc.log.Error("could not mark session as failed", logger.Uint64("sessionID", sessionID), zap.Error(err))
	}
}
