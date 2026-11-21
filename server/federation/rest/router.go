package rest

import (
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/options"

	"github.com/madnikulin50/lowcode/server/federation/rest/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
)

func MountRoutes(opts options.LimitOpt) func(r chi.Router) {
	return func(r chi.Router) {
		r.Group(func(r chi.Router) {
			handlers.NewNodeHandshake(NodeHandshake{}.New()).MountRoutes(r)
		})

		// Protect all _private_ routes
		r.Group(func(r chi.Router) {
			r.Use(auth.HttpTokenValidator("api"))

			handlers.NewPermissions(Permissions{}.New()).MountRoutes(r)

			handlers.NewNode(Node{}.New()).MountRoutes(r)
			handlers.NewManageStructure((ManageStructure{}.New())).MountRoutes(r)

			handlers.NewSyncData((SyncData{}.New(opts))).MountRoutes(r)
			handlers.NewSyncStructure((SyncStructure{}.New())).MountRoutes(r)
		})
	}
}
