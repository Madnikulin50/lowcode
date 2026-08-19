package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AgentPoller struct {
	engine  *Engine
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	client  *http.Client
}

func NewAgentPoller() *AgentPoller {
	return &AgentPoller{
		cancels: make(map[string]context.CancelFunc),
		client:  &http.Client{Timeout: 20 * time.Second},
	}
}

func (p *AgentPoller) SetEngine(e *Engine) {
	p.engine = e
}

func (p *AgentPoller) StartFromDetach(ctx context.Context, cfg DetachConfig, vars map[string]interface{}) error {
	jobID := strings.TrimSpace(fmt.Sprintf("%v", vars["jobID"]))
	if jobID == "" || jobID == "<nil>" {
		jobID = strings.TrimSpace(fmt.Sprintf("%v", vars["scanID"]))
	}
	if jobID == "" || jobID == "<nil>" {
		return fmt.Errorf("poll detach needs jobID or scanID")
	}
	interval := time.Duration(cfg.Interval) * time.Second
	if interval <= 0 {
		interval = 2 * time.Second
	}
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	until := parseUntil(cfg.Until)

	runCtx, cancel := context.WithTimeout(context.Background(), timeout)
	p.mu.Lock()
	if prev, ok := p.cancels[jobID]; ok {
		prev()
	}
	p.cancels[jobID] = cancel
	p.mu.Unlock()

	ident := pollIdentity{UserID: uint64FromAny(vars["userID"])}
	if CapturePollIdentity != nil {
		if uid, roles := CapturePollIdentity(ctx); uid > 0 {
			ident.UserID, ident.Roles = uid, roles
		}
	}
	go p.loop(runCtx, cancel, pollSpec{
		jobID:         jobID,
		ingestChainID: cfg.IngestChainID,
		statusURL:     cfg.StatusURL,
		itemsURL:      cfg.ItemsURL,
		interval:      interval,
		until:         until,
		base:          vars,
		ident:         ident,
	})
	return nil
}

func (p *AgentPoller) Cancel(jobID string) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.cancels[jobID]; ok {
		c()
		delete(p.cancels, jobID)
	}
}

type pollIdentity struct {
	UserID uint64
	Roles  []uint64
}

type pollSpec struct {
	jobID         string
	ingestChainID string
	statusURL     string
	itemsURL      string
	interval      time.Duration
	until         map[string]bool
	base          map[string]interface{}
	ident         pollIdentity
}

func (p *AgentPoller) loop(ctx context.Context, cancel context.CancelFunc, spec pollSpec) {
	defer func() {
		cancel()
		p.mu.Lock()
		delete(p.cancels, spec.jobID)
		p.mu.Unlock()
	}()
	ticker := time.NewTicker(spec.interval)
	defer ticker.Stop()

	runOnce := func() bool {
		st, err := p.getJSON(ctx, spec.statusURL)
		if err != nil {
			log.Printf("[poller] %s status: %v", spec.jobID[:min(8, len(spec.jobID))], err)
			return false
		}
		stMap, _ := st.(map[string]interface{})
		status := strings.ToLower(fmt.Sprintf("%v", firstMap(stMap, "status", "Status")))
		envelope := baseEnvelope(spec, st)
		if spec.until[status] {
			kind := "complete"
			if status == "error" || status == "failed" {
				kind = "failed"
			}
			envelope["kind"] = kind
			envelope["status"] = composeJobStatus(status)
			if spec.itemsURL != "" && kind == "complete" {
				items, err := p.getJSON(ctx, spec.itemsURL)
				if err != nil {
					log.Printf("[poller] %s items: %v", spec.jobID[:min(8, len(spec.jobID))], err)
				} else {
					envelope["items"] = jsonToItems(items)
					envelope["found"] = len(CollectItems(envelope["items"]))
				}
			}
			p.runIngest(ctx, spec, envelope)
			return true
		}
		envelope["kind"] = "progress"
		envelope["status"] = composeJobStatus(status)
		p.runIngest(ctx, spec, envelope)
		return false
	}

	if runOnce() {
		return
	}
	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				st, _ := p.getJSON(context.Background(), spec.statusURL)
				envelope := baseEnvelope(spec, st)
				envelope["kind"] = "failed"
				envelope["status"] = "failed"
				envelope["error"] = "poll timeout"
				p.runIngest(context.Background(), spec, envelope)
			}
			return
		case <-ticker.C:
			if runOnce() {
				return
			}
		}
	}
}

func (p *AgentPoller) runIngest(ctx context.Context, spec pollSpec, envelope map[string]interface{}) {
	if p.engine == nil {
		return
	}
	if RestorePollIdentity != nil {
		ctx = RestorePollIdentity(ctx, spec.ident.UserID, spec.ident.Roles)
	}
	kind, _ := envelope["kind"].(string)
	if kind == "complete" || kind == "failed" {
		// ingest will also cancel; avoid re-entrancy by cancelling after Run
	}
	if _, err := p.engine.Run(ctx, spec.ingestChainID, envelope); err != nil {
		log.Printf("[poller] ingest %s: %v", spec.ingestChainID, err)
	}
}

func (p *AgentPoller) getJSON(ctx context.Context, rawURL string) (interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func baseEnvelope(spec pollSpec, statusJSON interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range spec.base {
		out[k] = v
	}
	out["jobID"] = spec.jobID
	if emptyAny(out["scanRecordID"]) && !emptyAny(out["createdRecordID"]) {
		out["scanRecordID"] = out["createdRecordID"]
	}
	m, _ := statusJSON.(map[string]interface{})
	if m != nil {
		out["progress"] = firstMap(m, "progress", "Progress")
		out["found"] = firstMap(m, "found", "Found")
		out["error"] = firstNonEmptyStr(firstMap(m, "error", "Error"), firstMap(m, "message", "Message"))
		out["scanningIP"] = firstMap(m, "scanningIP", "ScanningIP")
		out["target"] = firstMap(m, "target", "Target")
		out["startedAt"] = firstMap(m, "startedAt", "StartedAt")
		out["finishedAt"] = firstMap(m, "finishedAt", "FinishedAt")
	}
	return out
}

func jsonToItems(v interface{}) []interface{} {
	if items := CollectItems(v); len(items) > 0 {
		return items
	}
	if m, ok := v.(map[string]interface{}); ok {
		for _, k := range []string{"items", "devices", "set"} {
			if items := CollectItems(m[k]); len(items) > 0 {
				return items
			}
		}
	}
	return nil
}

func composeJobStatus(agentStatus string) string {
	switch strings.ToLower(strings.TrimSpace(agentStatus)) {
	case "done", "completed", "complete":
		return "completed"
	case "error", "failed", "fail":
		return "failed"
	case "running", "pending":
		return "running"
	default:
		if agentStatus == "" {
			return "running"
		}
		return agentStatus
	}
}

func parseUntil(s string) map[string]bool {
	out := map[string]bool{}
	if strings.TrimSpace(s) == "" {
		s = "done,error,completed,failed"
	}
	for _, p := range strings.Split(s, ",") {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out[p] = true
		}
	}
	return out
}

func firstMap(m map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func firstNonEmptyStr(vals ...interface{}) string {
	for _, v := range vals {
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

var (
	CapturePollIdentity func(ctx context.Context) (userID uint64, roles []uint64)
	RestorePollIdentity func(ctx context.Context, userID uint64, roles []uint64) context.Context
)

var defaultPoller *AgentPoller

func SetDefaultPoller(p *AgentPoller) {
	defaultPoller = p
}

func CancelPoll(jobID string) {
	if defaultPoller != nil {
		defaultPoller.Cancel(jobID)
	}
}

func CancelPollIfTerminal(input map[string]interface{}) {
	if input == nil {
		return
	}
	kind, _ := input["kind"].(string)
	if kind != "complete" && kind != "failed" {
		return
	}
	id := strings.TrimSpace(fmt.Sprintf("%v", input["jobID"]))
	if id == "" || id == "<nil>" {
		id = strings.TrimSpace(fmt.Sprintf("%v", input["scanID"]))
	}
	CancelPoll(id)
}
