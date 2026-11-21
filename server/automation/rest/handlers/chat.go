package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type (
	ChatAPI interface {
		Ask(context.Context, *request.ChatAsk) (interface{}, error)
		AskStream(context.Context, *request.ChatAsk, service.ChatStreamFunc) error
	}
	Chat struct {
		Ask       func(http.ResponseWriter, *http.Request)
		AskStream func(http.ResponseWriter, *http.Request)
	}
)

func NewChat(h ChatAPI) *Chat {
	return &Chat{
		Ask: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewChatAsk()
			if err := params.Fill(r); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			value, err := h.Ask(r.Context(), params)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(value)
		},

		AskStream: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()

			flusher, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "streaming not supported", http.StatusInternalServerError)
				return
			}

			params := request.NewChatAsk()
			if err := params.Fill(r); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			writeMu := sync.Mutex{}
			writeEvent := func(payload map[string]any) error {
				writeMu.Lock()
				defer writeMu.Unlock()
				data, _ := json.Marshal(payload)
				_, err := fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				return err
			}

			ctx := chat.ContextWithStatus(r.Context(), func(status string) error {
				return writeEvent(map[string]any{
					"status": status,
					"done":   false,
				})
			})

			buffered, flush := chat.BufferStream(func(token string, reason string, done bool) error {
				return writeEvent(map[string]any{
					"token":  token,
					"reason": reason,
					"done":   done,
				})
			})
			defer func() { _ = flush() }()

			err := h.AskStream(ctx, params, buffered)
			_ = flush()

			if err != nil {
				_ = writeEvent(map[string]any{
					"error": err.Error(),
					"done":  true,
				})
			}
		},
	}
}

func (h *Chat) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/prompt", h.Ask)
		r.Post("/workflow/{workflowID}/prompt", h.Ask)
		r.Post("/prompt/stream", h.AskStream)
		r.Post("/workflow/{workflowID}/prompt/stream", h.AskStream)
	})
}
