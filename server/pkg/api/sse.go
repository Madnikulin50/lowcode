package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSE writes a Server-Sent Events stream, one JSON payload per event, in the
// framing the clients already read for chat streaming ("data: {...}\n\n").
type SSE struct {
	w  http.ResponseWriter
	f  http.Flusher
	mu sync.Mutex
}

// NewSSE starts an event stream on w. ok is false when the connection cannot
// stream (no flusher); nothing has been written then.
func NewSSE(w http.ResponseWriter) (s *SSE, ok bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // do not let a proxy hold events back
	return &SSE{w: w, f: f}, true
}

// Send writes one event and flushes it to the client.
func (s *SSE) Send(payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err = fmt.Fprintf(s.w, "data: %s\n\n", data); err != nil {
		return err
	}
	s.f.Flush()
	return nil
}
