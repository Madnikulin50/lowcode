package aiagent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContextWithModel(t *testing.T) {
	ctx := context.Background()
	require.Equal(t, "", modelFromContext(ctx))
	require.Equal(t, "", modelFromContext(ContextWithModel(ctx, "  ")), "blank is no override")
	require.Equal(t, "qwen3:8b", modelFromContext(ContextWithModel(ctx, " qwen3:8b ")))
}

// The override has to reach the client the agent actually calls - an earlier
// approach (a <model> tag in the prompt) was silently ignored by the agent
// runtime, so assert on the client's model, not on the context.
func TestAgentClientForRun_HonoursModelOverride(t *testing.T) {
	agent := New(nil, AgentConfig{Name: "a", Model: "configured-model"})

	require.Equal(t, "configured-model", agent.clientForRun(context.Background()).Model())
	require.Equal(t, "per-call-model", agent.clientForRun(ContextWithModel(context.Background(), "per-call-model")).Model())
}

func TestRunStructured_PassesModelOnTheContext(t *testing.T) {
	var seen string
	run := func(ctx context.Context, _, _ string, _ bool) (*AgentResult, error) {
		seen = modelFromContext(ctx)
		return &AgentResult{Success: true, Output: `{"a":"b"}`}, nil
	}

	_, err := RunStructured(context.Background(), run, StructuredRequest{Agent: "x", Prompt: "p", Model: "m1"})
	require.NoError(t, err)
	require.Equal(t, "m1", seen)
}
