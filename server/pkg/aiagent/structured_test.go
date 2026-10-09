package aiagent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// structured.go is the single contract every non-interactive AI surface shares
// (rulesgo ai.operation and the automation workflow ai* functions). These
// tests pin its parsing, schema validation, retry and tracing behaviour so a
// refactor on one surface cannot silently change what the others promise.

func TestExtractJSONObject(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"plain object", `{"a":1}`, `{"a":1}`, true},
		{"surrounding prose", `Sure, here you go: {"a":1} - hope it helps`, `{"a":1}`, true},
		{"json fence", "```json\n{\"a\":1}\n```", `{"a":1}`, true},
		{"nested objects", `{"a":{"b":{"c":1}}}`, `{"a":{"b":{"c":1}}}`, true},
		{"braces inside a string", "{\"a\":\"{ not real }\"}", "{\"a\":\"{ not real }\"}", true},
		{"escaped quote inside a string", "{\"a\":\"say \\\"hi\\\"\"}", "{\"a\":\"say \\\"hi\\\"\"}", true},
		{"first object wins", `{"a":1} trailing {"b":2}`, `{"a":1}`, true},
		{"no object", `no json here`, "", false},
		{"unbalanced", `{"a":1`, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ExtractJSONObject(tc.in)
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestValidateOutputSchema(t *testing.T) {
	schema := map[string]string{
		"answer": "string",
		"score":  "number",
		"done":   "boolean",
		"tags":   "array",
		"meta":   "object",
	}

	t.Run("accepts matching types", func(t *testing.T) {
		res := map[string]interface{}{
			"answer": "yes",
			"score":  float64(3),
			"done":   true,
			"tags":   []interface{}{"a"},
			"meta":   map[string]interface{}{"k": "v"},
		}
		require.NoError(t, ValidateOutputSchema(res, schema))
	})

	t.Run("reports a missing field", func(t *testing.T) {
		err := ValidateOutputSchema(map[string]interface{}{"answer": "yes"}, schema)
		require.Error(t, err)
		require.Contains(t, err.Error(), `"done"`)
	})

	t.Run("reports a wrong type", func(t *testing.T) {
		err := ValidateOutputSchema(map[string]interface{}{"answer": 42}, schema)
		require.Error(t, err)
		require.Contains(t, err.Error(), `"answer"`)
	})

	t.Run("ignores an unknown declared type", func(t *testing.T) {
		require.NoError(t, ValidateOutputSchema(
			map[string]interface{}{"x": 1},
			map[string]string{"x": "whatever"},
		))
	})
}

func TestBuildOperationPrompt(t *testing.T) {
	t.Run("adds inputs and a sorted output contract", func(t *testing.T) {
		got := BuildOperationPrompt("BASE",
			map[string]interface{}{"b": 1, "a": "x"},
			map[string]string{"z": "string", "a": "number"},
		)
		require.Contains(t, got, "BASE")
		require.Contains(t, got, "## Input parameters")
		require.Contains(t, got, `"a": "x"`)
		require.Contains(t, got, "## Required output")
		// Fields are emitted in a stable (sorted) order.
		require.Less(t, strings.Index(got, `"a": number`), strings.Index(got, `"z": string`))
	})

	t.Run("returns the bare prompt when there is nothing to add", func(t *testing.T) {
		require.Equal(t, "BASE", BuildOperationPrompt("BASE", nil, nil))
	})
}

func TestRetryPromptMentionsTheRejection(t *testing.T) {
	got := RetryPrompt("BASE", errors.New("bad json"))
	require.Contains(t, got, "BASE")
	require.Contains(t, got, "bad json")
	require.Contains(t, got, "rejected")
}

func TestRunStructured(t *testing.T) {
	t.Run("succeeds on the first attempt", func(t *testing.T) {
		run := func(_ context.Context, agent, prompt string, _ bool) (*AgentResult, error) {
			require.Equal(t, "assistant", agent)
			require.Contains(t, prompt, "BASE")
			return &AgentResult{Success: true, Output: `{"answer":"yes"}`, Model: "m", LLMCalls: 1}, nil
		}

		res, err := RunStructured(context.Background(), run, StructuredRequest{
			Agent:        "assistant",
			Prompt:       "BASE",
			OutputSchema: map[string]string{"answer": "string"},
		})
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{"answer": "yes"}, res.Result)
		require.Equal(t, 1, res.Attempts)
		require.Equal(t, 1, res.Trace.Attempts)
		require.Equal(t, "assistant", res.Trace.Agent)
		require.Empty(t, res.Trace.Rejected)
	})

	t.Run("retries after a non-JSON answer", func(t *testing.T) {
		var prompts []string
		run := func(_ context.Context, _, prompt string, _ bool) (*AgentResult, error) {
			prompts = append(prompts, prompt)
			if len(prompts) == 1 {
				return &AgentResult{Success: true, Output: "no json at all"}, nil
			}
			return &AgentResult{Success: true, Output: `{"answer":"yes"}`}, nil
		}

		res, err := RunStructured(context.Background(), run, StructuredRequest{
			Agent:        "assistant",
			Prompt:       "BASE",
			OutputSchema: map[string]string{"answer": "string"},
			MaxRetries:   1,
		})
		require.NoError(t, err)
		require.Equal(t, 2, res.Attempts)
		require.Len(t, prompts, 2)
		require.NotContains(t, prompts[0], "rejected")
		require.Contains(t, prompts[1], "rejected")
		require.Len(t, res.Trace.Rejected, 1)
	})

	t.Run("fails once retries are exhausted", func(t *testing.T) {
		run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
			return &AgentResult{Success: true, Output: "still not json"}, nil
		}

		res, err := RunStructured(context.Background(), run, StructuredRequest{
			Agent: "a", Prompt: "p", MaxRetries: 2,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "did not produce a valid response")
		require.NotNil(t, res)
		require.Nil(t, res.Result)
		require.Len(t, res.Trace.Rejected, 3) // maxRetries + 1 attempts
	})

	t.Run("rejects an answer that misses the schema", func(t *testing.T) {
		run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
			return &AgentResult{Success: true, Output: `{"answer":123}`}, nil
		}

		_, err := RunStructured(context.Background(), run, StructuredRequest{
			Agent:        "a",
			Prompt:       "p",
			OutputSchema: map[string]string{"answer": "string"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "valid response")
	})

	t.Run("blocks a mutating call made without permission", func(t *testing.T) {
		run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
			return &AgentResult{ConfirmNeeded: true, ConfirmCalls: []Call{{Name: "delete_module"}}}, nil
		}

		_, err := RunStructured(context.Background(), run, StructuredRequest{Agent: "a", Prompt: "p"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "delete_module")
	})

	t.Run("propagates an agent failure", func(t *testing.T) {
		run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
			return &AgentResult{Success: false, Error: "model down"}, nil
		}

		_, err := RunStructured(context.Background(), run, StructuredRequest{Agent: "a", Prompt: "p"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "model down")
	})

	t.Run("propagates a runner error", func(t *testing.T) {
		run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
			return nil, errors.New("boom")
		}

		_, err := RunStructured(context.Background(), run, StructuredRequest{Agent: "a", Prompt: "p"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "boom")
	})
}

func TestRunStructured_RequiresAgentAndPrompt(t *testing.T) {
	run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
		t.Fatal("runner must not be called when the request is invalid")
		return nil, nil
	}

	_, err := RunStructured(context.Background(), run, StructuredRequest{Prompt: "p"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "agent is required")

	_, err = RunStructured(context.Background(), run, StructuredRequest{Agent: "a"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "prompt is required")
}

func TestRunStructured_TracesAFailedStep(t *testing.T) {
	run := func(_ context.Context, _, _ string, _ bool) (*AgentResult, error) {
		return &AgentResult{Success: true, Output: "nope"}, nil
	}

	res, err := RunStructured(context.Background(), run, StructuredRequest{Agent: "x", Prompt: "p"})
	require.Error(t, err)
	require.NotNil(t, res)
	require.Equal(t, "x", res.Trace.Agent)
	require.Contains(t, res.Trace.Error, "did not produce a valid response")
	require.NotEmpty(t, res.Trace.Prompt)
}

func TestRegistryRunner_NilRegistry(t *testing.T) {
	_, err := RegistryRunner(nil)(context.Background(), "a", "p", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registry not available")
}
