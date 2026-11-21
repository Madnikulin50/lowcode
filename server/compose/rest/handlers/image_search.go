package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/pkg/api"
)

type ImageSearchAPI interface {
	Search(context.Context, *request.ImageSearchSearch) (interface{}, error)
}

type ImageSearch struct {
	Search func(http.ResponseWriter, *http.Request)
}

func NewImageSearch(h ImageSearchAPI) *ImageSearch {
	return &ImageSearch{
		Search: func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			params := request.NewImageSearchSearch()
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
	}
}

func (h *ImageSearch) MountRoutes(r chi.Router, middlewares ...func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(middlewares...)
		r.Get("/image/search", h.Search)
	})
}
