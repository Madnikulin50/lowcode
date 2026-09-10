package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// Call is a single tool invocation (native or XML). Params is JSON object text.
type Call struct {
	Name   string `json:"name"`
	Params string `json:"params"`
}

// DefaultNeedsConfirm is a name-based heuristic: create_* / delete_* require
// a user "да", as do the dynamically generated per-module record mutators
// (module_<handle>_create_record / _update_record / _delete_record — see
// getTools() in compose/service/chat.go) and *_restore / *_prune actions.
//
// This is a fallback for tool sources that don't (yet) declare
// chat.ToolDef.Mutating explicitly — e.g. remote/agent-kit tools resolved
// through aiagent.Catalog. Anything that *does* declare it should be
// checked with NeedsConfirmFromToolDefs instead, which trusts the flag
// rather than guessing from the name.
func DefaultNeedsConfirm(calls []Call) bool {
	for _, c := range calls {
		n := c.Name
		if strings.HasPrefix(n, "create_") || strings.HasPrefix(n, "delete_") {
			return true
		}
		if strings.HasSuffix(n, "_create_record") || strings.HasSuffix(n, "_update_record") || strings.HasSuffix(n, "_delete_record") {
			return true
		}
		if strings.HasSuffix(n, "_restore") || strings.HasSuffix(n, "_prune") {
			return true
		}
	}
	return false
}

// NeedsConfirmFromToolDefs builds a NeedsConfirm function that trusts each
// tool's explicit chat.ToolDef.Mutating flag instead of guessing from its
// name. defs should be the same tool list offered to the model for this
// request (aiagent.Options.Tools) — every callable tool name is expected to
// be in it. A tool absent from defs (should not normally happen, but keeps
// this safe for partially-migrated tool sources) falls back to
// DefaultNeedsConfirm for that call, so nothing loses existing confirmation
// coverage by omission.
func NeedsConfirmFromToolDefs(defs []chat.ToolDef) func([]Call) bool {
	mutating := make(map[string]bool, len(defs))
	for _, d := range defs {
		mutating[d.Name] = d.Mutating
	}
	return func(calls []Call) bool {
		var unknown []Call
		for _, c := range calls {
			isMutating, known := mutating[c.Name]
			if !known {
				unknown = append(unknown, c)
				continue
			}
			if isMutating {
				return true
			}
		}
		if len(unknown) > 0 {
			return DefaultNeedsConfirm(unknown)
		}
		return false
	}
}

func UserConfirmed(prompt string) bool {
	switch strings.ToLower(strings.TrimSpace(prompt)) {
	case "да", "yes", "y", "ok", "ок", "гоу", "do it", "создай", "подтверждаю", "confirm", "выполнить":
		return true
	}
	return false
}

func UserCancelled(prompt string) bool {
	switch strings.ToLower(strings.TrimSpace(prompt)) {
	case "нет", "no", "n", "отмена", "cancel", "не надо", "стоп":
		return true
	}
	return false
}

// CallsFromMessage collects native tool calls, then XML fallback.
func CallsFromMessage(msg *schema.Message) []Call {
	if msg == nil {
		return nil
	}
	if len(msg.ToolCalls) > 0 {
		out := make([]Call, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			out = append(out, Call{Name: tc.Function.Name, Params: tc.Function.Arguments})
		}
		return out
	}
	return CallsFromXML(msg.Content)
}

func CallsFromXML(content string) []Call {
	if !chat.HasToolCallsStr(content) {
		return nil
	}
	xmlCalls := chat.ParseToolCallsStr(content)
	out := make([]Call, 0, len(xmlCalls))
	for _, xc := range xmlCalls {
		paramJSON, err := json.Marshal(xc.Params)
		if err != nil {
			paramJSON = []byte("{}")
		}
		out = append(out, Call{Name: xc.Name, Params: string(paramJSON)})
	}
	return out
}

func CallsFromNative(toolCalls []schema.ToolCall) []Call {
	if len(toolCalls) == 0 {
		return nil
	}
	out := make([]Call, 0, len(toolCalls))
	for _, tc := range toolCalls {
		out = append(out, Call{Name: tc.Function.Name, Params: tc.Function.Arguments})
	}
	return out
}

func MergeCalls(native []schema.ToolCall, content string) []Call {
	if calls := CallsFromNative(native); len(calls) > 0 {
		return calls
	}
	return CallsFromXML(content)
}

func ExecCalls(ctx context.Context, calls []Call, tools []chat.ToolDef, extra map[string]string) string {
	if len(calls) == 0 {
		return ""
	}
	chat.EmitStatus(ctx, chat.StatusUsingTools)
	var results []string
	for _, call := range calls {
		params := paramsFromJSON(call.Params)
		for k, v := range extra {
			if _, ok := params[k]; !ok {
				params[k] = v
			}
		}
		found := false
		for _, t := range tools {
			if t.Name == call.Name {
				results = append(results, t.Handler(ctx, params))
				found = true
				break
			}
		}
		if !found {
			results = append(results, fmt.Sprintf("unknown tool: %s", call.Name))
		}
	}
	return strings.Join(results, "\n")
}

func toChatToolCalls(calls []Call) []chat.ToolCall {
	out := make([]chat.ToolCall, 0, len(calls))
	for _, c := range calls {
		out = append(out, chat.ToolCall{Name: c.Name, Params: paramsFromJSON(c.Params)})
	}
	return out
}

func paramsFromJSON(jsonStr string) map[string]string {
	if jsonStr == "" {
		return map[string]string{}
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return map[string]string{"_raw": jsonStr}
	}
	result := make(map[string]string, len(raw))
	for k, v := range raw {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func extraFromContext(data map[string]interface{}) map[string]string {
	if len(data) == 0 {
		return nil
	}
	out := make(map[string]string, len(data))
	for k, v := range data {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func stripFinal(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "FINAL:") {
		return strings.TrimSpace(strings.TrimPrefix(s, "FINAL:"))
	}
	return s
}
