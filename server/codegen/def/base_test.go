package def

import (
	"reflect"
	"testing"
)

func TestSplitWords(t *testing.T) {
	cases := []struct {
		in   string
		seps []string
		want []string
	}{
		{"shared_node_id", []string{"-", "_", "."}, []string{"shared", "node", "id"}},
		{"node-sync", []string{"-", "_", "."}, []string{"node", "sync"}},
		{"exposed.module", []string{"-", "_", "."}, []string{"exposed", "module"}},
		{"plain", []string{"-", "_", "."}, []string{"plain"}},
	}
	for _, c := range cases {
		got := splitWords(c.in, c.seps...)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitWords(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}

	if got := splitWords("", "-", "_", "."); len(got) != 0 {
		t.Errorf(`splitWords("") = %#v, want empty`, got)
	}
}

func TestTitleFirst(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"node": "Node",
		"sync": "Sync",
		"ID":   "ID",
	}
	for in, want := range cases {
		if got := titleFirst(in); got != want {
			t.Errorf("titleFirst(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPascal(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{[]string{"node"}, "Node"},
		{[]string{"node", "sync"}, "NodeSync"},
		{[]string{"exposed", "module"}, "ExposedModule"},
		{nil, ""},
	}
	for _, c := range cases {
		if got := pascal(c.in); got != c.want {
			t.Errorf("pascal(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCamel(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"Node":        "node",
		"NodeSync":    "nodeSync",
		"ExposedNode": "exposedNode",
	}
	for in, want := range cases {
		if got := camel(in); got != want {
			t.Errorf("camel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBaseResolve(t *testing.T) {
	cases := []struct {
		name string
		in   Base
		want ResolvedBase
	}{
		{
			name: "handle only",
			in:   Base{Handle: "node-sync"},
			want: ResolvedBase{
				Handle: "node-sync", Ident: "nodeSync", ExpIdent: "NodeSync",
				IdentPlural: "nodeSyncs", ExpIdentPlural: "NodeSyncs",
			},
		},
		{
			name: "explicit ident/expIdent override",
			in:   Base{Handle: "node", Ident: "customIdent", ExpIdent: "CustomExpIdent"},
			want: ResolvedBase{
				Handle: "node", Ident: "customIdent", ExpIdent: "CustomExpIdent",
				IdentPlural: "customIdents", ExpIdentPlural: "CustomExpIdents",
			},
		},
		{
			name: "dotted and underscored handle",
			in:   Base{Handle: "exposed_module.set"},
			want: ResolvedBase{
				Handle: "exposed_module.set", Ident: "exposedModuleSet", ExpIdent: "ExposedModuleSet",
				IdentPlural: "exposedModuleSets", ExpIdentPlural: "ExposedModuleSets",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.in.Resolve()
			if got != c.want {
				t.Errorf("Base%#v.Resolve() = %#v, want %#v", c.in, got, c.want)
			}
		})
	}
}
