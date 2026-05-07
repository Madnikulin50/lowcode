package rest

import (
	"context"
	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/automation/service"
)

type (
	Type struct {
		reg interface {
			Types() []string
		}
	}

	typeSetPayload struct {
		Set []string `json:"set"`
	}
)

func (Type) New() *Type {
	ctrl := &Type{reg: service.Registry()}
	return ctrl
}

func (ctrl Type) List(_ context.Context, _ *request.TypeList) (interface{}, error) {
	return typeSetPayload{Set: ctrl.reg.Types()}, nil
}
