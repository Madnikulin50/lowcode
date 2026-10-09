package aiagent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/stretchr/testify/require"
)

func withResolver(t *testing.T, fn PromptResolver) {
	t.Helper()
	SetPromptResolver(fn)
	t.Cleanup(func() { SetPromptResolver(nil) })
}

func TestResolvePrompt(t *testing.T) {
	var gotHandle string
	var gotVersion int
	withResolver(t, func(_ context.Context, handle string, version int) (*ResolvedPrompt, error) {
		gotHandle, gotVersion = handle, version
		if handle == "missing" {
			return nil, errors.New("no such prompt")
		}
		return &ResolvedPrompt{Text: "Summarise the ticket.", Ref: handle + "@2"}, nil
	})
	ctx := context.Background()

	// inline text is left alone
	text, ref, err := ResolvePrompt(ctx, "Just answer yes or no")
	require.NoError(t, err)
	require.Equal(t, "Just answer yes or no", text)
	require.Empty(t, ref)

	// active version
	text, ref, err = ResolvePrompt(ctx, "  @prompt:ticket_summary ")
	require.NoError(t, err)
	require.Equal(t, "Summarise the ticket.", text)
	require.Equal(t, "ticket_summary@2", ref)
	require.Equal(t, "ticket_summary", gotHandle)
	require.Equal(t, 0, gotVersion, "0 means the active version")

	// pinned
	_, _, err = ResolvePrompt(ctx, "@prompt:ticket_summary@7")
	require.NoError(t, err)
	require.Equal(t, 7, gotVersion)

	// a missing prompt is an error, never sent to the model as text
	_, _, err = ResolvePrompt(ctx, "@prompt:missing")
	require.ErrorContains(t, err, "no such prompt")
	require.ErrorContains(t, err, "@prompt:missing")

	// not a reference: a mention inside prose, or a malformed handle
	for _, s := range []string{"see @prompt:foo for details", "@prompt:Foo", "@prompt:", "@prompt:foo@x"} {
		text, ref, err = ResolvePrompt(ctx, s)
		require.NoError(t, err, s)
		require.Equal(t, s, text, s)
		require.Empty(t, ref, s)
		require.False(t, IsPromptRef(s), s)
	}
	require.True(t, IsPromptRef("@prompt:a_b-c.d@12"))
}

func TestResolvePrompt_NoLibrary(t *testing.T) {
	SetPromptResolver(nil)
	_, _, err := ResolvePrompt(context.Background(), "@prompt:x")
	require.ErrorContains(t, err, "not available")

	// but inline prompts never need it
	text, _, err := ResolvePrompt(context.Background(), "plain")
	require.NoError(t, err)
	require.Equal(t, "plain", text)
}

func TestRenderPrompt(t *testing.T) {
	out, used := RenderPrompt("Hello {{ name }}, total {{n}}, {{missing}}", map[string]interface{}{"name": "Acme", "n": 3})
	require.True(t, used)
	require.Equal(t, "Hello Acme, total 3, {{missing}}", out)

	out, used = RenderPrompt("No placeholders", map[string]interface{}{"name": "Acme"})
	require.False(t, used)
	require.Equal(t, "No placeholders", out)
}

// scripted answers keyed by a word that appears in the prompt
func scripted(answers map[string]string) Runner {
	return func(_ context.Context, _, prompt string, _ bool) (*AgentResult, error) {
		for key, out := range answers {
			if strings.Contains(prompt, key) {
				return &AgentResult{Success: true, Output: out, PromptTokens: 10, CompletionTokens: 2}, nil
			}
		}
		return &AgentResult{Success: true, Output: "I don't know"}, nil
	}
}

func TestRunEval_StructuredCases(t *testing.T) {
	report := RunEval(context.Background(),
		scripted(map[string]string{
			"server down":  `{"priority": "High"}`, // case differs from "high": still right
			"printer jam":  `{"priority": "high"}`, // wrong
			"coffee":       "not json at all",
			"office party": `{"priority": "low"}`,
		}),
		EvalSpec{
			Prompt:    "Rate the urgency of the ticket.",
			PromptRef: "triage@2",
			Cases: []EvalCase{
				{Name: "outage", Inputs: map[string]interface{}{"ticket": "server down"}, Expect: map[string]interface{}{"priority": "high"}},
				{Name: "minor", Inputs: map[string]interface{}{"ticket": "printer jam"}, Expect: map[string]interface{}{"priority": "low"}},
				{Name: "garbled", Inputs: map[string]interface{}{"ticket": "coffee"}, Expect: map[string]interface{}{"priority": "low"}},
				{Name: "social", Inputs: map[string]interface{}{"ticket": "office party"}, Expect: map[string]interface{}{"priority": "low"}},
			},
		})

	require.Equal(t, "triage@2", report.PromptRef)
	require.Equal(t, 4, report.Total)
	require.Equal(t, 2, report.Passed)
	require.InDelta(t, 0.5, report.PassRate, 0.001)

	byName := map[string]EvalResult{}
	for _, c := range report.Cases {
		byName[c.Case] = c
	}
	require.True(t, byName["outage"].Pass, "string comparison ignores case")
	require.True(t, byName["social"].Pass)

	require.False(t, byName["minor"].Pass)
	require.Contains(t, byName["minor"].Failures[0], "expected low, got high")

	require.False(t, byName["garbled"].Pass)
	require.NotEmpty(t, byName["garbled"].Error, "an unparseable answer is reported, not hidden")

	require.Equal(t, 10, byName["outage"].PromptTokens)
	require.Equal(t, 40, report.PromptTokens, "tokens add up across the four cases, failed attempts included")
}

func TestRunEval_ProseCasesAndPlaceholders(t *testing.T) {
	var prompts []string
	run := func(_ context.Context, _, prompt string, _ bool) (*AgentResult, error) {
		prompts = append(prompts, prompt)
		return &AgentResult{Success: true, Output: "The Acme invoice is overdue."}, nil
	}

	report := RunEval(context.Background(), run, EvalSpec{
		Prompt: "Write one sentence about {{customer}}'s invoice.",
		Cases: []EvalCase{
			{Name: "mentions", Inputs: map[string]interface{}{"customer": "Acme"}, Contains: []string{"acme", "OVERDUE"}},
			{Name: "misses", Inputs: map[string]interface{}{"customer": "Acme"}, Contains: []string{"refund"}},
			{Name: "empty", Inputs: map[string]interface{}{"customer": "Acme"}},
		},
	})

	require.Contains(t, prompts[0], "about Acme's invoice", "placeholders are filled")
	require.NotContains(t, prompts[0], "Input parameters", "inputs already in the text are not repeated")

	require.True(t, report.Cases[0].Pass)
	require.False(t, report.Cases[1].Pass)
	require.Contains(t, report.Cases[1].Failures[0], "refund")
	require.False(t, report.Cases[2].Pass, "a case that expects nothing proves nothing")
	require.Contains(t, report.Cases[2].Failures[0], "expects nothing")
}

func TestRunEval_InputsGoInAJSONBlockWhenThereAreNoPlaceholders(t *testing.T) {
	var seen string
	run := func(_ context.Context, _, prompt string, _ bool) (*AgentResult, error) {
		seen = prompt
		return &AgentResult{Success: true, Output: "ok"}, nil
	}
	RunEval(context.Background(), run, EvalSpec{
		Prompt: "Classify the ticket.",
		Cases:  []EvalCase{{Name: "c", Inputs: map[string]interface{}{"ticket": "server down"}, Contains: []string{"ok"}}},
	})
	require.Contains(t, seen, "Input parameters")
	require.Contains(t, seen, "server down")
}

func TestRunEval_RunnerErrorIsACaseFailure(t *testing.T) {
	report := RunEval(context.Background(),
		func(context.Context, string, string, bool) (*AgentResult, error) {
			return nil, errors.New("model down")
		},
		EvalSpec{Prompt: "p", Cases: []EvalCase{{Name: "a", Contains: []string{"x"}}, {Name: "b", Expect: map[string]interface{}{"k": "v"}}}})

	require.Equal(t, 0, report.Passed)
	for _, c := range report.Cases {
		require.Contains(t, c.Error, "model down")
	}
}

func TestRunEval_NoCases(t *testing.T) {
	report := RunEval(context.Background(), nil, EvalSpec{Prompt: "p"})
	require.Equal(t, 0, report.Total)
	require.Equal(t, 0.0, report.PassRate)
	require.Empty(t, report.Cases)
}

func TestSameValue(t *testing.T) {
	require.True(t, sameValue(3, float64(3)), "numbers compare by value after JSON")
	require.True(t, sameValue(true, true))
	require.False(t, sameValue(true, "true"))
	require.True(t, sameValue(" High ", "high"))
	require.False(t, sameValue("high", 1.0))
	require.True(t, sameValue([]interface{}{"a"}, []interface{}{"a"}))
}

func TestParamsFromJSON_KeepsCompositesAsJSON(t *testing.T) {
	got := paramsFromJSON(`{"n": 1000000, "ok": true, "ids": ["a"], "f": {"k": 1}, "s": "x", "none": null}`)
	require.Equal(t, "1000000", got["n"])
	require.Equal(t, "true", got["ok"])
	require.Equal(t, `["a"]`, got["ids"])
	require.JSONEq(t, `{"k":1}`, got["f"])
	require.Equal(t, "x", got["s"])
	require.Equal(t, "", got["none"])

	require.Equal(t, map[string]string{"_raw": "not json"}, paramsFromJSON("not json"))
}

func TestProgress(t *testing.T) {
	var got []Progress
	ctx := ContextWithProgress(context.Background(), func(p Progress) { got = append(got, p) })

	ReportProgress(ctx, Progress{Attempt: 2})
	chat.EmitStatus(ctx, chat.StatusWarming) // the statuses the model client already emits
	require.Equal(t, []Progress{{Attempt: 2}, {Status: chat.StatusWarming}}, got)

	// no watcher: everything is a no-op
	ReportProgress(context.Background(), Progress{Token: "x"})
	require.Nil(t, streamFromContext(context.Background()))
	require.Equal(t, context.Background(), ContextWithProgress(context.Background(), nil))
}

func TestStreamFromContext_PassesTokensAndStopsWhenTheWatcherGoes(t *testing.T) {
	var got []Progress
	ctx, cancel := context.WithCancel(ContextWithProgress(context.Background(), func(p Progress) { got = append(got, p) }))

	stream := streamFromContext(ctx)
	require.NoError(t, stream("Hel", "", false))
	require.NoError(t, stream("", "thinking", false))
	require.NoError(t, stream("", "", true), "an empty chunk is not an event")
	require.Equal(t, []Progress{{Token: "Hel"}, {Reason: "thinking"}}, got)

	cancel()
	require.Error(t, stream("lo", "", false), "generation must stop once the client has disconnected")
}

func TestProgressEvent(t *testing.T) {
	require.Equal(t, map[string]interface{}{"done": false, "status": "warming"}, Progress{Status: "warming"}.Event())
	require.Equal(t, map[string]interface{}{"done": false, "attempt": 2}, Progress{Attempt: 2}.Event())
	require.Equal(t, map[string]interface{}{"done": false, "token": "a", "reason": ""}, Progress{Token: "a"}.Event())
}

func TestRunStructured_ReportsEachAttempt(t *testing.T) {
	var attempts []int
	ctx := ContextWithProgress(context.Background(), func(p Progress) {
		if p.Attempt > 0 {
			attempts = append(attempts, p.Attempt)
		}
	})

	n := 0
	run := func(context.Context, string, string, bool) (*AgentResult, error) {
		n++
		if n < 3 {
			return &AgentResult{Success: true, Output: "no json yet"}, nil
		}
		return &AgentResult{Success: true, Output: `{"k":"v"}`}, nil
	}

	_, err := RunStructured(ctx, run, StructuredRequest{Agent: "a", Prompt: "p", OutputSchema: map[string]string{"k": "string"}, MaxRetries: 3})
	require.NoError(t, err)
	require.Equal(t, []int{1, 2, 3}, attempts, "a watcher can show that the model is on its second try")
}
