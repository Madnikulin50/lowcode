package aiagent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSpecYAML(t *testing.T) {
	s, err := ParseSpecYAML([]byte(`handle: ops
description: Ops
model: mcp.agent
toolkits:
  - cmdb
maxSteps: 6
prompt: Be brief.
`))
	if err != nil {
		t.Fatal(err)
	}
	if s.Handle != "ops" || s.MaxSteps != 6 || len(s.Toolkits) != 1 || s.Toolkits[0] != "cmdb" {
		t.Fatalf("%+v", s)
	}
	if !s.IsEnabled() {
		t.Fatal("nil enabled should be on")
	}
	if _, err := ParseSpecYAML([]byte(`description: no handle`)); err == nil {
		t.Fatal("expected missing handle error")
	}
}

func TestBuiltinSpecs(t *testing.T) {
	specs := BuiltinSpecs()
	got := map[string]AgentSpec{}
	for _, s := range specs {
		got[s.Handle] = s
	}
	for _, h := range []string{"crud-agent", "assistant", "cmdb-operator"} {
		s, ok := got[h]
		if !ok {
			t.Fatalf("missing builtin %s in %#v", h, got)
		}
		if s.Source != "builtin" {
			t.Fatalf("%s source=%s", h, s.Source)
		}
		if s.Confirm {
			t.Fatalf("%s should not confirm (MCP autonomous)", h)
		}
		if len(s.Toolkits) == 0 || s.Prompt == "" {
			t.Fatalf("%s incomplete: %+v", h, s)
		}
	}
}

func TestEffectiveSpecsSettingsOverlay(t *testing.T) {
	t.Cleanup(func() { SetExtrasProvider(nil) })
	SetExtrasProvider(func() []AgentSpec {
		return []AgentSpec{{Handle: "custom", Description: "c", Source: "settings"}}
	})
	got := map[string]AgentSpec{}
	for _, s := range EffectiveSpecs() {
		got[s.Handle] = s
	}
	if _, ok := got["crud-agent"]; !ok {
		t.Fatal("settings overlay must keep builtins")
	}
	if s, ok := got["custom"]; !ok || s.Source != "settings" {
		t.Fatalf("custom: %+v", got)
	}
}

func TestEffectiveSpecsEmptyExtrasKeepBuiltins(t *testing.T) {
	t.Cleanup(func() { SetExtrasProvider(nil) })
	SetExtrasProvider(func() []AgentSpec { return nil })
	if n := len(EffectiveSpecs()); n < 3 {
		t.Fatalf("expected builtins, got %d", n)
	}
}

func TestEffectiveSpecsDropsDisabled(t *testing.T) {
	t.Cleanup(func() { SetExtrasProvider(nil) })
	off := false
	SetExtrasProvider(func() []AgentSpec {
		return []AgentSpec{
			{Handle: "on", Description: "yes"},
			{Handle: "crud-agent", Enabled: &off},
			{Handle: "off", Enabled: &off, Description: "no"},
		}
	})
	got := map[string]bool{}
	for _, s := range EffectiveSpecs() {
		got[s.Handle] = true
	}
	if !got["on"] {
		t.Fatal("expected custom on")
	}
	if got["crud-agent"] {
		t.Fatal("disabled builtin should be gone")
	}
	if got["off"] {
		t.Fatal("disabled custom should be gone")
	}
	if !got["assistant"] {
		t.Fatal("other builtins stay")
	}
}

func TestOverlaySpecsFileWinsAndDisable(t *testing.T) {
	off := false
	out := overlaySpecs(
		[]AgentSpec{{Handle: "crud-agent", Description: "old"}, {Handle: "assistant"}},
		[]AgentSpec{
			{Handle: "crud-agent", Description: "new", Source: "file"},
			{Handle: "assistant", Enabled: &off},
			{Handle: "from-file", Source: "file"},
		},
	)
	if len(out) != 2 {
		t.Fatalf("%+v", out)
	}
	if out[0].Handle != "crud-agent" || out[0].Description != "new" {
		t.Fatalf("replace: %+v", out[0])
	}
	if out[1].Handle != "from-file" {
		t.Fatalf("append: %+v", out)
	}
}

func TestEffectiveSpecsDirOverlay(t *testing.T) {
	t.Cleanup(func() { SetExtrasProvider(nil) })
	dir := t.TempDir()
	raw := []byte("handle: from-file\ndescription: disk\n")
	if err := os.WriteFile(filepath.Join(dir, "x.yaml"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AI_AGENTS_DIR", dir)
	found := false
	for _, s := range EffectiveSpecs() {
		if s.Handle == "from-file" && s.Source == "file" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected file overlay")
	}
}

func TestMergeSpecsKeepDisabled(t *testing.T) {
	off := false
	out := mergeSpecs(
		[]AgentSpec{{Handle: "crud-agent", Description: "old"}},
		[]AgentSpec{{Handle: "crud-agent", Enabled: &off, Description: "off"}},
		true,
	)
	if len(out) != 1 || out[0].IsEnabled() || out[0].Description != "off" {
		t.Fatalf("%+v", out)
	}
}

func TestNewFromSpecConfirmAndCRUDWrapper(t *testing.T) {
	a := NewFromSpec(nil, AgentSpec{Handle: "x", Confirm: true, MaxSteps: 2, Prompt: "p"})
	if !a.cfg.Confirm || a.cfg.MaxSteps != 2 || a.cfg.SystemPrompt != "p" {
		t.Fatalf("%+v", a.cfg)
	}
	crud := NewCRUDAgent(nil, nil)
	if crud.Name() != "crud-agent" || crud.cfg.Confirm {
		t.Fatalf("%s confirm=%v", crud.Name(), crud.cfg.Confirm)
	}
}

func TestRegistryReloadFromSpecs(t *testing.T) {
	prev := DefaultRegistry()
	t.Cleanup(func() {
		SetDefaultRegistry(prev)
		SetExtrasProvider(nil)
	})
	SetExtrasProvider(nil)
	r := NewRegistry(nil)
	r.Reload(nil)
	if r.Get("crud-agent") == nil || r.Get("assistant") == nil || r.Get("cmdb-operator") == nil {
		t.Fatalf("missing builtins: %v", r.List())
	}
	infos := r.ListInfo()
	if len(infos) < 3 {
		t.Fatalf("infos=%v", infos)
	}
	SetExtrasProvider(func() []AgentSpec {
		return []AgentSpec{{Handle: "only-one", Description: "x"}}
	})
	r.Reload(nil)
	if r.Get("crud-agent") == nil {
		t.Fatal("settings overlay must keep crud-agent")
	}
	if r.Get("only-one") == nil {
		t.Fatal("expected only-one")
	}
}

func TestCatalogPayload(t *testing.T) {
	p := CatalogPayload()
	builtins, _ := p["builtins"].([]AgentSpec)
	if len(builtins) < 3 {
		t.Fatalf("builtins=%v", p["builtins"])
	}
}
