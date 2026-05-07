package handlers

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from system/rest.yaml

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/system/rest/request"
	"net/http"
)

type (
	// Internal API interface
	ActionlogAPI interface {
		List(context.Context, *request.ActionlogList) (interface{}, error)
	}

	// HTTP API interface
	Actionlog struct {
		List func(http.ResponseWriter, *http.Request)
	}
)

func NewActionlog(h ActionlogAPI) *Actionlog {
	return &Actionlog{
		List: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewActionlogList()
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

func (h Actionlog) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/actionlog/", h.List)
	})
}
