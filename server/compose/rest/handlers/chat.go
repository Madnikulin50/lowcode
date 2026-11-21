package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type (
	ChatAPI interface {
		Ask(context.Context, *request.ChatAsk) (interface{}, error)
		AskStream(context.Context, *request.ChatAsk, service.ChatStreamFunc) error
		Models(context.Context) (interface{}, error)
		DiscoverModels(context.Context) (interface{}, error)
		Agents(context.Context) (interface{}, error)
		ProbeToolkits(context.Context, string, string) (interface{}, error)
		WarmUp(context.Context, *request.ChatAsk) error
	}
	Chat struct {
		Ask            func(http.ResponseWriter, *http.Request)
		AskStream      func(http.ResponseWriter, *http.Request)
		Models         func(http.ResponseWriter, *http.Request)
		DiscoverModels func(http.ResponseWriter, *http.Request)
		Agents         func(http.ResponseWriter, *http.Request)
		ProbeToolkits  func(http.ResponseWriter, *http.Request)
		WarmUp         func(http.ResponseWriter, *http.Request)
	}
)

func NewChat(h ChatAPI) *Chat {
	return &Chat{
		WarmUp: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewChatAsk()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, map[string]any{"error": err.Error()})
				return
			}
			if err := h.WarmUp(r.Context(), params); err != nil {
				api.Send(w, r, map[string]any{"error": err.Error()})
				return
			}
			api.Send(w, r, map[string]any{"ok": true})
		},

		Models: func(w http.ResponseWriter, r *http.Request) {
			payload, err := h.Models(r.Context())
			if err != nil {
				api.Send(w, r, map[string]any{
					"error": err.Error(),
				})
				return
			}
			api.Send(w, r, payload)
		},

		DiscoverModels: func(w http.ResponseWriter, r *http.Request) {
			payload, err := h.DiscoverModels(r.Context())
			if err != nil {
				api.Send(w, r, map[string]any{
					"error": err.Error(),
				})
				return
			}
			api.Send(w, r, payload)
		},

		Agents: func(w http.ResponseWriter, r *http.Request) {
			payload, err := h.Agents(r.Context())
			if err != nil {
				api.Send(w, r, map[string]any{
					"error": err.Error(),
				})
				return
			}
			api.Send(w, r, payload)
		},

		ProbeToolkits: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			var in struct {
				URL   string `json:"url"`
				Token string `json:"token"`
			}
			_ = json.NewDecoder(r.Body).Decode(&in)
			if in.URL == "" {
				in.URL = r.URL.Query().Get("url")
			}
			if in.Token == "" {
				in.Token = r.URL.Query().Get("token")
			}
			payload, err := h.ProbeToolkits(r.Context(), in.URL, in.Token)
			if err != nil {
				api.Send(w, r, map[string]any{"error": err.Error()})
				return
			}
			api.Send(w, r, payload)
		},

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
		r.Get("/chat/models", h.Models)
		r.Get("/chat/models/discover", h.DiscoverModels)
		r.Get("/chat/agents", h.Agents)
		r.Get("/chat/toolkits/probe", h.ProbeToolkits)
		r.Post("/chat/toolkits/probe", h.ProbeToolkits)
		r.Post("/chat/warmup", h.WarmUp)
		r.Post("/namespace/{namespaceID}/prompt", h.Ask)
		r.Post("/namespace/{namespaceID}/page/{pageID}/prompt", h.Ask)
		r.Post("/namespace/{namespaceID}/prompt/stream", h.AskStream)
		r.Post("/namespace/{namespaceID}/page/{pageID}/prompt/stream", h.AskStream)
	})
}
