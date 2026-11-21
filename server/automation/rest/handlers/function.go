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
	FunctionAPI interface {
		List(context.Context, *request.FunctionList) (interface{}, error)
	}

	// HTTP API interface
	Function struct {
		List func(http.ResponseWriter, *http.Request)
	}
)

func NewFunction(h FunctionAPI) *Function {
	return &Function{
		List: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewFunctionList()
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

func (h Function) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/functions/", h.List)
	})
}
