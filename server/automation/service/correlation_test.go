package service

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
	"github.com/stretchr/testify/require"
)

// newWaitingSession builds a one-step wfexec session that immediately
// suspends on a prompt carrying a correlationKey (and, optionally, an
// owner) - exactly the shape bpmn_compile.go's compileIntermediateCatchEvent
// produces for a BPMN message intermediate catch event - and resumes with
// whatever input it's given, recording it in resumedWith.
func newWaitingSession(t *testing.T, ownerId uint64, correlationKey string) (*types.Session, *expr.Vars) {
	t.Helper()
	ctx := context.Background()
	resumedWith := &expr.Vars{}

	step := wfexec.NewGenericStep(func(ctx context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
		if r.Input == nil {
			payload := &expr.Vars{}
			require.NoError(t, payload.Set("correlationKey", correlationKey))
			return wfexec.Prompt(ownerId, "message:Test", payload), nil
		}
		r.Input.Copy(resumedWith, "status")
		return &expr.Vars{}, nil
	})

	g := wfexec.NewGraph()
	g.AddStep(step)

	ses := wfexec.NewSession(ctx, g)
	require.NoError(t, ses.Exec(ctx, step, nil))
	// Wait() only watches for Failed/Delayed/Completed - a prompt suspends
	// into SessionPrompted, so wait for that status explicitly.
	require.NoError(t, ses.WaitUntil(ctx, wfexec.SessionPrompted))
	require.True(t, ses.Prompted(), "session should be suspended on the prompt")

	return types.NewSession(ses), resumedWith
}

// noopPromptSender satisfies the unexported promptSender interface Resume
// notifies on when a prompt is answered - not what's under test here.
type noopPromptSender struct{}

func (noopPromptSender) Send(kind string, payload interface{}, userIDs ...uint64) error { return nil }

func withPoolSession(t *testing.T, sessions ...*types.Session) func() {
	t.Helper()
	pool := make(map[uint64]*types.Session, len(sessions))
	for _, s := range sessions {
		pool[s.ID] = s
	}
	restore := DefaultSession
	DefaultSession = &session{pool: pool, promptSender: noopPromptSender{}}
	return func() { DefaultSession = restore }
}

func TestResolveCorrelation_ResumesMatchingSession(t *testing.T) {
	req := require.New(t)
	ts, resumedWith := newWaitingSession(t, 42, "ORDER-123")
	defer withPoolSession(t, ts)()

	input := &expr.Vars{}
	req.NoError(input.Set("status", "confirmed"))

	req.NoError(ResolveCorrelation(context.Background(), "ORDER-123", input))

	// sync point: wait for the resumed step to actually run (Resume just
	// enqueues it), same as pkg/wfexec's own session tests do after Resume
	_, _, _, err := ts.WaitResults(context.Background())
	req.NoError(err)

	req.Equal("confirmed", resumedWith.Dict()["status"])
}

func TestResolveCorrelation_IgnoresNonMatchingSessions(t *testing.T) {
	req := require.New(t)
	other, _ := newWaitingSession(t, 42, "ORDER-999")
	target, resumedWith := newWaitingSession(t, 42, "ORDER-123")
	defer withPoolSession(t, other, target)()

	req.NoError(ResolveCorrelation(context.Background(), "ORDER-123", &expr.Vars{}))
	_ = resumedWith
}

func TestResolveCorrelation_ErrorsWhenNoMatch(t *testing.T) {
	req := require.New(t)
	ts, _ := newWaitingSession(t, 42, "ORDER-123")
	defer withPoolSession(t, ts)()

	err := ResolveCorrelation(context.Background(), "NO-SUCH-KEY", &expr.Vars{})
	req.Error(err)
}

// A message catch event's prompt is only reachable by ResolveCorrelation if
// it actually suspended - and wfexec itself refuses to suspend a prompt with
// no owner in the first place (session.go: "prompted... without an owner"),
// so the "no owner" branch in ResolveCorrelation is defense-in-depth for a
// state the engine already won't produce. This proves that invariant at the
// engine level: a step returning Prompt(0, ...) fails the session outright
// instead of ever appearing as a pending prompt.
func TestWfexecPrompt_RejectsMissingOwner(t *testing.T) {
	req := require.New(t)
	ctx := context.Background()

	step := wfexec.NewGenericStep(func(ctx context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
		return wfexec.Prompt(0, "message:Test", &expr.Vars{}), nil
	})
	g := wfexec.NewGraph()
	g.AddStep(step)

	ses := wfexec.NewSession(ctx, g)
	req.NoError(ses.Exec(ctx, step, nil))
	// WaitUntil returns the session's error once it reaches the expected
	// (failed) status - that error IS the assertion here.
	err := ses.WaitUntil(ctx, wfexec.SessionFailed)
	req.Error(err)
	req.Contains(err.Error(), "without an owner")
}

func TestResolveCorrelation_ErrorsOnEmptyKey(t *testing.T) {
	require.Error(t, ResolveCorrelation(context.Background(), "", &expr.Vars{}))
}
