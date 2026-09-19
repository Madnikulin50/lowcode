package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// Param describes one named argument a tool accepts. Every value arrives as
// a string (both the chat XML parser and MCP's WithString args work that
// way); Type is metadata used for prompt/schema generation only.
type Param struct {
	Name        string
	Type        string
	Required    bool
	Description string
}

// Def is the single source of truth for a tool that both the chat
// interface ([chatService.Chat], [chatService.getTools]) and the MCP server
// ([github.com/madnikulin50/lowcode/server/compose/mcp/handlers]) expose.
//
// One Handler resolves the request against the service layer and returns
// structured data (or an error); each transport renders that data however
// it needs — MCP always as JSON, chat with an optional bespoke Render for
// tools that predate this registry and have a hand-tuned reply format.
type Def struct {
	Name        string
	Description string
	Params      []Param

	// Mutating marks tools that change data. Chat requires an explicit "да"
	// from the user before calling these — see aiagent.DefaultNeedsConfirm.
	Mutating bool

	// Handler executes the tool. ctx carries the caller's identity and,
	// for chat, the resolved namespace via args["namespaceID"] (chat's
	// aiagent runtime injects it from ChatPromptArguments.Namespace).
	Handler func(ctx context.Context, args map[string]string) (any, error)

	// Render formats successful data for chat's human-readable replies.
	// nil uses defaultRender (raw string passthrough, else JSON) — the
	// same shape MCP always uses.
	Render func(data any) string
}

// defaultRender passes plain confirmation/status strings through as-is and
// JSON-encodes everything else. It is what MCP tool results always use, and
// what chat falls back to for tools without a bespoke Render.
func defaultRender(data any) string {
	if s, ok := data.(string); ok {
		return s
	}
	return toJSON(data)
}

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":%q}`, err.Error())
	}
	return string(b)
}

func (d Def) render(data any) string {
	if d.Render != nil {
		return d.Render(data)
	}
	return defaultRender(data)
}

// ToChatToolDefs adapts registry Defs into chat.ToolDef. Chat tool handlers
// never return a Go error — a Handler error becomes the reply text, matching
// how every hand-written chat tool in this package already reports failure.
func ToChatToolDefs(defs ...Def) []chat.ToolDef {
	out := make([]chat.ToolDef, len(defs))
	for i, d := range defs {
		out[i] = d.toChatToolDef()
	}
	return out
}

func (d Def) toChatToolDef() chat.ToolDef {
	params := make([]chat.ParamDef, len(d.Params))
	for i, p := range d.Params {
		params[i] = chat.ParamDef{Name: p.Name, Type: p.Type, Required: p.Required, Description: p.Description}
	}
	return chat.ToolDef{
		Name:        d.Name,
		Description: d.Description,
		Params:      params,
		Mutating:    d.Mutating,
		Handler: func(ctx context.Context, args map[string]string) string {
			data, err := d.Handler(ctx, args)
			if err != nil {
				return err.Error()
			}
			return d.render(data)
		},
	}
}

// byName picks one Def by name out of a slice built by a *ToolDefs()
// constructor — used where only a subset of a registry's tools should be
// offered to chat (see Chat() in chat.go for why the static chat tool list
// stays intentionally small).
func byName(defs []Def, name string) Def {
	for _, d := range defs {
		if d.Name == name {
			return d
		}
	}
	panic("service: unknown tool def " + name)
}
