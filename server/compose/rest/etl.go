package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/compose/rest/request"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
)

type (
	etlPayload struct {
		*types.ETLJob
	}

	etlSetPayload struct {
		Set    types.ETLJobSet    `json:"set"`
		Filter types.ETLJobFilter `json:"filter"`
	}

	ETL struct {
		svc service.ETLService
	}
)

func (ETL) New() *ETL {
	return &ETL{
		svc: service.DefaultETL,
	}
}

func (ctrl *ETL) List(ctx context.Context, r *request.ETLJobList) (interface{}, error) {
	var err error
	f := types.ETLJobFilter{
		NamespaceID: r.NamespaceID,
		ModuleID:    r.ModuleID,
		Query:       r.Query,
	}
	if f.Paging, err = filter.NewPaging(r.Limit, r.PageCursor); err != nil {
		return nil, err
	}

	set, fOut, err := ctrl.svc.Find(ctx, f)
	if err != nil {
		return nil, err
	}

	return &etlSetPayload{Set: set, Filter: fOut}, nil
}

func (ctrl *ETL) Create(ctx context.Context, r *request.ETLJobCreate) (interface{}, error) {
	j := &types.ETLJob{
		ModuleID:    r.ModuleID,
		NamespaceID: r.NamespaceID,
		Name:        r.Name,
		Enabled:     r.Enabled,
		Schedule:    r.Schedule,
		Source:      r.Source,
	}

	j, err := ctrl.svc.Create(ctx, j)
	if err != nil {
		return nil, err
	}

	return &etlPayload{ETLJob: j}, nil
}

func (ctrl *ETL) Read(ctx context.Context, r *request.ETLJobRead) (interface{}, error) {
	j, err := ctrl.svc.FindByID(ctx, r.JobID)
	if err != nil {
		return nil, err
	}
	return &etlPayload{ETLJob: j}, nil
}

func (ctrl *ETL) Update(ctx context.Context, r *request.ETLJobUpdate) (interface{}, error) {
	j := &types.ETLJob{
		ID:       r.JobID,
		Name:     r.Name,
		Enabled:  r.Enabled,
		Schedule: r.Schedule,
		Source:   r.Source,
	}

	j, err := ctrl.svc.Update(ctx, j)
	if err != nil {
		return nil, err
	}

	return &etlPayload{ETLJob: j}, nil
}

func (ctrl *ETL) Delete(ctx context.Context, r *request.ETLJobDelete) (interface{}, error) {
	return nil, ctrl.svc.Delete(ctx, r.JobID)
}

func (ctrl *ETL) Run(ctx context.Context, r *request.ETLJobRun) (interface{}, error) {
	return nil, ctrl.svc.Run(ctx, r.JobID)
}
