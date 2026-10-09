package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
)

func (Datasource) New() *Datasource {
	return &Datasource{}
}

type Datasource struct{}

func (ctrl *Datasource) Preview(ctx context.Context, r *request.DatasourcePreview) (interface{}, error) {
	return service.DefaultRecord.DatasourcePreview(ctx, r.NamespaceID, r.ModuleID, r.Datasource, r.Limit)
}
