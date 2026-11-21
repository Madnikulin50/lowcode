package service

import (
	"context"
	"fmt"
	"strings"

	atypes "github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

// LoopIncidentApply is a workflow function: route / status / escalate loop incidents.
func LoopIncidentApply() *atypes.Function {
	return &atypes.Function{
		Ref:    "loopIncidentApply",
		Kind:   "function",
		Labels: map[string]string{"compose": "step,workflow", "record": "step,workflow"},
		Meta: &atypes.FunctionMeta{
			Short:       "Incident routing / SLA / status",
			Description: "Assigns a person from incident_routing + retail_staff, changes status, or runs the SLA sweep.",
		},
		Parameters: []*atypes.Param{
			{
				Name:     "action",
				Types:    []string{"String"},
				Required: true,
				Meta:     &atypes.ParamMeta{Label: "route | in_progress | resolved | escalate | escalate_batch"},
			},
			{
				Name:  "record",
				Types: []string{"ID", "ComposeRecord"},
			},
		},
		Results: []*atypes.Param{
			{Name: "message", Types: []string{"String"}},
			{Name: "status", Types: []string{"String"}},
			{Name: "personID", Types: []string{"String"}},
			{Name: "escalated", Types: []string{"UnsignedInteger"}},
		},
		Handler: loopIncidentApply,
	}
}

func loopIncidentApply(ctx context.Context, in *expr.Vars) (*expr.Vars, error) {
	action := "route"
	if in != nil && in.Has("action") {
		aux := expr.Must(expr.Select(in, "action"))
		action = strings.TrimSpace(fmt.Sprint(aux.Get()))
	}

	recordID := uint64(0)
	if in != nil && in.Has("record") {
		aux := expr.Must(expr.Select(in, "record"))
		switch v := aux.Get().(type) {
		case *types.Record:
			if v != nil {
				recordID = v.ID
			}
		default:
			recordID = ParseRecordID(aux.Get())
		}
	}

	out := &expr.Vars{}
	set := func(message, status, person string, escalated uint64) error {
		if err := out.Set("message", expr.Must(expr.NewString(message))); err != nil {
			return err
		}
		if err := out.Set("status", expr.Must(expr.NewString(status))); err != nil {
			return err
		}
		if err := out.Set("personID", expr.Must(expr.NewString(person))); err != nil {
			return err
		}
		return out.Set("escalated", expr.Must(expr.NewUnsignedInteger(escalated)))
	}

	switch strings.ToLower(action) {
	case "route", "assign":
		res, err := RouteIncident(ctx, recordID)
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, res.Status, res.PersonID, 0)

	case "in_progress", "in-progress", "progress":
		res, err := SetIncidentStatus(ctx, recordID, "In Progress")
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, res.Status, res.PersonID, 0)

	case "resolved", "resolve":
		res, err := SetIncidentStatus(ctx, recordID, "Resolved")
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, res.Status, res.PersonID, 0)

	case "closed", "close":
		res, err := SetIncidentStatus(ctx, recordID, "Closed")
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, res.Status, res.PersonID, 0)

	case "escalate":
		res, err := EscalateIncident(ctx, recordID)
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, res.Status, res.PersonID, 1)

	case "escalate_batch", "escalate-batch", "sla":
		res, err := EscalateOverdueIncidents(ctx)
		if err != nil {
			return nil, err
		}
		return out, set(res.Message, "Escalated", "", uint64(res.Escalated))

	default:
		return nil, fmt.Errorf("unknown action %q", action)
	}
}
