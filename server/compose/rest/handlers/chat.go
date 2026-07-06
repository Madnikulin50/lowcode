package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
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
				api.Send(w, r, map[string]any{
					"response": err.Error(),
				})
				return
			}

			value, err := h.Ask(r.Context(), params)
			if err != nil {
				api.Send(w, r,
					map[string]any{
						"response": err.Error(),
					})
				return
			}

			api.Send(w, r, value)
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

			err := h.AskStream(r.Context(), params, func(token string, done bool) error {
				data, _ := json.Marshal(map[string]any{
					"token": token,
					"done":  done,
				})
				_, err := fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				return err
			})

			if err != nil {
				data, _ := json.Marshal(map[string]any{
					"error": err.Error(),
					"done":  true,
				})
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		},
	}
}

func (h *Chat) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/namespace/{namespaceID}/prompt", h.Ask)
		r.Post("/namespace/{namespaceID}/page/{pageID}/prompt", h.Ask)
		r.Post("/namespace/{namespaceID}/prompt/stream", h.AskStream)
		r.Post("/namespace/{namespaceID}/page/{pageID}/prompt/stream", h.AskStream)
	})
}
