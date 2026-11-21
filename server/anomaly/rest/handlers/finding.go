package handlers

// This file is auto-generated.
//
// Changes to this file may cause incorrect behavior and will be lost if
// the code is regenerated from anomaly/rest.yaml

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/anomaly/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
	"net/http"
)

type (
	// Internal API interface
	FindingAPI interface {
		Search(context.Context, *request.FindingSearch) (interface{}, error)
		UpdateStatus(context.Context, *request.FindingUpdateStatus) (interface{}, error)
	}

	// HTTP API interface
	Finding struct {
		Search       func(http.ResponseWriter, *http.Request)
		UpdateStatus func(http.ResponseWriter, *http.Request)
	}
)

func NewFinding(h FindingAPI) *Finding {
	return &Finding{
		Search: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewFindingSearch()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.Search(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
		UpdateStatus: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewFindingUpdateStatus()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.UpdateStatus(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
	}
}

func (h Finding) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/namespace/{namespaceID}/anomaly/", h.Search)
		r.Patch("/namespace/{namespaceID}/anomaly/{findingID}/status", h.UpdateStatus)
	})
}
