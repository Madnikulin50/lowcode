package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cortezaBase struct {
	mu       sync.Mutex
	url      string
	resolved bool
}

type Corteza struct {
	base        *cortezaBase
	token       string
	httpClient  *http.Client
	namespaceID uint64
	modules     map[string]uint64
}

func NewCorteza(baseURL, token string, namespaceID uint64, timeout time.Duration) *Corteza {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Corteza{
		base:        &cortezaBase{url: strings.TrimRight(strings.TrimSpace(baseURL), "/")},
		token:       token,
		httpClient:  &http.Client{Timeout: timeout},
		namespaceID: namespaceID,
		modules:     map[string]uint64{},
	}
}

func (c *Corteza) BaseURL() string {
	if c == nil || c.base == nil {
		return ""
	}
	c.base.mu.Lock()
	defer c.base.mu.Unlock()
	return c.base.url
}

func (c *Corteza) WithToken(token string) *Corteza {
	if strings.TrimSpace(token) == "" {
		return c
	}
	cp := *c
	cp.token = token
	cp.modules = map[string]uint64{}
	return &cp
}

func (c *Corteza) WithNamespace(id uint64) *Corteza {
	if id == 0 {
		return c
	}
	cp := *c
	cp.namespaceID = id
	cp.modules = map[string]uint64{}
	return &cp
}

// composeOriginCandidates turns --api values into origins that, when
// concatenated with /compose/..., hit the real Compose REST mount.
// GoLand with HTTP_WEBAPP_ENABLED=false serves APIs at /compose, not /api/compose
// (chi's HTML 404 is the Bootstrap 4.6 page under /api/...).
func composeOriginCandidates(raw string) []string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		raw = "http://127.0.0.1:3333"
	}
	origin := strings.TrimSuffix(raw, "/compose")
	origin = strings.TrimSuffix(origin, "/api")
	origin = strings.TrimRight(origin, "/")
	seen := map[string]struct{}{}
	out := make([]string, 0, 3)
	add := func(s string) {
		s = strings.TrimRight(s, "/")
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	add(origin)
	add(origin + "/api")
	add(raw)
	return out
}

func looksLikeHTML(body []byte) bool {
	s := strings.TrimSpace(strings.ToLower(string(body)))
	return strings.HasPrefix(s, "<!doctype") || strings.Contains(s, "<html")
}

func looksLikeComposeAPI(status int, body []byte) bool {
	if looksLikeHTML(body) {
		return false
	}
	if status == http.StatusNotFound {
		return false
	}
	return true
}

func apiErrorBody(raw []byte) string {
	s := strings.TrimSpace(string(raw))
	if looksLikeHTML(raw) {
		return "HTML 404 (Compose is not under this prefix; use --api=http://127.0.0.1:3333 not .../api)"
	}
	if i := strings.Index(s, "\n"); i > 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) > 300 {
		s = s[:300] + "..."
	}
	return s
}

func probeComposeOrigin(ctx context.Context, origin, token string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin+"/compose/namespace/?limit=1", nil)
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return looksLikeComposeAPI(resp.StatusCode, raw)
}

// Discover picks the origin that actually serves Compose REST.
// Safe to call repeatedly; WithToken copies share the resolved origin.
func (c *Corteza) Discover(ctx context.Context) error {
	if c == nil || c.base == nil {
		return fmt.Errorf("corteza client is nil")
	}
	c.base.mu.Lock()
	defer c.base.mu.Unlock()
	if c.base.resolved {
		return nil
	}
	given := c.base.url
	candidates := composeOriginCandidates(given)
	for _, origin := range candidates {
		if probeComposeOrigin(ctx, origin, c.token) {
			if origin != given {
				log.Printf("corteza API origin %s (from --api=%s)", origin, given)
			}
			c.base.url = origin
			c.base.resolved = true
			return nil
		}
	}
	return fmt.Errorf("no Compose API at %s (tried %s); pass --api=http://127.0.0.1:3333", given, strings.Join(candidates, ", "))
}

type composeModule struct {
	ID     uint64 `json:"moduleID,string"`
	Handle string `json:"handle"`
}

type composeNamespace struct {
	ID   uint64 `json:"namespaceID,string"`
	Slug string `json:"slug"`
}

type composeRecord struct {
	ID        uint64               `json:"recordID,string"`
	UpdatedAt string               `json:"updatedAt,omitempty"`
	Values    []composeRecordValue `json:"values"`
}

type composeRecordValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *Corteza) request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	if err := c.Discover(ctx); err != nil {
		return nil, err
	}
	var b io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL()+path, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 || looksLikeHTML(raw) {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, apiErrorBody(raw))
	}
	if msg := cortezaJSONError(raw); msg != "" {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, msg)
	}
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Response) > 0 {
		return envelope.Response, nil
	}
	return raw, nil
}

// cortezaJSONError reads {"error":...} from a 200 OK body. Compose in
// development often returns permission/validation failures with HTTP 200.
func cortezaJSONError(raw []byte) string {
	var env struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &env) != nil || len(env.Error) == 0 || string(env.Error) == "null" {
		return ""
	}
	var msg string
	if json.Unmarshal(env.Error, &msg) == nil && strings.TrimSpace(msg) != "" {
		return msg
	}
	var obj struct {
		Message string `json:"message"`
		Details []struct {
			Kind    string                 `json:"kind"`
			Message string                 `json:"message"`
			Meta    map[string]interface{} `json:"meta"`
		} `json:"details"`
	}
	if json.Unmarshal(env.Error, &obj) != nil {
		return apiErrorBody(env.Error)
	}
	parts := make([]string, 0, 1+len(obj.Details))
	if strings.TrimSpace(obj.Message) != "" {
		parts = append(parts, obj.Message)
	}
	for _, d := range obj.Details {
		field, _ := d.Meta["field"].(string)
		bit := d.Kind
		if field != "" {
			if bit != "" {
				bit = field + " " + bit
			} else {
				bit = field
			}
		}
		if d.Message != "" && d.Message != obj.Message {
			if bit != "" {
				bit += ": " + d.Message
			} else {
				bit = d.Message
			}
		}
		if bit != "" {
			parts = append(parts, bit)
		}
	}
	if len(parts) == 0 {
		return apiErrorBody(env.Error)
	}
	return strings.Join(parts, "; ")
}

func unwrapSet[T any](raw []byte) ([]T, error) {
	var wrapped struct {
		Set []T `json:"set"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Set != nil {
		return wrapped.Set, nil
	}
	var flat []T
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, err
	}
	return flat, nil
}

func (c *Corteza) ResolveNamespace(ctx context.Context) (uint64, error) {
	if c.namespaceID != 0 {
		return c.namespaceID, nil
	}
	raw, err := c.request(ctx, "GET", "/compose/namespace/?slug=invest&limit=50", nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeNamespace](raw)
	if err != nil {
		return 0, err
	}
	for _, ns := range set {
		if strings.EqualFold(ns.Slug, "invest") {
			c.namespaceID = ns.ID
			return ns.ID, nil
		}
	}
	return 0, fmt.Errorf("namespace slug invest not found; run agents/invest/compose/apply.mjs")
}

func (c *Corteza) ModuleByHandle(ctx context.Context, handle string) (uint64, error) {
	if id, ok := c.modules[handle]; ok {
		return id, nil
	}
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	raw, err := c.request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/?handle=%s&limit=50", nsID, url.QueryEscape(handle)), nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeModule](raw)
	if err != nil {
		return 0, err
	}
	for _, m := range set {
		if m.Handle == handle {
			c.modules[handle] = m.ID
			return m.ID, nil
		}
	}
	return 0, fmt.Errorf("module %s not found", handle)
}

func (c *Corteza) ListRecords(ctx context.Context, handle, query string) ([]composeRecord, error) {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return nil, err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf("/compose/namespace/%d/module/%d/record/?limit=500", nsID, modID)
	if query != "" {
		q += "&query=" + url.QueryEscape(query)
	}
	raw, err := c.request(ctx, "GET", q, nil)
	if err != nil {
		return nil, err
	}
	return unwrapSet[composeRecord](raw)
}

func (c *Corteza) GetRecord(ctx context.Context, handle string, recordID uint64) (*composeRecord, error) {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return nil, err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	raw, err := c.request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID), nil)
	if err != nil {
		return nil, err
	}
	var rec composeRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func overlayRecordValues(existing []composeRecordValue, patch map[string]string) []composeRecordValue {
	out := make([]composeRecordValue, 0, len(existing)+len(patch))
	replaced := map[string]bool{}
	for _, v := range existing {
		if nv, ok := patch[v.Name]; ok {
			if replaced[v.Name] {
				continue
			}
			replaced[v.Name] = true
			out = append(out, composeRecordValue{Name: v.Name, Value: nv})
			continue
		}
		out = append(out, v)
	}
	for k, v := range patch {
		if replaced[k] {
			continue
		}
		out = append(out, composeRecordValue{Name: k, Value: v})
	}
	return out
}

func (c *Corteza) UpdateValues(ctx context.Context, handle string, recordID uint64, values map[string]string) error {
	rec, err := c.GetRecord(ctx, handle, recordID)
	if err != nil {
		return err
	}
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return err
	}
	// Compose record POST replaces the value set: omitted required fields
	// (name/code/project) are treated as deleted → "3 issue(s) found".
	payload := overlayRecordValues(rec.Values, values)
	body := map[string]interface{}{"values": payload}
	if rec.UpdatedAt != "" {
		body["updatedAt"] = rec.UpdatedAt
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	_, err = c.request(ctx, "POST", path, body)
	return err
}

func (c *Corteza) CreateValues(ctx context.Context, handle string, values map[string]string) (uint64, error) {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return 0, err
	}
	payload := make([]composeRecordValue, 0, len(values))
	for k, v := range values {
		payload = append(payload, composeRecordValue{Name: k, Value: v})
	}
	raw, err := c.request(ctx, "POST", fmt.Sprintf("/compose/namespace/%d/module/%d/record/", nsID, modID), map[string]interface{}{"values": payload})
	if err != nil {
		return 0, err
	}
	var rec composeRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return 0, err
	}
	return rec.ID, nil
}

func recMap(r composeRecord) map[string]string {
	m := map[string]string{}
	for _, v := range r.Values {
		if _, ok := m[v.Name]; !ok {
			m[v.Name] = v.Value
		} else if v.Value != "" {
			m[v.Name] = m[v.Name] + "\n" + v.Value
		}
	}
	return m
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(strings.ReplaceAll(s, " ", ""))
	s = strings.ReplaceAll(s, ",", ".")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func splitIDs(s string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == ',' || r == ' ' }) {
		p = strings.TrimSpace(p)
		if p != "" && p != "0" {
			out = append(out, p)
		}
	}
	return out
}

func (c *Corteza) LoadWBS(ctx context.Context, projectID string) ([]WBSItem, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "wbs_items", q)
	if err != nil {
		return nil, err
	}
	items := make([]WBSItem, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		items = append(items, WBSItem{
			ID:              r.ID,
			ProjectID:       m["project"],
			Code:            m["code"],
			Name:            m["name"],
			ParentID:        m["parent"],
			Predecessors:    splitIDs(m["predecessor"]),
			StartPlanned:    parseTime(m["start_planned"]),
			EndPlanned:      parseTime(m["end_planned"]),
			BudgetPlanned:   parseFloat(m["budget_planned"]),
			ActualCost:      parseFloat(m["actual_cost"]),
			PercentComplete: parseFloat(m["percent_complete"]),
		})
	}
	return items, nil
}

func (c *Corteza) LoadFacts(ctx context.Context, projectID string) ([]ProgressFact, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "progress_facts", q)
	if err != nil {
		return nil, err
	}
	facts := make([]ProgressFact, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		facts = append(facts, ProgressFact{
			WBSID:   m["wbs"],
			Percent: parseFloat(m["percent"]),
			Cost:    parseFloat(m["cost"]),
		})
	}
	return facts, nil
}

func (c *Corteza) LoadDocuments(ctx context.Context, projectID string) ([]Document, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "documents", q)
	if err != nil {
		return nil, err
	}
	docs := make([]Document, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		docs = append(docs, Document{
			ID:       r.ID,
			Project:  m["project"],
			Title:    m["title"],
			Status:   m["status"],
			DueDate:  parseTime(m["due_date"]),
			Assignee: m["assignee"],
			Author:   m["author"],
			Contract: m["contract"],
			File:     m["file"],
			Number:   m["number"],
		})
	}
	return docs, nil
}

func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', 4, 64)
}

func (c *Corteza) SaveWBSMetrics(ctx context.Context, items []WBSItem) error {
	for _, it := range items {
		vals := map[string]string{
			"pv":               fmtNum(it.PV),
			"ev":               fmtNum(it.EV),
			"spi":              fmtNum(it.SPI),
			"cpi":              fmtNum(it.CPI),
			"eac":              fmtNum(it.EAC),
			"percent_complete": fmtNum(it.PercentComplete),
			"actual_cost":      fmtNum(it.ActualCost),
		}
		if err := c.UpdateValues(ctx, "wbs_items", it.ID, vals); err != nil {
			return fmt.Errorf("update wbs %d: %w", it.ID, err)
		}
	}
	return nil
}

func (c *Corteza) SaveWBSCritical(ctx context.Context, items []WBSItem) error {
	for _, it := range items {
		vals := map[string]string{
			"is_critical": boolStr(it.IsCritical),
			"total_float": fmtNum(it.TotalFloat),
		}
		if err := c.UpdateValues(ctx, "wbs_items", it.ID, vals); err != nil {
			return fmt.Errorf("update wbs critical %d: %w", it.ID, err)
		}
	}
	return nil
}

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func (c *Corteza) SaveProjectEVM(ctx context.Context, projectID string, r EVMResult) error {
	id, err := strconv.ParseUint(strings.TrimSpace(projectID), 10, 64)
	if err != nil || id == 0 {
		return fmt.Errorf("invalid projectID %q", projectID)
	}
	return c.UpdateValues(ctx, "projects", id, map[string]string{
		"spi":           fmtNum(r.SPI),
		"cpi":           fmtNum(r.CPI),
		"eac":           fmtNum(r.EAC),
		"budget_actual": fmtNum(r.AC),
	})
}

func (c *Corteza) EnsureRisk(ctx context.Context, a Alert) error {
	q := fmt.Sprintf("title = '%s'", strings.ReplaceAll(a.Title, "'", " "))
	existing, err := c.ListRecords(ctx, "risks", q)
	if err == nil && len(existing) > 0 {
		return nil
	}
	vals := map[string]string{
		"title":       a.Title,
		"project":     a.Project,
		"wbs":         a.WBS,
		"probability": "high",
		"impact":      "high",
		"status":      "open",
		"description": a.Detail,
		"score":       fmtNum(RiskScore("high", "high")),
	}
	_, err = c.CreateValues(ctx, "risks", compactValues(vals))
	return err
}

func (c *Corteza) LoadDocument(ctx context.Context, id uint64) (*Document, error) {
	rec, err := c.GetRecord(ctx, "documents", id)
	if err != nil {
		return nil, err
	}
	m := recMap(*rec)
	return &Document{
		ID: rec.ID, Project: m["project"], Title: m["title"], Status: m["status"],
		DueDate: parseTime(m["due_date"]), Assignee: m["assignee"], Author: m["author"],
		Contract: m["contract"], File: m["file"], Number: m["number"],
	}, nil
}

func (c *Corteza) LoadProject(ctx context.Context, id uint64) (*Project, error) {
	if id == 0 {
		return nil, fmt.Errorf("project id required")
	}
	rec, err := c.GetRecord(ctx, "projects", id)
	if err != nil {
		return nil, err
	}
	m := recMap(*rec)
	return &Project{
		ID: rec.ID, Name: m["name"], BudgetPlanned: parseFloat(m["budget_planned"]),
		BudgetActual: parseFloat(m["budget_actual"]), EAC: parseFloat(m["eac"]),
		EndPlanned: parseTime(m["end_planned"]), ConstructionType: m["construction_type"],
	}, nil
}

func (c *Corteza) LoadRFC(ctx context.Context, id uint64) (*RFC, error) {
	rec, err := c.GetRecord(ctx, "change_requests", id)
	if err != nil {
		return nil, err
	}
	m := recMap(*rec)
	return &RFC{
		ID: rec.ID, Project: m["project"], Title: m["title"], Status: m["status"],
		DeltaBudget: parseFloat(m["delta_budget"]), DeltaDays: parseFloat(m["delta_days"]),
		EACBefore: parseFloat(m["eac_before"]), EACAfter: parseFloat(m["eac_after"]),
		Simulated: boolish(m["simulated"]), EndAfter: parseTime(m["end_after"]),
		Author: m["author"],
	}, nil
}

func (c *Corteza) LoadMembers(ctx context.Context, projectID string) ([]Member, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "project_members", q)
	if err != nil {
		return nil, err
	}
	out := make([]Member, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		out = append(out, Member{User: m["user"], Role: m["role"]})
	}
	return out, nil
}

func (c *Corteza) LoadApprovals(ctx context.Context, documentID string) ([]Approval, error) {
	q := ""
	if documentID != "" && documentID != "0" {
		q = "document = " + documentID
	}
	recs, err := c.ListRecords(ctx, "approvals", q)
	if err != nil {
		return nil, err
	}
	out := make([]Approval, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		out = append(out, Approval{
			ID: r.ID, Document: m["document"], Approver: m["approver"],
			Decision: m["decision"], Step: parseFloat(m["step"]), Role: m["role"],
		})
	}
	return out, nil
}

func (c *Corteza) ListVersionNumbers(ctx context.Context, documentID string) ([]float64, error) {
	q := ""
	if documentID != "" && documentID != "0" {
		q = "document = " + documentID
	}
	recs, err := c.ListRecords(ctx, "document_versions", q)
	if err != nil {
		return nil, err
	}
	out := make([]float64, 0, len(recs))
	for _, r := range recs {
		out = append(out, parseFloat(recMap(r)["version"]))
	}
	return out, nil
}

func (c *Corteza) LoadBudgetLines(ctx context.Context, projectID string) ([]BudgetLine, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "budget_lines", q)
	if err != nil {
		return nil, err
	}
	out := make([]BudgetLine, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		out = append(out, BudgetLine{
			Project: m["project"], Article: m["article"],
			Reserve: parseFloat(m["reserve"]), Actual: parseFloat(m["actual"]),
			Planned: parseFloat(m["planned"]),
		})
	}
	return out, nil
}

func (c *Corteza) LoadRFCs(ctx context.Context, projectID string) ([]RFC, error) {
	q := ""
	if projectID != "" && projectID != "0" {
		q = "project = " + projectID
	}
	recs, err := c.ListRecords(ctx, "change_requests", q)
	if err != nil {
		return nil, err
	}
	out := make([]RFC, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		out = append(out, RFC{
			ID: r.ID, Project: m["project"], Title: m["title"], Status: m["status"],
			EndAfter: parseTime(m["end_after"]), Simulated: boolish(m["simulated"]),
		})
	}
	return out, nil
}

func (c *Corteza) LoadWBSTemplates(ctx context.Context, typeID string) ([]WBSTemplate, error) {
	q := ""
	if typeID != "" && typeID != "0" {
		q = "construction_type = " + typeID
	}
	recs, err := c.ListRecords(ctx, "wbs_templates", q)
	if err != nil {
		return nil, err
	}
	out := make([]WBSTemplate, 0, len(recs))
	for _, r := range recs {
		m := recMap(r)
		out = append(out, WBSTemplate{
			ID: r.ID, Type: m["construction_type"], Code: m["code"], Name: m["name"],
			Level: m["level"], ParentCode: m["parent_code"], PredecessorCode: m["predecessor_code"],
			BudgetPlanned: parseFloat(m["budget_planned"]), DurationDays: parseFloat(m["duration_days"]),
		})
	}
	return out, nil
}

func (c *Corteza) RefreshRiskScores(ctx context.Context) (int, error) {
	recs, err := c.ListRecords(ctx, "risks", "")
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range recs {
		m := recMap(r)
		score := RiskScore(m["probability"], m["impact"])
		if parseFloat(m["score"]) == score {
			continue
		}
		if err := c.UpdateValues(ctx, "risks", r.ID, map[string]string{"score": fmtNum(score)}); err != nil {
			continue
		}
		n++
	}
	return n, nil
}

func (c *Corteza) UserEmail(ctx context.Context, userID string) (string, error) {
	id := ParseID(userID)
	if id == 0 {
		return "", nil
	}
	if err := c.Discover(ctx); err != nil {
		return "", err
	}
	origin := strings.TrimSuffix(c.BaseURL(), "/compose")
	origin = strings.TrimSuffix(origin, "/api")
	origin = strings.TrimRight(origin, "/")
	for _, path := range []string{
		fmt.Sprintf("/system/users/%d", id),
		fmt.Sprintf("/api/system/users/%d", id),
	} {
		raw, err := c.requestAbs(ctx, "GET", origin+path, nil)
		if err != nil {
			continue
		}
		var u struct {
			Email string `json:"email"`
		}
		if json.Unmarshal(raw, &u) == nil && strings.Contains(u.Email, "@") {
			return u.Email, nil
		}
	}
	return "", nil
}

func (c *Corteza) requestAbs(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	var b io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 || looksLikeHTML(raw) {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, apiErrorBody(raw))
	}
	if msg := cortezaJSONError(raw); msg != "" {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, msg)
	}
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Response) > 0 {
		return envelope.Response, nil
	}
	return raw, nil
}
