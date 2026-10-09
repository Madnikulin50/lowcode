package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type (
	ETLAPI interface {
		List(context.Context, *request.ETLJobList) (interface{}, error)
		Create(context.Context, *request.ETLJobCreate) (interface{}, error)
		Read(context.Context, *request.ETLJobRead) (interface{}, error)
		Update(context.Context, *request.ETLJobUpdate) (interface{}, error)
		Delete(context.Context, *request.ETLJobDelete) (interface{}, error)
		Run(context.Context, *request.ETLJobRun) (interface{}, error)
	}

	ETL struct {
		List   func(http.ResponseWriter, *http.Request)
		Create func(http.ResponseWriter, *http.Request)
		Read   func(http.ResponseWriter, *http.Request)
		Update func(http.ResponseWriter, *http.Request)
		Delete func(http.ResponseWriter, *http.Request)
		Run    func(http.ResponseWriter, *http.Request)
	}
)

func NewETL(h ETLAPI) *ETL {
	return &ETL{
		List: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ETLJobList{}
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
		Create: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ETLJobCreate{}
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
		Read: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ETLJobRead{}
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}
			value, err := h.Read(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}
			api.Send(w, r, value)
		},
		Update: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ETLJobUpdate{}
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
			params := &request.ETLJobDelete{}
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
		Run: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ETLJobRun{}
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}
			value, err := h.Run(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}
			api.Send(w, r, value)
		},
	}
}

func (h *ETL) MountRoutes(r chi.Router) {
	r.Get("/namespace/{namespaceID}/etl", h.List)
	r.Post("/namespace/{namespaceID}/etl", h.Create)
	r.Get("/namespace/{namespaceID}/etl/{jobID}", h.Read)
	r.Put("/namespace/{namespaceID}/etl/{jobID}", h.Update)
	r.Delete("/namespace/{namespaceID}/etl/{jobID}", h.Delete)
	r.Post("/namespace/{namespaceID}/etl/{jobID}/run", h.Run)
}
