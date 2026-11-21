package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/jsruntime"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var JSRuntime *jsruntime.Runtime

func SetJSRuntime(rt *jsruntime.Runtime) {
	JSRuntime = rt
}

func initAIScripts(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("run_ai_script",
		mcp.WithDescription("Execute a JavaScript script with access to lowcode runtime (mcp, mail, http, skill, log, data). The script can create/update/search records, send emails, call HTTP APIs. For AI-generated scripts and custom skills."),
		mcp.WithString("script", mcp.Description("JavaScript code to execute. Has access to: runtime.mcp (CRUD), runtime.mail (send), runtime.http (get/post), runtime.skill (register/invoke), runtime.log (info/warn/error), runtime.data (transform), runtime.context (input)"), mcp.Required()),
		mcp.WithString("input", mcp.Description("JSON object with input data, available as runtime.context and as direct variables in the script")),
	), handleRunAIScript)

	s.AddTool(mcp.NewTool("generate_skill",
		mcp.WithDescription("Register a reusable AI skill (JavaScript function)"),
		mcp.WithString("name", mcp.Description("Skill name"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Skill description")),
		mcp.WithString("script", mcp.Description("JavaScript code that defines the skill function. Use runtime.skill.register('name', async function(input) { ... })"), mcp.Required()),
	), handleGenerateSkill)

	s.AddTool(mcp.NewTool("list_ai_skills",
		mcp.WithDescription("List registered AI skills"),
	), handleListAISkills)
}

func handleRunAIScript(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	script := getString(args, "script")
	inputJSON := getString(args, "input")

	if script == "" {
		return textResult("Error: 'script' is required"), nil
	}

	var input map[string]interface{}
	if inputJSON != "" {
		if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
			return textResult(fmt.Sprintf("Invalid input JSON: %v", err)), nil
		}
	}

	if JSRuntime == nil {
		JSRuntime = newDefaultJSRuntime()
	}

	result := JSRuntime.Run(ctx, script, input)
	return jsonResult(result), nil
}

func handleGenerateSkill(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	name := getString(args, "name")
	description := getString(args, "description")
	script := getString(args, "script")

	if name == "" || script == "" {
		return textResult("Error: 'name' and 'script' are required"), nil
	}

	if JSRuntime == nil {
		JSRuntime = newDefaultJSRuntime()
	}

	result := JSRuntime.Run(ctx, script, nil)
	return jsonResult(map[string]interface{}{
		"skill":       name,
		"description": description,
		"registered":  result.Error == "",
		"runResult":   result,
	}), nil
}

func handleListAISkills(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	return jsonResult(map[string]interface{}{
		"skills":  []string{},
		"message": "skill listing available after registration",
	}), nil
}

func newDefaultJSRuntime() *jsruntime.Runtime {
	return jsruntime.New(jsruntime.Services{
		MCPCall: func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"error": "MCP call bridge not configured"}, nil
		},
		RecordCreate: func(ctx context.Context, nsID, modID uint64, values map[string]interface{}) (string, string, error) {
			r := &types.Record{
				NamespaceID: nsID,
				ModuleID:    modID,
				Values:      make([]*types.RecordValue, 0),
			}
			for name, val := range values {
				r.Values = append(r.Values, &types.RecordValue{
					Name:  name,
					Value: fmt.Sprintf(`%v`, val),
				})
			}
			created, _, err := service.DefaultRecord.Create(ctx, r)
			if err != nil {
				return "", "", err
			}
			return fmt.Sprintf("%d", created.ID), created.CreatedAt.String(), nil
		},
		RecordUpdate: func(ctx context.Context, nsID, modID uint64, recordID string, values map[string]interface{}) (string, error) {
			var rid uint64
			fmt.Sscanf(recordID, "%d", &rid)
			existing, _, err := service.DefaultRecord.FindByID(ctx, nsID, modID, rid)
			if err != nil {
				return "", err
			}
			for name, val := range values {
				found := false
				for _, rv := range existing.Values {
					if rv.Name == name {
						rv.Value = fmt.Sprintf(`%v`, val)
						found = true
						break
					}
				}
				if !found {
					existing.Values = append(existing.Values, &types.RecordValue{
						Name:  name,
						Value: fmt.Sprintf(`%v`, val),
					})
				}
			}
			updated, _, err := service.DefaultRecord.Update(ctx, existing)
			if err != nil {
				return "", err
			}
			return updated.UpdatedAt.String(), nil
		},
		RecordDelete: func(ctx context.Context, nsID, modID uint64, recordID string) error {
			var rid uint64
			fmt.Sscanf(recordID, "%d", &rid)
			return service.DefaultRecord.DeleteByID(ctx, nsID, modID, rid)
		},
		RecordSearch: func(ctx context.Context, nsID, modID uint64, query string, limit int) ([]map[string]interface{}, error) {
			ff := types.RecordFilter{
				NamespaceID: nsID,
				ModuleID:    modID,
				Query:       query,
			}
			if limit > 0 {
				ff.Limit = uint(limit)
			}
			set, _, err := service.DefaultRecord.Find(ctx, ff)
			if err != nil {
				return nil, err
			}
			result := make([]map[string]interface{}, 0, len(set))
			for _, r := range set {
				record := map[string]interface{}{"recordID": fmt.Sprintf("%d", r.ID)}
				for _, v := range r.Values {
					record[v.Name] = strings.Trim(v.Value, `"`)
				}
				result = append(result, record)
			}
			return result, nil
		},
		MailSend: func(ctx context.Context, to []string, subject, body string, cc []string, contentType string) error {
			n := &types.EmailNotification{
				To:      to,
				Cc:      cc,
				Subject: subject,
			}
			if contentType == "html" {
				n.ContentHTML = body
			} else {
				n.ContentPlain = body
			}
			return service.DefaultNotification.SendEmail(ctx, n)
		},
	})
}
