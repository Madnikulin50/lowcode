package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type (
	ParamDef struct {
		Name        string
		Type        string
		Required    bool
		Description string
	}

	ToolDef struct {
		Name        string
		Description string
		Params      []ParamDef
		Handler     func(ctx context.Context, params map[string]string) string
	}
)

type toolAdapter struct {
	def ToolDef
}

func (a *toolAdapter) Info(_ context.Context) (*schema.ToolInfo, error) {
	params := make(map[string]*schema.ParameterInfo)

	for _, p := range a.def.Params {
		params[p.Name] = &schema.ParameterInfo{
			Desc:     p.Description,
			Type:     schema.String,
			Required: p.Required,
		}
	}

	return &schema.ToolInfo{
		Name:        a.def.Name,
		Desc:        a.def.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

func (a *toolAdapter) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...tool.Option) (string, error) {
	params := make(map[string]string)
	if argumentsInJSON != "" {
		raw := make(map[string]any)
		if err := json.Unmarshal([]byte(argumentsInJSON), &raw); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w", err)
		}
		for k, v := range raw {
			params[k] = fmt.Sprintf("%v", v)
		}
	}
	return a.def.Handler(ctx, params), nil
}

var _ tool.InvokableTool = (*toolAdapter)(nil)

func ToEinoTools(defs []ToolDef) []tool.BaseTool {
	tools := make([]tool.BaseTool, len(defs))
	for i, d := range defs {
		tools[i] = &toolAdapter{def: d}
	}
	return tools
}

func ToToolInfos(defs []ToolDef) ([]*schema.ToolInfo, error) {
	infos := make([]*schema.ToolInfo, len(defs))
	for i, d := range defs {
		params := make(map[string]*schema.ParameterInfo)
		for _, p := range d.Params {
			params[p.Name] = &schema.ParameterInfo{
				Desc:     p.Description,
				Type:     schema.String,
				Required: p.Required,
			}
		}
		infos[i] = &schema.ToolInfo{
			Name:        d.Name,
			Desc:        d.Description,
			ParamsOneOf: schema.NewParamsOneOfByParams(params),
		}
	}
	return infos, nil
}

func HasToolCalls(msg *schema.Message) bool {
	return len(msg.ToolCalls) > 0
}

func ParseToolCalls(msg *schema.Message) []schema.ToolCall {
	return msg.ToolCalls
}

func ToolSystemPrompt(tools []ToolDef) string {
	b := new(strings.Builder)
	b.WriteString("You have access to the following tools:\n\n")
	b.WriteString("To call a tool, output XML like this exactly:\n")
	b.WriteString("<tool name=\"tool_name\">\n")
	b.WriteString("<param name=\"param1\">value1</param>\n")
	b.WriteString("<param name=\"param2\">value2</param>\n")
	b.WriteString("</tool>\n\n")

	for i, t := range tools {
		fmt.Fprintf(b, "%d. %s - %s\n", i+1, t.Name, t.Description)
		if len(t.Params) > 0 {
			b.WriteString("   Parameters:\n")
			for _, p := range t.Params {
				req := ""
				if p.Required {
					req = " (required)"
				}
				fmt.Fprintf(b, "   - %s%s: %s\n", p.Name, req, p.Description)
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("Always ask the user to confirm before creating anything. List the parameters clearly before creating.\n")
	b.WriteString("For list_* tools (list_modules, list_charts, list_pages), call them immediately without asking for confirmation — they are read-only and provide necessary information to the user.")

	return b.String()
}

// string-based tool call parsing for backward compat with directListTool
type ToolCall struct {
	Name   string
	Params map[string]string
}

func HasToolCallsStr(response string) bool {
	return strings.Contains(response, "<tool ")
}

func ParseToolCallsStr(response string) []ToolCall {
	var calls []ToolCall
	remaining := response

	for {
		startIdx := strings.Index(remaining, "<tool")
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(remaining[startIdx:], "</tool>")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + len("</tool>")

		toolBlock := remaining[startIdx:endIdx]
		remaining = remaining[endIdx:]

		name := extractTagAttr(toolBlock, "tool", "name")
		if name == "" {
			continue
		}

		params := parseToolParams(toolBlock)
		calls = append(calls, ToolCall{Name: name, Params: params})
	}

	return calls
}

func extractTagAttr(xml, tag, attr string) string {
	start := strings.Index(xml, "<"+tag)
	if start == -1 {
		return ""
	}

	close := strings.Index(xml[start:], ">")
	if close == -1 {
		return ""
	}
	openTag := xml[start : start+close]

	attrPrefix := attr + `="`
	attrStart := strings.Index(openTag, attrPrefix)
	if attrStart == -1 {
		return ""
	}
	attrStart += len(attrPrefix)
	attrEnd := strings.Index(openTag[attrStart:], `"`)
	if attrEnd == -1 {
		return ""
	}
	return openTag[attrStart : attrStart+attrEnd]
}

func parseToolParams(xml string) map[string]string {
	params := make(map[string]string)
	remaining := xml

	for {
		startIdx := strings.Index(remaining, "<param")
		if startIdx == -1 {
			break
		}

		endIdx := strings.Index(remaining[startIdx:], "</param>")
		if endIdx == -1 {
			break
		}
		endIdx += startIdx + len("</param>")

		paramBlock := remaining[startIdx:endIdx]
		remaining = remaining[endIdx:]

		name := extractTagAttr(paramBlock, "param", "name")
		if name == "" {
			continue
		}

		value := extractTagBody(paramBlock, "param")
		params[name] = value
	}

	return params
}

func extractTagBody(xml, tag string) string {
	start := strings.Index(xml, ">")
	if start == -1 {
		return ""
	}
	start++

	end := strings.LastIndex(xml, "</"+tag+">")
	if end == -1 || end <= start {
		return ""
	}

	return strings.TrimSpace(xml[start:end])
}
