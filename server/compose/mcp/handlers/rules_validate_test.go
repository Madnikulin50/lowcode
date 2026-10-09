package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/mark3labs/mcp-go/server"
)

var chainTestCatalog = []rulesgo.NodeSchema{
	{Type: "mail", Description: "Send an email", Fields: []rulesgo.FieldSchema{{Key: "to", Widget: "string", Required: true}}},
	{Type: "ai", Description: "Ask an agent", Fields: []rulesgo.FieldSchema{
		{Key: "agent", Widget: "enum", Required: true, Options: []string{"assistant"}},
		{Key: "prompt", Widget: "textarea", Required: true},
	}},
}

func withChainCatalog(t *testing.T) {
	t.Helper()
	SetRuleChainCatalog(func() []rulesgo.NodeSchema { return chainTestCatalog })

	prev := RuleEngine
	RuleEngine = rulesgo.NewEngineWithPersistence(rulesgo.NewRegistry(), nil)
	t.Cleanup(func() {
		SetRuleChainCatalog(nil)
		RuleEngine = prev
	})
}

func TestRuleChainTools_AreRegistered(t *testing.T) {
	s := server.NewMCPServer("test", "0")
	initRules(context.Background(), s)
	for _, name := range []string{"rulechain_node_types", "validate_rule_chain", "create_rule_chain"} {
		if _, ok := s.ListTools()[name]; !ok {
			t.Errorf("tool %q not registered", name)
		}
	}
}

func TestNodeTypesTool(t *testing.T) {
	ctx := context.Background()

	res, _ := handleRuleChainNodeTypes(ctx, callWith(nil))
	if got := resultText(t, res); !strings.Contains(got, "not available") {
		t.Fatalf("without a catalog: %q", got)
	}

	withChainCatalog(t)
	res, _ = handleRuleChainNodeTypes(ctx, callWith(nil))
	got := resultText(t, res)
	for _, want := range []string{"mail - Send an email", "to (string, required)", "agent (enum, required; one of: assistant)"} {
		if !strings.Contains(got, want) {
			t.Errorf("catalog lacks %q:\n%s", want, got)
		}
	}
}

func TestValidateRuleChainTool(t *testing.T) {
	withChainCatalog(t)
	ctx := context.Background()

	res, _ := handleValidateRuleChain(ctx, callWith(map[string]interface{}{
		"nodes": `[{"id":"a","type":"mail","config":{"to":"x@example.com"}}]`,
	}))
	if got := resultText(t, res); !strings.Contains(got, `"valid":true`) || !strings.Contains(got, "nothing was saved") {
		t.Fatalf("valid chain: %s", got)
	}

	res, _ = handleValidateRuleChain(ctx, callWith(map[string]interface{}{
		"nodes": `[{"id":"a","type":"sms","config":{}},{"id":"b","type":"mail","config":{}}]`,
		"edges": `[{"from":"a","to":"nope"}]`,
	}))
	got := resultText(t, res)
	for _, want := range []string{`"valid":false`, `unknown node type`, `needs the setting`, `unknown node`} {
		if !strings.Contains(got, want) {
			t.Errorf("issues lack %q: %s", want, got)
		}
	}

	res, _ = handleValidateRuleChain(ctx, callWith(map[string]interface{}{"nodes": "{not json"}))
	if got := resultText(t, res); !strings.Contains(got, "nvalid nodes JSON") {
		t.Fatalf("bad JSON: %s", got)
	}

	SetRuleChainCatalog(nil)
	res, _ = handleValidateRuleChain(ctx, callWith(map[string]interface{}{"nodes": "[]"}))
	if got := resultText(t, res); !strings.Contains(got, "not available") {
		t.Fatalf("no catalog: %s", got)
	}
}

func TestCreateRuleChain_RefusesBrokenChains(t *testing.T) {
	withChainCatalog(t)
	ctx := context.Background()

	res, _ := handleCreateRuleChain(ctx, callWith(map[string]interface{}{
		"name":  "broken",
		"nodes": `[{"id":"a","type":"mail","config":{}}]`,
	}))
	got := resultText(t, res)
	if !strings.Contains(got, `"created":false`) || !strings.Contains(got, `needs the setting`) {
		t.Fatalf("a chain with errors must not be saved: %s", got)
	}
	if RuleEngine.Chain("rc_broken") != nil {
		t.Fatal("the broken chain was registered anyway")
	}

	res, _ = handleCreateRuleChain(ctx, callWith(map[string]interface{}{
		"name":  "fine",
		"nodes": `[{"id":"a","type":"mail","config":{"to":"x@example.com","cc":"y"}}]`,
	}))
	got = resultText(t, res)
	if !strings.Contains(got, `"created":true`) {
		t.Fatalf("a valid chain is saved: %s", got)
	}
	if !strings.Contains(got, `"warnings"`) || !strings.Contains(got, "is not a setting of mail") {
		t.Fatalf("warnings are reported alongside: %s", got)
	}
	if RuleEngine.Chain("rc_fine") == nil {
		t.Fatal("the valid chain was not registered")
	}
}

func TestCreateRuleChain_WithoutACatalogStillWorksAsBefore(t *testing.T) {
	prev := RuleEngine
	RuleEngine = rulesgo.NewEngineWithPersistence(rulesgo.NewRegistry(), nil)
	defer func() { RuleEngine = prev }()
	SetRuleChainCatalog(nil)

	res, _ := handleCreateRuleChain(context.Background(), callWith(map[string]interface{}{
		"name":  "legacy",
		"nodes": `[{"id":"a","type":"anything","config":{}}]`,
	}))
	got := resultText(t, res)
	if !strings.Contains(got, `"created":true`) || !strings.Contains(got, "was not checked") {
		t.Fatalf("with no catalog the old behaviour stays, and says it did not check: %s", got)
	}
	if RuleEngine.Chain("rc_legacy") == nil {
		t.Fatal("chain not registered")
	}
}
