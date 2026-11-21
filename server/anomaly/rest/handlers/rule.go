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
	RuleAPI interface {
		Search(context.Context, *request.RuleSearch) (interface{}, error)
		Create(context.Context, *request.RuleCreate) (interface{}, error)
		Update(context.Context, *request.RuleUpdate) (interface{}, error)
		Delete(context.Context, *request.RuleDelete) (interface{}, error)
	}

	// HTTP API interface
	Rule struct {
		Search func(http.ResponseWriter, *http.Request)
		Create func(http.ResponseWriter, *http.Request)
		Update func(http.ResponseWriter, *http.Request)
		Delete func(http.ResponseWriter, *http.Request)
	}
)

func NewRule(h RuleAPI) *Rule {
	return &Rule{
		Search: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewRuleSearch()
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
		Create: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewRuleCreate()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.Create(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
		Update: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewRuleUpdate()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.Update(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
		Delete: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewRuleDelete()
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}

			value, err := h.Delete(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}

			api.Send(w, r, value)
		},
	}
}

func (h Rule) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/namespace/{namespaceID}/anomaly/rule/", h.Search)
		r.Post("/namespace/{namespaceID}/anomaly/rule/", h.Create)
		r.Post("/namespace/{namespaceID}/anomaly/rule/{ruleID}", h.Update)
		r.Delete("/namespace/{namespaceID}/anomaly/rule/{ruleID}", h.Delete)
	})
}
