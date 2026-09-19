package aiagent

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// These exercise the exact mechanism Agent.RunConfirmed now wires into every
// caller (chat, MCP, and rulechain's ai/ai.operation nodes) by default:
// NeedsConfirmFromToolDefs + Options.Confirmed, run through the same fakeModel
// harness runtime_test.go already uses (Agent itself hardcodes *chat.Client,
// which can't be swapped for a fake without a live model - see agent.go).

func TestGuardrail_MutatingToolBlockedWithoutConfirmation(t *testing.T) {
	var executed bool
	mutating := chat.ToolDef{
		Name:     "delete_thing",
		Mutating: true,
		Params:   []chat.ParamDef{{Name: "id", Type: "string"}},
		Handler: func(_ context.Context, params map[string]string) string {
			executed = true
			return "deleted"
		},
	}
	tools := []chat.ToolDef{mutating}

	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `<tool name="delete_thing"><param name="id">42</param></tool>`},
		},
	}

	out := Run(context.Background(), Options{
		Client:       m,
		Messages:     []*schema.Message{schema.UserMessage("delete it")},
		Tools:        tools,
		MaxSteps:     3,
		NeedsConfirm: NeedsConfirmFromToolDefs(tools),
		// Confirmed left false: this is the "no human in the loop" case a
		// rulechain node or scheduled job is in.
	})

	if !out.ConfirmNeeded {
		t.Fatalf("expected ConfirmNeeded=true, got %+v", out)
	}
	if executed {
		t.Fatal("mutating tool handler must not run without confirmation")
	}
	if len(out.ConfirmCalls) != 1 || out.ConfirmCalls[0].Name != "delete_thing" {
		t.Fatalf("unexpected ConfirmCalls: %#v", out.ConfirmCalls)
	}
}

func TestGuardrail_MutatingToolRunsOnceConfirmed(t *testing.T) {
	var executed bool
	mutating := chat.ToolDef{
		Name:     "delete_thing",
		Mutating: true,
		Handler: func(_ context.Context, params map[string]string) string {
			executed = true
			return "deleted"
		},
	}
	tools := []chat.ToolDef{mutating}

	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `<tool name="delete_thing"><param name="id">42</param></tool>`},
			{content: "FINAL: done"},
		},
	}

	out := Run(context.Background(), Options{
		Client:       m,
		Messages:     []*schema.Message{schema.UserMessage("delete it")},
		Tools:        tools,
		MaxSteps:     3,
		NeedsConfirm: NeedsConfirmFromToolDefs(tools),
		Confirmed:    true, // stands in for the caller's explicit allowMutating:true
	})

	if out.ConfirmNeeded {
		t.Fatalf("expected ConfirmNeeded=false once confirmed, got %+v", out)
	}
	if !executed {
		t.Fatal("expected the mutating tool handler to run once confirmed")
	}
}

func TestGuardrail_ReadOnlyToolNeverNeedsConfirmation(t *testing.T) {
	var executed bool
	readOnly := chat.ToolDef{
		Name: "search_thing", // no Mutating flag
		Handler: func(_ context.Context, params map[string]string) string {
			executed = true
			return "found"
		},
	}
	tools := []chat.ToolDef{readOnly}

	m := &fakeModel{
		tools: true,
		gen: []genTurn{
			{content: `<tool name="search_thing"><param name="q">x</param></tool>`},
			{content: "FINAL: done"},
		},
	}

	out := Run(context.Background(), Options{
		Client:       m,
		Messages:     []*schema.Message{schema.UserMessage("search")},
		Tools:        tools,
		MaxSteps:     3,
		NeedsConfirm: NeedsConfirmFromToolDefs(tools),
	})

	if out.ConfirmNeeded {
		t.Fatalf("a read-only tool should never require confirmation, got %+v", out)
	}
	if !executed {
		t.Fatal("expected the read-only tool handler to run")
	}
}

// Registry.RunAgent/RunAgentConfirmed and Agent.Run/RunConfirmed are thin
// wrappers around this same mechanism (see agent.go, registry.go) - covered
// by build + the existing aiagent suite continuing to pass unchanged;
// there's no seam to swap in a fake model at that level (Agent hardcodes
// *chat.Client), so the wiring itself is exercised at the Options level here.
