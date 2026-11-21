package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

// ModuleToolDefs is the single definition of the module CRUD tools, shared
// by chat (a curated subset, see Chat() in chat.go) and MCP (all of them,
// registered verbatim — see mcp/handlers/modules.go).
func ModuleToolDefs() []Def {
	return []Def{
		{
			Name:        "read_module",
			Description: "Read a module by ID, including its fields and config",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				mID := parseUint64(args["moduleID"])
				mod, err := DefaultModule.FindByID(ctx, ns, mID)
				if err != nil {
					return nil, fmt.Errorf("module not found: %w", err)
				}
				return mod, nil
			},
		},
		{
			Name:        "list_modules",
			Description: "List all modules in the current namespace with their fields",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: ns})
				if err != nil {
					return nil, fmt.Errorf("failed to list modules: %w", err)
				}
				return set, nil
			},
			Render: renderModuleList,
		},
		{
			Name:        "search_modules",
			Description: "Search modules by name",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: ns, Query: args["query"]})
				if err != nil {
					return nil, fmt.Errorf("failed to search modules: %w", err)
				}
				return set, nil
			},
		},
		{
			Name:        "create_module",
			Description: "Create a new module (entity to store records) with fields",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "name", Type: "string", Required: true, Description: "Module display name"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle (auto-generated if empty)"},
				{Name: "fields", Type: "json", Required: true, Description: `JSON array. Each field: {"name":"...","kind":"String|Number|DateTime|Select|Bool|User|Record|File|URL|Email","label":"...","required":true/false}`},
			},
			Handler: createModuleTool,
			Render:  renderModuleCreated,
		},
		{
			Name:        "update_module",
			Description: "Update a module's name, handle, or fields",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
				{Name: "name", Type: "string", Required: false, Description: "New module name"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "fields", Type: "json", Required: false, Description: "JSON array of field objects"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				mID := parseUint64(args["moduleID"])
				mod, err := DefaultModule.FindByID(ctx, ns, mID)
				if err != nil {
					return nil, fmt.Errorf("module not found: %w", err)
				}
				if v := args["name"]; v != "" {
					mod.Name = v
				}
				if v := args["handle"]; v != "" {
					mod.Handle = v
				}
				if v := args["fields"]; v != "" {
					var fields []*types.ModuleField
					if err := json.Unmarshal([]byte(v), &fields); err != nil {
						return nil, fmt.Errorf("invalid fields JSON: %w", err)
					}
					mod.Fields = fields
				}
				updated, err := DefaultModule.Update(ctx, mod)
				if err != nil {
					return nil, fmt.Errorf("failed to update module: %w", err)
				}
				return updated, nil
			},
		},
		{
			Name:        "delete_module",
			Description: "Delete a module by ID",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "moduleID", Type: "string", Required: true, Description: "Module ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				mID := parseUint64(args["moduleID"])
				if err := DefaultModule.DeleteByID(ctx, ns, mID); err != nil {
					return nil, fmt.Errorf("failed to delete module: %w", err)
				}
				return "Module deleted", nil
			},
		},
	}
}

func createModuleTool(ctx context.Context, args map[string]string) (any, error) {
	ns := parseUint64(args["namespaceID"])
	name := args["name"]
	fieldsJSON := args["fields"]
	if ns == 0 || name == "" || fieldsJSON == "" {
		return nil, fmt.Errorf("missing required parameters (namespaceID='%s', name='%s')", args["namespaceID"], name)
	}

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return nil, fmt.Errorf("invalid fields JSON for module '%s': %w", name, err)
	}

	fieldSet := make(types.ModuleFieldSet, 0, len(fields))
	for _, f := range fields {
		fieldSet = append(fieldSet, &types.ModuleField{
			Name:     getFieldStr(f, "name"),
			Kind:     getFieldStr(f, "kind"),
			Label:    getFieldStr(f, "label"),
			Required: getFieldBool(f, "required"),
		})
	}

	mod := &types.Module{
		NamespaceID: ns,
		Name:        name,
		Handle:      args["handle"],
		Fields:      fieldSet,
		Config:      types.ModuleConfig{},
	}

	created, err := DefaultModule.Create(ctx, mod)
	if err != nil {
		return nil, fmt.Errorf("failed to create module '%s': %w", name, err)
	}
	return created, nil
}

func renderModuleCreated(data any) string {
	m := data.(*types.Module)
	return fmt.Sprintf("✅ Module '%s' created! ID: %d, Handle: %s, Fields: %d", m.Name, m.ID, m.Handle, len(m.Fields))
}

func renderModuleList(data any) string {
	set := data.(types.ModuleSet)
	if len(set) == 0 {
		return "No modules found in this namespace."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📦 **Modules (%d):**\n\n", len(set))
	for _, m := range set {
		fields := make([]string, 0, len(m.Fields))
		for _, f := range m.Fields {
			fields = append(fields, f.Name+" ("+f.Kind+")")
		}
		fs := strings.Join(fields, "\r\n")
		fmt.Fprintf(&b, "• **%s** (ID: %d, Handle: %s)\r\n", m.Name, m.ID, m.Handle)
		if fs != "" {
			fmt.Fprintf(&b, "\r\n *Fields*: %s\r\n\r\n", fs)
		}
	}
	return b.String()
}
