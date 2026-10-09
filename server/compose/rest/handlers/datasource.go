package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type (
	DatasourceAPI interface {
		Preview(context.Context, *request.DatasourcePreview) (interface{}, error)
	}

	Datasource struct {
		Preview func(http.ResponseWriter, *http.Request)
	}
)

func NewDatasource(h DatasourceAPI) *Datasource {
	return &Datasource{
		Preview: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.DatasourcePreview{}
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}
			value, err := h.Preview(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}
			api.Send(w, r, value)
		},
	}
}

func (h *Datasource) MountRoutes(r chi.Router) {
	r.Post("/datasource/preview", h.Preview)
}
