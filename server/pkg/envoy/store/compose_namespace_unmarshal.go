package store

import (
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/envoy"
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
)

func newComposeNamespace(ns *types.Namespace) *composeNamespace {
	return &composeNamespace{
		ns: ns,
	}
}

func (ns *composeNamespace) MarshalEnvoy() ([]resource.Interface, error) {
	return envoy.CollectNodes(
		resource.NewComposeNamespace(ns.ns),
	)
}
