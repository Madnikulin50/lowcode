package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	ttlcache "github.com/jellydator/ttlcache/v3"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type (
	ChatMessage struct {
		Role    string
		Content string
	}

	ChatFile struct {
		Name    string
		Content string
	}

	ChatStreamFunc = chat.StreamFunc

	chatService struct {
		clients *ttlcache.Cache[string, *chat.Client]
		tools   []chat.ToolDef
	}

	ChatPromptArguments struct {
		Chat      string
		Prompt    string
		Facts     []string
		Files     []ChatFile
		Messages  []ChatMessage
		Namespace uint64
		Module    uint64
		Page      uint64
		Record    uint64
	}
)

func Chat() *chatService {
	ttl := ttlcache.New[string, *chat.Client](
		ttlcache.WithTTL[string, *chat.Client](30 * time.Minute),
	)

	go ttl.Start()

	tools := []chat.ToolDef{
		{Name: "create_module", Description: "Create a new module (entity to store records)", Params: []chat.ParamDef{{Name: "name", Type: "string", Required: true, Description: "Module display name"}, {Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"}, {Name: "fields", Type: "json", Required: true, Description: `JSON array. Each field: {"name":"...","kind":"String|Number|DateTime|Select|Bool|User|Record|File|URL|Email","label":"...","required":true/false}`}}, Handler: createModule},
		{Name: "create_chart", Description: "Create a new chart", Params: []chat.ParamDef{{Name: "name", Type: "string", Required: true, Description: "Chart display name"}, {Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"}, {Name: "config", Type: "json", Required: true, Description: "JSON chart config with reports, dimensions, metrics"}}, Handler: createChart},
		{Name: "create_page", Description: "Create a new page", Params: []chat.ParamDef{{Name: "title", Type: "string", Required: true, Description: "Page title"}, {Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"}, {Name: "description", Type: "string", Required: false, Description: "Page description"}, {Name: "moduleID", Type: "string", Required: false, Description: "Module ID this page is for"}, {Name: "blocks", Type: "json", Required: false, Description: "JSON array of page blocks"}}, Handler: createPage},
		{Name: "list_modules", Description: "List all modules (entities/collections/tables) in the current namespace with their fields. Use this to show stores, products, tasks, or any data entities.", Handler: listModules},
		{Name: "list_charts", Description: "List all charts in the current namespace", Handler: listCharts},
		{Name: "list_pages", Description: "List all pages in the current namespace", Handler: listPages},
	}

	return &chatService{
		clients: ttl,
		tools:   tools,
	}
}

func (c *chatService) modelFromPrompt(prompt string) string {
	if strings.Contains(prompt, "<model>") {
		start := strings.Index(prompt, "<model>") + len("<model>")
		end := strings.Index(prompt[start:], "</model>")
		if end > 0 {
			return prompt[start : start+end]
		}
	}
	return "deepseek-r1" // "deepseek-v2"
}

func (c *chatService) buildMessages(ask *ChatPromptArguments) []*schema.Message {
	msgs := make([]*schema.Message, 0, len(ask.Messages)+2)

	hasSystem := false
	for _, m := range ask.Messages {
		role := schema.User
		if m.Role == "assistant" || m.Role == "system" {
			role = schema.RoleType(m.Role)
		}
		if m.Role == "system" {
			hasSystem = true
		}
		msgs = append(msgs, &schema.Message{Role: role, Content: m.Content})
	}
	if !hasSystem {
		msgs = append([]*schema.Message{
			schema.SystemMessage("You are an assistant for a database app. You MUST call a tool to get or manage data.\n\nCall a tool with XML:\n<tool name=\"tool_name\">\n<param name=\"param1\">value1</param>\n</tool>\n\nIMPORTANT: When the user asks to show/list/view specific items (e.g. stores, products, tasks), look through all available tools for one whose description mentions that item name and call it directly. Do NOT call list_modules for this — it lists entity types, not records.\n\nRules:\n- show/list stores/products/tasks/etc → find module_{id}_records tool matching the name\n- show/list what entities exist → list_modules\n- show/list pages → list_pages\n- show/list charts → list_charts\n- create page/module/chart → create_* tool\n\nAsk before creating. For listing, call tool immediately."),
		}, msgs...)
	}
	if ask.Prompt != "" {
		prompt := ask.Prompt
		if len(ask.Files) > 0 {
			var fileBlock string
			for _, f := range ask.Files {
				fileBlock += fmt.Sprintf("\n\nAttached data file '%s':\n```csv\n%s\n```", f.Name, f.Content)
			}
			prompt = fileBlock + "\n\n" + prompt
		}
		msgs = append(msgs, schema.UserMessage(prompt))
	}
	return msgs
}

func (c *chatService) getTools(ctx context.Context, namespaceID uint64) []chat.ToolDef {
	tools := make([]chat.ToolDef, len(c.tools))
	copy(tools, c.tools)

	if namespaceID == 0 {
		return tools
	}

	modules, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: namespaceID})
	if err == nil {
		for _, m := range modules {

			id := m.Handle
			if len(id) == 0 {
				id = fmt.Sprintf("%v", m.ID)
			}
			id = strings.ToLower(id)
			name := m.Name
			name = strings.ToLower(name)
			name = strings.Replace(name, "_", " ", -1)
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("show_module_%v", id),
				Description: fmt.Sprintf("Show properties and fields of module '%s'", name),
				Handler: func(ctx context.Context, params map[string]string) string {
					return showModuleByID(ctx, namespaceID, m.ID)
				},
			})
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("module_search_%v", id),
				Description: fmt.Sprintf("Search records in module '%s' by text query", name),
				Params: []chat.ParamDef{
					{Name: "query", Type: "string", Required: true, Description: "Search text to find in record fields"},
				},
				Handler: func(ctx context.Context, params map[string]string) string {
					return moduleSearch(ctx, namespaceID, m.ID, params["query"])
				},
			})
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("module_%v_records", id),
				Description: fmt.Sprintf("View all records in module '%s'", name),
				Handler: func(ctx context.Context, params map[string]string) string {
					return moduleRecords(ctx, namespaceID, m.ID)
				},
			})
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("module_%v_create_record", id),
				Description: fmt.Sprintf("Create a new record in module '%s'. Pass field values as JSON: {\"fieldName\":\"value\"}", name),
				Params: []chat.ParamDef{
					{Name: "values", Type: "json", Required: true, Description: `JSON object with field values, e.g. {"title":"New Item","price":"100"}`},
				},
				Handler: func(ctx context.Context, params map[string]string) string {
					return moduleCreateRecord(ctx, namespaceID, m.ID, params["values"])
				},
			})
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("module_%v_update_record", id),
				Description: fmt.Sprintf("Update a record in module '%s' by record ID. Pass recordID and field values.", name),
				Params: []chat.ParamDef{
					{Name: "recordID", Type: "string", Required: true, Description: "Record ID to update"},
					{Name: "values", Type: "json", Required: true, Description: `JSON object with field values to update, e.g. {"title":"New Title","price":"200"}`},
				},
				Handler: func(ctx context.Context, params map[string]string) string {
					return moduleUpdateRecord(ctx, namespaceID, m.ID, parseUint64(params["recordID"]), params["values"])
				},
			})
		}
	}

	charts, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: namespaceID})
	if err == nil {
		for _, ch := range charts {
			id := ch.ID
			name := ch.Name
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("show_chart_%d", id),
				Description: fmt.Sprintf("Show details of chart '%s'", name),
				Handler: func(ctx context.Context, params map[string]string) string {
					return showChartByID(ctx, namespaceID, id)
				},
			})
		}
	}

	pages, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: namespaceID})
	if err == nil {
		for _, p := range pages {
			id := p.ID
			title := p.Title
			tools = append(tools, chat.ToolDef{
				Name:        fmt.Sprintf("show_page_%d", id),
				Description: fmt.Sprintf("Show details of page '%s'", title),
				Handler: func(ctx context.Context, params map[string]string) string {
					return showPageByID(ctx, namespaceID, id)
				},
			})
		}
	}

	return tools
}

func (c *chatService) Ask(ctx context.Context, ask *ChatPromptArguments) (interface{}, error) {
	client, err := c.getClient(ask)
	if err != nil {
		return nil, err
	}

	ctx = c.chatEnvToContext(ask, ctx)
	allTools := c.getTools(ctx, ask.Namespace)

	if result := directListTool(c.tools, ctx, ask); result != "" {
		return map[string]any{"response": result}, nil
	}

	toolInfos, err := chat.ToToolInfos(allTools)
	if err != nil {
		return nil, fmt.Errorf("failed to build tool infos: %w", err)
	}

	msgs := c.buildMessages(ask)
	out, err := client.Generate(ctx, msgs, model.WithTools(toolInfos))
	if err != nil {
		return nil, err
	}

	content := out.Content

	// try native tool calls first, fall back to XML-based
	toolCalls := out.ToolCalls
	if len(toolCalls) == 0 && chat.HasToolCallsStr(content) {
		xmlCalls := chat.ParseToolCallsStr(content)
		toolCalls = make([]schema.ToolCall, len(xmlCalls))
		for i, xc := range xmlCalls {
			paramJSON, _ := json.Marshal(xc.Params)
			toolCalls[i] = schema.ToolCall{
				Function: schema.FunctionCall{
					Name:      xc.Name,
					Arguments: string(paramJSON),
				},
			}
		}
	}

	if len(toolCalls) > 0 {
		parsed := make([]CallParam, 0, len(toolCalls))
		for _, tc := range toolCalls {
			parsed = append(parsed, CallParam{
				Name:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
		if needsConfirm(parsed) && !userConfirmed(ask.Prompt) {
			return map[string]any{
				"response": content + "\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.",
			}, nil
		}
		result := execToolCalls(ctx, parsed, ask.Namespace, allTools)
		if result != "" {
			return map[string]any{
				"response": result,
			}, nil
		}
	}

	return map[string]any{
		"response": content,
	}, nil
}

func (c *chatService) AskStream(ctx context.Context, ask *ChatPromptArguments, stream chat.StreamFunc) error {
	ctx = c.chatEnvToContext(ask, ctx)
	allTools := c.getTools(ctx, ask.Namespace)

	if result := directListTool(c.tools, ctx, ask); result != "" {
		stream(result, "", false)
		return stream("", "", true)
	}

	client, err := c.getClient(ask)
	if err != nil {
		return err
	}

	toolInfos, err := chat.ToToolInfos(allTools)
	if err != nil {
		return fmt.Errorf("failed to build tool infos: %w", err)
	}

	msgs := c.buildMessages(ask)
	streamReader, err := client.Stream(ctx, msgs, model.WithTools(toolInfos))
	if err != nil {
		return err
	}
	defer streamReader.Close()

	var fullContent string
	var toolCalls []schema.ToolCall
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			stream("⚠ Error: "+err.Error(), "", false)
			return err
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			if err := stream(chunk.Content, "", false); err != nil {
				return err
			}
		}
		if chunk.ReasoningContent != "" {
			if err := stream("", chunk.ReasoningContent, false); err != nil {
				return err
			}
		}
		if chunk.ToolCalls != nil && len(chunk.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}
	}

	// if no native tool calls, check for XML-based tool calls in full content
	if len(toolCalls) == 0 && chat.HasToolCallsStr(fullContent) {
		xmlCalls := chat.ParseToolCallsStr(fullContent)
		toolCalls = make([]schema.ToolCall, len(xmlCalls))
		for i, xc := range xmlCalls {
			paramJSON, _ := json.Marshal(xc.Params)
			toolCalls[i] = schema.ToolCall{
				Function: schema.FunctionCall{
					Name:      xc.Name,
					Arguments: string(paramJSON),
				},
			}
		}
	}

	if len(toolCalls) > 0 {
		parsed := make([]CallParam, 0, len(toolCalls))
		for _, tc := range toolCalls {
			parsed = append(parsed, CallParam{
				Name:   tc.Function.Name,
				Params: tc.Function.Arguments,
			})
		}
		if needsConfirm(parsed) && !userConfirmed(ask.Prompt) {
			stream("\n\n⚠️ **Обнаружено предлагаемое действие.** Напишите **«да»** чтобы подтвердить.", "", false)
			return stream("", "", true)
		}
		result := execToolCalls(ctx, parsed, ask.Namespace, allTools)
		if result != "" {
			stream(result, "", false)
		}
		return stream("", "", true)
	}

	return stream("", "", true)
}

func directListTool(staticTools []chat.ToolDef, ctx context.Context, ask *ChatPromptArguments) string {
	return ""
	prompt := strings.ToLower(ask.Prompt)
	var toolNames []string

	for _, t := range staticTools {
		if !strings.HasPrefix(t.Name, "list_") {
			continue
		}
		keywords := keywordsForTool(t.Name)
		for _, kw := range keywords {
			if strings.Contains(prompt, kw) {
				toolNames = append(toolNames, t.Name)
				break
			}
		}
	}

	if len(toolNames) == 0 {
		return ""
	}

	var results []string
	for _, name := range toolNames {
		for _, t := range staticTools {
			if t.Name == name {
				results = append(results, t.Handler(ctx, nil))
			}
		}
	}

	return strings.Join(results, "\n\n")
}

func keywordsForTool(name string) []string {
	switch name {
	case "list_modules":
		return []string{"модул", "module", "сущност"}
	case "list_charts":
		return []string{"чарт", "chart", "график", "граф"}
	case "list_pages":
		return []string{"страниц", "страни", "page", "pages"}
	}
	return nil
}

type CallParam struct {
	Name   string
	Params string
}

func needsConfirm(calls []CallParam) bool {
	for _, call := range calls {
		if strings.HasPrefix(call.Name, "create_") || strings.HasPrefix(call.Name, "delete_") {
			return true
		}
	}
	return false
}

func execToolCalls(ctx context.Context, calls []CallParam, namespaceID uint64, tools []chat.ToolDef) string {
	var results []string
	for _, call := range calls {
		params := make(map[string]string)
		if call.Params != "" {
			raw := make(map[string]any)
			if err := json.Unmarshal([]byte(call.Params), &raw); err == nil {
				for k, v := range raw {
					params[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		if _, ok := params["namespaceID"]; !ok && namespaceID > 0 {
			params["namespaceID"] = fmt.Sprintf("%d", namespaceID)
		}
		for _, t := range tools {
			if t.Name == call.Name {
				results = append(results, t.Handler(ctx, params))
				break
			}
		}
	}
	if len(results) > 0 {
		return strings.Join(results, "\n")
	}
	return ""
}

func (c *chatService) chatEnvToContext(ask *ChatPromptArguments, ctx context.Context) context.Context {
	if ask.Namespace > 0 {
		ctx = context.WithValue(ctx, "namespaceID", ask.Namespace)
	}
	if ask.Page > 0 {
		ctx = context.WithValue(ctx, "pageID", ask.Namespace)
	}
	return ctx
}

func (c *chatService) getClient(ask *ChatPromptArguments) (*chat.Client, error) {
	client := c.clients.Get(ask.Chat)
	if client == nil {
		model := c.modelFromPrompt(ask.Prompt)
		cl, err := chat.NewClient(model)
		if err != nil {
			return nil, err
		}
		c.clients.Set(ask.Chat, cl, 30*time.Minute)
		return cl, nil
	}
	return client.Value(), nil
}

func userConfirmed(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch lower {
	case "да", "yes", "y", "ok", "ок", "гоу", "do it", "создай", "подтверждаю", "confirm", "выполнить":
		return true
	}
	return false
}

func createModule(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}

	name := params["name"]
	handle := params["handle"]
	fieldsJSON := params["fields"]

	if namespaceID == 0 || name == "" || fieldsJSON == "" {
		return fmt.Sprintf("Missing required parameters for create_module (namespaceID='%s', name='%s')", params["namespaceID"], name)
	}

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return fmt.Sprintf("Invalid fields JSON for module '%s': %v", name, err)
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

	module := &types.Module{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Fields:      fieldSet,
		Config:      types.ModuleConfig{},
	}

	created, err := DefaultModule.Create(ctx, module)
	if err != nil {
		return fmt.Sprintf("Failed to create module '%s': %v", name, err)
	}

	return fmt.Sprintf("✅ Module '%s' created! ID: %d, Handle: %s, Fields: %d", created.Name, created.ID, created.Handle, len(created.Fields))
}

func createChart(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}
	name := params["name"]
	handle := params["handle"]
	configJSON := params["config"]

	if namespaceID == 0 || name == "" || configJSON == "" {
		return fmt.Sprintf("Missing required parameters for create_chart (namespaceID='%s', name='%s')", params["namespaceID"], name)
	}

	var config types.ChartConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Sprintf("Invalid config JSON for chart '%s': %v", name, err)
	}

	chart := &types.Chart{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Config:      config,
	}

	created, err := DefaultChart.Create(ctx, chart)
	if err != nil {
		return fmt.Sprintf("Failed to create chart '%s': %v", name, err)
	}

	return fmt.Sprintf("✅ Chart '%s' created! ID: %d, Handle: %s", created.Name, created.ID, created.Handle)
}

func createPage(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}
	title := params["title"]
	handle := params["handle"]
	description := params["description"]
	moduleID := parseUint64(params["moduleID"])

	if namespaceID == 0 || title == "" {
		return fmt.Sprintf("Missing required parameters for create_page (namespaceID='%s', title='%s')", params["namespaceID"], title)
	}

	var blocks types.PageBlocks
	if b := params["blocks"]; b != "" {
		if err := json.Unmarshal([]byte(b), &blocks); err != nil {
			return fmt.Sprintf("Invalid blocks JSON for page '%s': %v", title, err)
		}
	}

	page := &types.Page{
		NamespaceID: namespaceID,
		Title:       title,
		Handle:      handle,
		Description: description,
		ModuleID:    moduleID,
		Visible:     true,
		Blocks:      blocks,
	}

	created, err := DefaultPage.Create(ctx, page)
	if err != nil {
		return fmt.Sprintf("Failed to create page '%s': %v", title, err)
	}

	return fmt.Sprintf("✅ Page '%s' created! ID: %d, Handle: %s", created.Title, created.ID, created.Handle)
}

func parseUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	var v uint64
	fmt.Sscanf(s, "%d", &v)
	return v
}

func getFieldStr(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFieldBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func listModules(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}
	if namespaceID == 0 {
		return "Missing required parameter: namespaceID"
	}

	set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: namespaceID})
	if err != nil {
		return fmt.Sprintf("Failed to list modules: %v", err)
	}

	if len(set) == 0 {
		return "No modules found in this namespace."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📦 **Modules (%d):**\n\n", len(set))
	for _, m := range set {
		fields := make([]string, 0, len(m.Fields))
		for _, f := range m.Fields {
			fields = append(fields, f.Name+" ("+f.Kind+")")
		}
		fs := strings.Join(fields, "\r\n")
		fmt.Fprintf(&b, "• **%s** (ID: %d, Handle: %s)\r\n", m.Name, m.ID, m.Handle)
		if fs != "" {
			fmt.Fprintf(&b, "\r\n *Fields*: %s\r\n\r\n", fs)
		}
	}
	return b.String()
}

func listCharts(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}
	if namespaceID == 0 {
		return "Missing required parameter: namespaceID"
	}

	set, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: namespaceID})
	if err != nil {
		return fmt.Sprintf("Failed to list charts: %v", err)
	}

	if len(set) == 0 {
		return "No charts found in this namespace."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 **Charts (%d):**\n\n", len(set))
	for _, c := range set {
		reports := len(c.Config.Reports)
		fmt.Fprintf(&b, "• **%s** (ID: %d, Handle: %s, Reports: %d)\n", c.Name, c.ID, c.Handle, reports)
	}
	return b.String()
}

func listPages(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value("namespaceID").(uint64)
	if !ok {
		namespaceID = parseUint64(params["namespaceID"])
	}
	if namespaceID == 0 {
		return "Missing required parameter: namespaceID"
	}

	set, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: namespaceID})
	if err != nil {
		return fmt.Sprintf("Failed to list pages: %v", err)
	}

	if len(set) == 0 {
		return "No pages found in this namespace."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📄 **Pages (%d):**\n\n", len(set))
	for _, p := range set {
		fmt.Fprintf(&b, "### %s\n", p.Title)
		fmt.Fprintf(&b, "- **ID:** %d\n", p.ID)
		fmt.Fprintf(&b, "- **Handle:** `%s`\n", p.Handle)
		if p.Description != "" {
			fmt.Fprintf(&b, "- **Description:** %s\n", p.Description)
		}
		if p.ModuleID > 0 {
			fmt.Fprintf(&b, "- **ModuleID:** %d\n", p.ModuleID)
		}
		if p.Visible {
			fmt.Fprintf(&b, "- **Visible:** ✅\n")
		}
		fmt.Fprintf(&b, "- **Blocks (%d):**\n", len(p.Blocks))
		for _, blk := range p.Blocks {
			title := blk.Title
			if title == "" {
				title = blk.Kind
			}
			fmt.Fprintf(&b, "  - `%s`", title)
			if blk.Kind != "" && blk.Kind != title {
				fmt.Fprintf(&b, " (_%s_)", blk.Kind)
			}
			if blk.Description != "" {
				fmt.Fprintf(&b, ": %s", blk.Description)
			}
			fmt.Fprintf(&b, "\n")
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func showModuleByID(ctx context.Context, namespaceID, moduleID uint64) string {
	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}
	if module == nil {
		return fmt.Sprintf("Module %d not found.", moduleID)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📦 **Module: %s** (ID: %d)\n", module.Name, module.ID)
	fmt.Fprintf(&b, "- **Handle:** `%s`\n", module.Handle)
	fmt.Fprintf(&b, "- **Fields (%d):**\n", len(module.Fields))
	for _, f := range module.Fields {
		fmt.Fprintf(&b, "  - `%s` (_%s_)", f.Name, f.Kind)
		if f.Label != "" {
			fmt.Fprintf(&b, " — %s", f.Label)
		}
		if f.Required {
			fmt.Fprintf(&b, " *required*")
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func showChartByID(ctx context.Context, namespaceID, chartID uint64) string {
	chart, err := DefaultChart.FindByID(ctx, namespaceID, chartID)
	if err != nil {
		return fmt.Sprintf("Failed to find chart %d: %v", chartID, err)
	}
	if chart == nil {
		return fmt.Sprintf("Chart %d not found.", chartID)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📊 **Chart: %s** (ID: %d)\n", chart.Name, chart.ID)
	fmt.Fprintf(&b, "- **Handle:** `%s`\n", chart.Handle)
	fmt.Fprintf(&b, "- **Reports (%d):**\n", len(chart.Config.Reports))
	for _, r := range chart.Config.Reports {
		fmt.Fprintf(&b, "  - reportID: %d (moduleID: %d)\n", r.ReportID, r.ModuleID)
	}
	return b.String()
}

func showPageByID(ctx context.Context, namespaceID, pageID uint64) string {
	page, err := DefaultPage.FindByID(ctx, namespaceID, pageID)
	if err != nil {
		return fmt.Sprintf("Failed to find page %d: %v", pageID, err)
	}
	if page == nil {
		return fmt.Sprintf("Page %d not found.", pageID)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📄 **Page: %s** (ID: %d)\n", page.Title, page.ID)
	fmt.Fprintf(&b, "- **Handle:** `%s`\n", page.Handle)
	if page.Description != "" {
		fmt.Fprintf(&b, "- **Description:** %s\n", page.Description)
	}
	if page.ModuleID > 0 {
		fmt.Fprintf(&b, "- **ModuleID:** %d\n", page.ModuleID)
	}
	fmt.Fprintf(&b, "- **Blocks (%d):**\n", len(page.Blocks))
	for _, blk := range page.Blocks {
		title := blk.Title
		if title == "" {
			title = blk.Kind
		}
		fmt.Fprintf(&b, "  - `%s`", title)
		if blk.Kind != "" && blk.Kind != title {
			fmt.Fprintf(&b, " (_%s_)", blk.Kind)
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func moduleSearch(ctx context.Context, namespaceID, moduleID uint64, query string) string {
	if query == "" {
		return "Missing required parameter: query"
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	textKinds := map[string]bool{"String": true, "Text": true, "URL": true, "Email": true}
	var conditions []string
	for _, f := range module.Fields {
		if textKinds[f.Kind] {
			escaped := strings.ReplaceAll(query, "'", "\\'")
			conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", f.Name, escaped))
		}
	}

	if len(conditions) == 0 {
		return fmt.Sprintf("Module '%s' has no text fields to search.", module.Name)
	}

	filter := types.RecordFilter{
		ModuleID:    moduleID,
		NamespaceID: namespaceID,
		Query:       strings.Join(conditions, " OR "),
	}

	set, _, err := DefaultRecord.Find(ctx, filter)
	if err != nil {
		return fmt.Sprintf("Search failed: %v", err)
	}

	if len(set) == 0 {
		return fmt.Sprintf("No records found in module '%s' matching '%s'.", module.Name, query)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "🔍 **Search results in '%s' (%d):**\n\n", module.Name, len(set))
	for _, r := range set {
		fmt.Fprintf(&b, "• **Record #%d** (ID: %d)\n", r.Revision, r.ID)
		for _, v := range r.Values {
			if v.Value != "" {
				fmt.Fprintf(&b, "  - `%s`: %s\n", v.Name, v.Value)
			}
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func moduleRecords(ctx context.Context, namespaceID, moduleID uint64) string {
	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	filter := types.RecordFilter{
		ModuleID:    moduleID,
		NamespaceID: namespaceID,
	}
	set, _, err := DefaultRecord.Find(ctx, filter)
	if err != nil {
		return fmt.Sprintf("Failed to list records: %v", err)
	}

	if len(set) == 0 {
		return fmt.Sprintf("No records in module '%s'.", module.Name)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "📋 **Records in '%s' (%d):**\n\n", module.Name, len(set))
	for _, r := range set {
		fmt.Fprintf(&b, "• **Record #%d** (ID: %d)\n", r.Revision, r.ID)
		for _, v := range r.Values {
			if v.Value != "" {
				fmt.Fprintf(&b, "  - `%s`: %s\n", v.Name, v.Value)
			}
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func moduleCreateRecord(ctx context.Context, namespaceID, moduleID uint64, valuesJSON string) string {
	if valuesJSON == "" {
		return "Please provide field values as JSON, e.g. {\"title\":\"New Item\"}."
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(valuesJSON), &raw); err != nil {
		return fmt.Sprintf("Invalid values JSON: %v. Please use format like {\"fieldName\":\"value\"}.", err)
	}

	values := make(types.RecordValueSet, 0, len(raw))
	for k, v := range raw {
		values = append(values, &types.RecordValue{
			Name:  k,
			Value: fmt.Sprintf("%v", v),
		})
	}

	rec := &types.Record{
		NamespaceID: namespaceID,
		ModuleID:    moduleID,
		Values:      values,
	}

	created, errs, err := DefaultRecord.Create(ctx, rec)
	if err != nil {
		return fmt.Sprintf("Failed to create record in '%s': %v", module.Name, err)
	}
	if errs != nil && len(errs.Set) > 0 {
		msg := fmt.Sprintf("Validation errors for '%s':", module.Name)
		for _, e := range errs.Set {
			msg += fmt.Sprintf("\n- %s: %s", e.Kind, e.Message)
		}
		return msg
	}

	id := created.ID
	var vals []string
	for _, v := range created.Values {
		if v.Value != "" {
			vals = append(vals, fmt.Sprintf("%s=%s", v.Name, v.Value))
		}
	}

	return fmt.Sprintf("Record created in '%s'! ID: %d, Values: %s", module.Name, id, strings.Join(vals, ", "))
}

func moduleUpdateRecord(ctx context.Context, namespaceID, moduleID, recordID uint64, valuesJSON string) string {
	if recordID == 0 {
		return "Please provide a record ID to update."
	}
	if valuesJSON == "" {
		return "Please provide field values to update as JSON."
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	existing, _, err := DefaultRecord.FindByID(ctx, namespaceID, moduleID, recordID)
	if err != nil {
		return fmt.Sprintf("Record %d not found: %v", recordID, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(valuesJSON), &raw); err != nil {
		return fmt.Sprintf("Invalid values JSON: %v. Please use format like {\"fieldName\":\"newValue\"}.", err)
	}

	values := make(types.RecordValueSet, 0, len(raw))
	for k, v := range raw {
		values = append(values, &types.RecordValue{
			Name:  k,
			Value: fmt.Sprintf("%v", v),
		})
	}

	existing.Values = values
	updated, errs, err := DefaultRecord.Update(ctx, existing)
	if err != nil {
		return fmt.Sprintf("Failed to update record %d: %v", recordID, err)
	}
	if errs != nil && len(errs.Set) > 0 {
		msg := fmt.Sprintf("Validation errors for record %d:", recordID)
		for _, e := range errs.Set {
			msg += fmt.Sprintf("\n- %s: %s", e.Kind, e.Message)
		}
		return msg
	}

	var vals []string
	for _, v := range updated.Values {
		if v.Value != "" {
			vals = append(vals, fmt.Sprintf("%s=%s", v.Name, v.Value))
		}
	}

	return fmt.Sprintf("Record %d updated in '%s'! Values: %s", recordID, module.Name, strings.Join(vals, ", "))
}
