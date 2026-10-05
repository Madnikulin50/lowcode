package rulesgo

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

var testCatalog = []NodeSchema{
	{Type: "condition", Description: "Branch on a field", Fields: []FieldSchema{
		{Key: "field", Widget: "string", Required: true},
		{Key: "operator", Widget: "enum", Required: true, Options: []string{"eq", "neq", "empty"}},
		{Key: "value", Widget: "string", VisibleIf: map[string][]string{"operator": {"eq", "neq"}}},
	}},
	{Type: "ai", Fields: []FieldSchema{
		{Key: "agent", Widget: "enum", Required: true, Options: []string{"assistant", "crud-agent"}},
		{Key: "prompt", Widget: "textarea", Required: true},
		{Key: "timeout", Widget: "number"},
		{Key: "optional", Widget: "bool"},
		{Key: "inputs", Widget: "keymap"},
		{Key: "tags", Widget: "stringlist"},
		{Key: "bands", Widget: "objectlist"},
		{Key: "level", Widget: "enum", Default: "low", Required: true, Options: []string{"low", "high"}},
	}},
	{Type: "mail", Fields: []FieldSchema{{Key: "to", Widget: "string", Required: true}}},
}

func node(id, typ, cfg string) ChainNode {
	n := ChainNode{ID: id, Type: typ}
	if cfg != "" {
		n.Config = json.RawMessage(cfg)
	}
	return n
}

func issuesOf(c *Chain, opts ...ValidateOptions) string {
	var o ValidateOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	var out []string
	for _, i := range ValidateChain(c, testCatalog, o) {
		out = append(out, i.Severity+": "+i.String())
	}
	return strings.Join(out, "\n")
}

func TestValidateChain_ValidChainHasNoIssues(t *testing.T) {
	c := &Chain{
		Name: "ok", EntryNode: "check",
		Nodes: []ChainNode{
			node("check", "condition", `{"field":"amount","operator":"eq","value":"10"}`),
			node("ask", "ai", `{"agent":"assistant","prompt":"Hello {{name}}","timeout":30,"optional":true,"inputs":{"a":"b"}}`),
			node("send", "mail", `{"to":"x@example.com"}`),
		},
		Edges: []ChainEdge{{From: "check", To: "ask", Condition: "passed"}, {From: "ask", To: "send"}},
	}
	if got := issuesOf(c); got != "" {
		t.Fatalf("unexpected issues:\n%s", got)
	}
}

func TestValidateChain_Problems(t *testing.T) {
	cases := map[string]struct {
		chain *Chain
		want  []string
	}{
		"empty chain": {&Chain{Name: "x"}, []string{"error: chain: the chain has no nodes"}},
		"no name":     {&Chain{Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`)}}, []string{"name is required"}},
		"unknown type": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "teleport", "")}},
			[]string{`node "a": unknown node type "teleport"`, "condition, mail"},
		},
		"missing required": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"assistant"}`)}},
			[]string{`needs the setting "prompt"`},
		},
		"missing required enum lists options": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "condition", `{"field":"f"}`)}},
			[]string{`needs the setting "operator" (one of: eq, neq, empty)`},
		},
		"bad enum value": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"wizard","prompt":"p"}`)}},
			[]string{`"agent" is "wizard"`, "assistant, crud-agent"},
		},
		"enum placeholder is allowed": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"{{which}}","prompt":"p"}`)}},
			nil,
		},
		"hidden field is not required": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "condition", `{"field":"f","operator":"empty"}`)}},
			nil,
		},
		"default satisfies required": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"assistant","prompt":"p"}`)}},
			nil,
		},
		"wrong types": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"assistant","prompt":"p","timeout":"soon","optional":"maybe","inputs":["a"],"tags":"x","bands":[1]}`)}},
			[]string{`"timeout" must be a number`, `"optional" must be true or false`, `"inputs" must be an object`, `"tags" must be a list`, `item 1 must be an object`},
		},
		"numeric strings are fine": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"assistant","prompt":"p","timeout":"30","optional":"true"}`)}},
			nil,
		},
		"config is not an object": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `[1]`)}},
			[]string{"config must be a JSON object"},
		},
		"unknown setting is a warning": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x","cc":"y"}`)}},
			[]string{`warning: node "a": "cc" is not a setting of mail`, "Its settings: to"},
		},
		"duplicate and missing ids": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`), node("a", "mail", `{"to":"x"}`), node("", "mail", `{"to":"x"}`)}},
			[]string{"the id is used twice", "has no id"},
		},
		"bad entry node": {
			&Chain{Name: "x", EntryNode: "zzz", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`)}},
			[]string{`entryNode "zzz" is not one of the nodes (a)`},
		},
		"bad edges": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`)},
				Edges: []ChainEdge{{From: "a", To: "b"}, {From: "q", To: "a"}, {From: "a", To: "a"}}},
			[]string{`unknown node "b"`, `unknown node "q"`, "cannot lead to itself"},
		},
		"expression as condition": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`), node("b", "mail", `{"to":"x"}`)},
				Edges: []ChainEdge{{From: "a", To: "b", Condition: "amount > 100"}}},
			[]string{"warning", "name of a variable"},
		},
		"unreachable node": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`), node("island", "mail", `{"to":"x"}`)}},
			[]string{`node "island": is not reachable`},
		},
		"loop": {
			&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x"}`), node("b", "mail", `{"to":"x"}`)},
				Edges: []ChainEdge{{From: "a", To: "b"}, {From: "b", To: "a"}}},
			[]string{"form a loop"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := issuesOf(c.chain)
			if len(c.want) == 0 {
				if got != "" {
					t.Fatalf("expected no issues, got:\n%s", got)
				}
				return
			}
			for _, phrase := range c.want {
				if !strings.Contains(got, phrase) {
					t.Errorf("issues do not contain %q:\n%s", phrase, got)
				}
			}
		})
	}
}

func TestValidateChain_SeverityDecidesWhetherItBlocks(t *testing.T) {
	warnOnly := ValidateChain(&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", `{"to":"x","cc":"y"}`)}}, testCatalog, ValidateOptions{})
	if len(warnOnly) == 0 || HasErrors(warnOnly) {
		t.Fatalf("an ignored setting is a warning, not an error: %+v", warnOnly)
	}

	broken := ValidateChain(&Chain{Name: "x", Nodes: []ChainNode{node("a", "mail", "")}}, testCatalog, ValidateOptions{})
	if !HasErrors(broken) {
		t.Fatalf("a missing required setting is an error: %+v", broken)
	}
}

func TestValidateChain_LibraryPromptReferences(t *testing.T) {
	check := func(ref string) error {
		if strings.Contains(ref, "gone") {
			return errors.New("prompt does not exist")
		}
		return nil
	}
	chain := func(prompt string) *Chain {
		return &Chain{Name: "x", Nodes: []ChainNode{node("a", "ai", `{"agent":"assistant","prompt":"`+prompt+`"}`)}}
	}

	if got := issuesOf(chain("@prompt:triage"), ValidateOptions{PromptExists: check}); got != "" {
		t.Fatalf("an existing prompt is fine: %s", got)
	}
	got := issuesOf(chain("@prompt:gone@2"), ValidateOptions{PromptExists: check})
	if !strings.Contains(got, "library prompt that cannot be used") || !strings.Contains(got, "does not exist") {
		t.Fatalf("a missing prompt is an error: %s", got)
	}
	// prose that merely mentions one is not a reference
	if got := issuesOf(chain("see @prompt:gone for context"), ValidateOptions{PromptExists: check}); got != "" {
		t.Fatalf("only a whole-field reference counts: %s", got)
	}
	// without a checker nothing is verified
	if got := issuesOf(chain("@prompt:gone")); got != "" {
		t.Fatalf("no checker, no check: %s", got)
	}
}

func TestDescribeNodeTypes(t *testing.T) {
	out := DescribeNodeTypes(testCatalog)
	for _, want := range []string{
		"ai - ", "condition - Branch on a field",
		"agent (enum, required; one of: assistant, crud-agent)",
		"level (enum, required; one of: low, high; default low)",
		"NAME of a variable", "@prompt:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("catalog lacks %q:\n%s", want, out)
		}
	}
	if strings.Index(out, "ai - ") > strings.Index(out, "condition - ") {
		t.Error("node types should be listed alphabetically")
	}
}
