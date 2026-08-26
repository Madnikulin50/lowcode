package service

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	ttlcache "github.com/jellydator/ttlcache/v3"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"go.uber.org/zap"

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

	pendingToolCalls struct {
		Calls     []CallParam
		Namespace uint64
	}

	chatService struct {
		clients *ttlcache.Cache[string, *chat.Client]
		pending *ttlcache.Cache[string, pendingToolCalls]
		catalog *ttlcache.Cache[uint64, nsCatalog]
		tools   []chat.ToolDef
	}

	nsCatalog struct {
		modules types.ModuleSet
		charts  types.ChartSet
		pages   types.PageSet
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
		Model     string
	}
)

const chatSystemPrompt = "Ты ассистент приложения с базой данных. Для данных всегда вызывай инструмент.\n\n" +
	"Вызов инструмента XML:\n<tool name=\"tool_name\">\n<param name=\"param1\">value1</param>\n</tool>\n\n" +
	"Графики и динамика:\n" +
	"- Если пользователь называет график с ТЕКУЩЕЙ страницы по заголовку блока (список Charts on this page) — вызови show_page_chart. НЕ вызывай sales_dynamics и НЕ visualize_report для этих заголовков.\n" +
	"- «покажи динамику продаж», выручка по месяцам, sales trend → sales_dynamics (линейный график), если это не заголовок блока на странице. НЕ вызывай create_chart и НЕ читай сырые чеки.\n" +
	"- новый график / диаграмма / динамика / сравни / chart / trend по данным модуля → visualize_report (агрегация, не больше 24 категорий).\n" +
	"- Никогда не вызывай module_*_records для больших таблиц вроде receipt_positions.\n" +
	"show_page_chart вернёт блок ```compose-chart — включи его в ответ как есть (UI покажет график страницы).\n" +
	"sales_dynamics / visualize_report вернут блок ```chart — включи его как есть.\n" +
	"```chart\n" +
	"{\"type\":\"line\",\"title\":\"...\",\"labels\":[...],\"series\":[{\"name\":\"...\",\"data\":[...]}]}\n" +
	"```\n" +
	"Типы: bar, line, pie, doughnut. Для динамики во времени — line. Только реальные числа из инструментов, не выдумывай.\n\n" +
	"Другие правила:\n" +
	"- показать магазины/товары/задачи → module_{handle}_records по имени сущности; не list_modules (это типы, не записи)\n" +
	"- какие сущности есть → list_modules; страницы → list_pages; сохранённые чарты → list_charts\n" +
	"- создать страницу/модуль/чарт → create_* (спроси подтверждение)\n" +
	"Для списков вызывай инструмент сразу."

const chatSystemPromptNoTools = "Ты ассистент приложения с базой данных. Отвечай по существу и на языке пользователя."

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
	tools = append(tools, chatVisualizeTools()...)

	pending := ttlcache.New[string, pendingToolCalls](
		ttlcache.WithTTL[string, pendingToolCalls](30 * time.Minute),
	)
	go pending.Start()

	catalog := ttlcache.New[uint64, nsCatalog](
		ttlcache.WithTTL[uint64, nsCatalog](30 * time.Second),
	)
	go catalog.Start()

	return &chatService{
		clients: ttl,
		pending: pending,
		catalog: catalog,
		tools:   tools,
	}
}

// Preload the default model into memory at startup so the first chat
// message is not delayed by model loading.
func init() {
	go func() {
		if err := chat.WarmUp(chat.DefaultModelName()); err != nil {
			logger.Default().Warn("failed to preload chat model", zap.Error(err))
		}
	}()
}

func (c *chatService) xmlToMap(xmlStr string) (map[string]string, error) {
	result := make(map[string]string)
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	var currentTag string

	for {
		token, err := decoder.Token()
		if err != nil {
			// io.EOF means we reached the end of the string
			break
		}

		switch element := token.(type) {
		case xml.StartElement:
			currentTag = element.Name.Local
			// Optional: Capture attributes if needed
			for _, attr := range element.Attr {
				result[currentTag+"_"+attr.Name.Local] = attr.Value
			}
		case xml.CharData:
			value := strings.TrimSpace(string(element))
			if value != "" {
				result[currentTag] = value
			}
		case xml.EndElement:
			currentTag = ""
		}
	}

	return result, nil
}

func (c *chatService) splitPrompt(prompt string) []*schema.Message {
	data, err := c.xmlToMap(prompt)
	if err != nil {
		return []*schema.Message{&schema.Message{Role: "user", Content: prompt}}
	}
	out := []*schema.Message{}
	for k, v := range data {
		if k == "" {
			continue
		}
		out = append(out, &schema.Message{Role: "system", Content: "##" + k + "\r\n" + v})
	}
	noTag, ok := data[""]
	if ok {
		out = append(out, &schema.Message{Role: "user", Content: noTag})
	}
	return out
}
func (c *chatService) modelFromPrompt(prompt string) string {
	if strings.Contains(prompt, "<model>") {
		start := strings.Index(prompt, "<model>") + len("<model>")
		end := strings.Index(prompt[start:], "</model>")
		if end > 0 {
			return prompt[start : start+end]
		}
	}
	return chat.DefaultModelName()
}

func (c *chatService) buildMessages(ctx context.Context, ask *ChatPromptArguments, useTools bool) []*schema.Message {
	if len(ask.Messages) > 10 {
		ask.Messages = ask.Messages[len(ask.Messages)-10:]
	}
	msgs := make([]*schema.Message, 0, len(ask.Messages)+2)

	hasSystem := false
	for _, m := range ask.Messages {
		role := schema.User
		if m.Role == "assistant" || m.Role == "system" {
			role = schema.RoleType(m.Role)
		}

		if m.Role == "user" {
			m := c.splitPrompt(m.Content)
			msgs = append(msgs, m...)
		} else {
			msgs = append(msgs, &schema.Message{Role: role, Content: m.Content})
		}

	}
	for _, m := range msgs {
		if m.Role == "system" {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		sys := chatSystemPromptNoTools
		if useTools {
			sys = chatSystemPrompt
			if extra := pageChartsSystemHint(ctx, ask); extra != "" {
				sys += "\n\n" + extra
			}
		}
		msgs = append([]*schema.Message{
			schema.SystemMessage(sys),
		}, msgs...)
	} else if useTools {
		if extra := pageChartsSystemHint(ctx, ask); extra != "" {
			msgs = append([]*schema.Message{
				schema.SystemMessage(extra),
			}, msgs...)
		}
	}

	// Current turn is sent as ask.Prompt (history is ask.Messages without it).
	// Dropping this made the model see only the system prompt and return empty
	// content → "Модель не сгенерировала ответ."
	prompt := strings.TrimSpace(ask.Prompt)
	if prompt != "" {
		if len(ask.Files) > 0 {
			var fileBlock string
			for _, f := range ask.Files {
				fileBlock += fmt.Sprintf("\n\nAttached data file '%s':\n```csv\n%s\n```", f.Name, f.Content)
			}
			prompt = fileBlock + "\n\n" + prompt
		}

		if ask.Namespace > 0 && DefaultRAG != nil {
			if ctx == nil {
				ctx = context.Background()
			}
			var allCtx string
			ragCtx := DefaultRAG.BuildContext(ctx, fmt.Sprint(ask.Namespace), prompt, 3)
			if ragCtx != "" {
				allCtx += "\nDocuments:\n" + ragCtx
			}
			if DefaultPagesRAG != nil {
				pagesCtx := DefaultPagesRAG.BuildContext(ctx, prompt, 3)
				if pagesCtx != "" {
					allCtx += "\nPublished pages:\n" + pagesCtx
				}
			}
			if allCtx != "" {
				if len(allCtx) > 4096 {
					allCtx = allCtx[:4096]
				}
				prompt = "Relevant information:\n" + allCtx + "\n\n" + prompt
			}
		}

		msgs = append(msgs, schema.UserMessage(prompt))
	}
	return msgs
}

func (c *chatService) loadCatalog(ctx context.Context, namespaceID uint64) nsCatalog {
	if c.catalog != nil {
		if item := c.catalog.Get(namespaceID); item != nil {
			return item.Value()
		}
	}
	cat := nsCatalog{}
	if modules, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: namespaceID}); err == nil {
		cat.modules = modules
	}
	if charts, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: namespaceID}); err == nil {
		cat.charts = charts
	}
	if pages, _, err := DefaultPage.Find(ctx, types.PageFilter{NamespaceID: namespaceID}); err == nil {
		cat.pages = pages
	}
	if c.catalog != nil {
		c.catalog.Set(namespaceID, cat, 30*time.Second)
	}
	return cat
}

func (c *chatService) getTools(ctx context.Context, namespaceID uint64, prompt string) []chat.ToolDef {
	tools := make([]chat.ToolDef, len(c.tools))
	copy(tools, c.tools)

	if namespaceID == 0 {
		return tools
	}

	keywords := promptKeywords(prompt)
	cat := c.loadCatalog(ctx, namespaceID)

	for _, m := range selectNamed(cat.modules, keywords, 15, 10, func(m *types.Module) string { return m.Name + " " + m.Handle }) {

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
		tools = append(tools, chat.ToolDef{
			Name:        fmt.Sprintf("module_%v_delete_record", id),
			Description: fmt.Sprintf("Delete a record by ID from module '%s'", name),
			Params: []chat.ParamDef{
				{Name: "recordID", Type: "string", Required: true, Description: "Record ID to delete"},
			},
			Handler: func(ctx context.Context, params map[string]string) string {
				return moduleDeleteRecord(ctx, namespaceID, m.ID, parseUint64(params["recordID"]))
			},
		})
	}

	for _, ch := range selectNamed(cat.charts, keywords, 15, 5, func(ch *types.Chart) string { return ch.Name }) {
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

	for _, p := range selectNamed(cat.pages, keywords, 15, 5, func(p *types.Page) string { return p.Title }) {
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

	return tools
}

// promptKeywords extracts search keywords from the user's message.
// Words shorter than 3 characters are ignored to avoid false matches.
func promptKeywords(prompt string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 8)
	for _, w := range strings.FieldsFunc(strings.ToLower(prompt), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if len(w) < 3 || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
		if len(out) >= 20 {
			break
		}
	}
	return expandChatKeywords(out, seen)
}

func expandChatKeywords(out []string, seen map[string]bool) []string {
	add := func(w string) {
		if w == "" || seen[w] {
			return
		}
		seen[w] = true
		out = append(out, w)
	}
	joined := strings.Join(out, " ")
	if strings.Contains(joined, "продаж") || strings.Contains(joined, "выруч") || strings.Contains(joined, "sales") || strings.Contains(joined, "revenue") {
		add("sales")
		add("receipt")
		add("margin")
		add("revenue")
	}
	return out
}

// selectNamed returns the entities whose name matches any prompt keyword.
// When nothing matches, the first maxFallback entities are returned so the
// model still has tools to work with. Keeping the tool list small makes
// prompt evaluation dramatically faster.
func selectNamed[T any](set []T, keywords []string, maxMatched, maxFallback int, name func(T) string) []T {
	if len(keywords) == 0 {
		if len(set) > maxFallback {
			set = set[:maxFallback]
		}
		return set
	}

	var matched []T
	for _, item := range set {
		haystack := strings.ToLower(name(item))
		for _, kw := range keywords {
			if strings.Contains(haystack, kw) {
				matched = append(matched, item)
				break
			}
		}
		if len(matched) >= maxMatched {
			break
		}
	}

	if len(matched) == 0 {
		if len(set) > maxFallback {
			set = set[:maxFallback]
		}
		return set
	}
	return matched
}

func (c *chatService) Ask(ctx context.Context, ask *ChatPromptArguments) (interface{}, error) {
	ctx = c.chatEnvToContext(ask, ctx)

	if ask.Chat != "" && c.pending != nil {
		if item := c.pending.Get(ask.Chat); item != nil {
			if userCancelled(ask.Prompt) {
				c.pending.Delete(ask.Chat)
				return map[string]any{"response": "Действие отменено."}, nil
			}
			if userConfirmed(ask.Prompt) {
				pending := item.Value()
				c.pending.Delete(ask.Chat)
				ns := ask.Namespace
				if ns == 0 {
					ns = pending.Namespace
				}
				execTools := c.getTools(ctx, ns, pendingPrompt(pending.Calls))
				result := execToolCalls(ctx, pending.Calls, ns, execTools)
				client, err := c.getClient(ask)
				if err != nil {
					return map[string]any{"response": result}, err
				}
				if err := chat.EnsureWarm(ctx, client.Model()); err != nil {
					return nil, err
				}
				msgs := c.buildMessages(ctx, ask, client.IsToolsSupported())
				if result != "" {
					cont, err := c.generateToolContinuation(ctx, client, msgs, "", result)
					if err == nil && strings.TrimSpace(cont) != "" {
						return map[string]any{"response": mergeChartAndComment(result, cont)}, nil
					}
					return map[string]any{"response": result}, nil
				}
				return map[string]any{"response": "Действие выполнено."}, nil
			}
		}
	}

	client, err := c.getClient(ask)
	if err != nil {
		return nil, err
	}
	// Warm before Generate so the 3m HTTP timeout covers inference only.
	if err := chat.EnsureWarm(ctx, client.Model()); err != nil {
		return nil, err
	}

	useTools := client.IsToolsSupported()
	var allTools []chat.ToolDef
	if useTools {
		allTools = c.getTools(ctx, ask.Namespace, ask.Prompt)
	}
	msgs := c.buildMessages(ctx, ask, useTools)

	if result := showPageChartFastPath(ctx, ask); result != "" {
		cont, err := c.generateToolContinuation(ctx, client, msgs, "", result)
		if err == nil && strings.TrimSpace(cont) != "" {
			return map[string]any{"response": mergeChartAndComment(result, cont)}, nil
		}
		return map[string]any{"response": result}, nil
	}

	if result := salesDynamicsFastPath(ctx, ask); result != "" {
		cont, err := c.generateToolContinuation(ctx, client, msgs, "", result)
		if err == nil && strings.TrimSpace(cont) != "" {
			return map[string]any{"response": mergeChartAndComment(result, cont)}, nil
		}
		return map[string]any{"response": result}, nil
	}

	if result := directListTool(c.tools, ctx, ask); result != "" {
		return map[string]any{"response": result}, nil
	}

	var opts []model.Option
	if useTools {
		toolInfos, err := chat.ToToolInfos(allTools)
		if err != nil {
			return nil, fmt.Errorf("failed to build tool infos: %w", err)
		}
		opts = append(opts, model.WithTools(toolInfos))
	}
	out, err := client.Generate(ctx, msgs, opts...)
	if err != nil {
		return nil, err
	}

	content := out.Content
	if content == "" && out.ReasoningContent != "" {
		content = out.ReasoningContent
	}
	if content == "" {
		content = "Модель не сгенерировала ответ."
	}

	if !useTools {
		return map[string]any{"response": content}, nil
	}

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
			c.storePending(ask.Chat, parsed, ask.Namespace)
			return map[string]any{
				"response": content + confirmFence(parsed),
			}, nil
		}
		result := execToolCalls(ctx, parsed, ask.Namespace, allTools)
		if result != "" {
			cont, err := c.generateToolContinuation(ctx, client, msgs, content, result)
			if err == nil && strings.TrimSpace(cont) != "" {
				return map[string]any{"response": mergeChartAndComment(result, cont)}, nil
			}
			return map[string]any{"response": result}, nil
		}
	}

	return map[string]any{
		"response": content,
	}, nil
}

func (c *chatService) AskStream(ctx context.Context, ask *ChatPromptArguments, stream chat.StreamFunc) error {
	ctx = c.chatEnvToContext(ask, ctx)

	if handled, err := c.handlePendingStream(ctx, ask, stream); handled {
		return err
	}

	client, err := c.getClient(ask)
	if err != nil {
		return err
	}
	// Warm before Stream so the 3m HTTP timeout covers inference only,
	// not cold model load (which can take minutes on CPU).
	if err := chat.EnsureWarm(ctx, client.Model()); err != nil {
		return err
	}

	useTools := client.IsToolsSupported()
	chat.EmitToolsCapability(ctx, useTools)
	var allTools []chat.ToolDef
	if useTools {
		allTools = c.getTools(ctx, ask.Namespace, ask.Prompt)
	}
	msgs := c.buildMessages(ctx, ask, useTools)

	if result := showPageChartFastPath(ctx, ask); result != "" {
		if err := c.streamToolContinuation(ctx, client, msgs, "", result, stream); err != nil {
			return err
		}
		return stream("", "", true)
	}

	if result := salesDynamicsFastPath(ctx, ask); result != "" {
		if err := c.streamToolContinuation(ctx, client, msgs, "", result, stream); err != nil {
			return err
		}
		return stream("", "", true)
	}

	if result := directListTool(c.tools, ctx, ask); result != "" {
		stream(result, "", false)
		return stream("", "", true)
	}

	var opts []model.Option
	if useTools {
		toolInfos, err := chat.ToToolInfos(allTools)
		if err != nil {
			return fmt.Errorf("failed to build tool infos: %w", err)
		}
		opts = append(opts, model.WithTools(toolInfos))
	}
	streamReader, err := client.Stream(ctx, msgs, opts...)
	if err != nil {
		return err
	}

	fullContent, fullReasoning, toolCalls, err := pumpChatStream(ctx, streamReader, stream)
	if err != nil {
		return err
	}

	effective := strings.TrimSpace(fullContent)
	if effective == "" {
		effective = strings.TrimSpace(fullReasoning)
	}

	if !useTools {
		if err := streamFallbackAnswer(stream, fullContent, fullReasoning); err != nil {
			return err
		}
		return stream("", "", true)
	}

	// if no native tool calls, check for XML-based tool calls in full content
	if len(toolCalls) == 0 && chat.HasToolCallsStr(effective) {
		xmlCalls := chat.ParseToolCallsStr(effective)
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
		if strings.TrimSpace(fullContent) == "" {
			fullContent = effective
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
			c.storePending(ask.Chat, parsed, ask.Namespace)
			stream(confirmFence(parsed), "", false)
			return stream("", "", true)
		}
		result := execToolCalls(ctx, parsed, ask.Namespace, allTools)
		if result != "" {
			if err := c.streamToolContinuation(ctx, client, msgs, fullContent, result, stream); err != nil {
				return err
			}
		}
		return stream("", "", true)
	}

	if err := streamFallbackAnswer(stream, fullContent, fullReasoning); err != nil {
		return err
	}
	return stream("", "", true)
}

func streamFallbackAnswer(stream chat.StreamFunc, content, reasoning string) error {
	if strings.TrimSpace(content) != "" {
		return nil
	}
	if strings.TrimSpace(reasoning) != "" {
		return stream(reasoning, "", false)
	}
	return stream("Модель не сгенерировала ответ.", "", false)
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
	if len(calls) > 0 {
		chat.EmitStatus(ctx, chat.StatusUsingTools)
	}
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

func pumpChatStream(ctx context.Context, streamReader *schema.StreamReader[*schema.Message], stream chat.StreamFunc) (fullContent, fullReasoning string, toolCalls []schema.ToolCall, err error) {
	defer streamReader.Close()
	for {
		select {
		case <-ctx.Done():
			return fullContent, fullReasoning, toolCalls, ctx.Err()
		default:
		}
		chunk, recvErr := streamReader.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				return fullContent, fullReasoning, toolCalls, nil
			}
			if chat.IsTimeout(recvErr) {
				note := "Превышено время ожидания ответа модели."
				if strings.TrimSpace(fullContent) != "" || strings.TrimSpace(fullReasoning) != "" {
					_ = stream("\n\n"+note, "", false)
				} else {
					_ = stream(note, "", false)
					fullContent = note
				}
				return fullContent, fullReasoning, toolCalls, nil
			}
			stream("⚠ Error: "+recvErr.Error(), "", false)
			return fullContent, fullReasoning, toolCalls, recvErr
		}
		if chunk.Content != "" {
			fullContent += chunk.Content
			emit := chunk.Content
			if idx := strings.Index(fullContent, "<tool "); idx >= 0 {
				already := len(fullContent) - len(chunk.Content)
				if already >= idx {
					emit = ""
				} else {
					emit = fullContent[already:idx]
				}
			}
			if emit != "" {
				if err := stream(emit, "", false); err != nil {
					return fullContent, fullReasoning, toolCalls, err
				}
			}
		}
		if chunk.ReasoningContent != "" {
			fullReasoning += chunk.ReasoningContent
			if err := stream("", chunk.ReasoningContent, false); err != nil {
				return fullContent, fullReasoning, toolCalls, err
			}
		}
		if len(chunk.ToolCalls) > 0 {
			if len(toolCalls) == 0 {
				chat.EmitStatus(ctx, chat.StatusUsingTools)
			}
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}
	}
}

func continuationMessages(msgs []*schema.Message, assistantContent, toolResult string) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs)+2)
	out = append(out, msgs...)
	if stripped := stripToolXML(assistantContent); stripped != "" {
		out = append(out, schema.AssistantMessage(stripped, nil))
	}
	fence := extractChartFence(toolResult)
	payload := truncateRunes(toolResult, 16000)
	var instruction string
	if fence != "" {
		if strings.Contains(strings.ToLower(fence), "compose-chart") {
			instruction = "A page chart was already shown to the user. Write a short comment in the user's language. Do NOT repeat the ```compose-chart fence or dump JSON. Do not call tools. Do not invent numbers."
		} else {
			instruction = "A ```chart block from the tool was already shown to the user. Write a short comment in the user's language. Do NOT repeat the ```chart fence or dump JSON. Do not call tools. Do not invent numbers."
		}
	} else {
		instruction = "Tool results (use only this data; never invent numbers):\n\n" + payload +
			"\n\nWrite a short answer in the user's language. If a comparison, trend, or share visualization helps, append a ```chart fenced JSON block (type bar|line|pie|doughnut) with real labels and series from these results. Valid JSON, no comments. Do not call tools."
	}
	out = append(out, schema.UserMessage(instruction))
	return out
}

func stripToolXML(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || chat.HasToolCallsStr(s) {
		return ""
	}
	return s
}

func salesDynamicsFastPath(ctx context.Context, ask *ChatPromptArguments) string {
	if ask == nil || !wantsSalesDynamics(ask.Prompt) {
		return ""
	}
	result := salesDynamics(ctx, nil)
	if extractChartFence(result) == "" {
		return ""
	}
	return result
}

func mergeChartAndComment(toolResult, comment string) string {
	fence := extractChartFence(toolResult)
	comment = strings.TrimSpace(comment)
	if fence == "" {
		if comment != "" {
			return comment
		}
		return toolResult
	}
	if comment == "" || extractChartFence(comment) != "" {
		return fence
	}
	return fence + "\n\n" + comment
}

func truncateRunes(s string, n int) string {
	if n <= 0 || s == "" {
		return s
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "\n…"
}

func (c *chatService) generateToolContinuation(ctx context.Context, client *chat.Client, msgs []*schema.Message, assistantContent, toolResult string) (string, error) {
	out, err := client.Generate(ctx, continuationMessages(msgs, assistantContent, toolResult))
	if err != nil || out == nil {
		return "", err
	}
	content := out.Content
	if content == "" && out.ReasoningContent != "" {
		content = out.ReasoningContent
	}
	return content, nil
}

func (c *chatService) streamToolContinuation(ctx context.Context, client *chat.Client, msgs []*schema.Message, assistantContent, toolResult string, stream chat.StreamFunc) error {
	fence := extractChartFence(toolResult)
	shownFence := false
	if fence != "" {
		// Stream the fence first so the UI can mount a chart even if the
		// continuation model is told not to repeat it (commentary-only).
		if err := stream("\n"+fence+"\n", "", false); err != nil {
			return err
		}
		shownFence = true
	}

	streamReader, err := client.Stream(ctx, continuationMessages(msgs, assistantContent, toolResult))
	if err != nil {
		if !shownFence && toolResult != "" {
			stream(toolResult, "", false)
		}
		return nil
	}

	cont, reasoning, _, err := pumpChatStream(ctx, streamReader, stream)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cont) != "" {
		return nil
	}
	if strings.TrimSpace(reasoning) != "" {
		return stream(reasoning, "", false)
	}
	if !shownFence && toolResult != "" {
		stream(toolResult, "", false)
	}
	return nil
}

func (c *chatService) chatEnvToContext(ask *ChatPromptArguments, ctx context.Context) context.Context {
	if ask.Namespace > 0 {
		ctx = context.WithValue(ctx, "namespaceID", ask.Namespace)
	}
	if ask.Page > 0 {
		ctx = context.WithValue(ctx, "pageID", ask.Page)
	}
	if ask.Module > 0 {
		ctx = context.WithValue(ctx, "moduleID", ask.Module)
	}
	return ctx
}

func (c *chatService) getClient(ask *ChatPromptArguments) (*chat.Client, error) {
	model := ask.Model
	if model == "" {
		model = c.modelFromPrompt(ask.Prompt)
	}

	key := ask.Chat + "|" + model
	client := c.clients.Get(key)
	if client == nil {
		cl, err := chat.NewClient(model)
		if err != nil {
			return nil, err
		}
		c.clients.Set(key, cl, 30*time.Minute)
		return cl, nil
	}
	return client.Value(), nil
}

func (c *chatService) Models(ctx context.Context) ([]string, error) {
	return chat.AvailableModels()
}

func (c *chatService) ModelsInfo(ctx context.Context) (map[string]any, error) {
	models, err := chat.AvailableModels()
	if err != nil {
		return nil, err
	}
	cfg := chat.CurrentConfig()
	return map[string]any{
		"models":     models,
		"default":    chat.ResolveInstalledModel(chat.DefaultModelName()),
		"tools":      chat.ModelsToolsMap(models),
		"enabled":    cfg.Enabled,
		"ollamaURL":  chat.EffectiveOllamaURL(),
		"ollamaFrom": ollamaURLSource(),
		"roles": map[string]string{
			chat.RoleComposeChat:    cfg.Roles.ComposeChat,
			chat.RoleMCPAgent:       cfg.Roles.MCPAgent,
			chat.RoleAutomationChat: cfg.Roles.AutomationChat,
			chat.RoleRulesgoAI:      cfg.Roles.RulesgoAI,
		},
	}, nil
}

func ollamaURLSource() string {
	cfg := chat.CurrentConfig()
	if strings.TrimSpace(cfg.OllamaURL) != "" {
		return "settings"
	}
	if strings.TrimSpace(os.Getenv("OLLAMA_URL")) != "" {
		return "OLLAMA_URL"
	}
	if strings.TrimSpace(os.Getenv("OLLAMA_HOST")) != "" {
		return "OLLAMA_HOST"
	}
	return "default"
}

func (c *chatService) DiscoverModels(ctx context.Context) ([]string, error) {
	return chat.DiscoverModels()
}

func (c *chatService) WarmUp(ctx context.Context, model string) error {
	if model == "" {
		model = chat.DefaultModelName()
	}
	return chat.WarmUp(model)
}

func userConfirmed(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch lower {
	case "да", "yes", "y", "ok", "ок", "гоу", "do it", "создай", "подтверждаю", "confirm", "выполнить":
		return true
	}
	return false
}

func userCancelled(prompt string) bool {
	lower := strings.ToLower(strings.TrimSpace(prompt))
	switch lower {
	case "нет", "no", "n", "отмена", "cancel", "не надо", "стоп":
		return true
	}
	return false
}

func (c *chatService) storePending(chatID string, calls []CallParam, namespaceID uint64) {
	if chatID == "" || c.pending == nil || len(calls) == 0 {
		return
	}
	c.pending.Set(chatID, pendingToolCalls{Calls: calls, Namespace: namespaceID}, ttlcache.DefaultTTL)
}

func callSummary(call CallParam) string {
	raw := make(map[string]any)
	if call.Params != "" {
		_ = json.Unmarshal([]byte(call.Params), &raw)
	}
	label := strings.ReplaceAll(call.Name, "_", " ")
	for _, key := range []string{"name", "title", "handle"} {
		if v, ok := raw[key]; ok {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" && s != "<nil>" {
				return label + ": " + s
			}
		}
	}
	return label
}

func confirmFence(calls []CallParam) string {
	type tool struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
	}
	payload := struct {
		Tools []tool `json:"tools"`
	}{}
	for _, call := range calls {
		payload.Tools = append(payload.Tools, tool{Name: call.Name, Summary: callSummary(call)})
	}
	body, err := json.Marshal(payload)
	if err != nil {
		body = []byte(`{"tools":[]}`)
	}
	return "\n\nПредлагаемое действие требует подтверждения.\n```chat-confirm\n" + string(body) + "\n```\n"
}

func pendingPrompt(calls []CallParam) string {
	var b strings.Builder
	for _, call := range calls {
		b.WriteString(call.Name)
		b.WriteByte(' ')
		b.WriteString(call.Params)
		b.WriteByte(' ')
	}
	return b.String()
}

func (c *chatService) handlePendingStream(ctx context.Context, ask *ChatPromptArguments, stream chat.StreamFunc) (bool, error) {
	if ask.Chat == "" || c.pending == nil {
		return false, nil
	}
	item := c.pending.Get(ask.Chat)
	if item == nil {
		return false, nil
	}
	if userCancelled(ask.Prompt) {
		c.pending.Delete(ask.Chat)
		if err := stream("Действие отменено.", "", false); err != nil {
			return true, err
		}
		return true, stream("", "", true)
	}
	if !userConfirmed(ask.Prompt) {
		return false, nil
	}
	pending := item.Value()
	c.pending.Delete(ask.Chat)

	ns := ask.Namespace
	if ns == 0 {
		ns = pending.Namespace
	}
	execTools := c.getTools(ctx, ns, pendingPrompt(pending.Calls))
	result := execToolCalls(ctx, pending.Calls, ns, execTools)

	client, err := c.getClient(ask)
	if err != nil {
		if result != "" {
			_ = stream(result, "", false)
		} else {
			_ = stream("Действие выполнено.", "", false)
		}
		return true, stream("", "", true)
	}
	if err := chat.EnsureWarm(ctx, client.Model()); err != nil {
		return true, err
	}
	if result != "" {
		msgs := c.buildMessages(ctx, ask, client.IsToolsSupported())
		if err := c.streamToolContinuation(ctx, client, msgs, "", result, stream); err != nil {
			return true, err
		}
	} else {
		_ = stream("Действие выполнено.", "", false)
	}
	return true, stream("", "", true)
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

var chatFieldIdent = regexp.MustCompile(`^[A-Za-z][0-9A-Za-z_-]*$`)

func chatSearchQuery(query string) string {
	var b strings.Builder
	for _, r := range query {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || r == '-' || r == '.' {
			b.WriteRune(r)
		}
	}
	s := strings.TrimSpace(strings.Join(strings.Fields(b.String()), " "))
	s = strings.ReplaceAll(s, `'`, `''`)
	return s
}

func moduleSearch(ctx context.Context, namespaceID, moduleID uint64, query string) string {
	if query == "" {
		return "Missing required parameter: query"
	}

	q := chatSearchQuery(query)
	if q == "" {
		return "Search query is empty after sanitization."
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	textKinds := map[string]bool{"String": true, "Text": true, "URL": true, "Email": true}
	var conditions []string
	for _, f := range module.Fields {
		if !textKinds[f.Kind] || !chatFieldIdent.MatchString(f.Name) {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%s%%'", f.Name, q))
	}

	if len(conditions) == 0 {
		return fmt.Sprintf("Module '%s' has no text fields to search.", module.Name)
	}

	filter := types.RecordFilter{
		ModuleID:    moduleID,
		NamespaceID: namespaceID,
		Query:       strings.Join(conditions, " OR "),
	}
	filter.Limit = 50

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
	filter.Limit = 50
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
	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	if valuesJSON == "" {
		var fieldList []string
		for _, f := range module.Fields {
			req := ""
			if f.Required {
				req = " (required)"
			}
			fieldList = append(fieldList, fmt.Sprintf("  - %s (%s)%s", f.Name, f.Kind, req))
		}
		return fmt.Sprintf("Module '%s' has these fields:\n%s\n\nPlease provide field values as JSON, e.g. {\"%s\":\"value\"}.", module.Name, strings.Join(fieldList, "\n"), module.Fields[0].Name)
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

func moduleDeleteRecord(ctx context.Context, namespaceID, moduleID, recordID uint64) string {
	if recordID == 0 {
		return "Please provide a record ID to delete."
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	if err := DefaultRecord.DeleteByID(ctx, namespaceID, moduleID, recordID); err != nil {
		return fmt.Sprintf("Failed to delete record %d: %v", recordID, err)
	}

	return fmt.Sprintf("Record %d deleted from '%s'.", recordID, module.Name)
}
