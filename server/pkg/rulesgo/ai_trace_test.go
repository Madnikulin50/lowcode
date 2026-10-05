package rulesgo

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
)

func runAIChain(t *testing.T, nodeType, config string, call func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error)) *ChainResult {
	t.Helper()

	registry := NewRegistry()
	registry.Register("ai", &aiExecutor{call: call})
	registry.Register("ai.operation", &aiOperationExecutor{call: call})
	engine := NewEngine(registry)
	engine.RegisterChain(&Chain{
		ID:        "c",
		EntryNode: "n",
		Nodes:     []ChainNode{{ID: "n", Type: nodeType, Config: []byte(config)}},
	})

	res, err := engine.Run(context.Background(), "c", map[string]interface{}{"name": "Acme"})
	if err != nil {
		t.Fatalf("engine error: %v", err)
	}
	if len(res.Nodes) != 1 {
		t.Fatalf("expected 1 node result, got %d", len(res.Nodes))
	}
	return res
}

func TestEngine_RecordsNodeDuration(t *testing.T) {
	res := runAIChain(t, "ai", `{"agent":"assistant","prompt":"hi"}`,
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			time.Sleep(20 * time.Millisecond)
			return &AIOperationResult{Output: "hello"}, nil
		})

	if d := res.Nodes[0].DurationMs; d < 20 {
		t.Fatalf("DurationMs = %d, want >= 20", d)
	}
}

func TestAINode_TraceOnSuccess(t *testing.T) {
	res := runAIChain(t, "ai", `{"agent":"assistant","prompt":"Greet {{name}}","model":"qwen3:8b"}`,
		func(_ context.Context, _, _, model string, _ bool) (*AIOperationResult, error) {
			if model != "qwen3:8b" {
				t.Errorf("model passed to the call = %q", model)
			}
			return &AIOperationResult{Output: "Hello Acme", Model: "qwen3:8b", LLMCalls: 2, PromptTokens: 40, CompletionTokens: 5, Tools: []string{"search_records"}}, nil
		})

	tr := res.Nodes[0].Trace
	if tr == nil {
		t.Fatal("no trace on the node result")
	}
	if tr["agent"] != "assistant" || tr["model"] != "qwen3:8b" {
		t.Fatalf("trace = %v", tr)
	}
	if p, _ := tr["prompt"].(string); p != "Greet Acme" {
		t.Fatalf("the trace should hold the rendered prompt, got %q", p)
	}
	if tr["response"] != "Hello Acme" || tr["promptTokens"] != float64(40) || tr["llmCalls"] != float64(2) {
		t.Fatalf("trace = %v", tr)
	}
	if tools, _ := tr["tools"].([]interface{}); len(tools) != 1 || tools[0] != "search_records" {
		t.Fatalf("tools = %v", tr["tools"])
	}
	if _, has := tr["error"]; has {
		t.Fatalf("a successful call must not carry an error: %v", tr)
	}
}

func TestAINode_TraceSurvivesFailure(t *testing.T) {
	res := runAIChain(t, "ai", `{"agent":"assistant","prompt":"hi"}`,
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			return nil, errors.New("model overloaded")
		})

	if res.Success {
		t.Fatal("chain should fail")
	}
	n := res.Nodes[0]
	if n.Output != nil {
		t.Fatalf("a failed node has no output, got %v", n.Output)
	}
	if n.Trace == nil || !strings.Contains(n.Trace["error"].(string), "model overloaded") {
		t.Fatalf("the failed call's trace must be kept, got %v", n.Trace)
	}
	if n.Trace["prompt"] != "hi" {
		t.Fatalf("trace = %v", n.Trace)
	}
}

func TestAINode_OptionalFailureStillTraced(t *testing.T) {
	res := runAIChain(t, "ai", `{"agent":"assistant","prompt":"hi","optional":true}`,
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			return nil, errors.New("down")
		})

	if !res.Success {
		t.Fatalf("optional AI failure must not fail the chain: %s", res.Error)
	}
	if res.Nodes[0].Trace["error"] == nil {
		t.Fatalf("trace should say the call failed: %v", res.Nodes[0].Trace)
	}
}

func TestAIOperation_TraceShowsRetries(t *testing.T) {
	calls := 0
	res := runAIChain(t, "ai.operation", `{"agent":"assistant","prompt":"score it","outputSchema":{"risk":"number"},"maxRetries":1}`,
		func(_ context.Context, _, prompt, _ string, _ bool) (*AIOperationResult, error) {
			calls++
			if calls == 1 {
				return &AIOperationResult{Output: "I think it is risky", LLMCalls: 1, PromptTokens: 30}, nil
			}
			if !strings.Contains(prompt, "rejected") {
				t.Errorf("retry prompt should carry the rejection, got %q", prompt)
			}
			return &AIOperationResult{Output: `{"risk": 8}`, LLMCalls: 1, PromptTokens: 45}, nil
		})

	if !res.Success {
		t.Fatalf("chain failed: %s", res.Error)
	}
	tr := res.Nodes[0].Trace
	if tr["attempts"] != float64(2) || tr["promptTokens"] != float64(75) {
		t.Fatalf("trace = %v", tr)
	}
	rejected, _ := tr["rejected"].([]interface{})
	if len(rejected) != 1 || !strings.Contains(rejected[0].(string), "JSON") {
		t.Fatalf("rejected = %v", tr["rejected"])
	}
}

func TestAIOperation_ModelReachesTheCall(t *testing.T) {
	var viaArg, viaCtx string
	runAIChain(t, "ai.operation", `{"agent":"assistant","prompt":"p","model":"m2"}`,
		func(ctx context.Context, _, _, model string, _ bool) (*AIOperationResult, error) {
			viaArg, viaCtx = model, aiagent.ModelFromContext(ctx)
			return &AIOperationResult{Output: `{}`}, nil
		})

	if viaArg != "m2" || viaCtx != "m2" {
		t.Fatalf("model arg = %q, ctx = %q", viaArg, viaCtx)
	}
}

func TestEngine_ExecuteNode_StandAlone(t *testing.T) {
	registry := NewRegistry()
	registry.Register("ai", &aiExecutor{call: func(_ context.Context, _, prompt, _ string, _ bool) (*AIOperationResult, error) {
		return &AIOperationResult{Output: "echo: " + prompt, Model: "m"}, nil
	}})
	engine := NewEngine(registry)

	res, err := engine.ExecuteNode(context.Background(),
		ChainNode{ID: "t", Type: "ai", Config: []byte(`{"agent":"assistant","prompt":"Hello {{name}}"}`)},
		map[string]interface{}{"name": "Acme"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" || res.Output["response"] != "echo: Hello Acme" {
		t.Fatalf("res = %+v", res)
	}
	if res.Trace == nil || res.Trace["model"] != "m" {
		t.Fatalf("trace = %v", res.Trace)
	}

	// a node that fails still reports through the result, trace included
	registry.Register("ai", &aiExecutor{call: func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
		return nil, errors.New("boom")
	}})
	res, err = engine.ExecuteNode(context.Background(),
		ChainNode{ID: "t", Type: "ai", Config: []byte(`{"agent":"a","prompt":"p"}`)}, nil)
	if err != nil {
		t.Fatalf("a failing node is not an engine error: %v", err)
	}
	if !strings.Contains(res.Error, "boom") || res.Trace == nil {
		t.Fatalf("res = %+v", res)
	}

	if _, err = engine.ExecuteNode(context.Background(), ChainNode{Type: "nope"}, nil); err == nil {
		t.Fatal("unknown node type must be an error")
	}
}

func withPromptLibrary(t *testing.T, prompts map[string]string) {
	t.Helper()
	aiagent.SetPromptResolver(func(_ context.Context, handle string, version int) (*aiagent.ResolvedPrompt, error) {
		text, ok := prompts[handle]
		if !ok {
			return nil, errors.New("prompt does not exist")
		}
		if version == 0 {
			version = 1
		}
		return &aiagent.ResolvedPrompt{Text: text, Ref: handle + "@" + string(rune('0'+version))}, nil
	})
	t.Cleanup(func() { aiagent.SetPromptResolver(nil) })
}

func TestAINode_LibraryPrompt(t *testing.T) {
	withPromptLibrary(t, map[string]string{"greet": "Greet {{name}} politely."})

	var sent string
	res := runAIChain(t, "ai", `{"agent":"assistant","prompt":"@prompt:greet"}`,
		func(_ context.Context, _, prompt, _ string, _ bool) (*AIOperationResult, error) {
			sent = prompt
			return &AIOperationResult{Output: "Hello Acme"}, nil
		})

	if sent != "Greet Acme politely." {
		t.Fatalf("the library text should be sent with the chain's variables filled in, got %q", sent)
	}
	if ref := res.Nodes[0].Trace["promptRef"]; ref != "greet@1" {
		t.Fatalf("the trace should say which prompt version ran, got %v", res.Nodes[0].Trace)
	}
}

func TestAIOperation_LibraryPrompt(t *testing.T) {
	withPromptLibrary(t, map[string]string{"score": "Score {{name}} from 1 to 10."})

	var sent string
	res := runAIChain(t, "ai.operation", `{"agent":"assistant","prompt":"@prompt:score@2","outputSchema":{"score":"number"}}`,
		func(_ context.Context, _, prompt, _ string, _ bool) (*AIOperationResult, error) {
			sent = prompt
			return &AIOperationResult{Output: `{"score": 7}`}, nil
		})

	if !strings.Contains(sent, "Score Acme from 1 to 10.") {
		t.Fatalf("prompt sent = %q", sent)
	}
	if ref := res.Nodes[0].Trace["promptRef"]; ref != "score@2" {
		t.Fatalf("pinned version should show in the trace, got %v", res.Nodes[0].Trace)
	}
}

func TestAINodes_MissingLibraryPromptFailsTheNode(t *testing.T) {
	withPromptLibrary(t, map[string]string{})

	for _, nodeType := range []string{"ai", "ai.operation"} {
		called := false
		res := runAIChain(t, nodeType, `{"agent":"assistant","prompt":"@prompt:gone"}`,
			func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
				called = true
				return &AIOperationResult{Output: `{}`}, nil
			})

		if res.Success || called {
			t.Fatalf("%s: a missing prompt must fail the node without calling the model (success=%v called=%v)", nodeType, res.Success, called)
		}
		if !strings.Contains(res.Error, "does not exist") {
			t.Fatalf("%s: error = %q", nodeType, res.Error)
		}
	}
}

// two AI nodes in a row, with the chain's own limits
func budgetChain(t *testing.T, config string, nodeType string, call func(context.Context, string, string, string, bool) (*AIOperationResult, error)) *ChainResult {
	t.Helper()
	registry := NewRegistry()
	registry.Register("ai", &aiExecutor{call: call})
	registry.Register("ai.operation", &aiOperationExecutor{call: call})
	engine := NewEngine(registry)

	nodeCfg := `{"agent":"assistant","prompt":"hi"}`
	if nodeType == "ai.operation" {
		nodeCfg = `{"agent":"assistant","prompt":"hi","outputSchema":{"k":"string"}}`
	}
	engine.RegisterChain(&Chain{
		ID: "b", EntryNode: "one", Config: []byte(config),
		Nodes: []ChainNode{
			{ID: "one", Type: nodeType, Config: []byte(nodeCfg)},
			{ID: "two", Type: nodeType, Config: []byte(nodeCfg)},
			{ID: "three", Type: nodeType, Config: []byte(nodeCfg)},
		},
		Edges: []ChainEdge{{From: "one", To: "two"}, {From: "two", To: "three"}},
	})

	res, err := engine.Run(context.Background(), "b", nil)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestChainAIBudget_CallLimitStopsTheRun(t *testing.T) {
	calls := 0
	res := budgetChain(t, `{"aiBudget":{"maxLLMCalls":2}}`, "ai",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "ok"}, nil
		})

	if res.Success {
		t.Fatal("the third AI node should have been stopped by the budget")
	}
	if calls != 2 {
		t.Fatalf("model called %d times, want 2", calls)
	}
	if !strings.Contains(res.Error, "budget") || !strings.Contains(res.Error, "2 of 2 LLM calls") {
		t.Fatalf("error = %q", res.Error)
	}
	if len(res.Nodes) != 3 || res.Nodes[2].Trace["error"] == nil {
		t.Fatalf("the refused node should still leave a trace: %+v", res.Nodes)
	}
}

func TestChainAIBudget_TokenLimit(t *testing.T) {
	calls := 0
	res := budgetChain(t, `{"aiBudget":{"maxTokens":100}}`, "ai",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "ok", PromptTokens: 80, CompletionTokens: 10}, nil
		})

	// 90 after the first (within 100), 180 after the second (over), third refused
	if calls != 2 || res.Success || !strings.Contains(res.Error, "180 of 100 tokens") {
		t.Fatalf("calls=%d success=%v error=%q", calls, res.Success, res.Error)
	}
}

func TestChainAIBudget_AppliesToAIOperationToo(t *testing.T) {
	calls := 0
	res := budgetChain(t, `{"aiBudget":{"maxLLMCalls":1}}`, "ai.operation",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: `{"k":"v"}`, LLMCalls: 1}, nil
		})

	if calls != 1 || res.Success || !strings.Contains(res.Error, "budget") {
		t.Fatalf("calls=%d success=%v error=%q", calls, res.Success, res.Error)
	}
}

func TestChainAIBudget_RetriesDrawOnTheSameBudget(t *testing.T) {
	calls := 0
	registry := NewRegistry()
	registry.Register("ai.operation", &aiOperationExecutor{call: func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
		calls++
		return &AIOperationResult{Output: "never JSON", LLMCalls: 1}, nil
	}})
	engine := NewEngine(registry)
	engine.RegisterChain(&Chain{
		ID: "r", EntryNode: "n", Config: []byte(`{"aiBudget":{"maxLLMCalls":2}}`),
		Nodes: []ChainNode{{ID: "n", Type: "ai.operation", Config: []byte(`{"agent":"a","prompt":"p","outputSchema":{"k":"string"},"maxRetries":5}`)}},
	})

	res, _ := engine.Run(context.Background(), "r", nil)
	if calls != 2 {
		t.Fatalf("a retry loop must not outrun the budget: model called %d times", calls)
	}
	if res.Success || !strings.Contains(res.Error, "budget") {
		t.Fatalf("error = %q", res.Error)
	}
}

func TestChainAIBudget_OptionalNodeSurvivesAnExhaustedBudget(t *testing.T) {
	registry := NewRegistry()
	registry.Register("ai", &aiExecutor{call: func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
		return &AIOperationResult{Output: "ok"}, nil
	}})
	engine := NewEngine(registry)
	engine.RegisterChain(&Chain{
		ID: "o", EntryNode: "a", Config: []byte(`{"aiBudget":{"maxLLMCalls":1}}`),
		Nodes: []ChainNode{
			{ID: "a", Type: "ai", Config: []byte(`{"agent":"x","prompt":"p"}`)},
			{ID: "b", Type: "ai", Config: []byte(`{"agent":"x","prompt":"p","optional":true}`)},
		},
		Edges: []ChainEdge{{From: "a", To: "b"}},
	})

	res, _ := engine.Run(context.Background(), "o", nil)
	if !res.Success {
		t.Fatalf("an optional AI node should let the chain go on when the budget is gone: %s", res.Error)
	}
}

func TestChainAIBudget_NoLimitsMeansNone(t *testing.T) {
	calls := 0
	res := budgetChain(t, ``, "ai",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "ok", PromptTokens: 1_000_000}, nil
		})
	if !res.Success || calls != 3 {
		t.Fatalf("calls=%d success=%v error=%q", calls, res.Success, res.Error)
	}
}

func TestChainAIBudget_EnvDefault(t *testing.T) {
	t.Setenv("AI_MAX_LLM_CALLS_PER_RUN", "1")
	calls := 0
	res := budgetChain(t, ``, "ai",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "ok"}, nil
		})
	if calls != 1 || res.Success {
		t.Fatalf("the platform default should apply to a chain with no limits of its own: calls=%d", calls)
	}

	// the chain's own limits win over the default
	calls = 0
	res = budgetChain(t, `{"aiBudget":{"maxLLMCalls":3}}`, "ai",
		func(context.Context, string, string, string, bool) (*AIOperationResult, error) {
			calls++
			return &AIOperationResult{Output: "ok"}, nil
		})
	if calls != 3 || !res.Success {
		t.Fatalf("calls=%d success=%v", calls, res.Success)
	}
}
