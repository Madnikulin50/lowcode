package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Client is a Compose REST client with origin discovery
// (`/` vs `/api` — same problem backup/cmdb hit independently).
type Client struct {
	base        *origin
	token       string
	httpClient  *http.Client
	namespaceID uint64
	slug        string
	modules     map[string]uint64
}

type origin struct {
	mu       sync.Mutex
	url      string
	resolved bool
}

func NewClient(baseURL, token string, namespaceID uint64) *Client {
	return &Client{
		base:        &origin{url: strings.TrimRight(strings.TrimSpace(baseURL), "/")},
		token:       token,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		namespaceID: namespaceID,
		modules:     map[string]uint64{},
	}
}

func (c *Client) WithSlug(slug string) *Client {
	if c == nil {
		return nil
	}
	c.slug = slug
	return c
}

func (c *Client) BaseURL() string {
	if c == nil || c.base == nil {
		return ""
	}
	c.base.mu.Lock()
	defer c.base.mu.Unlock()
	return c.base.url
}

func (c *Client) HasToken() bool {
	return c != nil && strings.TrimSpace(c.token) != ""
}

func (c *Client) WithToken(token string) *Client {
	if strings.TrimSpace(token) == "" {
		return c
	}
	cp := *c
	cp.token = token
	cp.modules = map[string]uint64{}
	return &cp
}

func (c *Client) WithNamespace(id uint64) *Client {
	if id == 0 {
		return c
	}
	cp := *c
	cp.namespaceID = id
	cp.modules = map[string]uint64{}
	return &cp
}

func composeOriginCandidates(raw string) []string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		raw = "http://127.0.0.1:3333"
	}
	origin := strings.TrimSuffix(raw, "/compose")
	origin = strings.TrimSuffix(origin, "/api")
	origin = strings.TrimRight(origin, "/")
	seen := map[string]struct{}{}
	var out []string
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
	if looksLikeHTML(raw) || resp.StatusCode == http.StatusNotFound {
		return false
	}
	return true
}

func (c *Client) Discover(ctx context.Context) error {
	if c == nil || c.base == nil {
		return fmt.Errorf("compose client is nil")
	}
	c.base.mu.Lock()
	defer c.base.mu.Unlock()
	if c.base.resolved {
		return nil
	}
	given := c.base.url
	for _, origin := range composeOriginCandidates(given) {
		if probeComposeOrigin(ctx, origin, c.token) {
			if origin != given {
				log.Printf("sdk compose origin %s (from %s)", origin, given)
			}
			c.base.url = origin
			c.base.resolved = true
			return nil
		}
	}
	return fmt.Errorf("no Compose API at %s (tried %s)", given, strings.Join(composeOriginCandidates(given), ", "))
}

type composeModule struct {
	ID     uint64 `json:"moduleID,string"`
	Handle string `json:"handle"`
}

type composeNamespace struct {
	ID   uint64 `json:"namespaceID,string"`
	Slug string `json:"slug"`
}

type Record struct {
	ID        uint64        `json:"recordID,string"`
	UpdatedAt string        `json:"updatedAt,omitempty"`
	Values    []RecordValue `json:"values"`
}

type RecordValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func RecordMap(r Record) map[string]string {
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

func (c *Client) Request(ctx context.Context, method, path string, body any) ([]byte, error) {
	if err := c.Discover(ctx); err != nil {
		return nil, err
	}
	var rdr io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL()+path, rdr)
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
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(raw))
		if looksLikeHTML(raw) {
			msg = "HTML 404 (Compose is not under this prefix; use --api=http://127.0.0.1:3333 not .../api)"
		}
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

func (c *Client) ResolveNamespace(ctx context.Context) (uint64, error) {
	if c.namespaceID != 0 {
		return c.namespaceID, nil
	}
	slug := c.slug
	q := "/compose/namespace/?limit=50"
	if slug != "" {
		q = "/compose/namespace/?slug=" + url.QueryEscape(slug) + "&limit=50"
	}
	raw, err := c.Request(ctx, "GET", q, nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeNamespace](raw)
	if err != nil {
		return 0, err
	}
	for _, ns := range set {
		if slug == "" || strings.EqualFold(ns.Slug, slug) {
			c.namespaceID = ns.ID
			return ns.ID, nil
		}
	}
	return 0, fmt.Errorf("namespace slug %q not found", slug)
}

func (c *Client) ModuleByHandle(ctx context.Context, handle string) (uint64, error) {
	if id, ok := c.modules[handle]; ok {
		return id, nil
	}
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	raw, err := c.Request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/?handle=%s&limit=50", nsID, url.QueryEscape(handle)), nil)
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

func (c *Client) ListRecords(ctx context.Context, handle, query string) ([]Record, error) {
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
	raw, err := c.Request(ctx, "GET", q, nil)
	if err != nil {
		return nil, err
	}
	return unwrapSet[Record](raw)
}

func (c *Client) GetRecord(ctx context.Context, handle string, recordID uint64) (*Record, error) {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return nil, err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return nil, err
	}
	raw, err := c.Request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID), nil)
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (c *Client) CreateValues(ctx context.Context, handle string, values map[string]string) (uint64, error) {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return 0, err
	}
	payload := make([]RecordValue, 0, len(values))
	for k, v := range values {
		payload = append(payload, RecordValue{Name: k, Value: v})
	}
	raw, err := c.Request(ctx, "POST", fmt.Sprintf("/compose/namespace/%d/module/%d/record/", nsID, modID), map[string]any{"values": payload})
	if err != nil {
		return 0, err
	}
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		return 0, err
	}
	return rec.ID, nil
}

func (c *Client) UpdateValues(ctx context.Context, handle string, recordID uint64, values map[string]string) error {
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
	payload := make([]RecordValue, 0, len(values))
	for k, v := range values {
		payload = append(payload, RecordValue{Name: k, Value: v})
	}
	body := map[string]any{"values": payload}
	if rec.UpdatedAt != "" {
		body["updatedAt"] = rec.UpdatedAt
	}
	_, err = c.Request(ctx, "POST", fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID), body)
	return err
}

func (c *Client) DeleteRecord(ctx context.Context, handle string, recordID uint64) error {
	nsID, err := c.ResolveNamespace(ctx)
	if err != nil {
		return err
	}
	modID, err := c.ModuleByHandle(ctx, handle)
	if err != nil {
		return err
	}
	_, err = c.Request(ctx, "DELETE", fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID), nil)
	return err
}

// Heartbeat upserts a row in the given module (backup `agents` pattern).
func (c *Client) Heartbeat(ctx context.Context, module string, ident Identity) error {
	if c == nil || !c.HasToken() || module == "" {
		return nil
	}
	set, err := c.ListRecords(ctx, module, "")
	if err != nil {
		return err
	}
	vals := map[string]string{
		"name":         ident.Name,
		"url":          ident.PublicURL,
		"hostname":     ident.Hostname,
		"capabilities": strings.Join(ident.Capabilities, ","),
		"last_seen":    time.Now().UTC().Format(time.RFC3339),
		"status":       "online",
	}
	for _, rec := range set {
		m := RecordMap(rec)
		if m["hostname"] == ident.Hostname || m["url"] == ident.PublicURL {
			return c.UpdateValues(ctx, module, rec.ID, vals)
		}
	}
	_, err = c.CreateValues(ctx, module, vals)
	return err
}
