package rest

import (
	"github.com/go-chi/chi/v5"

	"github.com/madnikulin50/lowcode/server/discovery/rest/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/options"
)

func MountRoutes(discoveryOpts options.DiscoveryOpt) func(r chi.Router) {
	return func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.HttpTokenValidator("discovery"))

			handlers.NewResources(Resources()).MountRoutes(r)
			handlers.NewFeed(Feed()).MountRoutes(r)
			handlers.NewMappings(Mappings(discoveryOpts)).MountRoutes(r)
		})
	}
}
