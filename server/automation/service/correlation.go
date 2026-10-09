package service

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

// ResolveCorrelation finds the pending prompt whose "correlationKey"
// argument matches key, and resumes that specific suspended session with
// input.
//
// This is the automated counterpart to a person answering the interactive
// pending-prompts modal (CPromptModal.vue / wf-prompts store) - here the
// trigger is an external event (typically a Kafka/RabbitMQ message matched
// by a broker subscriber, see pkg/rulesgo.BrokerSubscriber) instead of a
// click. It's what a compiled BPMN message intermediate catch event
// (bpmn_compile.go's compileIntermediateCatchEvent) is waiting for: that
// compiler emits a plain prompt step carrying a correlationKey argument,
// reusing the engine's existing suspend/resume machinery unchanged - the
// only new thing here is *what* resumes it.
//
// Resume runs as the prompt's own owner identity (whoever the process
// assigned this wait to via the userTask/message event's "owner"
// argument), never as an escalated system identity - the same guardrail
// this session established for the ai/ai.operation nodes: an automated
// resolver never gets to act with more authority than the human it's
// standing in for. A prompt with no owner set can't be resolved this way
// (see the error below) - assign an owner if you want message-driven
// resume.
func ResolveCorrelation(ctx context.Context, key string, input *expr.Vars) error {
	if key == "" {
		return fmt.Errorf("correlation: empty key")
	}
	if DefaultSession == nil {
		return fmt.Errorf("correlation: session service not available")
	}

	for _, p := range DefaultSession.AllPendingPrompts() {
		if p.Payload == nil || !p.Payload.Has("correlationKey") {
			continue
		}
		v, err := expr.Select(p.Payload, "correlationKey")
		if err != nil {
			continue
		}
		if fmt.Sprintf("%v", v.Get()) != key {
			continue
		}
		if p.OwnerId == 0 {
			return fmt.Errorf("correlation: pending prompt for key %q has no owner - set an \"owner\" property alongside correlationKey to allow message-driven resume", key)
		}

		return DefaultSession.Resume(p.SessionID, p.StateID, auth.Authenticated(p.OwnerId), input)
	}

	return fmt.Errorf("correlation: no pending prompt is waiting on key %q", key)
}
