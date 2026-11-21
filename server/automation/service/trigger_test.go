package service

import (
	"github.com/madnikulin50/lowcode/server/automation/types"
	sysEvent "github.com/madnikulin50/lowcode/server/system/service/event"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateWorkflowTriggersEmpty(t *testing.T) {
	var (
		req = require.New(t)

		issues = validateWorkflowTriggers(
			&types.Workflow{},
		)
	)

	req.Empty(issues)
}
func TestValidateWorkflowTriggersRunAs(t *testing.T) {
	var (
		req = require.New(t)
		soi = sysEvent.SystemOnInterval()

		issues = validateWorkflowTriggers(
			&types.Workflow{},
			&types.Trigger{
				Enabled:      true,
				ResourceType: soi.ResourceType(),
				EventType:    soi.EventType(),
			},
		)
	)

	req.Len(issues, 1)
	req.Contains(issues[0].String(), "requires run-as to be set")
}

func TestValidateWorkflowTriggersSubWorkflow(t *testing.T) {
	var (
		req = require.New(t)

		issues = validateWorkflowTriggers(
			&types.Workflow{Meta: &types.WorkflowMeta{SubWorkflow: true}},
			&types.Trigger{Enabled: true},
		)
	)

	req.Len(issues, 1)
	req.Contains(issues[0].String(), "marked as sub-workflow")
}

func TestValidateWorkflowTriggersRunAsForBackgroundEvents(t *testing.T) {
	for _, tc := range []struct{ resource, event string }{
		{"anomaly:finding", "onCreate"},
		{"anomaly:finding", "onReopen"},
		{"risk:assessment", "onAssessed"},
		{"risk:assessment", "onEscalated"},
	} {
		trigger := &types.Trigger{Enabled: true, ResourceType: tc.resource, EventType: tc.event}

		issues := validateWorkflowTriggers(&types.Workflow{}, trigger)
		require.Len(t, issues, 1, "%s/%s without run-as", tc.resource, tc.event)

		issues = validateWorkflowTriggers(&types.Workflow{RunAs: 7}, trigger)
		require.Empty(t, issues, "%s/%s with run-as", tc.resource, tc.event)
	}
}
