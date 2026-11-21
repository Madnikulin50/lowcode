package rest

import (
	"context"

	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
)

type (
	EventTypes struct {
		reg interface {
			Types() []string
		}
	}

	eventTypePayload struct {
		Set []eventTypeDef `json:"set"`
	}

	eventTypeDef struct {
		ResourceType string                   `json:"resourceType"`
		EventType    string                   `json:"eventType"`
		Properties   []eventTypePropertyDef   `json:"properties"`
		Constraints  []eventTypeConstraintDef `json:"constraints"`
	}

	eventTypePropertyDef struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		Immutable bool   `json:"immutable"`
	}

	eventTypeConstraintDef struct {
		Name string `json:"name"`
	}
)

func (EventTypes) New() *EventTypes {
	ctrl := &EventTypes{reg: service.Registry()}
	return ctrl
}

func (ctrl EventTypes) List(_ context.Context, _ *request.EventTypesList) (interface{}, error) {
	return eventTypePayload{Set: append(getEventTypeDefinitions(), customEventTypeDefinitions()...)}, nil
}

// customEventTypeDefinitions are trigger events raised by hand-written code
// (pkg/wfevent) rather than generated from events.yaml; they live outside
// eventTypes.gen.go so that regenerating it cannot drop them.
func customEventTypeDefinitions() []eventTypeDef {
	defs := wfevent.Definitions()
	out := make([]eventTypeDef, 0, len(defs))
	for _, d := range defs {
		def := eventTypeDef{
			ResourceType: d.ResourceType,
			EventType:    d.EventType,
			Properties:   make([]eventTypePropertyDef, 0, len(d.Properties)),
			Constraints:  make([]eventTypeConstraintDef, 0, len(d.Constraints)),
		}
		for _, p := range d.Properties {
			def.Properties = append(def.Properties, eventTypePropertyDef{Name: p.Name, Type: p.Type, Immutable: true})
		}
		for _, c := range d.Constraints {
			def.Constraints = append(def.Constraints, eventTypeConstraintDef{Name: c})
		}
		out = append(out, def)
	}
	return out
}
