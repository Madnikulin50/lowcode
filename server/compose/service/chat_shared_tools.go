package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

// nsID resolves the namespace for a tool call: chat's aiagent runtime
// injects the active namespace into ctx (see chatEnvToContext in chat.go),
// so that takes priority; a "namespaceID" arg is the fallback for callers
// that pass it explicitly (MCP always does — see ModuleToolDefs et al.,
// which read args["namespaceID"] directly instead of using this helper).
func nsID(ctx context.Context, params map[string]string) uint64 {
	if v := ctx.Value(chat.EnvNamespaceID); v != nil {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	if params == nil {
		return 0
	}
	return parseUint64(params["namespaceID"])
}

// chatRecordSearchToolDef and chatMailToolDefs are registered into the
// aiagent catalog by RegisterComposeToolKits (toolkits.go) — unlike
// modules/charts/pages, record search and mail aren't CRUD on a compose
// entity, so they stay outside the shared Def registry.
func chatRecordSearchToolDef() chat.ToolDef {
	return chat.ToolDef{
		Name:        "search_records",
		Description: "Search records in a module by text query (max 200 results)",
		Params: []chat.ParamDef{
			{Name: "query", Type: "string", Required: true, Description: "Search query"},
			{Name: "moduleID", Type: "string", Required: true, Description: "Module ID to search in"},
			{Name: "limit", Type: "string", Required: false, Description: "Max results"},
		},
		Handler: chatSearchRecords,
	}
}

func chatSearchRecords(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	modID := parseUint64(params["moduleID"])
	if modID == 0 {
		return "moduleID is required"
	}
	mod, err := DefaultModule.FindByID(ctx, ns, modID)
	if err != nil || mod == nil {
		return fmt.Sprintf("module not found: %v", err)
	}
	q := SanitizeRecordSearchQuery(params["query"])
	if q == "" {
		return "Search query is empty after sanitization."
	}
	ql := BuildRecordTextSearchQL(mod.Fields, q)
	if ql == "" {
		return "[]"
	}
	var limit uint
	if params["limit"] != "" {
		fmt.Sscanf(params["limit"], "%d", &limit)
	}
	ff := types.RecordFilter{NamespaceID: ns, ModuleID: modID, Query: ql}
	ff.Limit = ClampRecordSearchLimit(limit)
	set, _, err := DefaultRecord.Find(ctx, ff)
	if err != nil {
		return fmt.Sprintf("Search error: %v", err)
	}
	rows := make([]map[string]interface{}, 0, len(set))
	for _, r := range set {
		row := map[string]interface{}{"recordID": fmt.Sprintf("%d", r.ID)}
		for _, v := range r.Values {
			row[v.Name] = v.Value
		}
		rows = append(rows, row)
	}
	return toJSON(rows)
}

func chatMailToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "send_mail",
			Description: "Send an email notification",
			Params: []chat.ParamDef{
				{Name: "to", Type: "string", Required: true, Description: "Recipient email addresses (comma-separated)"},
				{Name: "subject", Type: "string", Required: true, Description: "Email subject"},
				{Name: "body", Type: "string", Required: true, Description: "Email body (plain text or HTML)"},
				{Name: "cc", Type: "string", Required: false, Description: "CC email addresses"},
				{Name: "contentType", Type: "string", Required: false, Description: "html or plain (default: html)"},
			},
			Handler: func(ctx context.Context, params map[string]string) string {
				to := splitTrimService(params["to"])
				if len(to) == 0 {
					return "Error: 'to' is required"
				}
				cc := splitTrimService(params["cc"])
				contentType := params["contentType"]
				if contentType == "" {
					contentType = "html"
				}
				n := &types.EmailNotification{
					To:      to,
					Cc:      cc,
					Subject: params["subject"],
				}
				if contentType == "html" {
					n.ContentHTML = params["body"]
				} else {
					n.ContentPlain = params["body"]
				}
				if err := DefaultNotification.SendEmail(ctx, n); err != nil {
					return fmt.Sprintf("Failed to send email: %v", err)
				}
				return fmt.Sprintf("Email sent to %v with subject '%s'", to, params["subject"])
			},
		},
	}
}

func splitTrimService(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
