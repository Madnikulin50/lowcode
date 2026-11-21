package store

import (
	"github.com/madnikulin50/lowcode/server/pkg/envoy"
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
	"github.com/madnikulin50/lowcode/server/system/types"
)

func newRole(rl *types.Role) *role {
	return &role{
		rl: rl,
	}
}

func (rl *role) MarshalEnvoy() ([]resource.Interface, error) {
	return envoy.CollectNodes(
		resource.NewRole(rl.rl),
	)
}
