package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type (
	ChatAPI interface {
		Ask(context.Context, *request.ChatAsk) (interface{}, error)
	}
	Chat struct {
		Ask func(http.ResponseWriter, *http.Request)
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
	}
}

func (h *Chat) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/namespace/{namespaceID}/prompt", h.Ask)
		r.Post("/namespace/{namespaceID}/page/{pageID}/prompt", h.Ask)
	})
}
