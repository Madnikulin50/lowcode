package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSSE(t *testing.T) {
	rec := httptest.NewRecorder()
	sse, ok := NewSSE(rec)
	require.True(t, ok)

	require.NoError(t, sse.Send(map[string]interface{}{"token": "Hel", "done": false}))
	require.NoError(t, sse.Send(map[string]interface{}{"result": map[string]interface{}{"n": 1}, "done": true}))

	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	require.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	require.True(t, rec.Flushed, "each event must be flushed or the client sees nothing until the end")
	require.Equal(t, "data: {\"done\":false,\"token\":\"Hel\"}\n\ndata: {\"done\":true,\"result\":{\"n\":1}}\n\n", rec.Body.String())

	require.Error(t, sse.Send(make(chan int)), "an unserialisable payload is an error, not a half-written event")
	require.NotContains(t, rec.Body.String(), "chan")
}

type noFlush struct{ http.ResponseWriter }

func TestSSE_NoFlusher(t *testing.T) {
	// a writer that cannot stream: report it before anything is written
	w := noFlush{httptest.NewRecorder()}
	_, ok := NewSSE(struct{ http.ResponseWriter }{w})
	require.False(t, ok)
	require.False(t, strings.Contains(w.Header().Get("Content-Type"), "event-stream"))
}
