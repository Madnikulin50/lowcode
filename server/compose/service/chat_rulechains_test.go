package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

func chainTool(t *testing.T, name string) func(context.Context, map[string]string) string {
	t.Helper()
	for _, d := range chatRuleChainToolDefs() {
		if d.Name == name {
			return d.Handler
		}
	}
	t.Fatalf("no tool %q", name)
	return nil
}

func withChainEnv(t *testing.T) {
	t.Helper()
	SetRuleChainCatalog(func() []rulesgo.NodeSchema {
		return []rulesgo.NodeSchema{{Type: "mail", Description: "Send an email", Fields: []rulesgo.FieldSchema{{Key: "to", Widget: "string", Required: true}}}}
	})
	prev := DefaultRuleEngine
	DefaultRuleEngine = rulesgo.NewEngineWithPersistence(rulesgo.NewRegistry(), nil)
	t.Cleanup(func() {
		SetRuleChainCatalog(nil)
		DefaultRuleEngine = prev
	})
}

func TestChatRuleChainTools_Confirmation(t *testing.T) {
	for _, d := range chatRuleChainToolDefs() {
		wantMutating := d.Name == "create_rule_chain"
		if d.Mutating != wantMutating {
			t.Errorf("%s: Mutating = %v, want %v", d.Name, d.Mutating, wantMutating)
		}
		if got := aiagent.DefaultNeedsConfirm([]aiagent.Call{{Name: d.Name}}); got != wantMutating {
			t.Errorf("%s: the assistant asks first = %v, want %v", d.Name, got, wantMutating)
		}
	}

	cat := aiagent.NewCatalog()
	RegisterComposeToolKits(cat)
	if _, ok := cat.Get("rulechains"); !ok {
		t.Fatal("rulechains toolkit is not registered")
	}
	found := false
	for _, n := range aiagent.AssistantKitNames() {
		found = found || n == "rulechains"
	}
	if !found {
		t.Error("the assistant does not offer the rulechains toolkit")
	}
}

func TestChatRuleChainTools_ValidateAndCreate(t *testing.T) {
	withChainEnv(t)
	ctx := context.Background()

	if got := chainTool(t, "list_rule_chain_node_types")(ctx, nil); !strings.Contains(got, "mail - Send an email") {
		t.Fatalf("node types: %s", got)
	}

	good := map[string]string{"name": "notify", "nodes": `[{"id":"m","type":"mail","config":{"to":"a@example.com"}}]`}
	if got := chainTool(t, "validate_rule_chain")(ctx, good); !strings.HasPrefix(got, "OK: 1 nodes, 0 edges") {
		t.Fatalf("valid draft: %s", got)
	}
	if DefaultRuleEngine.Chain("rc_notify") != nil {
		t.Fatal("validate must not save")
	}

	bad := map[string]string{"name": "oops", "nodes": `[{"id":"m","type":"mail","config":{}},{"id":"x","type":"sms"}]`}
	got := chainTool(t, "validate_rule_chain")(ctx, bad)
	if !strings.Contains(got, "Fix the errors") || !strings.Contains(got, `needs the setting "to"`) || !strings.Contains(got, `unknown node type "sms"`) {
		t.Fatalf("broken draft: %s", got)
	}

	if got := chainTool(t, "create_rule_chain")(ctx, bad); !strings.Contains(got, "Not saved") {
		t.Fatalf("a broken draft must not be saved: %s", got)
	}
	if DefaultRuleEngine.Chain("rc_oops") != nil {
		t.Fatal("the broken chain was saved")
	}

	if got := chainTool(t, "create_rule_chain")(ctx, good); !strings.Contains(got, `Saved rule chain "rc_notify"`) {
		t.Fatalf("good draft: %s", got)
	}
	if DefaultRuleEngine.Chain("rc_notify") == nil {
		t.Fatal("the good chain was not saved")
	}

	if got := chainTool(t, "validate_rule_chain")(ctx, map[string]string{"name": "x", "nodes": "[oops"}); !strings.Contains(got, "invalid nodes JSON") {
		t.Fatalf("unreadable JSON: %s", got)
	}
}

func TestChatRuleChainTools_NoCatalog(t *testing.T) {
	SetRuleChainCatalog(nil)
	ctx := context.Background()
	for _, name := range []string{"list_rule_chain_node_types", "validate_rule_chain", "create_rule_chain"} {
		if got := chainTool(t, name)(ctx, map[string]string{"name": "x", "nodes": "[]"}); !strings.Contains(got, "not available") {
			t.Errorf("%s without a catalog: %q", name, got)
		}
	}
}
