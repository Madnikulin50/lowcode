package rest

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/stretchr/testify/require"
)

func selectOptions(t *testing.T, p *types.Param) []map[string]interface{} {
	t.Helper()
	require.NotNil(t, p.Meta, "param %s has no meta", p.Name)
	input := p.Meta.Visual["input"].(map[string]interface{})
	require.Equal(t, "select", input["type"])
	return input["properties"].(map[string]interface{})["options"].([]map[string]interface{})
}

func TestWithLiveAIChoices(t *testing.T) {
	aiFn := &types.Function{
		Ref:    "aiAsk",
		Labels: map[string]string{"ai": "step"},
		Parameters: []*types.Param{
			{Name: "prompt", Types: []string{"String"}, Meta: &types.ParamMeta{Label: "Prompt"}},
			{Name: "agent", Types: []string{"String"}, Meta: &types.ParamMeta{Label: "Agent"}},
			{Name: "model", Types: []string{"String"}},
		},
	}
	required := &types.Function{
		Ref:        "aiRunCalls",
		Labels:     map[string]string{"ai": "step"},
		Parameters: []*types.Param{{Name: "agent", Types: []string{"String"}, Required: true}},
	}
	plain := &types.Function{Ref: "logInfo", Parameters: []*types.Param{{Name: "agent"}}}

	out := withLiveAIChoices(context.Background(), []*types.Function{aiFn, required, plain})
	require.Len(t, out, 3)

	// agent: the built-in agents are offered, with a "none" first
	agent := out[0].Parameters[1]
	opts := selectOptions(t, agent)
	require.Equal(t, "", opts[0]["value"])
	var handles []string
	for _, o := range opts {
		handles = append(handles, o["value"].(string))
	}
	require.Contains(t, handles, "assistant")
	require.Equal(t, "Agent", agent.Meta.Label, "existing meta is kept")

	// model: roles are always offered, even with no Ollama around
	var models []string
	for _, o := range selectOptions(t, out[0].Parameters[2]) {
		models = append(models, o["value"].(string))
	}
	require.Contains(t, models, "rulesgo.ai")

	// untouched parameters are the very same objects
	require.Same(t, aiFn.Parameters[0], out[0].Parameters[0])

	// a required choice has no "none"
	for _, o := range selectOptions(t, out[1].Parameters[0]) {
		require.NotEqual(t, "", o["value"])
	}

	// not an AI function: untouched
	require.Same(t, plain, out[2])

	// the registry's own definitions were not modified
	require.Nil(t, aiFn.Parameters[2].Meta)
	require.Nil(t, aiFn.Parameters[1].Meta.Visual)
	require.NotSame(t, aiFn, out[0])
}
