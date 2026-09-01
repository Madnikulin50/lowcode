package rulesgo

import (
	"encoding/json"
	"strings"
)

// ChainTrigger is a record-event subscription stored in chain Config.
type ChainTrigger struct {
	ResourceType string `json:"resourceType"`
	EventType    string `json:"eventType"`
	ModuleHandle string `json:"moduleHandle"`
	Async        bool   `json:"async"`
	FileField    string `json:"fileField"`
}

type chainConfigEnvelope struct {
	Triggers []ChainTrigger `json:"triggers"`
}

// ParseChainTriggers reads Config.triggers (missing/invalid → empty).
func ParseChainTriggers(c *Chain) []ChainTrigger {
	if c == nil || len(c.Config) == 0 {
		return nil
	}
	var env chainConfigEnvelope
	if err := json.Unmarshal(c.Config, &env); err != nil {
		return nil
	}
	var out []ChainTrigger
	for _, t := range env.Triggers {
		t.ResourceType = strings.TrimSpace(t.ResourceType)
		t.EventType = strings.TrimSpace(t.EventType)
		t.ModuleHandle = strings.TrimSpace(t.ModuleHandle)
		t.FileField = strings.TrimSpace(t.FileField)
		if t.ResourceType == "" {
			t.ResourceType = "compose:record"
		}
		if t.EventType == "" {
			t.EventType = "afterCreate,afterUpdate"
		}
		if t.FileField == "" {
			t.FileField = "file"
		}
		out = append(out, t)
	}
	return out
}

func (t ChainTrigger) MatchesEvent(resourceType, eventType, moduleHandle string) bool {
	if t.ResourceType != "" && !strings.EqualFold(t.ResourceType, resourceType) {
		return false
	}
	if t.ModuleHandle != "" && !strings.EqualFold(t.ModuleHandle, moduleHandle) {
		return false
	}
	if t.EventType == "" {
		return true
	}
	want := strings.ToLower(strings.TrimSpace(eventType))
	for _, p := range strings.Split(t.EventType, ",") {
		if strings.EqualFold(strings.TrimSpace(p), want) {
			return true
		}
	}
	return false
}
