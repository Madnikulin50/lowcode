package types

import (
	"context"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/wfexec"
)

type (
	errorHandlerStep struct {
		wfexec.StepIdentifier
		handler wfexec.Step
		results *expr.Vars
	}
)

func ErrorHandlerStep(h wfexec.Step, rr *expr.Vars) *errorHandlerStep {
	return &errorHandlerStep{handler: h, results: rr}
}

// Exec errorHandler step
func (h errorHandlerStep) Exec(ctx context.Context, r *wfexec.ExecRequest) (wfexec.ExecResponse, error) {
	return wfexec.ErrorHandler(h.handler, h.results), nil
}
