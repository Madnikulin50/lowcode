package app

import (
	"context"

	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/rbac"
)

func (app *CortezaApp) InitExpr(ctx context.Context) (err error) {
	expr.Init(rbac.AllFunctions, expr.AllFunctions)

	return
}
