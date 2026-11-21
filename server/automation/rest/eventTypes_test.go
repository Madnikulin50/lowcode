package rest

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/rest/request"
	"github.com/stretchr/testify/require"
)

func TestEventTypesList_IncludesCustomEvents(t *testing.T) {
	res, err := EventTypes{}.List(context.Background(), &request.EventTypesList{})
	require.NoError(t, err)

	set := res.(eventTypePayload).Set
	find := func(resource, event string) *eventTypeDef {
		for i := range set {
			if set[i].ResourceType == resource && set[i].EventType == event {
				return &set[i]
			}
		}
		return nil
	}

	require.NotNil(t, find("compose", "onManual"), "generated events are still listed")

	f := find("anomaly:finding", "onCreate")
	require.NotNil(t, f)
	require.NotEmpty(t, f.Properties)
	require.Contains(t, f.Constraints, eventTypeConstraintDef{Name: "finding.severity"})

	require.NotNil(t, find("risk:assessment", "onEscalated"))
}
