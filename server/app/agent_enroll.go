package app

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/options"
)

// mountAgentEnroll wires a minimal, opt-in self-enrollment endpoint for the
// bundled agents (backup/cmdb/invest/...). Given a shared secret that only
// needs to be the same string in this server's env and in each agent's
// (AGENT_SHARED_SECRET), it hands back a working API token for the
// server's built-in "corteza-service" identity — no live user session to
// steal a refresh token from, no OAuth2 client to register by hand, no
// external provisioning script. Works from a cold `docker-compose up`.
//
// Entirely optional: if AGENT_SHARED_SECRET is unset, this route is not
// mounted at all and agents keep working exactly as before (unauthenticated,
// or with a manually-set TOKEN).
func (app *CortezaApp) mountAgentEnroll(r chi.Router, baseUrl string) {
	secret := strings.TrimSpace(os.Getenv("AGENT_SHARED_SECRET"))
	if secret == "" {
		return
	}

	r.Post(options.CleanBase(baseUrl, "agents/enroll"), func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		given := strings.TrimSpace(r.Header.Get("X-Agent-Secret"))
		if len(given) != len(secret) || subtle.ConstantTimeCompare([]byte(given), []byte(secret)) != 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid secret"})
			return
		}

		ident := auth.ServiceUserOrNil()
		if ident == nil || auth.TokenIssuer == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "server not ready"})
			return
		}

		tok, err := auth.TokenIssuer.Issue(r.Context(),
			auth.WithIdentity(ident),
			auth.WithScope("api", "profile"),
			auth.WithAudience("agent"),
			auth.WithExpiration(365*24*time.Hour),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not issue token"})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": string(tok)})
	})

	app.Log.Info("agent self-enrollment endpoint mounted (AGENT_SHARED_SECRET set)")
}
