package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type (
	ConnectorAPI interface {
		Test(context.Context, *request.ConnectorTest) (interface{}, error)
	}

	Connector struct {
		Test func(http.ResponseWriter, *http.Request)
	}
)

func NewConnector(h ConnectorAPI) *Connector {
	return &Connector{
		Test: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := &request.ConnectorTest{}
			if err := params.Fill(r); err != nil {
				api.Send(w, r, err)
				return
			}
			value, err := h.Test(r.Context(), params)
			if err != nil {
				api.Send(w, r, err)
				return
			}
			api.Send(w, r, value)
		},
	}
}

func (h *Connector) MountRoutes(r chi.Router) {
	r.Post("/connector/test", h.Test)
}
