// 1C:Enterprise integration node (1c.sync).
//
// 1C's HTTP/OData service (the standard "odata/standard.odata" endpoint
// published by any 1C:Enterprise base) is plain HTTP + JSON + Basic Auth -
// there's no dedicated client library to wrap, unlike Kafka/RabbitMQ. So
// this node talks to it directly with net/http:
//
//  1. GET  <url>/<entity>?$filter=...   - look up an existing record by
//     business key (matchBy), not by 1C's internal Ref_Key GUID
//  2. PATCH <url>/<entity>(guid'<Ref_Key>')  - update it, if found
//  3. POST <url>/<entity>                    - create it, if not found
//
// This is the flagship integration case call out in the platform review:
// two-way sync of nomenclature/counterparties/stock between 1C and other
// systems ("Стройконтроль" and friends), matching a business key instead of
// 1C's internal GUID.
package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/vault"
)

// OneCConfig is the config shape for the 1c.sync node.
type OneCConfig struct {
	URL      string `json:"url"`    // OData service root, e.g. http://host/base/odata/standard.odata
	Entity   string `json:"entity"` // e.g. Catalog_Номенклатура, Catalog_Контрагенты
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// PasswordSecretRef points Password at a Vault secret instead of storing
	// the 1C service account password in plaintext - see vault.ResolveField.
	PasswordSecretRef string `json:"passwordSecretRef,omitempty"`

	// MatchBy: 1C field name -> template value, used to build the $filter
	// that looks up an existing record by business key (e.g. "Код" or
	// "Артикул"), not by 1C's internal Ref_Key GUID.
	MatchBy map[string]string `json:"matchBy,omitempty"`
	// Fields: 1C field name -> template value, written on both create and
	// update (MatchBy fields are also sent on create, so the new record is
	// identifiable the same way on the next run).
	Fields map[string]string `json:"fields,omitempty"`

	TimeoutSeconds int `json:"timeoutSeconds,omitempty"`
}

type oneCSyncExecutor struct{}

func (n *oneCSyncExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[OneCConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.URL = resolveTemplateValue(cfg.URL, ec)
	cfg.Entity = resolveTemplateValue(cfg.Entity, ec)
	cfg.Username = resolveTemplateValue(cfg.Username, ec)

	passwordRef := resolveTemplateValue(cfg.PasswordSecretRef, ec)
	password, err := vault.ResolveField(ctx, passwordRef, resolveTemplateValue(cfg.Password, ec))
	if err != nil {
		return nil, fmt.Errorf("1c.sync: resolve password: %w", err)
	}

	if cfg.URL == "" {
		return nil, fmt.Errorf("1c.sync: url is required")
	}
	if cfg.Entity == "" {
		return nil, fmt.Errorf("1c.sync: entity is required")
	}
	if len(cfg.MatchBy) == 0 {
		return nil, fmt.Errorf("1c.sync: matchBy is required")
	}

	matchBy := make(map[string]string, len(cfg.MatchBy))
	for k, v := range cfg.MatchBy {
		matchBy[k] = resolveTemplateValue(v, ec)
	}
	fields := make(map[string]interface{}, len(cfg.Fields))
	for k, v := range cfg.Fields {
		fields[k] = resolveTemplateValue(v, ec)
	}

	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 20
	}
	reqCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	base := strings.TrimRight(cfg.URL, "/") + "/" + strings.TrimLeft(cfg.Entity, "/")
	filterExpr := buildODataFilter(matchBy)

	found, err := n.find(reqCtx, base, filterExpr, cfg.Username, password)
	if err != nil {
		return nil, fmt.Errorf("1c.sync: lookup: %w", err)
	}

	if found != nil {
		refKey, _ := found["Ref_Key"].(string)
		if refKey == "" {
			return nil, fmt.Errorf("1c.sync: matched record has no Ref_Key")
		}
		updated, err := n.patch(reqCtx, base, refKey, fields, cfg.Username, password)
		if err != nil {
			return nil, fmt.Errorf("1c.sync: update: %w", err)
		}
		return map[string]interface{}{
			"success": true,
			"action":  "updated",
			"entity":  cfg.Entity,
			"ref":     refKey,
			"record":  updated,
		}, nil
	}

	createFields := make(map[string]interface{}, len(fields)+len(matchBy))
	for k, v := range matchBy {
		createFields[k] = v
	}
	for k, v := range fields {
		createFields[k] = v
	}
	created, err := n.create(reqCtx, base, createFields, cfg.Username, password)
	if err != nil {
		return nil, fmt.Errorf("1c.sync: create: %w", err)
	}
	refKey, _ := created["Ref_Key"].(string)

	return map[string]interface{}{
		"success": true,
		"action":  "created",
		"entity":  cfg.Entity,
		"ref":     refKey,
		"record":  created,
	}, nil
}

func (n *oneCSyncExecutor) find(ctx context.Context, base, filterExpr, user, pass string) (map[string]interface{}, error) {
	q := url.Values{}
	q.Set("$format", "json")
	q.Set("$top", "1")
	if filterExpr != "" {
		q.Set("$filter", filterExpr)
	}

	body, status, err := doODataRequest(ctx, http.MethodGet, base+"?"+q.Encode(), nil, user, pass)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, strings.TrimSpace(string(body)))
	}

	items, err := extractODataItems(body)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return items[0], nil
}

func (n *oneCSyncExecutor) patch(ctx context.Context, base, refKey string, fields map[string]interface{}, user, pass string) (map[string]interface{}, error) {
	reqURL := fmt.Sprintf("%s(guid'%s')?$format=json", base, refKey)
	payload, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal fields: %w", err)
	}

	body, status, err := doODataRequest(ctx, http.MethodPatch, reqURL, payload, user, pass)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, strings.TrimSpace(string(body)))
	}

	var record map[string]interface{}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &record) // 1C commonly returns 204 No Content on PATCH
	}
	if record == nil {
		record = map[string]interface{}{"Ref_Key": refKey}
	}
	return record, nil
}

func (n *oneCSyncExecutor) create(ctx context.Context, base string, fields map[string]interface{}, user, pass string) (map[string]interface{}, error) {
	payload, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal fields: %w", err)
	}

	body, status, err := doODataRequest(ctx, http.MethodPost, base+"?$format=json", payload, user, pass)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, strings.TrimSpace(string(body)))
	}

	var record map[string]interface{}
	if err := json.Unmarshal(body, &record); err != nil {
		return nil, fmt.Errorf("decode created record: %w", err)
	}
	return record, nil
}

// doODataRequest is shared by the 1c.sync node and the 1c-odata DAL
// connector (compose/service/connector_1c.go can't import rulesgo, so it has
// its own tiny copy - kept intentionally simple to avoid a shared internal
// package for two ~20-line call sites).
func doODataRequest(ctx context.Context, method, reqURL string, body []byte, user, pass string) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if user != "" {
		req.SetBasicAuth(user, pass)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	return raw, resp.StatusCode, nil
}

// buildODataFilter builds a "field eq 'value' and field2 eq 'value2'" OData
// $filter expression from a business-key match map. Keys are sorted for a
// deterministic, testable filter string.
func buildODataFilter(matchBy map[string]string) string {
	if len(matchBy) == 0 {
		return ""
	}
	keys := make([]string, 0, len(matchBy))
	for k := range matchBy {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s eq '%s'", k, odataEscapeString(matchBy[k])))
	}
	return strings.Join(parts, " and ")
}

// odataEscapeString escapes a string for use inside an OData string literal
// ('...'): the only special character is the single quote, doubled.
func odataEscapeString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// extractODataItems reads the "value" array 1C's modern OData JSON responses
// use, falling back to the older "d.results" shape for older 1C versions.
func extractODataItems(raw []byte) ([]map[string]interface{}, error) {
	var probe struct {
		Value []map[string]interface{} `json:"value"`
		D     *struct {
			Results []map[string]interface{} `json:"results"`
		} `json:"d"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("decode odata response: %w", err)
	}
	if probe.Value != nil {
		return probe.Value, nil
	}
	if probe.D != nil {
		return probe.D.Results, nil
	}
	return nil, nil
}
