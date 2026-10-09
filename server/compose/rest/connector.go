package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
)

func (Connector) New() *Connector {
	return &Connector{}
}

type Connector struct{}

func (ctrl *Connector) Test(ctx context.Context, r *request.ConnectorTest) (interface{}, error) {
	c := service.Connector()
	if err := c.Test(ctx, r.Connector); err != nil {
		return map[string]any{"success": false, "error": err.Error()}, nil
	}
	return map[string]any{"success": true}, nil
}
