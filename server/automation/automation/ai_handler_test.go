package automation

import (
	"context"
	"fmt"
	"strings"
	"testing"

	atypes "github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

func fakeAI(answers ...string) (aiHandler, *[]string) {
	var prompts []string
	i := 0
	return aiHandler{run: func(_ context.Context, agent, prompt string, _ bool) (*aiagent.AgentResult, error) {
		prompts = append(prompts, prompt)
		a := answers[min(i, len(answers)-1)]
		i++
		return &aiagent.AgentResult{Success: true, Output: a}, nil
	}}, &prompts
}

func TestAiAskDefaultsToToollessAgent(t *testing.T) {
	var gotAgent string
	h := aiHandler{run: func(_ context.Context, agent, _ string, _ bool) (*aiagent.AgentResult, error) {
		gotAgent = agent
		return &aiagent.AgentResult{Success: true, Output: "hi"}, nil
	}}
	res, err := h.ask(context.Background(), &aiAskArgs{Prompt: "say hi"})
	if err != nil || res.Text != "hi" || gotAgent != aiDefaultAgent {
		t.Fatalf("res=%v err=%v agent=%q", res, err, gotAgent)
	}
}

func TestAiAskBlocksMutation(t *testing.T) {
	h := aiHandler{run: func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{ConfirmNeeded: true, ConfirmCalls: []aiagent.Call{{Name: "create_module"}}}, nil
	}}
	if _, err := h.ask(context.Background(), &aiAskArgs{Prompt: "x"}); err == nil || !strings.Contains(err.Error(), "create_module") {
		t.Fatalf("expected confirm error, got %v", err)
	}
}

func TestAiExtractRetriesThenSucceeds(t *testing.T) {
	h, prompts := fakeAI("not json", "```json\n{\"total\": 42}\n```")
	res, err := h.extract(context.Background(), &aiExtractArgs{
		Prompt:       "get total",
		OutputSchema: map[string]string{"total": "number"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(*prompts) != 2 || !strings.Contains((*prompts)[1], "rejected") {
		t.Fatalf("expected a retry prompt, got %v", *prompts)
	}
	if v, _ := res.Result.Dict()["total"]; v == nil {
		t.Fatalf("total missing: %v", res.Result.Dict())
	}
}

func TestAiExtractGivesUp(t *testing.T) {
	h, _ := fakeAI("nope")
	_, err := h.extract(context.Background(), &aiExtractArgs{Prompt: "p", OutputSchema: map[string]string{"a": "string"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAiClassify(t *testing.T) {
	h, _ := fakeAI(`{"label":"URGENT","confidence":0.9,"reason":"deadline"}`)
	res, err := h.classify(context.Background(), &aiClassifyArgs{Text: "asap!", Labels: "urgent, normal"})
	if err != nil || res.Label != "urgent" || res.Confidence != 0.9 {
		t.Fatalf("res=%+v err=%v", res, err)
	}

	h, _ = fakeAI(`{"label":"banana","confidence":1,"reason":"x"}`)
	if _, err = h.classify(context.Background(), &aiClassifyArgs{Text: "t", Labels: "a,b"}); err == nil {
		t.Fatal("expected unknown-label error")
	}
	if _, err = h.classify(context.Background(), &aiClassifyArgs{Text: "t", Labels: "only"}); err == nil {
		t.Fatal("expected too-few-labels error")
	}
}

type testAIReg struct{ ff []*atypes.Function }

func (r *testAIReg) AddFunctions(ff ...*atypes.Function) { r.ff = append(r.ff, ff...) }
func (r *testAIReg) Type(ref string) expr.Type {
	switch ref {
	case "String":
		return &expr.String{}
	case "Float":
		return &expr.Float{}
	case "Vars":
		return &expr.Vars{}
	}
	return &expr.Any{}
}

func TestAiFunctionsEndToEnd(t *testing.T) {
	reg := &testAIReg{}
	newAiHandler(reg, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true, Output: `{"label":"a","confidence":0.5,"reason":"r"}`}, nil
	}, nil)
	if len(reg.ff) != 6 {
		t.Fatalf("expected 6 functions, got %d", len(reg.ff))
	}

	for _, f := range reg.ff {
		if f.Ref != "aiClassify" {
			continue
		}
		in, err := expr.NewVars(map[string]interface{}{"text": "t", "labels": "a,b"})
		if err != nil {
			t.Fatal(err)
		}
		out, err := f.Handler(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		if got := out.Dict()["label"]; got != "a" {
			t.Fatalf("label=%v", got)
		}
		return
	}
	t.Fatal("aiClassify not registered")
}

func TestAiRagSearch(t *testing.T) {
	h := aiHandler{}
	if _, err := h.ragSearch(context.Background(), &aiRagSearchArgs{Namespace: "1", Query: "q"}); err == nil {
		t.Fatal("expected not-available error")
	}

	SetRAGSearch(func(_ context.Context, ns, q string, k int) ([]RAGHit, error) {
		if k != 5 {
			t.Fatalf("default topK = %d", k)
		}
		return []RAGHit{{Text: "alpha"}, {Text: "beta"}}, nil
	})
	defer SetRAGSearch(nil)

	res, err := h.ragSearch(context.Background(), &aiRagSearchArgs{Namespace: "1", Query: "q"})
	if err != nil || res.Count != 2 || !strings.Contains(res.Text, "beta") {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestAiToolGuards(t *testing.T) {
	h := aiHandler{}
	if _, err := h.tool(context.Background(), &aiToolArgs{Tool: "no_such_tool"}); err == nil {
		t.Fatal("expected unknown tool error")
	}

	aiagent.DefaultCatalog().Register(aiagent.ToolKit{Name: "test.wf", Tools: []chat.ToolDef{
		{Name: "wf_read", Handler: func(context.Context, map[string]string) string { return "read-ok" }},
		{Name: "wf_write", Mutating: true, Handler: func(context.Context, map[string]string) string { return "written" }},
	}})
	defer aiagent.DefaultCatalog().Unregister("test.wf")

	if res, err := h.tool(context.Background(), &aiToolArgs{Tool: "wf_read"}); err != nil || res.Result != "read-ok" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if _, err := h.tool(context.Background(), &aiToolArgs{Tool: "wf_write"}); err == nil {
		t.Fatal("mutating tool must be blocked without allowMutating")
	}
	if res, err := h.tool(context.Background(), &aiToolArgs{Tool: "wf_write", AllowMutating: true}); err != nil || res.Result != "written" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestAiAskDeferConfirmReturnsPending(t *testing.T) {
	h := aiHandler{run: func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{ConfirmNeeded: true, ConfirmCalls: []aiagent.Call{{Name: "create_module", Params: `{"handle":"x"}`}}}, nil
	}}
	res, err := h.ask(context.Background(), &aiAskArgs{Prompt: "p", DeferConfirm: true})
	if err != nil || !res.NeedsApproval || !strings.Contains(res.PendingCalls, "create_module") || !strings.Contains(res.Summary, "create_module") {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestAiRunCalls(t *testing.T) {
	var gotAgent string
	var gotCalls []aiagent.Call
	h := aiHandler{execApproved: func(_ context.Context, agent string, calls []aiagent.Call) (string, error) {
		gotAgent, gotCalls = agent, calls
		return "done", nil
	}}

	res, err := h.runCalls(context.Background(), &aiRunCallsArgs{Agent: "crud-agent", Calls: `[{"name":"create_module","params":"{}"}]`})
	if err != nil || res.Result != "done" || gotAgent != "crud-agent" || len(gotCalls) != 1 || gotCalls[0].Name != "create_module" {
		t.Fatalf("res=%+v err=%v agent=%q calls=%v", res, err, gotAgent, gotCalls)
	}

	for _, bad := range []string{`not json`, `[]`} {
		if _, err := h.runCalls(context.Background(), &aiRunCallsArgs{Agent: "a", Calls: bad}); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}

func TestAiStepTimeout(t *testing.T) {
	h := aiHandler{run: func(ctx context.Context, _, _ string, _ bool) (*aiagent.AgentResult, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	_, err := h.ask(context.Background(), &aiAskArgs{Prompt: "p", TimeoutSec: 1, hasTimeoutSec: true})
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("expected deadline error, got %v", err)
	}
}

func TestAiStepsReturnTrace(t *testing.T) {
	h := aiHandler{run: func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true, Output: `{"label":"a","confidence":0.5,"reason":"r"}`,
			Model: "m", LLMCalls: 1, PromptTokens: 11, CompletionTokens: 3}, nil
	}}

	check := func(name string, tr *expr.Vars) {
		t.Helper()
		if tr == nil {
			t.Fatalf("%s: no trace", name)
		}
		d := tr.Dict()
		if d["model"] != "m" || d["agent"] == nil {
			t.Fatalf("%s: trace = %v", name, d)
		}
		if p, _ := d["prompt"].(string); !strings.Contains(p, "hello") {
			t.Fatalf("%s: prompt not traced: %v", name, d["prompt"])
		}
	}

	ask, err := h.ask(context.Background(), &aiAskArgs{Prompt: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	check("ask", ask.Trace)

	ext, err := h.extract(context.Background(), &aiExtractArgs{Prompt: "hello", OutputSchema: map[string]string{"label": "string"}})
	if err != nil {
		t.Fatal(err)
	}
	check("extract", ext.Trace)

	cls, err := h.classify(context.Background(), &aiClassifyArgs{Text: "hello", Labels: "a,b"})
	if err != nil {
		t.Fatal(err)
	}
	check("classify", cls.Trace)
}

func TestAiExtractFailureMentionsLastResponse(t *testing.T) {
	h, _ := fakeAI("I would rather not")
	_, err := h.extract(context.Background(), &aiExtractArgs{Prompt: "p", OutputSchema: map[string]string{"a": "string"}})
	if err == nil || !strings.Contains(err.Error(), "I would rather not") {
		t.Fatalf("error should carry the model's last words, got %v", err)
	}
}

func TestAiAskPassesModelToTheAgentRun(t *testing.T) {
	var seen string
	h := aiHandler{run: func(ctx context.Context, _, _ string, _ bool) (*aiagent.AgentResult, error) {
		seen = aiagent.ModelFromContext(ctx)
		return &aiagent.AgentResult{Success: true, Output: "ok"}, nil
	}}
	if _, err := h.ask(context.Background(), &aiAskArgs{Prompt: "p", Model: "qwen3:8b"}); err != nil {
		t.Fatal(err)
	}
	if seen != "qwen3:8b" {
		t.Fatalf("model seen by the run = %q", seen)
	}
}

func TestAiStepsUseLibraryPrompts(t *testing.T) {
	aiagent.SetPromptResolver(func(_ context.Context, handle string, _ int) (*aiagent.ResolvedPrompt, error) {
		if handle != "triage" {
			return nil, fmt.Errorf("prompt does not exist")
		}
		return &aiagent.ResolvedPrompt{Text: "Rate the urgency of the ticket.", Ref: "triage@4"}, nil
	})
	defer aiagent.SetPromptResolver(nil)

	var sent string
	h := aiHandler{run: func(_ context.Context, _, prompt string, _ bool) (*aiagent.AgentResult, error) {
		sent = prompt
		return &aiagent.AgentResult{Success: true, Output: `{"priority":"high","label":"a","confidence":1,"reason":"r"}`}, nil
	}}

	ask, err := h.ask(context.Background(), &aiAskArgs{Prompt: "@prompt:triage"})
	if err != nil || !strings.Contains(sent, "Rate the urgency") || ask.Trace.Dict()["promptRef"] != "triage@4" {
		t.Fatalf("ask: err=%v sent=%q trace=%v", err, sent, ask.Trace.Dict())
	}

	ext, err := h.extract(context.Background(), &aiExtractArgs{Prompt: "@prompt:triage", OutputSchema: map[string]string{"priority": "string"}})
	if err != nil || ext.Trace.Dict()["promptRef"] != "triage@4" {
		t.Fatalf("extract: err=%v trace=%v", err, ext.Trace.Dict())
	}

	cls, err := h.classify(context.Background(), &aiClassifyArgs{Text: "x", Labels: "a,b", Instruction: "@prompt:triage"})
	if err != nil || cls.Trace.Dict()["promptRef"] != "triage@4" || !strings.Contains(sent, "Rate the urgency") {
		t.Fatalf("classify: err=%v trace=%v", err, cls.Trace.Dict())
	}

	// a reference to a prompt that is not there must not be sent as text
	sent = ""
	for name, call := range map[string]func() error{
		"ask": func() error { _, e := h.ask(context.Background(), &aiAskArgs{Prompt: "@prompt:gone"}); return e },
		"extract": func() error {
			_, e := h.extract(context.Background(), &aiExtractArgs{Prompt: "@prompt:gone"})
			return e
		},
		"classify": func() error {
			_, e := h.classify(context.Background(), &aiClassifyArgs{Text: "x", Labels: "a,b", Instruction: "@prompt:gone"})
			return e
		},
	} {
		if err := call(); err == nil || !strings.Contains(err.Error(), "does not exist") {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if sent != "" {
		t.Fatalf("the model was called with %q", sent)
	}
}
