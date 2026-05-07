package app

import (
	"strings"

	searcherRest "github.com/madnikulin50/lowcode/extra/server-discovery/searcher/rest"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (app *CortezaDiscoveryApp) MountHttpRoutes(r chi.Router) {
	var (
		ho = app.Opt.HTTPServer
	)

	func() {
		if !app.Opt.Searcher.Enabled {
			app.Log.Info("JSON REST API disabled")
			return
		}

		r.Route(options.CleanBase(ho.ApiBaseUrl), func(r chi.Router) {
			var fullPathAPI = "/" + strings.TrimPrefix(options.CleanBase(ho.BaseUrl, ho.ApiBaseUrl), "/")

			app.Log.Info(
				"JSON REST API enabled",
				zap.String("baseUrl", fullPathAPI),
			)

			r.Route("/", searcherRest.MountRoutes(app.Log))
		})
	}()

}
