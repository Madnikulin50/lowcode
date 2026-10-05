package aiagent

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/stretchr/testify/require"
)

func TestCallTrace_AccumulatesAcrossAttempts(t *testing.T) {
	answers := []*AgentResult{
		{Success: true, Output: "not json", Model: "m1", LLMCalls: 1, PromptTokens: 100, CompletionTokens: 10,
			Steps: []AgentStep{{Tools: []chat.ToolCall{{Name: "search_records"}}}}},
		{Success: true, Output: `{"total": 5}`, Model: "m1", LLMCalls: 2, PromptTokens: 120, CompletionTokens: 20},
	}
	i := 0
	run := func(context.Context, string, string, bool) (*AgentResult, error) {
		a := answers[i]
		i++
		return a, nil
	}

	res, err := RunStructured(context.Background(), run, StructuredRequest{
		Agent: "a", Prompt: "p", OutputSchema: map[string]string{"total": "number"}, MaxRetries: 1,
	})
	require.NoError(t, err)

	tr := res.Trace
	require.Equal(t, "a", tr.Agent)
	require.Equal(t, "m1", tr.Model)
	require.Equal(t, 2, tr.Attempts)
	require.Equal(t, 3, tr.LLMCalls)
	require.Equal(t, 220, tr.PromptTokens)
	require.Equal(t, 30, tr.CompletionTokens)
	require.Equal(t, []string{"search_records"}, tr.Tools)
	require.Contains(t, tr.Prompt, "Required output", "the prompt that was sent, not only the instruction")
	require.Equal(t, `{"total": 5}`, tr.Response, "the last response")
	require.Empty(t, tr.Error)
}

func TestCallTrace_SurvivesFailure(t *testing.T) {
	run := func(context.Context, string, string, bool) (*AgentResult, error) {
		return &AgentResult{Success: true, Output: "I refuse to answer in JSON", LLMCalls: 1}, nil
	}

	res, err := RunStructured(context.Background(), run, StructuredRequest{
		Agent: "a", Prompt: "p", OutputSchema: map[string]string{"x": "string"}, MaxRetries: 1,
	})
	require.Error(t, err)
	require.NotNil(t, res, "a failed step must still return its trace")
	require.Equal(t, 2, res.Trace.Attempts)
	require.Equal(t, "I refuse to answer in JSON", res.Trace.Response)
	require.Contains(t, res.Trace.Error, "did not produce a valid response")
}

func TestCallTrace_MissingInputsStillTraced(t *testing.T) {
	res, err := RunStructured(context.Background(), nil, StructuredRequest{Prompt: "p"})
	require.Error(t, err)
	require.NotNil(t, res)
	require.Contains(t, res.Trace.Error, "agent is required")
}

func TestPreview(t *testing.T) {
	require.Equal(t, "short", Preview("  short \n"))

	long := strings.Repeat("я", TracePreviewLimit+50)
	got := Preview(long)
	require.True(t, strings.HasSuffix(got, "…"))
	require.Equal(t, TracePreviewLimit+1, len([]rune(got)), "cut by characters, not bytes")
}

func TestCallTrace_Map(t *testing.T) {
	m := CallTrace{Agent: "a", Attempts: 1, DurationMs: 12, PromptTokens: 7}.Map()
	require.Equal(t, "a", m["agent"])
	require.EqualValues(t, 12, m["durationMs"])
	require.EqualValues(t, 7, m["promptTokens"])
	require.NotContains(t, m, "error", "empty fields are omitted")
}
