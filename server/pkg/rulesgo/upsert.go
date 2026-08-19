package rulesgo

import (
	"context"
	"fmt"
	"net"
	"strings"
)

type upsertConfig struct {
	ModuleID     flexibleID             `json:"moduleID"`
	ModuleHandle string                 `json:"moduleHandle,omitempty"`
	NamespaceID  flexibleID             `json:"namespaceID"`
	MatchBy      []string               `json:"matchBy"`
	Fields       map[string]interface{} `json:"fields,omitempty"`
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
				return nil, fmt.Errorf("module %q: %w", handle, err)
			}
			modID = id
		}
	}
	if nsID == 0 || modID == 0 {
		return nil, fmt.Errorf("namespaceID and moduleID are required")
	}

	fields := resolveFieldTemplates(cfg.Fields, ec)
	matchBy := cfg.MatchBy
	if len(matchBy) == 0 {
		matchBy = []string{"id"}
	}
	if !hasUpsertIdentity(matchBy, fields) {
		return map[string]interface{}{"skipped": true, "reason": "no identity fields"}, nil
	}

	existingID, err := n.findExisting(ctx, nsID, modID, matchBy, fields)
	if err != nil {
		return nil, err
	}
	if existingID != "" {
		updatedAt, err := n.svc.Update(ctx, nsID, modID, existingID, fields)
		if err != nil {
			return nil, fmt.Errorf("upsert update: %w", err)
		}
		ec.Set("upsertRecordID", existingID)
		return map[string]interface{}{"recordID": existingID, "created": false, "updatedAt": updatedAt}, nil
	}
	id, createdAt, err := n.svc.Create(ctx, nsID, modID, fields)
	if err != nil {
		return nil, fmt.Errorf("upsert create: %w", err)
	}
	ec.Set("upsertRecordID", id)
	return map[string]interface{}{"recordID": id, "created": true, "createdAt": createdAt}, nil
}

func (n *upsertExecutor) findExisting(ctx context.Context, nsID, modID uint64, matchBy []string, fields map[string]interface{}) (string, error) {
	for _, field := range matchBy {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		raw, _ := fields[field]
		value := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if value == "" || value == "<nil>" {
			continue
		}
		if (field == "hostname" || strings.HasSuffix(field, "_hostname")) && genericMatchValue(field, value) {
			continue
		}
		if strings.Contains(field, "mac") {
			value = normalizeMatchMAC(value)
		}
		q := field + " = '" + strings.ReplaceAll(value, "'", "\\'") + "'"
		set, err := n.svc.Search(ctx, nsID, modID, q, 5)
		if err != nil {
			return "", err
		}
		for _, rec := range set {
			got := strings.TrimSpace(fmt.Sprintf("%v", rec[field]))
			if strings.Contains(field, "mac") {
				got = normalizeMatchMAC(got)
			}
			if strings.EqualFold(got, value) {
				if id := recordIDFromMap(rec); id != "" {
					return id, nil
				}
			}
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

func genericMatchValue(field, value string) bool {
	if field != "hostname" && !strings.HasSuffix(field, "hostname") {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(value))
	switch s {
	case "", "_gateway", "gateway", "localhost", "unknown", "none", "null", "router":
		return true
	}
	return net.ParseIP(s) != nil
}

func hasUpsertIdentity(matchBy []string, fields map[string]interface{}) bool {
	for _, field := range matchBy {
		raw, _ := fields[strings.TrimSpace(field)]
		value := strings.TrimSpace(fmt.Sprintf("%v", raw))
		if value == "" || value == "<nil>" {
			continue
		}
		if (field == "hostname" || strings.HasSuffix(field, "hostname")) && genericMatchValue(field, value) {
			continue
		}
		return true
	}
	return false
}

func emptyAny(v interface{}) bool {
	if v == nil {
		return true
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	return s == "" || s == "<nil>"
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
