package rulesgo

import (
	"testing"
)

func TestAttachmentIDsFromValue(t *testing.T) {
	got := AttachmentIDsFromValue("11\n22,33")
	if len(got) != 3 || got[0] != 11 || got[2] != 33 {
		t.Fatalf("%v", got)
	}
}

func TestParseChainTriggers(t *testing.T) {
	c := &Chain{Config: []byte(`{"triggers":[{"resourceType":"compose:record","eventType":"afterCreate,afterUpdate","moduleHandle":"documents","async":true}]}`)}
	tt := ParseChainTriggers(c)
	if len(tt) != 1 || tt[0].ModuleHandle != "documents" {
		t.Fatalf("%+v", tt)
	}
	if !tt[0].MatchesEvent("compose:record", "afterUpdate", "documents") {
		t.Fatal("should match")
	}
	if tt[0].MatchesEvent("compose:record", "afterUpdate", "projects") {
		t.Fatal("module mismatch")
	}
}
