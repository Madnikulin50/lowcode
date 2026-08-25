package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func AllChatToolDefs() []chat.ToolDef {
	defs := make([]chat.ToolDef, 0, 18)
	defs = append(defs, chatModuleToolDefs()...)
	defs = append(defs, chatChartToolDefs()...)
	defs = append(defs, chatPageToolDefs()...)
	defs = append(defs, newChatTools()...)
	return defs
}

func chatModuleToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "read_module",
			Description: "Read a module by ID, including its fields and config",
			Params: []chat.ParamDef{
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
			},
			Handler: chatReadModule,
		},
		{
			Name:        "list_modules",
			Description: "List all modules in the current namespace with their fields",
			Handler:     chatListModules,
		},
		{
			Name:        "search_modules",
			Description: "Search modules by name",
			Params: []chat.ParamDef{
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: chatSearchModules,
		},
		{
			Name:        "create_module",
			Description: "Create a new module (entity to store records) with fields",
			Params: []chat.ParamDef{
				{Name: "name", Type: "string", Required: true, Description: "Module display name"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle (auto-generated if empty)"},
				{Name: "fields", Type: "json", Required: true, Description: `JSON array. Each field: {"name":"...","kind":"String|Number|DateTime|Select|Bool|User|Record|File|URL|Email","label":"...","required":true/false}`},
			},
			Handler: chatCreateModule,
		},
		{
			Name:        "update_module",
			Description: "Update a module's name, handle, or fields",
			Params: []chat.ParamDef{
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
				{Name: "name", Type: "string", Required: false, Description: "New module name"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "fields", Type: "json", Required: false, Description: "JSON array of field objects"},
			},
			Handler: chatUpdateModule,
		},
		{
			Name:        "delete_module",
			Description: "Delete a module by ID",
			Params: []chat.ParamDef{
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
			},
			Handler: chatDeleteModule,
		},
	}
}

func chatChartToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "read_chart",
			Description: "Read a chart by ID, including its config and reports",
			Params: []chat.ParamDef{
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
			},
			Handler: chatReadChart,
		},
		{
			Name:        "list_charts",
			Description: "List all charts in the current namespace",
			Handler:     chatListCharts,
		},
		{
			Name:        "search_charts",
			Description: "Search charts by name",
			Params: []chat.ParamDef{
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: chatSearchCharts,
		},
		{
			Name:        "create_chart",
			Description: "Create a new chart with reports, dimensions, and metrics",
			Params: []chat.ParamDef{
				{Name: "name", Type: "string", Required: true, Description: "Chart display name"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"},
				{Name: "config", Type: "json", Required: true, Description: `JSON chart config with reports, dimensions, metrics`},
			},
			Handler: chatCreateChart,
		},
		{
			Name:        "update_chart",
			Description: "Update a chart's name, handle, or config",
			Params: []chat.ParamDef{
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
				{Name: "name", Type: "string", Required: false, Description: "New chart name"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "config", Type: "json", Required: false, Description: "JSON chart config"},
			},
			Handler: chatUpdateChart,
		},
		{
			Name:        "delete_chart",
			Description: "Delete a chart by ID",
			Params: []chat.ParamDef{
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
			},
			Handler: chatDeleteChart,
		},
	}
}

func chatPageToolDefs() []chat.ToolDef {
	return []chat.ToolDef{
		{
			Name:        "read_page",
			Description: "Read a page by ID, including its blocks",
			Params: []chat.ParamDef{
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
			},
			Handler: chatReadPage,
		},
		{
			Name:        "list_pages",
			Description: "List all pages in the current namespace",
			Handler:     chatListPages,
		},
		{
			Name:        "search_pages",
			Description: "Search pages by title",
			Params: []chat.ParamDef{
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: chatSearchPages,
		},
		{
			Name:        "create_page",
			Description: "Create a new page",
			Params: []chat.ParamDef{
				{Name: "title", Type: "string", Required: true, Description: "Page title"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "Page description"},
				{Name: "moduleID", Type: "string", Required: false, Description: "Module ID this page is for"},
				{Name: "blocks", Type: "json", Required: false, Description: `JSON array of page blocks`},
			},
			Handler: chatCreatePage,
		},
		{
			Name:        "update_page",
			Description: "Update a page's title, handle, description, or blocks",
			Params: []chat.ParamDef{
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
				{Name: "title", Type: "string", Required: false, Description: "New page title"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "New page description"},
				{Name: "blocks", Type: "json", Required: false, Description: "JSON array of page blocks"},
			},
			Handler: chatUpdatePage,
		},
		{
			Name:        "delete_page",
			Description: "Delete a page by ID",
			Params: []chat.ParamDef{
				{Name: "pageID", Type: "string", Required: true, Description: "Page ID"},
			},
			Handler: chatDeletePage,
		},
	}
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}
	return string(b)
}

func nsID(ctx context.Context, params map[string]string) uint64 {
	if v := ctx.Value("namespaceID"); v != nil {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	if params == nil {
		return 0
	}
	return parseUint64(params["namespaceID"])
}

func chatReadModule(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	mID := parseUint64(params["moduleID"])
	if ns == 0 || mID == 0 {
		return "Missing required parameters: namespaceID or moduleID"
	}
	mod, err := DefaultModule.FindByID(ctx, ns, mID)
	if err != nil {
		return fmt.Sprintf("Failed to read module: %v", err)
	}
	return toJSON(mod)
}

func chatListModules(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing required parameter: namespaceID"
	}
	set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: ns})
	if err != nil {
		return fmt.Sprintf("Failed to list modules: %v", err)
	}
	if len(set) == 0 {
		return "No modules found in this namespace."
	}
	return toJSON(set)
}

func chatSearchModules(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	query := params["query"]
	if ns == 0 || query == "" {
		return "Missing required parameters: namespaceID or query"
	}
	set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{
		NamespaceID: ns,
		Query:       query,
	})
	if err != nil {
		return fmt.Sprintf("Failed to search modules: %v", err)
	}
	if len(set) == 0 {
		return "No modules found matching query."
	}
	return toJSON(set)
}

func chatCreateModule(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	name := params["name"]
	handle := params["handle"]
	fieldsJSON := params["fields"]
	if ns == 0 {
		return "Cannot determine namespace."
	}
	if name == "" {
		return "Please specify a name for the module (e.g. 'Tasks')."
	}
	if fieldsJSON == "" {
		return "Please specify fields for the module as a JSON array, e.g. [{\"name\":\"title\",\"kind\":\"String\",\"required\":true}]."
	}
	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return fmt.Sprintf("The fields JSON for module '%s' has an error: %v. Each field needs 'name' and 'kind'.", name, err)
	}
	fieldSet := make(types.ModuleFieldSet, 0, len(fields))
	for _, f := range fields {
		fieldSet = append(fieldSet, &types.ModuleField{
			Name:     getFieldStr(f, "name"),
			Kind:     getFieldStr(f, "kind"),
			Label:    getFieldStr(f, "label"),
			Required: getFieldBool(f, "required"),
		})
	}
	mod := &types.Module{
		NamespaceID: ns,
		Name:        name,
		Handle:      handle,
		Fields:      fieldSet,
		Config:      types.ModuleConfig{},
	}
	created, err := DefaultModule.Create(ctx, mod)
	if err != nil {
		return fmt.Sprintf("Failed to create module '%s': %v", name, err)
	}
	return fmt.Sprintf("Module '%s' created! ID: %d, Handle: %s, Fields: %d", created.Name, created.ID, created.Handle, len(created.Fields))
}

func chatUpdateModule(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	mID := parseUint64(params["moduleID"])
	if ns == 0 || mID == 0 {
		return "Missing required parameters: namespaceID or moduleID"
	}
	mod, err := DefaultModule.FindByID(ctx, ns, mID)
	if err != nil {
		return fmt.Sprintf("Module not found: %v", err)
	}
	if v := params["name"]; v != "" {
		mod.Name = v
	}
	if v := params["handle"]; v != "" {
		mod.Handle = v
	}
	if v := params["fields"]; v != "" {
		var fields []*types.ModuleField
		if err := json.Unmarshal([]byte(v), &fields); err != nil {
			return fmt.Sprintf("Invalid fields JSON: %v", err)
		}
		mod.Fields = fields
	}
	updated, err := DefaultModule.Update(ctx, mod)
	if err != nil {
		return fmt.Sprintf("Failed to update module: %v", err)
	}
	return toJSON(updated)
}

func chatDeleteModule(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	mID := parseUint64(params["moduleID"])
	if ns == 0 || mID == 0 {
		return "Missing required parameters: namespaceID or moduleID"
	}
	if err := DefaultModule.DeleteByID(ctx, ns, mID); err != nil {
		return fmt.Sprintf("Failed to delete module: %v", err)
	}
	return "Module deleted."
}

func chatReadChart(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	cID := parseUint64(params["chartID"])
	if ns == 0 || cID == 0 {
		return "Missing required parameters: namespaceID or chartID"
	}
	ch, err := DefaultChart.FindByID(ctx, ns, cID)
	if err != nil {
		return fmt.Sprintf("Failed to read chart: %v", err)
	}
	return toJSON(ch)
}

func chatListCharts(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing required parameter: namespaceID"
	}
	set, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: ns})
	if err != nil {
		return fmt.Sprintf("Failed to list charts: %v", err)
	}
	if len(set) == 0 {
		return "No charts found in this namespace."
	}
	return toJSON(set)
}

func chatSearchCharts(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	query := params["query"]
	if ns == 0 || query == "" {
		return "Missing required parameters: namespaceID or query"
	}
	set, _, err := DefaultChart.Find(ctx, types.ChartFilter{
		NamespaceID: ns,
		Query:       query,
	})
	if err != nil {
		return fmt.Sprintf("Failed to search charts: %v", err)
	}
	if len(set) == 0 {
		return "No charts found matching query."
	}
	return toJSON(set)
}

func chatCreateChart(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	name := params["name"]
	handle := params["handle"]
	configJSON := params["config"]
	if ns == 0 {
		return "Cannot determine namespace."
	}
	if name == "" {
		return "Please specify a name for the chart (e.g. 'Task Statistics')."
	}
	if configJSON == "" {
		return "Please specify a chart config as JSON with reports, dimensions, and metrics."
	}
	var config types.ChartConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Sprintf("The chart config JSON has an error: %v. Please provide valid JSON with 'reports' array.", err)
	}
	ch := &types.Chart{
		NamespaceID: ns,
		Name:        name,
		Handle:      handle,
		Config:      config,
	}
	created, err := DefaultChart.Create(ctx, ch)
	if err != nil {
		return fmt.Sprintf("Failed to create chart '%s': %v", name, err)
	}
	return fmt.Sprintf("Chart '%s' created! ID: %d, Handle: %s", created.Name, created.ID, created.Handle)
}

func chatUpdateChart(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	cID := parseUint64(params["chartID"])
	if ns == 0 || cID == 0 {
		return "Missing required parameters: namespaceID or chartID"
	}
	ch, err := DefaultChart.FindByID(ctx, ns, cID)
	if err != nil {
		return fmt.Sprintf("Chart not found: %v", err)
	}
	if v := params["name"]; v != "" {
		ch.Name = v
	}
	if v := params["handle"]; v != "" {
		ch.Handle = v
	}
	if v := params["config"]; v != "" {
		var cfg types.ChartConfig
		if err := json.Unmarshal([]byte(v), &cfg); err != nil {
			return fmt.Sprintf("Invalid config JSON: %v", err)
		}
		ch.Config = cfg
	}
	updated, err := DefaultChart.Update(ctx, ch)
	if err != nil {
		return fmt.Sprintf("Failed to update chart: %v", err)
	}
	return toJSON(updated)
}

func chatDeleteChart(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	cID := parseUint64(params["chartID"])
	if ns == 0 || cID == 0 {
		return "Missing required parameters: namespaceID or chartID"
	}
	if err := DefaultChart.DeleteByID(ctx, ns, cID); err != nil {
		return fmt.Sprintf("Failed to delete chart: %v", err)
	}
	return "Chart deleted."
}

func chatReadPage(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	pID := parseUint64(params["pageID"])
	if ns == 0 || pID == 0 {
		return "Missing required parameters: namespaceID or pageID"
	}
	p, err := DefaultPage.FindByID(ctx, ns, pID)
	if err != nil {
		return fmt.Sprintf("Failed to read page: %v", err)
	}
	return toJSON(p)
}

func chatListPages(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	if ns == 0 {
		return "Missing required parameter: namespaceID"
	}
	set, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: ns})
	if err != nil {
		return fmt.Sprintf("Failed to list pages: %v", err)
	}
	if len(set) == 0 {
		return "No pages found in this namespace."
	}
	return toJSON(set)
}

func chatSearchPages(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	query := params["query"]
	if ns == 0 || query == "" {
		return "Missing required parameters: namespaceID or query"
	}
	set, _, err := DefaultPage.Find(ctx, types.PageFilter{
		NamespaceID: ns,
		Query:       query,
	})
	if err != nil {
		return fmt.Sprintf("Failed to search pages: %v", err)
	}
	if len(set) == 0 {
		return "No pages found matching query."
	}
	return toJSON(set)
}

func chatCreatePage(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	title := params["title"]
	handle := params["handle"]
	description := params["description"]
	moduleID := parseUint64(params["moduleID"])
	if ns == 0 {
		return "Cannot determine namespace."
	}
	if title == "" {
		return "Please specify a title for the page (e.g. 'Main Page')."
	}
	var blocks types.PageBlocks
	if b := params["blocks"]; b != "" {
		if err := json.Unmarshal([]byte(b), &blocks); err != nil {
			return fmt.Sprintf("The blocks JSON for page '%s' has an error: %v. Please provide a valid JSON array.", title, err)
		}
	}
	p := &types.Page{
		NamespaceID: ns,
		Title:       title,
		Handle:      handle,
		Description: description,
		ModuleID:    moduleID,
		Visible:     true,
		Weight:      0,
		Blocks:      blocks,
	}
	created, err := DefaultPage.Create(ctx, p)
	if err != nil {
		return fmt.Sprintf("Failed to create page '%s': %v", title, err)
	}
	return fmt.Sprintf("Page '%s' created! ID: %d, Handle: %s", created.Title, created.ID, created.Handle)
}

func chatUpdatePage(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	pID := parseUint64(params["pageID"])
	if ns == 0 || pID == 0 {
		return "Missing required parameters: namespaceID or pageID"
	}
	p, err := DefaultPage.FindByID(ctx, ns, pID)
	if err != nil {
		return fmt.Sprintf("Page not found: %v", err)
	}
	if v := params["title"]; v != "" {
		p.Title = v
	}
	if v := params["handle"]; v != "" {
		p.Handle = v
	}
	if v := params["description"]; v != "" {
		p.Description = v
	}
	if v := params["blocks"]; v != "" {
		var blocks types.PageBlocks
		if err := json.Unmarshal([]byte(v), &blocks); err != nil {
			return fmt.Sprintf("Invalid blocks JSON: %v", err)
		}
		p.Blocks = blocks
	}
	updated, err := DefaultPage.Update(ctx, p)
	if err != nil {
		return fmt.Sprintf("Failed to update page: %v", err)
	}
	return toJSON(updated)
}

func chatDeletePage(ctx context.Context, params map[string]string) string {
	ns := nsID(ctx, params)
	pID := parseUint64(params["pageID"])
	if ns == 0 || pID == 0 {
		return "Missing required parameters: namespaceID or pageID"
	}
	if err := DefaultPage.DeleteByID(ctx, ns, pID, types.PageChildrenOnDeleteForce); err != nil {
		return fmt.Sprintf("Failed to delete page: %v", err)
	}
	return "Page deleted."
}

func newChatTools() []chat.ToolDef {
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
		{
			Name:        "run_script",
			Description: "Execute a JavaScript snippet with access to lowcode runtime (mcp, mail, http, log). Use for custom data processing, automation, or integration logic.",
			Params: []chat.ParamDef{
				{Name: "script", Type: "string", Required: true, Description: "JavaScript code to execute. Has access to runtime.mcp (CRUD), runtime.mail (send), runtime.http (get/post), runtime.log (info/warn/error), runtime.context (input data)."},
				{Name: "input", Type: "string", Required: false, Description: "JSON input data for the script"},
			},
			Handler: func(ctx context.Context, params map[string]string) string {
				_ = ctx
				_ = params
				return `Script execution via chat requires MCP tool 'run_ai_script'. Use the MCP interface for script execution.`
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
