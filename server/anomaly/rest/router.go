package rest

import (
	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/anomaly/rest/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
)

func MountRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.HttpTokenValidator("api"))

			handlers.NewRule(Rule{}.New()).MountRoutes(r)
			handlers.NewFinding(Finding{}.New()).MountRoutes(r)
		})
	}
}
