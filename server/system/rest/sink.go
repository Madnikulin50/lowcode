package rest

import (
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"net/http"
)

type Sink struct {
	svc interface {
		ProcessRequest(w http.ResponseWriter, r *http.Request)
	}

	sign auth.Signer
}

func (ctrl *Sink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl.svc.ProcessRequest(w, r)
}
