package wfexec

import (
	"fmt"
	"sort"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

// Suspended-state snapshots let a session that is waiting on a delay or a
// prompt survive a process restart: the waiting state is reduced to plain
// data (step IDs instead of step objects, identity IDs instead of identity
// objects), stored by the caller, and later put back into a freshly created
// session over the same graph with Restore.
//
// Only what the session is *waiting on* is captured. Other parallel branches
// that were mid-flight when the snapshot was taken are not, and iterator
// (loop) state cannot be serialized - states inside a loop are reported as
// not Resumable so the caller can fail the session cleanly instead of
// resuming it with a half-remembered loop.

const (
	SnapshotDelayed  = "delayed"
	SnapshotPrompted = "prompted"
)

type (
	SuspendedSnapshot struct {
		Kind    string `json:"kind"` // SnapshotDelayed | SnapshotPrompted
		StateID uint64 `json:"stateID,string"`

		StepID           uint64 `json:"stepID,string"`
		ParentStepID     uint64 `json:"parentStepID,string,omitempty"`
		ErrHandlerStepID uint64 `json:"errHandlerStepID,string,omitempty"`
		ErrHandled       bool   `json:"errHandled,omitempty"`

		// Identity the state runs as
		OwnerID    uint64   `json:"ownerID,string,omitempty"`
		OwnerRoles []uint64 `json:"ownerRoles,omitempty"`

		CreatedAt time.Time  `json:"createdAt"`
		Scope     *expr.Vars `json:"scope"`
		Results   *expr.Vars `json:"results,omitempty"`

		// delayed
		ResumeAt *time.Time `json:"resumeAt,omitempty"`

		// prompted
		PromptRef     string     `json:"promptRef,omitempty"`
		PromptOwnerID uint64     `json:"promptOwnerID,string,omitempty"`
		PromptPayload *expr.Vars `json:"promptPayload,omitempty"`

		// Resumable is false for states inside a loop, which cannot be
		// restored faithfully.
		Resumable bool `json:"resumable"`
	}
)

// Snapshots returns every state the session is currently suspended on,
// ordered by state ID.
func (s *Session) Snapshots() []SuspendedSnapshot {
	s.mux.RLock()
	defer s.mux.RUnlock()

	out := make([]SuspendedSnapshot, 0, len(s.delayed)+len(s.prompted))

	for _, d := range s.delayed {
		snap := snapshotOf(d.state)
		snap.Kind = SnapshotDelayed
		if !d.resumeAt.IsZero() {
			at := d.resumeAt
			snap.ResumeAt = &at
		}
		out = append(out, snap)
	}

	for _, p := range s.prompted {
		snap := snapshotOf(p.state)
		snap.Kind = SnapshotPrompted
		snap.PromptRef = p.ref
		snap.PromptOwnerID = p.ownerId
		snap.PromptPayload = p.payload
		out = append(out, snap)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].StateID < out[j].StateID })
	return out
}

func snapshotOf(st *State) SuspendedSnapshot {
	snap := SuspendedSnapshot{
		StateID:    st.stateId,
		ErrHandled: st.errHandled,
		CreatedAt:  st.created,
		Scope:      st.scope,
		Results:    st.results,
		Resumable:  len(st.loops) == 0,
	}

	if st.step != nil {
		snap.StepID = st.step.ID()
	}
	if st.parent != nil {
		snap.ParentStepID = st.parent.ID()
	}
	if st.errHandler != nil {
		snap.ErrHandlerStepID = st.errHandler.ID()
	}
	if st.owner != nil {
		snap.OwnerID = st.owner.Identity()
		snap.OwnerRoles = st.owner.Roles()
	}

	return snap
}

// Restore puts previously captured snapshots back into the session, which
// must be a fresh one (see SetSessionID) over the same graph the snapshots
// were taken from. Delayed states resume at their original time - right
// away when it already passed - and prompts become pending again.
func (s *Session) Restore(snaps []SuspendedSnapshot) error {
	// resolve everything before touching the session so a bad snapshot
	// leaves it unchanged
	type restored struct {
		d *delayed
		p *prompted
	}
	res := make([]restored, 0, len(snaps))

	for _, snap := range snaps {
		if !snap.Resumable {
			return fmt.Errorf("state %d is inside a loop and cannot be restored", snap.StateID)
		}

		step := s.g.StepByID(snap.StepID)
		if step == nil {
			return fmt.Errorf("state %d: step %d is not in the workflow graph (changed since it was suspended?)", snap.StateID, snap.StepID)
		}

		st := &State{
			stateId:    snap.StateID,
			sessionId:  s.id,
			created:    snap.CreatedAt,
			step:       step,
			scope:      snap.Scope,
			results:    snap.Results,
			errHandled: snap.ErrHandled,
			loops:      make([]Iterator, 0, 4),
		}
		if st.scope == nil {
			st.scope = &expr.Vars{}
		}
		if snap.OwnerID != 0 {
			st.owner = auth.Authenticated(snap.OwnerID, snap.OwnerRoles...)
		}
		if snap.ParentStepID != 0 {
			st.parent = s.g.StepByID(snap.ParentStepID)
		}
		if snap.ErrHandlerStepID != 0 {
			if st.errHandler = s.g.StepByID(snap.ErrHandlerStepID); st.errHandler == nil {
				return fmt.Errorf("state %d: error handler step %d is not in the workflow graph", snap.StateID, snap.ErrHandlerStepID)
			}
		}

		switch snap.Kind {
		case SnapshotDelayed:
			d := &delayed{state: st}
			if snap.ResumeAt != nil {
				d.resumeAt = *snap.ResumeAt
			}
			res = append(res, restored{d: d})

		case SnapshotPrompted:
			res = append(res, restored{p: &prompted{
				state:   st,
				ref:     snap.PromptRef,
				ownerId: snap.PromptOwnerID,
				payload: snap.PromptPayload,
			}})

		default:
			return fmt.Errorf("state %d: unknown snapshot kind %q", snap.StateID, snap.Kind)
		}
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	for _, r := range res {
		switch {
		case r.d != nil:
			s.delayed[r.d.state.stateId] = r.d
		case r.p != nil:
			s.prompted[r.p.state.stateId] = r.p
		}
	}

	return nil
}

// SetSessionID makes a session reuse an existing identifier - used when a
// persisted session is brought back after a restart.
func SetSessionID(id uint64) SessionOpt {
	return func(s *Session) {
		s.id = id
	}
}

// ResolveTypes resolves the typed variables of a snapshot that went through
// JSON. Vars decode into unresolved values until they are matched against a
// type registry; Restore does not do that itself because the registry is the
// caller's (see automation/service.Registry).
func (snap *SuspendedSnapshot) ResolveTypes(res func(typ string) expr.Type) error {
	for _, v := range []*expr.Vars{snap.Scope, snap.Results, snap.PromptPayload} {
		if v == nil {
			continue
		}
		if err := v.ResolveTypes(res); err != nil {
			return fmt.Errorf("state %d: %w", snap.StateID, err)
		}
	}
	return nil
}
