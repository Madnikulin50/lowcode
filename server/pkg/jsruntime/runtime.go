package jsruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

type Runtime struct {
	mu       sync.RWMutex
	skills   map[string]goja.Callable
	services Services
}

type Services struct {
	MCPCall      func(ctx context.Context, tool string, params map[string]interface{}) (map[string]interface{}, error)
	RecordCreate func(ctx context.Context, nsID, modID uint64, values map[string]interface{}) (string, string, error)
	RecordUpdate func(ctx context.Context, nsID, modID uint64, recordID string, values map[string]interface{}) (string, error)
	RecordDelete func(ctx context.Context, nsID, modID uint64, recordID string) error
	RecordSearch func(ctx context.Context, nsID, modID uint64, query string, limit int) ([]map[string]interface{}, error)
	MailSend     func(ctx context.Context, to []string, subject, body string, cc []string, contentType string) error
}

type ScriptResult struct {
	Output   interface{} `json:"output"`
	Logs     []string    `json:"logs"`
	Duration string      `json:"duration"`
	Error    string      `json:"error,omitempty"`
}

func New(services Services) *Runtime {
	return &Runtime{
		skills:   make(map[string]goja.Callable),
		services: services,
	}
}

func (r *Runtime) RegisterSkill(name string, fn goja.Callable) {
	r.mu.Lock()
	r.skills[name] = fn
	r.mu.Unlock()
}

func (r *Runtime) Run(ctx context.Context, script string, input map[string]interface{}) *ScriptResult {
	start := time.Now()
	vm := goja.New()
	logs := &safeSlice{}

	r.setupRuntime(ctx, vm, logs, input)

	result := &ScriptResult{
		Logs: make([]string, 0),
	}

	defer func() {
		if rec := recover(); rec != nil {
			result.Error = fmt.Sprintf("panic: %v", rec)
		}
		result.Duration = time.Since(start).String()
		logs.mu.Lock()
		result.Logs = logs.items
		logs.mu.Unlock()
	}()

	// Wrap script in async function to support async/await
	wrapped := fmt.Sprintf(`
		(async function() {
			%s
		})();
	`, script)

	val, err := vm.RunString(wrapped)
	if err != nil {
		// Try running without async wrapper
		val, err = vm.RunString(script)
		if err != nil {
			result.Error = err.Error()
			return result
		}
	}

	if val != nil && !goja.IsUndefined(val) && !goja.IsNull(val) {
		result.Output = val.Export()
	}

	return result
}

func (r *Runtime) setupRuntime(ctx context.Context, vm *goja.Runtime, logs *safeSlice, input map[string]interface{}) {
	runtimeObj := vm.NewObject()

	// --- mcp ---
	mcpObj := vm.NewObject()
	mcpObj.Set("call", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("mcp.call requires tool name and params"))
		}
		tool := call.Arguments[0].String()
		paramsRaw := call.Arguments[1].Export()
		params, ok := paramsRaw.(map[string]interface{})
		if !ok {
			panic(vm.ToValue("mcp.call params must be an object"))
		}

		if r.services.MCPCall == nil {
			return vm.ToValue(map[string]interface{}{"error": "MCP call not configured"})
		}

		result, err := r.services.MCPCall(ctx, tool, params)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(result)
	})

	mcpObj.Set("createRecord", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(vm.ToValue("mcp.createRecord requires nsID, modID, data"))
		}
		nsID := uint64(call.Arguments[0].ToInteger())
		modID := uint64(call.Arguments[1].ToInteger())
		values, _ := call.Arguments[2].Export().(map[string]interface{})

		if r.services.RecordCreate == nil {
			return vm.ToValue(map[string]interface{}{"error": "not configured"})
		}
		id, createdAt, err := r.services.RecordCreate(ctx, nsID, modID, values)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"recordID": id, "createdAt": createdAt})
	})

	mcpObj.Set("updateRecord", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 4 {
			panic(vm.ToValue("mcp.updateRecord requires nsID, modID, recordID, data"))
		}
		nsID := uint64(call.Arguments[0].ToInteger())
		modID := uint64(call.Arguments[1].ToInteger())
		recordID := call.Arguments[2].String()
		values, _ := call.Arguments[3].Export().(map[string]interface{})

		if r.services.RecordUpdate == nil {
			return vm.ToValue(map[string]interface{}{"error": "not configured"})
		}
		updatedAt, err := r.services.RecordUpdate(ctx, nsID, modID, recordID, values)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"recordID": recordID, "updatedAt": updatedAt})
	})

	mcpObj.Set("deleteRecord", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(vm.ToValue("mcp.deleteRecord requires nsID, modID, recordID"))
		}
		nsID := uint64(call.Arguments[0].ToInteger())
		modID := uint64(call.Arguments[1].ToInteger())
		recordID := call.Arguments[2].String()

		if r.services.RecordDelete == nil {
			return vm.ToValue(map[string]interface{}{"error": "not configured"})
		}
		err := r.services.RecordDelete(ctx, nsID, modID, recordID)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"deleted": true})
	})

	mcpObj.Set("searchRecords", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("mcp.searchRecords requires nsID, modID"))
		}
		nsID := uint64(call.Arguments[0].ToInteger())
		modID := uint64(call.Arguments[1].ToInteger())
		query := ""
		limit := 100
		if len(call.Arguments) > 2 {
			query = call.Arguments[2].String()
		}
		if len(call.Arguments) > 3 {
			limit = int(call.Arguments[3].ToInteger())
		}

		if r.services.RecordSearch == nil {
			return vm.ToValue(map[string]interface{}{"error": "not configured"})
		}
		records, err := r.services.RecordSearch(ctx, nsID, modID, query, limit)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(records)
	})

	runtimeObj.Set("mcp", mcpObj)

	// --- mail ---
	mailObj := vm.NewObject()
	mailObj.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(vm.ToValue("mail.send requires options"))
		}
		opts, _ := call.Arguments[0].Export().(map[string]interface{})
		to := toStringSlice(opts["to"])
		subject := opts["subject"]
		body := opts["body"]
		cc := toStringSlice(opts["cc"])
		contentType := "html"
		if v, ok := opts["contentType"].(string); ok {
			contentType = v
		}

		if r.services.MailSend == nil {
			return vm.ToValue(map[string]interface{}{"error": "not configured"})
		}
		err := r.services.MailSend(ctx, to, fmt.Sprint(subject), fmt.Sprint(body), cc, contentType)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"sent": true})
	})
	runtimeObj.Set("mail", mailObj)

	// --- http ---
	httpObj := vm.NewObject()
	httpObj.Set("get", func(call goja.FunctionCall) goja.Value {
		url := call.Arguments[0].String()
		result, err := jsHTTPGet(ctx, url)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(result)
	})
	httpObj.Set("post", func(call goja.FunctionCall) goja.Value {
		url := call.Arguments[0].String()
		body := ""
		if len(call.Arguments) > 1 {
			if b, ok := call.Arguments[1].Export().(string); ok {
				body = b
			} else {
				data, _ := json.Marshal(call.Arguments[1].Export())
				body = string(data)
			}
		}
		result, err := jsHTTPPost(ctx, url, body)
		if err != nil {
			return vm.ToValue(map[string]interface{}{"error": err.Error()})
		}
		return vm.ToValue(result)
	})
	runtimeObj.Set("http", httpObj)

	// --- skill ---
	skillObj := vm.NewObject()
	skillObj.Set("register", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("skill.register requires name and function"))
		}
		name := call.Arguments[0].String()
		fn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			panic(vm.ToValue("skill.register: second argument must be a function"))
		}
		r.RegisterSkill(name, fn)
		return goja.Undefined()
	})
	skillObj.Set("invoke", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("skill.invoke requires name and input"))
		}
		name := call.Arguments[0].String()
		r.mu.RLock()
		fn, ok := r.skills[name]
		r.mu.RUnlock()
		if !ok {
			return vm.ToValue(map[string]interface{}{"error": "skill not found: " + name})
		}
		input := call.Arguments[1]
		result, err := fn(input)
		if err != nil {
			panic(vm.ToValue(err.Error()))
		}
		return result
	})
	runtimeObj.Set("skill", skillObj)

	// --- log ---
	logObj := vm.NewObject()
	logObj.Set("info", func(call goja.FunctionCall) goja.Value {
		msg := argumentsToString(call)
		logs.append("INFO: " + msg)
		return goja.Undefined()
	})
	logObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		msg := argumentsToString(call)
		logs.append("WARN: " + msg)
		return goja.Undefined()
	})
	logObj.Set("error", func(call goja.FunctionCall) goja.Value {
		msg := argumentsToString(call)
		logs.append("ERROR: " + msg)
		return goja.Undefined()
	})
	runtimeObj.Set("log", logObj)

	// --- data ---
	dataObj := vm.NewObject()
	dataObj.Set("transform", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		data := call.Arguments[0].Export()
		template := call.Arguments[1].String()
		result := simpleTransform(data, template)
		return vm.ToValue(result)
	})
	runtimeObj.Set("data", dataObj)

	// --- context (input data) ---
	contextObj := vm.NewObject()
	if input != nil {
		for k, v := range input {
			contextObj.Set(k, v)
		}
	}
	runtimeObj.Set("context", contextObj)

	// --- console ---
	consoleObj := vm.NewObject()
	consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		logs.append(argumentsToString(call))
		return goja.Undefined()
	})
	consoleObj.Set("error", func(call goja.FunctionCall) goja.Value {
		logs.append("ERR:" + argumentsToString(call))
		return goja.Undefined()
	})
	vm.Set("console", consoleObj)

	vm.Set("runtime", runtimeObj)

	// Expose context keys as direct variables
	if input != nil {
		for k, v := range input {
			vm.Set(k, v)
		}
	}

	// setTimeout polyfill for async
	vm.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		panic(vm.NewGoError(fmt.Errorf("setTimeout not available in sync mode; use async/await with runtime.http or runtime.mcp")))
	})

	vm.Set("fetch", func(call goja.FunctionCall) goja.Value {
		url := call.Arguments[0].String()
		var opts map[string]interface{}
		if len(call.Arguments) > 1 {
			opts, _ = call.Arguments[1].Export().(map[string]interface{})
		}
		var body string
		method := "GET"
		if opts != nil {
			if m, ok := opts["method"].(string); ok {
				method = m
			}
			if b, ok := opts["body"].(string); ok {
				body = b
			} else if b2 := opts["body"]; b2 != nil {
				data, _ := json.Marshal(b2)
				body = string(data)
			}
		}

		var result map[string]interface{}
		var err error
		if method == "GET" {
			result, err = jsHTTPGet(ctx, url)
		} else {
			result, err = jsHTTPPost(ctx, url, body)
		}
		if err != nil {
			return vm.ToValue(map[string]interface{}{"ok": false, "error": err.Error()})
		}
		return vm.ToValue(map[string]interface{}{"ok": true, "json": func() goja.Value { return vm.ToValue(result) }, "data": result})
	})
}

func toStringSlice(v interface{}) []string {
	switch arr := v.(type) {
	case []string:
		return arr
	case []interface{}:
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			result = append(result, fmt.Sprint(item))
		}
		return result
	case string:
		return strings.Split(arr, ",")
	}
	return nil
}

func argumentsToString(call goja.FunctionCall) string {
	parts := make([]string, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		parts = append(parts, arg.String())
	}
	return strings.Join(parts, " ")
}

func jsHTTPGet(ctx context.Context, url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	result := map[string]interface{}{
		"status": resp.StatusCode,
	}
	var jsonBody interface{}
	if err := json.Unmarshal(body, &jsonBody); err == nil {
		result["body"] = jsonBody
	} else {
		bodyStr := string(body)
		if len(bodyStr) > 10000 {
			bodyStr = bodyStr[:10000] + "..."
		}
		result["body"] = bodyStr
	}
	return result, nil
}

func jsHTTPPost(ctx context.Context, url, body string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))

	result := map[string]interface{}{
		"status": resp.StatusCode,
	}
	var jsonBody interface{}
	if err := json.Unmarshal(respBody, &jsonBody); err == nil {
		result["body"] = jsonBody
	} else {
		bodyStr := string(respBody)
		if len(bodyStr) > 10000 {
			bodyStr = bodyStr[:10000] + "..."
		}
		result["body"] = bodyStr
	}
	return result, nil
}

func simpleTransform(data interface{}, template string) map[string]interface{} {
	result := make(map[string]interface{})
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		result["_raw"] = data
		return result
	}

	fields := strings.Split(template, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		parts := strings.SplitN(field, ":", 2)
		if len(parts) == 2 {
			src := strings.TrimSpace(parts[0])
			dst := strings.TrimSpace(parts[1])
			if v, ok := dataMap[src]; ok {
				result[dst] = v
			}
		} else {
			if v, ok := dataMap[field]; ok {
				result[field] = v
			}
		}
	}
	return result
}

type safeSlice struct {
	mu    sync.Mutex
	items []string
}

func (s *safeSlice) append(item string) {
	s.mu.Lock()
	s.items = append(s.items, item)
	s.mu.Unlock()
}
