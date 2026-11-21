package aiagent

import (
	"context"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// Progress reporting: a caller that wants to watch an AI step as it runs - the
// "test step" and "try this node" buttons - puts a ProgressFunc on the context.
// What happens while the step runs then reaches it as it happens, instead of
// as one reply at the end of a wait that can take minutes on a local model.
//
// Without a ProgressFunc nothing changes: the agent generates its answer in
// one piece, as before.

// Progress is one thing that happened. Exactly one field is set.
type Progress struct {
	// Status: the model is warming up, ready, using tools (chat.Status*)
	Status string
	// Token is a piece of the answer as it is generated
	Token string
	// Reason is a piece of the model's reasoning, for models that expose it
	Reason string
	// Attempt (>0) is a new attempt after an answer was rejected
	Attempt int
}

type ProgressFunc func(Progress)

type progressCtxKey struct{}

// ContextWithProgress makes fn receive the progress of every AI step run
// under ctx. fn is called from the goroutine running the step.
func ContextWithProgress(ctx context.Context, fn ProgressFunc) context.Context {
	if fn == nil {
		return ctx
	}
	ctx = context.WithValue(ctx, progressCtxKey{}, fn)
	return chat.ContextWithStatus(ctx, func(status string) error {
		fn(Progress{Status: status})
		return nil
	})
}

func progressFromContext(ctx context.Context) ProgressFunc {
	fn, _ := ctx.Value(progressCtxKey{}).(ProgressFunc)
	return fn
}

// ReportProgress sends p to the context's ProgressFunc, if there is one.
func ReportProgress(ctx context.Context, p Progress) {
	if fn := progressFromContext(ctx); fn != nil {
		fn(p)
	}
}

// streamFromContext adapts the context's ProgressFunc to the streaming
// callback of the agent runtime; nil when nobody is watching.
func streamFromContext(ctx context.Context) chat.StreamFunc {
	fn := progressFromContext(ctx)
	if fn == nil {
		return nil
	}
	return func(text, reasoning string, _ bool) error {
		if text != "" || reasoning != "" {
			fn(Progress{Token: text, Reason: reasoning})
		}
		return ctx.Err() // stop generating once the watcher has gone
	}
}

// Event is the progress as a JSON event for a client stream, using the field
// names the chat stream already has: status, token, reason - plus attempt.
func (p Progress) Event() map[string]interface{} {
	ev := map[string]interface{}{"done": false}
	switch {
	case p.Status != "":
		ev["status"] = p.Status
	case p.Attempt > 0:
		ev["attempt"] = p.Attempt
	default:
		ev["token"], ev["reason"] = p.Token, p.Reason
	}
	return ev
}
