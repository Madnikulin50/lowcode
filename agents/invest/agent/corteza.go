package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Corteza struct {
	baseURL     string
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
		baseURL:     strings.TrimRight(baseURL, "/"),
		token:       token,
		httpClient:  &http.Client{Timeout: timeout},
		namespaceID: namespaceID,
		modules:     map[string]uint64{},
	}
}

func (c *Corteza) WithToken(token string) *Corteza {
	if token == "" {
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
	var b io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		b = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(raw))
	}
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Response) > 0 {
		return envelope.Response, nil
	}
	return raw, nil
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
	payload := make([]composeRecordValue, 0, len(values))
	for k, v := range values {
		payload = append(payload, composeRecordValue{Name: k, Value: v})
	}
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
			ID:      r.ID,
			Project: m["project"],
			Title:   m["title"],
			Status:  m["status"],
			DueDate: parseTime(m["due_date"]),
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
	id, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil || id == 0 {
		return nil
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
	_, err = c.CreateValues(ctx, "risks", map[string]string{
		"title":       a.Title,
		"project":     a.Project,
		"wbs":         a.WBS,
		"probability": "high",
		"impact":      "high",
		"status":      "open",
		"description": a.Detail,
	})
	return err
}
