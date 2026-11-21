package rulesgo

import (
	"context"
	"sync"
)

// A node reports how it did its work (an AI node: prompt, response, model,
// tokens) through a sink carried on its context rather than through its
// return value: a node that fails returns no output, yet the trace of a
// failed AI call is exactly what you want to read.
type (
	traceSink struct {
		mux   sync.Mutex
		trace map[string]interface{}
	}

	traceSinkKey struct{}
)

func withTraceSink(ctx context.Context) (context.Context, *traceSink) {
	s := &traceSink{}
	return context.WithValue(ctx, traceSinkKey{}, s), s
}

// recordTrace stores trace for the node running under ctx; a no-op when the
// node runs outside the engine (a unit test, say).
func recordTrace(ctx context.Context, trace map[string]interface{}) {
	if s, _ := ctx.Value(traceSinkKey{}).(*traceSink); s != nil {
		s.mux.Lock()
		s.trace = trace
		s.mux.Unlock()
	}
}

func (s *traceSink) get() map[string]interface{} {
	if s == nil {
		return nil
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.trace
}
