package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/madnikulin50/lowcode/server/store"
)

func ensureCRUDAuth(ctx context.Context) context.Context {
	if ident := auth.GetIdentityFromContext(ctx); ident != nil && ident.Valid() {
		return ctx
	}
	return auth.SetIdentityToContext(ctx, auth.Authenticated(1))
}

type composeCRUD struct{}

func (composeCRUD) LookupModule(ctx context.Context, namespaceID uint64, handle string) (uint64, error) {
	if service.DefaultStore == nil {
		return 0, fmt.Errorf("store not initialized")
	}
	if namespaceID == 0 {
		return 0, fmt.Errorf("namespaceID is required")
	}
	m, err := store.LookupComposeModuleByNamespaceIDHandle(ctx, service.DefaultStore, namespaceID, handle)
	if err != nil {
		return 0, err
	}
	if m == nil {
		return 0, fmt.Errorf("not found")
	}
	return m.ID, nil
}

func (composeCRUD) Create(ctx context.Context, namespaceID, moduleID uint64, values map[string]interface{}) (recordID string, createdAt string, err error) {
	ctx = ensureCRUDAuth(ctx)
	if service.DefaultRecord == nil {
		return "", "", fmt.Errorf("record service not initialized")
	}
	r := &types.Record{NamespaceID: namespaceID, ModuleID: moduleID, Values: mapToRecordValues(values)}
	created, dd, err := service.DefaultRecord.Create(ctx, r)
	if err != nil {
		return "", "", err
	}
	if !dd.IsValid() {
		return "", "", dd
	}
	if created == nil {
		return "", "", fmt.Errorf("create returned no record")
	}
	return fmt.Sprintf("%d", created.ID), created.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"), nil
}

func (composeCRUD) Update(ctx context.Context, namespaceID, moduleID uint64, recordID string, values map[string]interface{}) (updatedAt string, err error) {
	ctx = ensureCRUDAuth(ctx)
	if service.DefaultRecord == nil {
		return "", fmt.Errorf("record service not initialized")
	}
	var rid uint64
	fmt.Sscanf(recordID, "%d", &rid)
	if rid == 0 {
		return "", fmt.Errorf("invalid recordID")
	}
	existing, _, err := service.DefaultRecord.FindByID(ctx, namespaceID, moduleID, rid)
	if err != nil {
		return "", err
	}
	for _, rv := range mapToRecordValues(values) {
		found := false
		for _, cur := range existing.Values {
			if cur.Name == rv.Name {
				cur.Value = rv.Value
				cur.Updated = true
				found = true
				break
			}
		}
		if !found {
			existing.Values = append(existing.Values, rv)
		}
	}
	updated, dd, err := service.DefaultRecord.Update(ctx, existing)
	if err != nil {
		return "", err
	}
	if !dd.IsValid() {
		return "", dd
	}
	if updated.UpdatedAt != nil {
		return updated.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"), nil
	}
	return "", nil
}

func (composeCRUD) Delete(ctx context.Context, namespaceID, moduleID uint64, recordID string) error {
	ctx = ensureCRUDAuth(ctx)
	if service.DefaultRecord == nil {
		return fmt.Errorf("record service not initialized")
	}
	var rid uint64
	fmt.Sscanf(recordID, "%d", &rid)
	if rid == 0 {
		return fmt.Errorf("invalid recordID")
	}
	return service.DefaultRecord.DeleteByID(ctx, namespaceID, moduleID, rid)
}

func (composeCRUD) Search(ctx context.Context, namespaceID, moduleID uint64, query string, limit int) ([]map[string]interface{}, error) {
	ctx = ensureCRUDAuth(ctx)
	if service.DefaultRecord == nil {
		return nil, fmt.Errorf("record service not initialized")
	}
	ff := types.RecordFilter{NamespaceID: namespaceID, ModuleID: moduleID, Query: query}
	if limit > 0 {
		ff.Limit = uint(limit)
	}
	set, _, err := service.DefaultRecord.Find(ctx, ff)
	if err != nil {
		return nil, err
	}
	return recordSetToMap(set), nil
}

func mapToRecordValues(values map[string]interface{}) types.RecordValueSet {
	out := make(types.RecordValueSet, 0, len(values))
	for name, val := range values {
		if name == "" {
			continue
		}
		out = append(out, &types.RecordValue{Name: name, Value: stringifyCRUDValue(val)})
	}
	return out
}

func stringifyCRUDValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case bool:
		if t {
			return "1"
		}
		return "0"
	case string:
		return t
	case []interface{}, []string, map[string]interface{}:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", t)
	}
}

var _ rulesgo.CRUDService = composeCRUD{}
