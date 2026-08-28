package rest

type nodeTypeDef struct {
	Type         string          `json:"type"`
	Label        string          `json:"label"`
	Description  string          `json:"description"`
	ConfigFields []nodeTypeField `json:"configFields"`
}

type nodeTypeField struct {
	Key         string              `json:"key"`
	Widget      string              `json:"widget"`
	Label       string              `json:"label"`
	Help        string              `json:"help,omitempty"`
	Required    bool                `json:"required,omitempty"`
	Template    bool                `json:"template,omitempty"`
	Placeholder string              `json:"placeholder,omitempty"`
	Default     interface{}         `json:"default,omitempty"`
	Options     []string            `json:"options,omitempty"`
	Rows        int                 `json:"rows,omitempty"`
	Lang        string              `json:"lang,omitempty"`
	VisibleIf   map[string][]string `json:"visibleIf,omitempty"`
	ItemFields  []nodeTypeField     `json:"itemFields,omitempty"`
}

func nf(key, widget, label string, extra ...func(*nodeTypeField)) nodeTypeField {
	f := nodeTypeField{Key: key, Widget: widget, Label: label}
	for _, fn := range extra {
		fn(&f)
	}
	return f
}

func req(f *nodeTypeField)  { f.Required = true }
func tmpl(f *nodeTypeField) { f.Template = true }

func def(v interface{}) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.Default = v }
}
func opts(o ...string) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.Options = o }
}
func help(s string) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.Help = s }
}
func rows(n int) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.Rows = n }
}
func lang(s string) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.Lang = s }
}
func visIf(key string, vals ...string) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.VisibleIf = map[string][]string{key: vals} }
}
func items(fields ...nodeTypeField) func(*nodeTypeField) {
	return func(f *nodeTypeField) { f.ItemFields = fields }
}

func nodeTypes() []nodeTypeDef {
	out := builtinNodeTypes()
	seen := make(map[string]int, len(out))
	for i, n := range out {
		seen[n.Type] = i
	}
	merge := func(extra []nodeTypeDef) {
		for _, n := range extra {
			if i, ok := seen[n.Type]; ok {
				out[i] = n
				continue
			}
			seen[n.Type] = len(out)
			out = append(out, n)
		}
	}
	merge(agentNodeTypes())
	merge(fetchLiveAgentNodeTypes())
	return out
}

func builtinNodeTypes() []nodeTypeDef {
	return []nodeTypeDef{
		{
			Type:        "condition",
			Label:       "Condition",
			Description: "Evaluate a condition (eq, neq, gt, lt, contains, empty, notEmpty)",
			ConfigFields: []nodeTypeField{
				nf("field", "string", "Field", req, tmpl, help("Variable or field name in the execution context")),
				nf("operator", "enum", "Operator", req, opts("eq", "neq", "gt", "lt", "gte", "lte", "contains", "empty", "notEmpty")),
				nf("value", "string", "Value", tmpl, visIf("operator", "eq", "neq", "gt", "lt", "gte", "lte", "contains")),
			},
		},
		{
			Type:        "crud",
			Label:       "CRUD Record",
			Description: "Create, update, delete, or search records in a module",
			ConfigFields: []nodeTypeField{
				nf("operation", "enum", "Operation", req, def("create"), opts("create", "update", "delete", "search")),
				nf("moduleHandle", "string", "Module handle", tmpl, help("Preferred over module ID")),
				nf("moduleID", "string", "Module ID"),
				nf("namespaceID", "string", "Namespace ID"),
				nf("recordID", "string", "Record ID", tmpl, visIf("operation", "update", "delete")),
				nf("fields", "keymap", "Fields", visIf("operation", "create", "update"), help("Field name → value; supports {{templates}}")),
				nf("query", "string", "Query", tmpl, visIf("operation", "search")),
				nf("limit", "number", "Limit", visIf("operation", "search")),
				nf("omitEmpty", "bool", "Omit empty fields", visIf("operation", "update")),
				nf("continueOnError", "bool", "Continue on error", visIf("operation", "update")),
			},
		},
		{
			Type:        "crud.upsert",
			Label:       "Upsert Record",
			Description: "Find a record by matchBy fields and update it, or create it",
			ConfigFields: []nodeTypeField{
				nf("moduleHandle", "string", "Module handle", tmpl, help("Preferred over module ID")),
				nf("moduleID", "string", "Module ID"),
				nf("namespaceID", "string", "Namespace ID"),
				nf("matchBy", "stringlist", "Match by", help("Ordered fields used to find an existing record")),
				nf("matchAll", "bool", "Require all match fields"),
				nf("fields", "keymap", "Fields", help("Field templates, {{item.ip}} inside foreach")),
				nf("omitEmpty", "bool", "Omit empty fields"),
				nf("continueOnError", "bool", "Continue on error"),
				nf("resultVar", "string", "Result variable", help("Context key for the resulting record ID")),
			},
		},
		{
			Type:        "foreach",
			Label:       "For Each Item",
			Description: "Loop body nodes once per item in an array (items / devices)",
			ConfigFields: []nodeTypeField{
				nf("items", "string", "Items variable", def("items"), help("Context variable holding the array")),
				nf("itemVar", "string", "Item prefix", def("item")),
				nf("maxItems", "number", "Max items"),
				nf("failFast", "bool", "Stop on first error"),
			},
		},
		{
			Type:        "detach",
			Label:       "Detach (poll)",
			Description: "Start a background poller that feeds the ingest chain; does not block Run",
			ConfigFields: []nodeTypeField{
				nf("kind", "enum", "Kind", def("poll"), opts("poll")),
				nf("ingestChainID", "string", "Ingest chain ID", req),
				nf("statusUrl", "string", "Status URL", tmpl),
				nf("itemsUrl", "string", "Items URL", tmpl),
				nf("interval", "number", "Interval (seconds)", def(2)),
				nf("timeout", "number", "Timeout (seconds)", def(900)),
				nf("until", "string", "Stop statuses", help("Comma-separated statuses that stop polling")),
			},
		},
		{
			Type:        "mail",
			Label:       "Send Email",
			Description: "Send an email notification",
			ConfigFields: []nodeTypeField{
				nf("to", "string", "To", req, tmpl),
				nf("subject", "string", "Subject", req, tmpl),
				nf("body", "textarea", "Body", req, tmpl, rows(6), help("HTML supported")),
				nf("cc", "string", "CC", tmpl),
				nf("contentType", "enum", "Content type", def("html"), opts("html", "plain")),
			},
		},
		{
			Type:        "http",
			Label:       "HTTP Request",
			Description: "Make an HTTP request to an external API",
			ConfigFields: []nodeTypeField{
				nf("url", "string", "URL", req, tmpl),
				nf("method", "enum", "Method", def("GET"), opts("GET", "POST", "PUT", "PATCH", "DELETE")),
				nf("headers", "keymap", "Headers"),
				nf("body", "textarea", "Body", tmpl, rows(6)),
				nf("timeout", "number", "Timeout (seconds)", def(30)),
			},
		},
		{
			Type:        "ai",
			Label:       "AI Agent",
			Description: "Call an AI agent (crud-agent, assistant) with a prompt",
			ConfigFields: []nodeTypeField{
				nf("agent", "enum", "Agent", req, opts("crud-agent", "assistant")),
				nf("prompt", "textarea", "Prompt", req, tmpl, rows(6), help("Supports {{variable}} templates")),
				nf("model", "string", "Model", help("Default: qwen3:8b / CHAT_MODEL")),
				nf("maxTokens", "number", "Max tokens"),
			},
		},
		{
			Type:        "script",
			Label:       "JavaScript",
			Description: "Execute JavaScript code with lowcode runtime API",
			ConfigFields: []nodeTypeField{
				nf("code", "code", "Code", req, lang("javascript"), rows(12), help("runtime.mcp, runtime.mail, runtime.http, runtime.log")),
			},
		},
		{
			Type:        "gonec",
			Label:       "Go Code",
			Description: "Compile and execute Go code in a sandbox",
			ConfigFields: []nodeTypeField{
				nf("code", "code", "Code", req, lang("golang"), rows(12)),
				nf("timeout", "number", "Timeout (seconds)"),
			},
		},
		{
			Type:        "workflow",
			Label:       "Trigger Workflow",
			Description: "Trigger a Corteza workflow",
			ConfigFields: []nodeTypeField{
				nf("workflowID", "string", "Workflow ID", req),
				nf("payload", "textarea", "Payload", rows(4), help("JSON payload passed to the workflow")),
			},
		},
		{
			Type:        "fork",
			Label:       "Fork",
			Description: "Split execution into multiple parallel branches",
			ConfigFields: []nodeTypeField{
				nf("branches", "number", "Branches", def(2), help("Minimum 2")),
			},
		},
		{
			Type:        "score.matrix",
			Label:       "Risk Matrix",
			Description: "5×5 (or NxN) likelihood × impact → score",
			ConfigFields: []nodeTypeField{
				nf("likelihoodField", "string", "Likelihood field", def("likelihood")),
				nf("impactField", "string", "Impact field", def("impact")),
				nf("likelihood", "string", "Likelihood value", tmpl, help("Literal or {{template}}; overrides the field when set")),
				nf("impact", "string", "Impact value", tmpl),
				nf("scale", "number", "Scale", def(5), help("Clamp 1..scale")),
				nf("formula", "enum", "Formula", def("product"), opts("product", "sum"), help("Ignored if a custom matrix is set")),
				nf("matrix", "json", "Custom matrix", help("Optional NxN number array [L-1][I-1]")),
				nf("outScore", "string", "Score variable", def("score")),
				nf("outX", "string", "Likelihood output"),
				nf("outY", "string", "Impact output"),
			},
		},
		{
			Type:        "score.weighted",
			Label:       "Weighted Score",
			Description: "Σ weightᵢ · normalize(fieldᵢ / maxᵢ) → score 0..100",
			ConfigFields: []nodeTypeField{
				nf("factors", "objectlist", "Factors", req, items(
					nf("field", "string", "Field", req),
					nf("weight", "number", "Weight"),
					nf("max", "number", "Max"),
					nf("invert", "bool", "Invert"),
				)),
				nf("normalize", "bool", "Normalize", def(true), help("Scale to 0..scaleMax")),
				nf("scaleMax", "number", "Scale max", def(100)),
				nf("outScore", "string", "Score variable", def("score")),
			},
		},
		{
			Type:        "risk.band",
			Label:       "Risk Band",
			Description: "Map score → level; optional residual = score × (1 − control)",
			ConfigFields: []nodeTypeField{
				nf("scoreField", "string", "Score field", def("score")),
				nf("controlField", "string", "Control field", help("0..1 control effectiveness")),
				nf("bands", "objectlist", "Bands", items(
					nf("name", "string", "Name", req),
					nf("max", "number", "Max", help("Inclusive upper bound")),
				)),
				nf("criticalLevels", "stringlist", "Critical levels", help("Levels that set is_critical")),
				nf("outLevel", "string", "Level variable", def("level")),
				nf("outResidual", "string", "Residual variable", def("residualScore")),
				nf("outCriticalFlag", "string", "Critical flag variable", def("is_critical")),
			},
		},
	}
}
