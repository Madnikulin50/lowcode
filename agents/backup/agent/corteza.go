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

func NewCorteza(baseURL, token string, namespaceID uint64) *Corteza {
	return &Corteza{
		base:        &cortezaBase{url: strings.TrimRight(strings.TrimSpace(baseURL), "/")},
		token:       token,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
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

func (c *Corteza) HasToken() bool {
	return c != nil && strings.TrimSpace(c.token) != ""
}

func composeAPIError(raw []byte) string {
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &envelope) != nil || len(envelope.Error) == 0 {
		return ""
	}
	var obj map[string]interface{}
	if json.Unmarshal(envelope.Error, &obj) == nil {
		if m, _ := obj["message"].(string); strings.TrimSpace(m) != "" {
			return strings.TrimSpace(m)
		}
	}
	var s string
	if json.Unmarshal(envelope.Error, &s) == nil {
		return strings.TrimSpace(s)
	}
	return ""
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
	add(origin)          // http://host:3333 + /compose/...  (dev)
	add(origin + "/api") // http://host:3333/api + /compose/... (webapps on)
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
	if msg := composeAPIError(raw); msg != "" {
		code := resp.StatusCode
		if code < 400 {
			code = http.StatusForbidden
		}
		return nil, fmt.Errorf("API error %d: %s", code, msg)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, apiErrorBody(raw))
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

func (c *Corteza) ResolveNamespace(ctx context.Context) (uint64, error) {
	if c.namespaceID != 0 {
		return c.namespaceID, nil
	}
	raw, err := c.request(ctx, "GET", "/compose/namespace/?slug=backup&limit=50", nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeNamespace](raw)
	if err != nil {
		return 0, err
	}
	for _, ns := range set {
		if strings.EqualFold(ns.Slug, "backup") {
			c.namespaceID = ns.ID
			return ns.ID, nil
		}
	}
	return 0, fmt.Errorf("namespace slug backup not found; run agents/backup/compose/apply.mjs")
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
	_, err = c.request(ctx, "POST", fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID), body)
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

func (c *Corteza) LoadSource(ctx context.Context, id uint64) (*Source, error) {
	rec, err := c.GetRecord(ctx, "sources", id)
	if err != nil {
		return nil, err
	}
	return sourceFrom(rec), nil
}

func (c *Corteza) LoadPolicy(ctx context.Context, id uint64) (*Policy, error) {
	rec, err := c.GetRecord(ctx, "policies", id)
	if err != nil {
		return nil, err
	}
	return policyFrom(rec), nil
}

func (c *Corteza) LoadCredential(ctx context.Context, id uint64) (*Credential, error) {
	if id == 0 {
		return &Credential{}, nil
	}
	rec, err := c.GetRecord(ctx, "credentials", id)
	if err != nil {
		return nil, err
	}
	m := recMap(*rec)
	return &Credential{
		ID:       rec.ID,
		Name:     m["name"],
		Handle:   m["handle"],
		Kind:     m["kind"],
		Username: m["username"],
		Extra:    m["extra"],
	}, nil
}

func (c *Corteza) LoadSnapshot(ctx context.Context, id uint64) (*Snapshot, error) {
	rec, err := c.GetRecord(ctx, "snapshots", id)
	if err != nil {
		return nil, err
	}
	m := recMap(*rec)
	exp, _ := time.Parse(time.RFC3339, m["expires_at"])
	return &Snapshot{
		ID:         rec.ID,
		JobID:      ParseUint(m["job"]),
		SourceID:   ParseUint(m["source"]),
		PolicyID:   ParseUint(m["policy"]),
		S3Bucket:   m["s3_bucket"],
		S3Key:      m["s3_key"],
		SizeBytes:  int64(ParseUint(m["size_bytes"])),
		Checksum:   m["checksum"],
		Files:      int(ParseUint(m["files_count"])),
		Kind:       m["kind"],
		Engine:     m["engine"],
		ResticID:   m["restic_id"],
		ExpiresAt:  exp,
		Restorable: ParseBool(m["restorable"]),
	}, nil
}

func (c *Corteza) ListEnabledPolicies(ctx context.Context) ([]Policy, error) {
	set, err := c.ListRecords(ctx, "policies", "")
	if err != nil {
		return nil, err
	}
	var out []Policy
	for i := range set {
		p := policyFrom(&set[i])
		if p.Enabled {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (c *Corteza) FindPolicyForSource(ctx context.Context, sourceID uint64) (*Policy, error) {
	set, err := c.ListRecords(ctx, "policies", fmt.Sprintf("source = %d", sourceID))
	if err != nil {
		return nil, err
	}
	for i := range set {
		p := policyFrom(&set[i])
		if p.Enabled {
			return p, nil
		}
	}
	if len(set) > 0 {
		return policyFrom(&set[0]), nil
	}
	return nil, fmt.Errorf("no policy for source %d", sourceID)
}

func sourceFrom(rec *composeRecord) *Source {
	m := recMap(*rec)
	return &Source{
		ID:       rec.ID,
		Name:     m["name"],
		Type:     SourceType(m["type"]),
		AgentID:  ParseUint(m["agent"]),
		CredID:   ParseUint(m["credential"]),
		Path:     m["path"],
		Host:     m["host"],
		Share:    m["share"],
		SMBPath:  m["smb_path"],
		DBEngine: m["db_engine"],
		DBName:   m["db_name"],
		DBPort:   ParseInt(m["db_port"]),
		S3Bucket: m["s3_bucket"],
		S3Prefix: m["s3_prefix"],
		S3Region: m["s3_region"],
		S3Secure: ParseBool(m["s3_secure"]),
		Enabled:  ParseBool(m["enabled"]),
		Notes:    m["notes"],
	}
}

func policyFrom(rec *composeRecord) *Policy {
	m := recMap(*rec)
	last, _ := time.Parse(time.RFC3339, m["last_run"])
	ret := ParseInt(m["retention_days"])
	return &Policy{
		ID:            rec.ID,
		Name:          m["name"],
		SourceID:      ParseUint(m["source"]),
		Cron:          firstNonEmpty(m["cron"], "0 2 * * *"),
		RetentionDays: ret,
		Incremental:   ParseBool(m["incremental"]),
		DestPrefix:    m["dest_prefix"],
		Enabled:       m["enabled"] == "" || ParseBool(m["enabled"]),
		LastRun:       last,
	}
}

func (c *Corteza) UpsertAgent(ctx context.Context, name, publicURL, hostname, caps string) error {
	set, err := c.ListRecords(ctx, "agents", "")
	if err != nil {
		return err
	}
	vals := map[string]string{
		"name":         name,
		"url":          publicURL,
		"hostname":     hostname,
		"capabilities": caps,
		"last_seen":    time.Now().UTC().Format(time.RFC3339),
		"status":       "online",
	}
	for _, rec := range set {
		m := recMap(rec)
		if m["hostname"] == hostname || m["url"] == publicURL {
			return c.UpdateValues(ctx, "agents", rec.ID, vals)
		}
	}
	_, err = c.CreateValues(ctx, "agents", vals)
	return err
}

func fmtInt(n int64) string   { return strconv.FormatInt(n, 10) }
func fmtUint(n uint64) string { return strconv.FormatUint(n, 10) }
