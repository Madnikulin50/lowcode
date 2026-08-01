package app

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/assets"
	automationRest "github.com/madnikulin50/lowcode/server/automation/rest"
	composeRest "github.com/madnikulin50/lowcode/server/compose/rest"
	discoveryRest "github.com/madnikulin50/lowcode/server/discovery/rest"
	"github.com/madnikulin50/lowcode/server/docs"
	federationRest "github.com/madnikulin50/lowcode/server/federation/rest"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/pkg/webapp"
	systemRest "github.com/madnikulin50/lowcode/server/system/rest"
	"github.com/madnikulin50/lowcode/server/system/scim"
	"github.com/madnikulin50/lowcode/server/system/service"
	"go.uber.org/zap"
)

func (app *CortezaApp) mountHttpRoutes(r chi.Router) {
	var (
		ho = app.Opt.HTTPServer
	)

	func() {
		// asset serving has some overlap with auth assets, web-console and webapp serving
		// and might be joined with one or more of them in the later version

		var (
			url   = options.CleanBase(ho.BaseUrl, "assets")
			aPath = ho.AssetsPath
			files = assets.Files(app.Log, aPath)
		)

		r.Handle(url+"/*", http.StripPrefix(url+"/", http.FileServer(http.FS(files))))

		if aPath != "" {
			app.Log.Info("custom web assets mounted", zap.String("url", url), zap.String("path", aPath))
		} else {
			app.Log.Info("embedded web assets mounted", zap.String("url", url))
		}
	}()

	func() {
		if ho.WebappEnabled && ho.ApiEnabled && ho.ApiBaseUrl == ho.WebappBaseUrl {
			app.Log.
				Warn("client web applications and api can not use the same base URL: '" + ho.WebappBaseUrl + "'")
			ho.WebappEnabled = false
		}

		if !ho.WebappEnabled {
			r.Get(options.CleanBase(ho.BaseUrl, "custom.css"), func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "text/css")

				stylesheet := service.FetchCSS()

				_, _ = fmt.Fprint(w, stylesheet)
			})
			r.Get(options.CleanBase(ho.WebappBaseUrl, "custom.css"), func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Content-Type", "text/css")
				//DB_DSN=postgres://postgres:Zse45rdx@127.0.0.1:5432/test3?sslmode\=disable
				stylesheet := service.FetchCSS()

				_, _ = fmt.Fprint(w, stylesheet)
			})

			/*r.Get(options.CleanBase(ho.BaseUrl, "config.js"), func(w http.ResponseWriter, r *http.Request) {

				// Assure the content-type
				// The presence of the X-Content-Type-Options: nosniff header breaks web applications
				w.Header().Add("Content-Type", "text/javascript")

				const line = "window.%s = '%s';\n"
				_, _ = fmt.Fprintf(w, line, "CortezaAPI", options.CleanBase(ho.ApiBaseUrl, ho.ApiBaseUrl, "api"))
				_, _ = fmt.Fprintf(w, line, "CortezaAuth", options.CleanBase(ho.ApiBaseUrl, ho.ApiBaseUrl, "auth"))
			})*/

			app.Log.Info("client web applications disabled")
			return
		}

		r.Route(options.CleanBase(ho.WebappBaseUrl), webapp.MakeWebappServer(app.Log, ho, app.Opt.Auth, app.Opt.Discovery, app.Opt.Sentry))

		app.Log.Info(
			"client web applications enabled",
			zap.String("baseUrl", options.CleanBase(ho.BaseUrl, ho.WebappBaseUrl)),
			zap.String("baseDir", ho.WebappBaseDir),
			zap.Strings("apps", strings.Split(ho.WebappList, ",")),
		)
	}()

	// Auth server
	app.AuthService.MountHttpRoutes(ho.BaseUrl, r)

	func() {
		if !ho.ApiEnabled {
			app.Log.Info("JSON REST API disabled")
			return
		}

		r.Route(options.CleanBase(ho.ApiBaseUrl), func(r chi.Router) {
			var fullpathAPI = "/" + strings.TrimPrefix(options.CleanBase(ho.BaseUrl, ho.ApiBaseUrl), "/")

			app.Log.Info(
				"JSON REST API enabled",
				zap.String("baseUrl", fullpathAPI),
			)

			r.Route("/system", systemRest.MountRoutes())
			r.Route("/automation", automationRest.MountRoutes())
			r.Route("/compose", composeRest.MountRoutes())
			r.Route("/websocket", app.WsServer.MountRoutes)

			if app.Opt.Discovery.Enabled {
				r.Route("/discovery", discoveryRest.MountRoutes(app.Opt.Discovery))
			}

			if app.Opt.Federation.Enabled {
				r.Route("/federation", federationRest.MountRoutes(app.Opt.Limit))
			}

			var fullpathDocs = options.CleanBase(ho.BaseUrl, ho.ApiBaseUrl, "docs")
			app.Log.Info(
				"API docs enabled",
				zap.String("baseUrl", fullpathDocs),
			)

			r.Handle("/docs", http.RedirectHandler(fullpathDocs+"/", http.StatusPermanentRedirect))
			r.Handle("/docs*", http.StripPrefix(fullpathDocs, http.FileServer(docs.GetFS())))

			var fullpathGateway = options.CleanBase(ho.BaseUrl, ho.ApiBaseUrl, "gateway")
			r.Handle("/gateway*", http.StripPrefix(fullpathGateway, app.ApigwService))
		})
	}()

	func() {
		if !app.Opt.SCIM.Enabled {
			return
		}

		if app.Opt.SCIM.Secret == "" {
			app.Log.
				Error("SCIM secret empty")
		}

		var (
			baseUrl         = app.Opt.SCIM.BaseURL
			extIdValidation *regexp.Regexp
			err             error
		)

		if len(app.Opt.SCIM.ExternalIdValidation) > 0 {
			extIdValidation, err = regexp.Compile(app.Opt.SCIM.ExternalIdValidation)
		}

		if err != nil {
			app.Log.Error("failed to compile SCIM external ID validation", zap.Error(err))
			return
		}

		app.Log.Debug(
			"SCIM enabled",
			zap.String("baseUrl", path.Join(app.Opt.HTTPServer.BaseUrl, baseUrl)),
			logger.Mask("secret", app.Opt.SCIM.Secret),
		)

		r.Route(baseUrl, func(r chi.Router) {
			if !app.Opt.Environment.IsDevelopment() {
				r.Use(scim.Guard(app.Opt.SCIM))
			}

			scim.Routes(r, scim.Config{
				ExternalIdAsPrimary: app.Opt.SCIM.ExternalIdAsPrimary,
				ExternalIdValidator: extIdValidation,
			})
		})
	}()

	func() {
		r.Handle("/.well-known/openid-configuration", app.AuthService.WellKnownOpenIDConfiguration())
	}()
}
