package handlers

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from compose/rest.yaml

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"net/http"
)

type (
	// Internal API interface
	NotificationAPI interface {
		EmailSend(context.Context, *request.NotificationEmailSend) (interface{}, error)
	}

	// HTTP API interface
	Notification struct {
		EmailSend func(http.ResponseWriter, *http.Request)
	}
)

func NewNotification(h NotificationAPI) *Notification {
	return &Notification{
		EmailSend: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewNotificationEmailSend()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.EmailSend(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
	}
}

func (h Notification) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/notification/email", h.EmailSend)
	})
}
