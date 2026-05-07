package helpers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/steinfletcher/apitest"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
)

func BindAuthMiddleware(r chi.Router) {
	r.Use(
		auth.HttpTokenVerifier,
		auth.HttpTokenValidator(),
	)
}

func ReqHeaderRawAuthBearer(token []byte) apitest.Intercept {
	return func(req *http.Request) {
		req.Header.Set("Authorization", "Bearer "+string(token))
	}
}
