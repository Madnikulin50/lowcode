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

// DefaultNeedsConfirm matches compose chat: create_* / delete_* require a user "да".
func DefaultNeedsConfirm(calls []Call) bool {
	for _, c := range calls {
		n := c.Name
		if strings.HasPrefix(n, "create_") || strings.HasPrefix(n, "delete_") {
			return true
		}
		if strings.HasSuffix(n, "_restore") || strings.HasSuffix(n, "_prune") {
			return true
		}
	}
	return false
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
