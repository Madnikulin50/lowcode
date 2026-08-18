package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type CRUDService interface {
	Create(ctx context.Context, namespaceID, moduleID uint64, values map[string]interface{}) (recordID string, createdAt string, err error)
	Update(ctx context.Context, namespaceID, moduleID uint64, recordID string, values map[string]interface{}) (updatedAt string, err error)
	Delete(ctx context.Context, namespaceID, moduleID uint64, recordID string) error
	Search(ctx context.Context, namespaceID, moduleID uint64, query string, limit int) ([]map[string]interface{}, error)
}

type MailService interface {
	Send(ctx context.Context, to []string, subject, body string, cc []string, contentType string) error
}

// --- CRUD Node ---

type crudConfig struct {
	Operation    string                 `json:"operation"`
	ModuleID     flexibleID             `json:"moduleID"`
	ModuleHandle string                 `json:"moduleHandle,omitempty"`
	NamespaceID  flexibleID             `json:"namespaceID"`
	RecordID     string                 `json:"recordID,omitempty"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
	Query        string                 `json:"query,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
}

// flexibleID accepts Corteza snowflake IDs as JSON strings or numbers.
// JS Number() cannot hold them; apply.mjs and the page trigger send strings.
type flexibleID uint64

func (id *flexibleID) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "" || s == "null" || s == `""` {
		*id = 0
		return nil
	}
	if len(s) > 0 && s[0] == '"' {
		var str string
		if err := json.Unmarshal(b, &str); err != nil {
			return err
		}
		str = strings.TrimSpace(str)
		if str == "" {
			*id = 0
			return nil
		}
		n, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		*id = flexibleID(n)
		return nil
	}
	var n uint64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*id = flexibleID(n)
	return nil
}

type moduleResolver interface {
	LookupModule(ctx context.Context, namespaceID uint64, handle string) (uint64, error)
}

type crudExecutor struct {
	svc CRUDService
}

func (n *crudExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[crudConfig](node.Config)
	if err != nil {
		return nil, err
	}

	if cfg.RecordID != "" {
		cfg.RecordID = resolveTemplateValue(cfg.RecordID, ec)
	}
	if cfg.Query != "" {
		cfg.Query = resolveTemplateValue(cfg.Query, ec)
	}

	nsID := uint64(cfg.NamespaceID)
	if v := uint64FromAny(ec.Get("namespaceID")); v > 0 {
		nsID = v
	}
	modID := uint64(cfg.ModuleID)
	if cfg.ModuleHandle != "" {
		handle := resolveTemplateValue(cfg.ModuleHandle, ec)
		if resolver, ok := n.svc.(moduleResolver); ok {
			id, err := resolver.LookupModule(ctx, nsID, handle)
			if err != nil {
				return nil, fmt.Errorf("module %q: %w", handle, err)
			}
			modID = id
		}
	}

	switch cfg.Operation {
	case "create":
		if n.svc == nil {
			return map[string]interface{}{"status": "crud_service_not_configured"}, nil
		}
		id, createdAt, err := n.svc.Create(ctx, nsID, modID, resolveFieldTemplates(cfg.Fields, ec))
		if err != nil {
			return nil, fmt.Errorf("create failed: %w", err)
		}
		ec.Set("createdRecordID", id)
		return map[string]interface{}{"recordID": id, "createdAt": createdAt}, nil
	case "update":
		if n.svc == nil {
			return map[string]interface{}{"status": "crud_service_not_configured"}, nil
		}
		updatedAt, err := n.svc.Update(ctx, nsID, modID, cfg.RecordID, resolveFieldTemplates(cfg.Fields, ec))
		if err != nil {
			return nil, fmt.Errorf("update failed: %w", err)
		}
		return map[string]interface{}{"recordID": cfg.RecordID, "updatedAt": updatedAt}, nil
	case "delete":
		if n.svc == nil {
			return map[string]interface{}{"status": "crud_service_not_configured"}, nil
		}
		if err := n.svc.Delete(ctx, nsID, modID, cfg.RecordID); err != nil {
			return nil, fmt.Errorf("delete failed: %w", err)
		}
		return map[string]interface{}{"recordID": cfg.RecordID, "deleted": true}, nil
	case "search":
		if n.svc == nil {
			return map[string]interface{}{"status": "crud_service_not_configured"}, nil
		}
		records, err := n.svc.Search(ctx, nsID, modID, cfg.Query, cfg.Limit)
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
		return map[string]interface{}{"records": records, "total": len(records)}, nil
	default:
		return nil, fmt.Errorf("unknown CRUD operation: %s", cfg.Operation)
	}
}

// --- Mail Node ---

type mailConfig struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	Cc          string `json:"cc,omitempty"`
	ContentType string `json:"contentType,omitempty"`
}

type mailExecutor struct {
	svc MailService
}

func (n *mailExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[mailConfig](node.Config)
	if err != nil {
		return nil, err
	}

	to := splitTrimStr(resolveTemplateValue(cfg.To, ec))
	if len(to) == 0 {
		return nil, fmt.Errorf("at least one recipient required")
	}

	subject := resolveTemplateValue(cfg.Subject, ec)
	body := resolveTemplateValue(cfg.Body, ec)
	cc := splitTrimStr(resolveTemplateValue(cfg.Cc, ec))
	contentType := cfg.ContentType
	if contentType == "" {
		contentType = "html"
	}

	if n.svc == nil {
		return map[string]interface{}{
			"status":  "mail_service_not_configured",
			"to":      to,
			"subject": subject,
		}, nil
	}

	if err := n.svc.Send(ctx, to, subject, body, cc, contentType); err != nil {
		return nil, fmt.Errorf("send mail failed: %w", err)
	}

	return map[string]interface{}{"sent": true, "to": to, "subject": subject}, nil
}

// --- HTTP Node ---

type httpConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

type httpExecutor struct{}

func (n *httpExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[httpConfig](node.Config)
	if err != nil {
		return nil, err
	}

	url := resolveTemplateValue(cfg.URL, ec)
	method := resolveTemplateValue(cfg.Method, ec)
	if method == "" {
		method = "GET"
	}
	body := resolveTemplateValue(cfg.Body, ec)

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	for k, v := range cfg.Headers {
		req.Header.Set(resolveTemplateValue(k, ec), resolveTemplateValue(v, ec))
	}
	if body != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	result := map[string]interface{}{"statusCode": resp.StatusCode, "status": resp.Status}
	var jsonBody interface{}
	if err := json.Unmarshal(respBody, &jsonBody); err == nil {
		result["body"] = jsonBody
		if m, ok := jsonBody.(map[string]interface{}); ok {
			if id, ok := m["id"]; ok {
				ec.Set("scanID", fmt.Sprintf("%v", id))
			}
		}
	} else {
		bodyStr := string(respBody)
		if len(bodyStr) > 10000 {
			bodyStr = bodyStr[:10000] + "..."
			result["truncated"] = true
		}
		result["body"] = bodyStr
	}
	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncateHTTPBody(respBody))
	}
	return result, nil
}

func truncateHTTPBody(b []byte) string {
	s := string(b)
	if len(s) > 500 {
		return s[:500] + "..."
	}
	return s
}

// --- Condition Node ---

type conditionConfig struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

type conditionExecutor struct{}

func (n *conditionExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[conditionConfig](node.Config)
	if err != nil {
		return nil, err
	}

	field := resolveTemplateValue(cfg.Field, ec)
	fieldVal := ec.Get(field)
	compareVal := resolveTemplateValue(cfg.Value, ec)

	passed := evaluateCondition(fieldVal, compareVal, cfg.Operator)
	// Empty string when false so conditional edges (GetString != "") skip the branch,
	// matching risk.band's is_critical flag semantics.
	passStr := ""
	if passed {
		passStr = "true"
	}
	ec.Set(node.ID+"_result", passStr)

	passedOut := "false"
	if passed {
		passedOut = "true"
	}
	return map[string]interface{}{"field": field, "operator": cfg.Operator, "value": compareVal, "result": passed, "passed": passedOut}, nil
}

// --- AI Node ---

type aiConfig struct {
	Agent     string `json:"agent"`
	Prompt    string `json:"prompt"`
	Model     string `json:"model,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty"`
}

type aiExecutor struct {
	call func(ctx context.Context, agent, prompt, model string) (string, error)
}

func (n *aiExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[aiConfig](node.Config)
	if err != nil {
		return nil, err
	}

	prompt := resolveTemplateValue(cfg.Prompt, ec)
	agent := resolveTemplateValue(cfg.Agent, ec)
	model := resolveTemplateValue(cfg.Model, ec)

	if prompt == "" {
		return nil, fmt.Errorf("prompt is required for AI node")
	}

	if n.call == nil {
		return map[string]interface{}{"agent": agent, "status": "not_configured"}, nil
	}

	response, err := n.call(ctx, agent, prompt, model)
	if err != nil {
		return nil, fmt.Errorf("AI call failed: %w", err)
	}

	ec.Set("ai_response", response)
	return map[string]interface{}{"agent": agent, "response": response}, nil
}

// --- Workflow Node ---

type wfConfig struct {
	WorkflowID string `json:"workflowID"`
	Payload    string `json:"payload"`
}

type wfExecutor struct{}

func (n *wfExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[wfConfig](node.Config)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"workflowID": resolveTemplateValue(cfg.WorkflowID, ec),
		"status":     "not_implemented",
	}, nil
}

// --- Fork Node ---

type forkConfig struct {
	Branches int `json:"branches"`
}

type forkExecutor struct{}

func (n *forkExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[forkConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if cfg.Branches < 2 {
		cfg.Branches = 2
	}
	ids := make([]string, cfg.Branches)
	for i := 0; i < cfg.Branches; i++ {
		ids[i] = fmt.Sprintf("branch_%d", i+1)
	}
	return map[string]interface{}{"branches": cfg.Branches, "ids": ids}, nil
}

// --- Script Node ---

type scriptConfig struct {
	Code string `json:"code"`
}

type scriptExecutor struct {
	exec func(ctx context.Context, code string, ec *ExecutionContext) (map[string]interface{}, error)
}

func (n *scriptExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[scriptConfig](node.Config)
	if err != nil {
		return nil, err
	}
	code := resolveTemplateValue(cfg.Code, ec)
	if n.exec != nil {
		return n.exec(ctx, code, ec)
	}
	return map[string]interface{}{"code": code, "status": "not_configured"}, nil
}

// --- Gonec Node ---

type gonecConfig struct {
	Code    string `json:"code"`
	Timeout int    `json:"timeout,omitempty"`
}

type gonecExecutor struct {
	runner func(ctx context.Context, code string) (string, error)
}

func (n *gonecExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[gonecConfig](node.Config)
	if err != nil {
		return nil, err
	}
	code := resolveTemplateValue(cfg.Code, ec)

	if n.runner == nil {
		return map[string]interface{}{"code": code, "status": "not_configured"}, nil
	}

	result, err := n.runner(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("gonec execution failed: %w", err)
	}

	return map[string]interface{}{
		"code":   code,
		"result": result,
	}, nil
}

// --- Helpers ---

func resolveTemplateValue(val string, ec *ExecutionContext) string {
	if len(val) > 4 && val[:2] == "{{" && val[len(val)-2:] == "}}" {
		key := val[2 : len(val)-2]
		if v := ec.Get(key); v != nil {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}

	result := val
	for i := 0; i < len(result); {
		start := strings.Index(result[i:], "{{")
		if start < 0 {
			break
		}
		start += i
		end := strings.Index(result[start+2:], "}}")
		if end < 0 {
			break
		}
		end += start + 2
		key := result[start+2 : end]
		if v := ec.Get(key); v != nil {
			replacement := fmt.Sprintf("%v", v)
			result = result[:start] + replacement + result[end+2:]
			i = start + len(replacement)
		} else {
			i = end + 2
		}
	}
	return result
}

func resolveFieldTemplates(fields map[string]interface{}, ec *ExecutionContext) map[string]interface{} {
	if fields == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})
	for k, v := range fields {
		switch s := v.(type) {
		case string:
			result[k] = resolveTemplateValue(s, ec)
		default:
			result[k] = v
		}
	}
	return result
}

func uint64FromAny(v interface{}) uint64 {
	switch t := v.(type) {
	case nil:
		return 0
	case uint64:
		return t
	case int:
		if t > 0 {
			return uint64(t)
		}
	case int64:
		if t > 0 {
			return uint64(t)
		}
	case float64:
		if t > 0 {
			return uint64(t)
		}
	case json.Number:
		n, _ := t.Int64()
		if n > 0 {
			return uint64(n)
		}
	case string:
		n, _ := strconv.ParseUint(strings.TrimSpace(t), 10, 64)
		return n
	default:
		n, _ := strconv.ParseUint(strings.TrimSpace(fmt.Sprintf("%v", t)), 10, 64)
		return n
	}
	return 0
}

func splitTrimStr(s string) []string {
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

func evaluateCondition(fieldVal interface{}, compareVal, operator string) bool {
	fieldStr := fmt.Sprintf("%v", fieldVal)
	switch operator {
	case "empty":
		return fieldStr == ""
	case "notEmpty":
		return fieldStr != ""
	case "eq":
		return fieldStr == compareVal
	case "neq":
		return fieldStr != compareVal
	case "contains":
		return strings.Contains(fieldStr, compareVal)
	case "gt", "gte", "lt", "lte":
		f, e1 := strconv.ParseFloat(fieldStr, 64)
		v, e2 := strconv.ParseFloat(compareVal, 64)
		if e1 != nil || e2 != nil {
			return false
		}
		switch operator {
		case "gt":
			return f > v
		case "gte":
			return f >= v
		case "lt":
			return f < v
		case "lte":
			return f <= v
		}
	}
	return false
}
