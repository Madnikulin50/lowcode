package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/pkg/api"
	"github.com/madnikulin50/lowcode/server/pkg/corredor"
	"github.com/madnikulin50/lowcode/server/system/rest/request"
	"github.com/madnikulin50/lowcode/server/system/service/event"
)

type (
	Automation struct{}
)

func (Automation) New() *Automation {
	return &Automation{}
}

func (ctrl *Automation) List(ctx context.Context, r *request.AutomationList) (interface{}, error) {
	return corredor.GenericListHandler(
		ctx,
		corredor.Service(),
		corredor.Filter{
			ResourceTypePrefixes: r.ResourceTypePrefixes,
			ExcludeInvalid:       r.ExcludeInvalid,
			ResourceTypes:        r.ResourceTypes,
			EventTypes:           r.EventTypes,
			ExcludeServerScripts: r.ExcludeServerScripts,
			ExcludeClientScripts: r.ExcludeClientScripts,
		},
		"system",
	)
}

func (ctrl *Automation) Bundle(ctx context.Context, r *request.AutomationBundle) (interface{}, error) {
	return corredor.GenericBundleHandler(
		ctx,
		corredor.Service(),
		r.Bundle,
		r.Type,
		r.Ext,
	)
}

func (ctrl *Automation) TriggerScript(ctx context.Context, r *request.AutomationTriggerScript) (interface{}, error) {
	return api.OK(), corredor.Service().Exec(ctx, r.Script, corredor.ExtendScriptArgs(event.SystemOnManual(), r.Args))
}
