package handlers

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from automation/rest.yaml

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"net/http"
)

type (
	// Internal API interface
	EventTypesAPI interface {
		List(context.Context, *request.EventTypesList) (interface{}, error)
	}

	// HTTP API interface
	EventTypes struct {
		List func(http.ResponseWriter, *http.Request)
	}
)

func NewEventTypes(h EventTypesAPI) *EventTypes {
	return &EventTypes{
		List: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewEventTypesList()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.List(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
	}
}

func (h EventTypes) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/event-types/", h.List)
	})
}
