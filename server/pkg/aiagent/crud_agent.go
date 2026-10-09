package aiagent

import (
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func NewCRUDAgent(client *chat.Client, tools []chat.ToolDef) *Agent {
	a := NewFromSpec(client, mustBuiltin("crud-agent"))
	a.cfg.Tools = tools
	return a
}

func NewAssistantAgent(client *chat.Client, tools []chat.ToolDef) *Agent {
	a := NewFromSpec(client, mustBuiltin("assistant"))
	a.cfg.Tools = tools
	return a
}

func mustBuiltin(handle string) AgentSpec {
	for _, s := range BuiltinSpecs() {
		if s.Handle == handle {
			return s
		}
	}
	return AgentSpec{Handle: handle, MaxSteps: 5}
}
