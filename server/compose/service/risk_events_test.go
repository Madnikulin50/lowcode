package service

import (
	"context"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/eventbus"
	"github.com/madnikulin50/lowcode/server/pkg/riskengine"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
	"github.com/stretchr/testify/require"
)

func TestPublishRiskAssessment_Escalation(t *testing.T) {
	var (
		assessed  = make(chan struct{}, 4)
		escalated = make(chan struct{}, 4)
	)
	assessedReg := eventbus.Service().Register(
		func(context.Context, eventbus.Event) error { assessed <- struct{}{}; return nil },
		eventbus.For(wfevent.ResourceRiskAssessment), eventbus.On(wfevent.OnAssessed))
	escalatedReg := eventbus.Service().Register(
		func(context.Context, eventbus.Event) error { escalated <- struct{}{}; return nil },
		eventbus.For(wfevent.ResourceRiskAssessment), eventbus.On(wfevent.OnEscalated))
	defer eventbus.Service().Unregister(assessedReg)
	defer eventbus.Service().Unregister(escalatedReg)

	binding := &riskengine.RiskSubjectBinding{ID: 1, ModuleID: 2, NamespaceID: 3, ModelID: 4}
	fire := func(level, previous riskengine.Severity) (gotAssessed, gotEscalated bool) {
		publishRiskAssessment(context.Background(), binding, &riskengine.RiskAssessment{ID: 9, BindingID: 1, Level: level}, previous)

		drain := func(ch chan struct{}) bool {
			select {
			case <-ch:
				return true
			case <-time.After(200 * time.Millisecond):
				return false
			}
		}
		return drain(assessed), drain(escalated)
	}

	a, e := fire(riskengine.SeverityHigh, riskengine.SeverityLow)
	require.True(t, a)
	require.True(t, e, "low -> high is an escalation")

	a, e = fire(riskengine.SeverityHigh, riskengine.SeverityHigh)
	require.True(t, a)
	require.False(t, e, "unchanged level is not an escalation")

	a, e = fire(riskengine.SeverityLow, riskengine.SeverityCritical)
	require.True(t, a)
	require.False(t, e, "a drop is not an escalation")

	a, e = fire(riskengine.SeverityMedium, "")
	require.True(t, a)
	require.True(t, e, "a first assessment above low counts as an escalation")

	a, e = fire(riskengine.SeverityLow, "")
	require.True(t, a)
	require.False(t, e, "a first low assessment does not")
}
