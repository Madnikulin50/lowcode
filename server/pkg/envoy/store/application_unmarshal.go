package store

import (
	"github.com/madnikulin50/lowcode/server/pkg/envoy"
	"github.com/madnikulin50/lowcode/server/pkg/envoy/resource"
	"github.com/madnikulin50/lowcode/server/system/types"
)

func newApplication(app *types.Application, ux *userIndex) *application {
	return &application{
		app: app,
		ux:  ux,
	}
}

func (app *application) MarshalEnvoy() ([]resource.Interface, error) {
	rs := resource.NewApplication(app.app)
	syncUserStamps(rs.Userstamps(), app.ux)

	return envoy.CollectNodes(
		rs,
	)
}
