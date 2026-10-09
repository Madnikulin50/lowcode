package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func fieldOf(t *testing.T, defs []nodeTypeDef, nodeType, key string) *nodeTypeField {
	t.Helper()
	for i := range defs {
		if defs[i].Type != nodeType {
			continue
		}
		for j := range defs[i].ConfigFields {
			if defs[i].ConfigFields[j].Key == key {
				return &defs[i].ConfigFields[j]
			}
		}
	}
	return nil
}

func TestEnrichAINodes(t *testing.T) {
	defs := enrichAINodes(builtinNodeTypes(),
		[]agentChoice{{"zeta-agent", "Does Z"}, {"alpha-agent", ""}},
		[]string{"rulesgo.ai", "qwen3:8b"})

	for _, nodeType := range []string{"ai", "ai.operation"} {
		agent := fieldOf(t, defs, nodeType, "agent")
		if agent == nil || len(agent.Options) != 2 || agent.Options[0] != "alpha-agent" {
			t.Fatalf("%s agent options = %+v, want the live agents, sorted", nodeType, agent)
		}
		if agent.OptionLabels["zeta-agent"] != "Does Z" || len(agent.OptionLabels) != 1 {
			t.Fatalf("%s option labels = %v", nodeType, agent.OptionLabels)
		}

		model := fieldOf(t, defs, nodeType, "model")
		if model == nil || len(model.Suggestions) != 2 || model.Widget != "string" {
			t.Fatalf("%s model = %+v: suggestions only, it must stay free text", nodeType, model)
		}
	}

	schema := fieldOf(t, defs, "ai.operation", "outputSchema")
	if schema == nil || len(schema.ValueOptions) != 5 {
		t.Fatalf("outputSchema value options = %+v", schema)
	}

	// settings the runtime supports that the form never offered
	if fieldOf(t, defs, "ai", "timeout") == nil || fieldOf(t, defs, "ai", "optional") == nil {
		t.Fatal("ai node should expose timeout and optional")
	}

	// other nodes are left alone
	if f := fieldOf(t, defs, "http", "url"); f != nil && len(f.Options) > 0 {
		t.Fatalf("non-AI node changed: %+v", f)
	}
}

func TestEnrichAINodes_KeepsStaticAgentsWhenNoneKnown(t *testing.T) {
	defs := enrichAINodes(builtinNodeTypes(), nil, nil)
	agent := fieldOf(t, defs, "ai", "agent")
	if agent == nil || len(agent.Options) == 0 {
		t.Fatalf("with no live agents the static fallback must remain, got %+v", agent)
	}
}

func TestEnrichAINodes_DoesNotLeakIntoTheNextCall(t *testing.T) {
	enrichAINodes(builtinNodeTypes(), []agentChoice{{"only-me", ""}}, nil)

	again := enrichAINodes(builtinNodeTypes(), nil, nil)
	for _, o := range fieldOf(t, again, "ai", "agent").Options {
		if o == "only-me" {
			t.Fatal("a previous call's agents leaked into the shared catalog")
		}
	}
}

func TestWithoutMutation(t *testing.T) {
	got, err := withoutMutation(json.RawMessage(`{"agent":"a","allowMutating":true}`))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(got, &m)
	if m["allowMutating"] != false || m["agent"] != "a" {
		t.Fatalf("config = %v", m)
	}

	got, _ = withoutMutation(nil)
	_ = json.Unmarshal(got, &m)
	if m["allowMutating"] != false {
		t.Fatalf("an empty config must still forbid mutation: %v", m)
	}

	if _, err = withoutMutation(json.RawMessage(`[1]`)); err == nil {
		t.Fatal("a non-object config must be rejected")
	}
}

func TestNodeTest_RejectsNonAINodes(t *testing.T) {
	allowNodeTesting(t, true)
	for _, body := range []string{
		`{"type":"http","config":{}}`,
		`{"type":"crud","config":{}}`,
		`{"type":"","config":{}}`,
		`not json`,
	} {
		rec := httptest.NewRecorder()
		RuleChainAdmin{}.NodeTest(rec, httptest.NewRequest(http.MethodPost, "/admin/rulechain/node-test", bytes.NewBufferString(body)))

		out := rec.Body.String()
		if !strings.Contains(out, "only AI nodes") && !strings.Contains(out, "invalid JSON") {
			t.Errorf("body %q: response %q should refuse the request", body, out)
		}
	}
}

func TestNodeTest_RunsAnAINodeWithMutationForcedOff(t *testing.T) {
	allowNodeTesting(t, true)
	var sawMutating = true
	registry := rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{
		AICall: func(_ context.Context, _, prompt, _ string, allowMutating bool) (*rulesgo.AIOperationResult, error) {
			sawMutating = allowMutating
			return &rulesgo.AIOperationResult{Output: prompt}, nil
		},
	})

	prev := handlers.RuleEngine
	handlers.SetRuleEngine(rulesgo.NewEngineWithPersistence(registry, nil))
	defer handlers.SetRuleEngine(prev)

	rec := httptest.NewRecorder()
	body := `{"type":"ai","config":{"agent":"assistant","prompt":"Hi {{name}}","allowMutating":true},"input":{"name":"Acme"}}`
	RuleChainAdmin{}.NodeTest(rec, httptest.NewRequest(http.MethodPost, "/admin/rulechain/node-test", bytes.NewBufferString(body)))

	if sawMutating {
		t.Fatal("the test button must never allow data-changing tool calls, even if the saved node does")
	}

	var resp struct {
		Response rulesgo.NodeResult `json:"response"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	if resp.Response.Output["response"] != "Hi Acme" {
		t.Fatalf("response = %s", rec.Body.String())
	}
}

// The validator is only as good as the catalog it is given; check it against
// the real one with chains of the kind an assistant writes.
func TestRuleChainValidator_AgainstTheRealCatalog(t *testing.T) {
	catalog := ruleChainSchemas()
	if len(catalog) < 10 {
		t.Fatalf("catalog looks truncated: %d node types", len(catalog))
	}

	issuesFor := func(nodes, edges string) []rulesgo.ChainIssue {
		var c rulesgo.Chain
		c.Name = "t"
		if err := json.Unmarshal([]byte(nodes), &c.Nodes); err != nil {
			t.Fatal(err)
		}
		if edges != "" {
			if err := json.Unmarshal([]byte(edges), &c.Edges); err != nil {
				t.Fatal(err)
			}
		}
		return rulesgo.ValidateChain(&c, catalog, rulesgo.ValidateOptions{})
	}

	// a realistic chain: branch on a field, ask the AI, mail the result
	good := issuesFor(`[
	  {"id":"check","type":"condition","config":{"field":"amount","operator":"gt","value":"1000"}},
	  {"id":"ask","type":"ai.operation","config":{"agent":"assistant","prompt":"Assess {{customer}}","outputSchema":{"risk":"number"},"maxRetries":1}},
	  {"id":"notify","type":"mail","config":{"to":"risk@example.com","subject":"Review","body":"{{ask.result}}"}}
	]`, `[{"from":"check","to":"ask","condition":"passed"},{"from":"ask","to":"notify"}]`)
	for _, i := range good {
		if i.Severity == "error" {
			t.Errorf("a sound chain was rejected: %s", i)
		}
	}

	// the usual mistakes
	bad := issuesFor(`[
	  {"id":"ask","type":"ai","config":{"agent":"gandalf"}},
	  {"id":"x","type":"quantum-leap","config":{}}
	]`, "")
	text := ""
	for _, i := range bad {
		text += i.String() + "\n"
	}
	for _, want := range []string{`"agent" is "gandalf"`, `needs the setting "prompt"`, `unknown node type "quantum-leap"`} {
		if !strings.Contains(text, want) {
			t.Errorf("issues lack %q:\n%s", want, text)
		}
	}
	if !strings.Contains(text, "assistant") {
		t.Errorf("the allowed agents should be listed from the live registry:\n%s", text)
	}
}

func TestRuleChainSchemas_CoverEveryNodeType(t *testing.T) {
	defs := nodeTypes()
	schemas := ruleChainSchemas()
	if len(schemas) != len(defs) {
		t.Fatalf("%d node types in the catalog, %d schemas", len(defs), len(schemas))
	}
	for i, d := range defs {
		if schemas[i].Type != d.Type || len(schemas[i].Fields) != len(d.ConfigFields) {
			t.Errorf("%s: schema does not match the catalog entry", d.Type)
		}
	}
}

func TestNodeTest_StreamsProgressThenTheResult(t *testing.T) {
	allowNodeTesting(t, true)
	registry := rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{
		AICall: func(ctx context.Context, _, prompt string, _ string, _ bool) (*rulesgo.AIOperationResult, error) {
			aiagent.ReportProgress(ctx, aiagent.Progress{Token: "Hi "})
			aiagent.ReportProgress(ctx, aiagent.Progress{Token: "Acme"})
			return &rulesgo.AIOperationResult{Output: prompt}, nil
		},
	})
	prev := handlers.RuleEngine
	handlers.SetRuleEngine(rulesgo.NewEngineWithPersistence(registry, nil))
	defer handlers.SetRuleEngine(prev)

	rec := httptest.NewRecorder()
	body := `{"type":"ai","config":{"agent":"assistant","prompt":"Hi {{name}}"},"input":{"name":"Acme"}}`
	RuleChainAdmin{}.NodeTest(rec, httptest.NewRequest(http.MethodPost, "/admin/rulechain/node-test?stream=1", bytes.NewBufferString(body)))

	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("not a stream: %s", rec.Body.String())
	}

	var events []map[string]interface{}
	for _, block := range strings.Split(strings.TrimSpace(rec.Body.String()), "\n\n") {
		var ev map[string]interface{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(block, "data: ")), &ev); err != nil {
			t.Fatalf("bad event %q: %v", block, err)
		}
		events = append(events, ev)
	}

	if len(events) != 3 || events[0]["token"] != "Hi " || events[1]["token"] != "Acme" {
		t.Fatalf("events = %v", events)
	}
	last := events[2]
	node, _ := last["result"].(map[string]interface{})
	if last["done"] != true || node == nil || node["output"] == nil || node["trace"] == nil {
		t.Fatalf("final event = %v", last)
	}
}

func TestNodeTest_StreamModeRefusesNonAINodes(t *testing.T) {
	allowNodeTesting(t, true)
	rec := httptest.NewRecorder()
	RuleChainAdmin{}.NodeTest(rec, httptest.NewRequest(http.MethodPost, "/admin/rulechain/node-test?stream=1", bytes.NewBufferString(`{"type":"http","config":{}}`)))
	if rec.Header().Get("Content-Type") == "text/event-stream" || !strings.Contains(rec.Body.String(), "only AI nodes") {
		t.Fatalf("got %q", rec.Body.String())
	}
}

func allowNodeTesting(t *testing.T, allowed bool) {
	t.Helper()
	prev := canTestChainNode
	canTestChainNode = func(context.Context) bool { return allowed }
	t.Cleanup(func() { canTestChainNode = prev })
}

func TestNodeTest_NeedsPermission(t *testing.T) {
	allowNodeTesting(t, false)

	called := false
	registry := rulesgo.DefaultRegistry(&rulesgo.DefaultConfig{
		AICall: func(context.Context, string, string, string, bool) (*rulesgo.AIOperationResult, error) {
			called = true
			return &rulesgo.AIOperationResult{Output: "x"}, nil
		},
	})
	prev := handlers.RuleEngine
	handlers.SetRuleEngine(rulesgo.NewEngineWithPersistence(registry, nil))
	defer handlers.SetRuleEngine(prev)

	for _, url := range []string{"/admin/rulechain/node-test", "/admin/rulechain/node-test?stream=1"} {
		rec := httptest.NewRecorder()
		RuleChainAdmin{}.NodeTest(rec, httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(`{"type":"ai","config":{"agent":"a","prompt":"p"}}`)))

		if called {
			t.Fatalf("%s: an unauthorised caller reached the model", url)
		}
		if rec.Header().Get("Content-Type") == "text/event-stream" || !strings.Contains(rec.Body.String(), "not allowed") {
			t.Fatalf("%s: got %q", url, rec.Body.String())
		}
	}
}
