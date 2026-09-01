package service

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	ttlcache "github.com/jellydator/ttlcache/v3"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/logger"
	"go.uber.org/zap"

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
		Calls     []aiagent.Call
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
	tools = append(tools, aiagent.DefaultCatalog().Resolve(aiagent.AssistantKitNames()...)...)

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
			if aiagent.UserCancelled(ask.Prompt) {
				c.pending.Delete(ask.Chat)
				return map[string]any{"response": "Действие отменено."}, nil
			}
			if aiagent.UserConfirmed(ask.Prompt) {
				pending := item.Value()
				c.pending.Delete(ask.Chat)
				client, err := c.getClient(ask)
				if err != nil {
					return nil, err
				}
				ns := ask.Namespace
				if ns == 0 {
					ns = pending.Namespace
				}
				opt := c.chatRuntimeOpts(ctx, ask, client, ns, pendingPrompt(pending.Calls), nil)
				opt.Continue = confirmContinue
				out := aiagent.ContinueFromTools(ctx, opt, pending.Calls)
				if out.Err != nil {
					return nil, out.Err
				}
				if strings.TrimSpace(out.Output) == "" {
					return map[string]any{"response": "Действие выполнено."}, nil
				}
				return map[string]any{"response": out.Output}, nil
			}
		}
	}

	client, err := c.getClient(ask)
	if err != nil {
		return nil, err
	}
	opt := c.chatRuntimeOpts(ctx, ask, client, ask.Namespace, ask.Prompt, nil)

	if result := showPageChartFastPath(ctx, ask); result != "" {
		out := aiagent.ContinueAfterResult(ctx, opt, result)
		if out.Err != nil {
			return nil, out.Err
		}
		if strings.TrimSpace(out.Output) == "" {
			return map[string]any{"response": result}, nil
		}
		return map[string]any{"response": out.Output}, nil
	}

	if result := salesDynamicsFastPath(ctx, ask); result != "" {
		out := aiagent.ContinueAfterResult(ctx, opt, result)
		if out.Err != nil {
			return nil, out.Err
		}
		if strings.TrimSpace(out.Output) == "" {
			return map[string]any{"response": result}, nil
		}
		return map[string]any{"response": out.Output}, nil
	}

	out := aiagent.Run(ctx, opt)
	if out.Err != nil {
		return nil, out.Err
	}
	if out.ConfirmNeeded {
		c.storePending(ask.Chat, out.ConfirmCalls, ask.Namespace)
		return map[string]any{"response": out.Output + confirmFence(out.ConfirmCalls)}, nil
	}
	return map[string]any{"response": out.Output}, nil
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
	opt := c.chatRuntimeOpts(ctx, ask, client, ask.Namespace, ask.Prompt, stream)
	chat.EmitToolsCapability(ctx, client.IsToolsSupported())

	if result := showPageChartFastPath(ctx, ask); result != "" {
		out := aiagent.ContinueAfterResult(ctx, opt, result)
		if out.Err != nil {
			return out.Err
		}
		return stream("", "", true)
	}

	if result := salesDynamicsFastPath(ctx, ask); result != "" {
		out := aiagent.ContinueAfterResult(ctx, opt, result)
		if out.Err != nil {
			return out.Err
		}
		return stream("", "", true)
	}

	out := aiagent.Run(ctx, opt)
	if out.Err != nil {
		return out.Err
	}
	if out.ConfirmNeeded {
		c.storePending(ask.Chat, out.ConfirmCalls, ask.Namespace)
		if err := stream(confirmFence(out.ConfirmCalls), "", false); err != nil {
			return err
		}
	}
	return stream("", "", true)
}

func (c *chatService) chatRuntimeOpts(ctx context.Context, ask *ChatPromptArguments, client *chat.Client, namespaceID uint64, toolPrompt string, stream chat.StreamFunc) aiagent.Options {
	extra := map[string]string{}
	if namespaceID > 0 {
		extra["namespaceID"] = fmt.Sprintf("%d", namespaceID)
	}
	useTools := client.IsToolsSupported()
	tools := []chat.ToolDef(nil)
	if useTools {
		tools = c.getTools(ctx, namespaceID, toolPrompt)
	}
	msgs := c.buildMessages(ctx, ask, useTools)
	return aiagent.Options{
		Client:       client,
		Messages:     msgs,
		Tools:        tools,
		MaxSteps:     6,
		ExtraParams:  extra,
		Confirmed:    aiagent.UserConfirmed(ask.Prompt),
		NeedsConfirm: aiagent.DefaultNeedsConfirm,
		Continue:     chatContinue,
		Stream:       stream,
		HideToolXML:  stream != nil,
		EmptyAnswer:  "Модель не сгенерировала ответ.",
	}
}

func chatContinue(_ string, toolResult string, _ []aiagent.Call) aiagent.ContinueHint {
	fence := extractChartFence(toolResult)
	if fence != "" {
		instruction := "A ```chart block from the tool was already shown to the user. Write a short comment in the user's language. Do NOT repeat the ```chart fence or dump JSON. Do not call tools. Do not invent numbers."
		if strings.Contains(strings.ToLower(fence), "compose-chart") {
			instruction = "A page chart was already shown to the user. Write a short comment in the user's language. Do NOT repeat the ```compose-chart fence or dump JSON. Do not call tools. Do not invent numbers."
		}
		return aiagent.ContinueHint{
			UserMessage:  instruction,
			DisableTools: true,
			StreamPrefix: "\n" + fence + "\n",
		}
	}
	payload := truncateRunes(toolResult, 16000)
	return aiagent.ContinueHint{
		UserMessage: "Tool results (use only this data; never invent numbers):\n\n" + payload +
			"\n\nIf you need more data, call another tool. Otherwise write a short answer in the user's language. If a comparison, trend, or share visualization helps, append a ```chart fenced JSON block (type bar|line|pie|doughnut) with real labels and series from these results. Valid JSON, no comments.",
	}
}

func confirmContinue(assistant, toolResult string, calls []aiagent.Call) aiagent.ContinueHint {
	h := chatContinue(assistant, toolResult, calls)
	h.DisableTools = true
	return h
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

func (c *chatService) chatEnvToContext(ask *ChatPromptArguments, ctx context.Context) context.Context {
	if ask.Namespace > 0 {
		ctx = context.WithValue(ctx, chat.EnvNamespaceID, ask.Namespace)
	}
	if ask.Page > 0 {
		ctx = context.WithValue(ctx, chat.EnvPageID, ask.Page)
	}
	if ask.Module > 0 {
		ctx = context.WithValue(ctx, chat.EnvModuleID, ask.Module)
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

type CallParam = aiagent.Call

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
	if aiagent.UserCancelled(ask.Prompt) {
		c.pending.Delete(ask.Chat)
		if err := stream("Действие отменено.", "", false); err != nil {
			return true, err
		}
		return true, stream("", "", true)
	}
	if !aiagent.UserConfirmed(ask.Prompt) {
		return false, nil
	}
	pending := item.Value()
	c.pending.Delete(ask.Chat)

	ns := ask.Namespace
	if ns == 0 {
		ns = pending.Namespace
	}
	client, err := c.getClient(ask)
	if err != nil {
		return true, err
	}
	opt := c.chatRuntimeOpts(ctx, ask, client, ns, pendingPrompt(pending.Calls), stream)
	opt.Continue = confirmContinue
	out := aiagent.ContinueFromTools(ctx, opt, pending.Calls)
	if out.Err != nil {
		return true, out.Err
	}
	if strings.TrimSpace(out.Output) == "" {
		_ = stream("Действие выполнено.", "", false)
	}
	return true, stream("", "", true)
}

func createModule(ctx context.Context, params map[string]string) string {
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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
	namespaceID, ok := ctx.Value(chat.EnvNamespaceID).(uint64)
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

	q := SanitizeRecordSearchQuery(query)
	if q == "" {
		return "Search query is empty after sanitization."
	}

	module, err := DefaultModule.FindByID(ctx, namespaceID, moduleID)
	if err != nil || module == nil {
		return fmt.Sprintf("Failed to find module %d: %v", moduleID, err)
	}

	ql := BuildRecordTextSearchQL(module.Fields, q)
	if ql == "" {
		return fmt.Sprintf("Module '%s' has no text fields to search.", module.Name)
	}

	filter := types.RecordFilter{
		ModuleID:    moduleID,
		NamespaceID: namespaceID,
		Query:       ql,
	}
	filter.Limit = defaultRecordSearchLimit

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
		appendChatRecordPreview(&b, r)
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
	filter.Limit = defaultRecordSearchLimit
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
		appendChatRecordPreview(&b, r)
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func appendChatRecordPreview(b *strings.Builder, r *types.Record) {
	fmt.Fprintf(b, "• **Record #%d** (ID: %d)\n", r.Revision, r.ID)
	const maxFields = 4
	n := 0
	prefer := []string{"name", "title", "label", "handle", "email"}
	used := map[string]bool{}
	valOf := map[string]*types.RecordValue{}
	for _, v := range r.Values {
		if v != nil && v.Value != "" {
			valOf[strings.ToLower(v.Name)] = v
		}
	}
	write := func(v *types.RecordValue) {
		if v == nil || used[v.Name] || n >= maxFields {
			return
		}
		used[v.Name] = true
		val := v.Value
		if len(val) > 100 {
			val = val[:97] + "..."
		}
		fmt.Fprintf(b, "  - `%s`: %s\n", v.Name, val)
		n++
	}
	for _, name := range prefer {
		write(valOf[name])
	}
	if n < maxFields {
		for _, v := range r.Values {
			write(v)
			if n >= maxFields {
				break
			}
		}
	}
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
