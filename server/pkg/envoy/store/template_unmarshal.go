package store

import (
	"github.com/madnikulin50/lowcode/server/pkg/envoy"
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
	"github.com/madnikulin50/lowcode/server/system/types"
)

func newTemplate(t *types.Template) *template {
	return &template{
		t: t,
	}
}

func (u *template) MarshalEnvoy() ([]resource.Interface, error) {
	return envoy.CollectNodes(
		resource.NewTemplate(u.t),
	)
}
