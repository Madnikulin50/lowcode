package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/madnikulin50/lowcode/server/pkg/filter"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initModuleRecords(ctx context.Context, s *server.MCPServer) {
	if !mcpTransportEnabled() {
		return
	}
	modules, _, err := service.DefaultModule.Find(ctx, types.ModuleFilter{})
	if err != nil || len(modules) == 0 {
		return
	}

	for _, m := range modules {
		recordID := m.Handle
		if len(recordID) == 0 {
			recordID = fmt.Sprintf("%v", m.ID)
		}
		recordID = strings.ToLower(recordID)

		s.AddTool(mcp.NewTool("module_"+recordID+"_records",
			mcp.WithDescription(fmt.Sprintf("List records in module '%s' (max %d)", m.Name, defaultMCPRecordLimit)),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		), makeListRecordsHandler(m.NamespaceID, m.ID))

		s.AddTool(mcp.NewTool("module_"+recordID+"_search",
			mcp.WithDescription(fmt.Sprintf("Search records in module '%s' by text query", m.Name)),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
			mcp.WithString("query", mcp.Description("Search text"), mcp.Required()),
		), makeSearchRecordsHandler(m.NamespaceID, m.ID))

		s.AddTool(mcp.NewTool("module_"+recordID+"_create_record",
			mcp.WithDescription(fmt.Sprintf("Create a record in module '%s'", m.Name)),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
			mcp.WithString("values",
				mcp.Description(fmt.Sprintf(`JSON object with field values. Module '%s' fields: %s`, m.Name, moduleFieldsHint(m))),
				mcp.Required(),
			),
		), makeCreateRecordHandler(m.NamespaceID, m.ID))

		s.AddTool(mcp.NewTool("module_"+recordID+"_update_record",
			mcp.WithDescription(fmt.Sprintf("Update a record by ID in module '%s'", m.Name)),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
			mcp.WithString("recordID", mcp.Description("Record ID"), mcp.Required()),
			mcp.WithString("values",
				mcp.Description("JSON object with field values to update"),
				mcp.Required(),
			),
		), makeUpdateRecordHandler(m.NamespaceID, m.ID))

		s.AddTool(mcp.NewTool("module_"+recordID+"_delete_record",
			mcp.WithDescription(fmt.Sprintf("Delete a record by ID from module '%s'", m.Name)),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
			mcp.WithString("recordID", mcp.Description("Record ID"), mcp.Required()),
		), makeDeleteRecordHandler(m.NamespaceID, m.ID))
	}
}

func moduleFieldsHint(m *types.Module) string {
	var parts []string
	for _, f := range m.Fields {
		req := ""
		if f.Required {
			req = " required"
		}
		parts = append(parts, fmt.Sprintf("%s (%s%s)", f.Name, f.Kind, req))
	}
	return strings.Join(parts, ", ")
}

func makeListRecordsHandler(nsID, modID uint64) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		set, _, err := service.DefaultRecord.Find(ctx, types.RecordFilter{
			ModuleID:    modID,
			NamespaceID: nsID,
			Paging:      filter.Paging{Limit: defaultMCPRecordLimit},
		})
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(set), nil
	}
}

func makeSearchRecordsHandler(nsID, modID uint64) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		args := argsMap(request)
		mod, err := service.DefaultModule.FindByID(ctx, nsID, modID)
		if err != nil || mod == nil {
			return errorResult(fmt.Errorf("module not found: %w", err)), nil
		}

		q := service.SanitizeRecordSearchQuery(getString(args, "query"))
		if q == "" {
			return textResult("Search query is empty after sanitization."), nil
		}
		ql := service.BuildRecordTextSearchQL(mod.Fields, q)
		if ql == "" {
			return jsonResult([]types.Record{}), nil
		}

		set, _, err := service.DefaultRecord.Find(ctx, types.RecordFilter{
			ModuleID:    modID,
			NamespaceID: nsID,
			Query:       ql,
			Paging:      filter.Paging{Limit: defaultMCPRecordLimit},
		})
		if err != nil {
			return errorResult(err), nil
		}
		return jsonResult(set), nil
	}
}

func makeCreateRecordHandler(nsID, modID uint64) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		args := argsMap(request)

		valuesJSON := getString(args, "values")
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(valuesJSON), &raw); err != nil {
			return textResult(fmt.Sprintf("Invalid values JSON: %v", err)), nil
		}

		values := make(types.RecordValueSet, 0, len(raw))
		for k, v := range raw {
			values = append(values, &types.RecordValue{
				Name:  k,
				Value: fmt.Sprintf("%v", v),
			})
		}

		rec := &types.Record{
			NamespaceID: nsID,
			ModuleID:    modID,
			Values:      values,
		}

		created, errs, err := service.DefaultRecord.Create(ctx, rec)
		if err != nil {
			return errorResult(err), nil
		}
		if errs != nil && len(errs.Set) > 0 {
			var msgs []string
			for _, e := range errs.Set {
				msgs = append(msgs, fmt.Sprintf("%s: %s", e.Kind, e.Message))
			}
			return textResult("Validation errors: " + strings.Join(msgs, "; ")), nil
		}
		return jsonResult(created), nil
	}
}

func makeUpdateRecordHandler(nsID, modID uint64) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		args := argsMap(request)

		recordID := parseUint64(args, "recordID")
		valuesJSON := getString(args, "values")

		existing, _, err := service.DefaultRecord.FindByID(ctx, nsID, modID, recordID)
		if err != nil {
			return errorResult(fmt.Errorf("record not found: %w", err)), nil
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(valuesJSON), &raw); err != nil {
			return textResult(fmt.Sprintf("Invalid values JSON: %v", err)), nil
		}

		values := make(types.RecordValueSet, 0, len(raw))
		for k, v := range raw {
			values = append(values, &types.RecordValue{
				Name:  k,
				Value: fmt.Sprintf("%v", v),
			})
		}

		existing.Values = values
		updated, errs, err := service.DefaultRecord.Update(ctx, existing)
		if err != nil {
			return errorResult(err), nil
		}
		if errs != nil && len(errs.Set) > 0 {
			var msgs []string
			for _, e := range errs.Set {
				msgs = append(msgs, fmt.Sprintf("%s: %s", e.Kind, e.Message))
			}
			return textResult("Validation errors: " + strings.Join(msgs, "; ")), nil
		}
		return jsonResult(updated), nil
	}
}

func makeDeleteRecordHandler(nsID, modID uint64) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx = withAuth(ctx)
		args := argsMap(request)
		recordID := parseUint64(args, "recordID")
		if err := service.DefaultRecord.DeleteByID(ctx, nsID, modID, recordID); err != nil {
			return errorResult(err), nil
		}
		return textResult("Record deleted"), nil
	}
}
