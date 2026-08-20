package rulesgo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type upsertConfig struct {
	ModuleID        flexibleID             `json:"moduleID"`
	ModuleHandle    string                 `json:"moduleHandle,omitempty"`
	NamespaceID     flexibleID             `json:"namespaceID"`
	MatchBy         []string               `json:"matchBy"`
	MatchAll        bool                   `json:"matchAll,omitempty"`
	Fields          map[string]interface{} `json:"fields,omitempty"`
	OmitEmpty       bool                   `json:"omitEmpty,omitempty"`
	ContinueOnError bool                   `json:"continueOnError,omitempty"`
	ResultVar       string                 `json:"resultVar,omitempty"`
}

type upsertExecutor struct {
	svc CRUDService
}

func (n *upsertExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	cfg, err := ParseNodeConfig[upsertConfig](node.Config)
	if err != nil {
		return nil, err
	}
	if n.svc == nil {
		return map[string]interface{}{"status": "crud_service_not_configured"}, nil
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
				return n.upsertErr(cfg, fmt.Errorf("module %q: %w", handle, err))
			}
			modID = id
		}
	}
	if nsID == 0 || modID == 0 {
		return n.upsertErr(cfg, fmt.Errorf("namespaceID and moduleID are required"))
	}

	fields := resolveFieldTemplates(cfg.Fields, ec)
	if cfg.OmitEmpty {
		fields = dropEmptyFields(fields)
	}
	matchBy := cfg.MatchBy
	if len(matchBy) == 0 {
		matchBy = []string{"id"}
	}
	if cfg.MatchAll {
		if !hasAllUpsertIdentity(matchBy, fields) {
			return map[string]interface{}{"skipped": true, "reason": "no identity fields"}, nil
		}
	} else if !hasUpsertIdentity(matchBy, fields) {
		return map[string]interface{}{"skipped": true, "reason": "no identity fields"}, nil
	}

	existingID, err := n.findExisting(ctx, nsID, modID, matchBy, fields, cfg.MatchAll)
	if err != nil {
		return n.upsertErr(cfg, err)
	}
	if existingID != "" {
		updatedAt, err := n.svc.Update(ctx, nsID, modID, existingID, fields)
		if err != nil {
			return n.upsertErr(cfg, fmt.Errorf("upsert update: %w", err))
		}
		n.setResultID(ec, cfg, existingID)
		return map[string]interface{}{"recordID": existingID, "created": false, "updatedAt": updatedAt}, nil
	}
	id, createdAt, err := n.svc.Create(ctx, nsID, modID, fields)
	if err != nil {
		return n.upsertErr(cfg, fmt.Errorf("upsert create: %w", err))
	}
	n.setResultID(ec, cfg, id)
	return map[string]interface{}{"recordID": id, "created": true, "createdAt": createdAt}, nil
}

func (n *upsertExecutor) upsertErr(cfg upsertConfig, err error) (map[string]interface{}, error) {
	if cfg.ContinueOnError {
		return map[string]interface{}{"skipped": true, "error": err.Error()}, nil
	}
	return nil, err
}

func (n *upsertExecutor) setResultID(ec *ExecutionContext, cfg upsertConfig, id string) {
	ec.Set("upsertRecordID", id)
	if key := strings.TrimSpace(cfg.ResultVar); key != "" {
		ec.Set(key, id)
	}
}

func (n *upsertExecutor) findExisting(ctx context.Context, nsID, modID uint64, matchBy []string, fields map[string]interface{}, matchAll bool) (string, error) {
	if matchAll {
		return n.findExistingAll(ctx, nsID, modID, matchBy, fields)
	}
	for _, field := range matchBy {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, ok := identityValue(field, fields[field])
		if !ok {
			continue
		}
		q := field + " = '" + strings.ReplaceAll(value, "'", "\\'") + "'"
		set, err := n.svc.Search(ctx, nsID, modID, q, 5)
		if err != nil {
			return "", err
		}
		for _, rec := range set {
			if fieldValuesEqual(field, rec[field], value) {
				if id := recordIDFromMap(rec); id != "" {
					return id, nil
				}
			}
		}
	}
	return "", nil
}

func (n *upsertExecutor) findExistingAll(ctx context.Context, nsID, modID uint64, matchBy []string, fields map[string]interface{}) (string, error) {
	want := map[string]string{}
	var primaryField, primaryValue string
	for _, field := range matchBy {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		value, ok := identityValue(field, fields[field])
		if !ok {
			continue
		}
		want[field] = value
		if primaryField == "" {
			primaryField, primaryValue = field, value
		}
	}
	if primaryField == "" {
		return "", nil
	}
	q := primaryField + " = '" + strings.ReplaceAll(primaryValue, "'", "\\'") + "'"
	set, err := n.svc.Search(ctx, nsID, modID, q, 200)
	if err != nil {
		return "", err
	}
	for _, rec := range set {
		matched := true
		for field, value := range want {
			if !fieldValuesEqual(field, rec[field], value) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		if id := recordIDFromMap(rec); id != "" {
			return id, nil
		}
	}
	return "", nil
}

func recordIDFromMap(rec map[string]interface{}) string {
	for _, k := range []string{"recordID", "RecordID", "id", "ID"} {
		if v, ok := rec[k]; ok && v != nil {
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" && s != "0" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func genericHostname(value string) bool {
	s := strings.ToLower(strings.TrimSpace(value))
	switch s {
	case "", "_gateway", "gateway", "localhost", "unknown", "none", "null", "router":
		return true
	}
	return false
}

func hasUpsertIdentity(matchBy []string, fields map[string]interface{}) bool {
	for _, field := range matchBy {
		if _, ok := identityValue(field, fields[strings.TrimSpace(field)]); ok {
			return true
		}
	}
	return false
}

func hasAllUpsertIdentity(matchBy []string, fields map[string]interface{}) bool {
	n := 0
	for _, field := range matchBy {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if _, ok := identityValue(field, fields[field]); !ok {
			return false
		}
		n++
	}
	return n > 0
}

func identityValue(field string, raw interface{}) (string, bool) {
	value := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if value == "" || value == "<nil>" || leftoverTemplate(value) {
		return "", false
	}
	if (field == "hostname" || strings.HasSuffix(field, "hostname")) && genericHostname(value) {
		return "", false
	}
	if strings.Contains(field, "mac") {
		value = normalizeMatchMAC(value)
	}
	return value, true
}

func fieldValuesEqual(field string, gotRaw interface{}, want string) bool {
	got := strings.TrimSpace(fmt.Sprintf("%v", gotRaw))
	if strings.Contains(field, "mac") {
		got = normalizeMatchMAC(got)
	}
	if strings.EqualFold(got, want) {
		return true
	}
	gf, gErr := strconv.ParseFloat(got, 64)
	wf, wErr := strconv.ParseFloat(want, 64)
	return gErr == nil && wErr == nil && gf == wf
}

func leftoverTemplate(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}")
}

func emptyAny(v interface{}) bool {
	if v == nil {
		return true
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	return s == "" || s == "<nil>" || leftoverTemplate(s)
}

func normalizeMatchMAC(mac string) string {
	s := strings.ToLower(strings.TrimSpace(mac))
	s = strings.ReplaceAll(s, "-", ":")
	s = strings.ReplaceAll(s, ".", ":")
	parts := strings.Split(s, ":")
	if len(parts) != 6 {
		return s
	}
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = "0" + p
		}
	}
	return strings.Join(parts, ":")
}
