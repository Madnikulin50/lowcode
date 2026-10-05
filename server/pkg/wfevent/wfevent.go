// Package wfevent holds workflow trigger events that are raised by code
// outside the generated event sets (compose/system): anomaly findings and
// risk assessments. A workflow subscribes with an ordinary trigger
// (resource type + event type + constraints) and receives the event data as
// its input variables.
//
// Events carry plain data, so producers (anomaly scanner, risk engine) do
// not depend on the automation component and vice versa.
package wfevent

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

const (
	ResourceAnomalyFinding = "anomaly:finding"
	ResourceRiskAssessment = "risk:assessment"

	OnCreate    = "onCreate"
	OnReopen    = "onReopen"
	OnAssessed  = "onAssessed"
	OnEscalated = "onEscalated"

	// variable names the workflow receives
	PropFinding    = "finding"
	PropRule       = "rule"
	PropAssessment = "assessment"
	PropBinding    = "binding"
)

type (
	// Event is a workflow trigger event described by plain data. props maps
	// a variable name to its fields, e.g. "finding" -> {severity: "high"}.
	Event struct {
		resource  string
		eventType string
		props     map[string]map[string]interface{}
	}

	// Def describes an event type for the workflow editor.
	Def struct {
		ResourceType string
		EventType    string
		Properties   []Prop
		Constraints  []string
	}

	Prop struct {
		Name string
		Type string
	}
)

func New(resource, eventType string, props map[string]map[string]interface{}) *Event {
	return &Event{resource: resource, eventType: eventType, props: props}
}

func (e Event) ResourceType() string { return e.resource }
func (e Event) EventType() string    { return e.eventType }

// Match resolves a constraint named "<prop>.<field>" (e.g. finding.severity)
// against the event's data. Unknown names never match.
func (e Event) Match(c eventbus.ConstraintMatcher) bool {
	var prop, field string
	for i := 0; i < len(c.Name()); i++ {
		if c.Name()[i] == '.' {
			prop, field = c.Name()[:i], c.Name()[i+1:]
			break
		}
	}
	if prop == "" {
		return false
	}

	fields, ok := e.props[prop]
	if !ok {
		return false
	}
	v, ok := fields[field]
	if !ok || v == nil {
		return false
	}
	return c.Match(fmt.Sprintf("%v", v))
}

// EncodeVars turns the event into workflow input variables.
func (e Event) EncodeVars() (*expr.Vars, error) {
	vars := &expr.Vars{}
	for name, fields := range e.props {
		if err := vars.Set(name, fields); err != nil {
			return nil, fmt.Errorf("could not encode %q: %w", name, err)
		}
	}
	return vars, nil
}

// Emit dispatches the event to subscribed workflows without waiting for them,
// so a slow workflow cannot stall the producer (a scan loop, say).
//
// Producers such as the anomaly scanner run without a user; workflows need an
// invoker, so the system service user is supplied when the context has none.
func Emit(ctx context.Context, e *Event) {
	if e == nil {
		return
	}
	if auth.GetIdentityFromContext(ctx) == nil {
		if u := auth.ServiceUserOrNil(); u != nil {
			ctx = auth.SetIdentityToContext(ctx, u)
		}
	}
	eventbus.Service().Dispatch(ctx, e)
}

// Definitions lists the events for the workflow editor's trigger picker.
func Definitions() []Def {
	finding := []Prop{{PropFinding, "Any"}, {PropRule, "Any"}}
	findingConstraints := []string{
		"finding.severity", "finding.status", "finding.moduleID", "finding.namespaceID", "finding.ruleID",
		"rule.detector", "rule.field",
	}
	risk := []Prop{{PropAssessment, "Any"}, {PropBinding, "Any"}}
	riskConstraints := []string{
		"assessment.level", "assessment.previousLevel", "assessment.bindingID", "assessment.modelID",
		"binding.moduleID", "binding.namespaceID",
	}

	return []Def{
		{ResourceAnomalyFinding, OnCreate, finding, findingConstraints},
		{ResourceAnomalyFinding, OnReopen, finding, findingConstraints},
		{ResourceRiskAssessment, OnAssessed, risk, riskConstraints},
		{ResourceRiskAssessment, OnEscalated, risk, riskConstraints},
	}
}

// RequireRunAs lists the events raised with no user behind them: a workflow
// triggered by them must have run-as configured.
func RequireRunAs() []eventbus.Event {
	out := make([]eventbus.Event, 0, 4)
	for _, d := range Definitions() {
		out = append(out, New(d.ResourceType, d.EventType, nil))
	}
	return out
}
