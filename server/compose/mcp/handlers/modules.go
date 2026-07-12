package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/compose/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func initModules(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("read_module",
		mcp.WithDescription("Read a module by ID, including its fields"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("moduleID", mcp.Description("Module ID"), mcp.Required()),
	), handleReadModule)

	s.AddTool(mcp.NewTool("list_modules",
		mcp.WithDescription("List all modules in a namespace"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
	), handleListModules)

	s.AddTool(mcp.NewTool("search_modules",
		mcp.WithDescription("Search modules by name"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("query", mcp.Description("Search query"), mcp.Required()),
	), handleSearchModules)

	s.AddTool(mcp.NewTool("create_module",
		mcp.WithDescription("Create a new module with fields"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Module name"), mcp.Required()),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("fields",
			mcp.Description(`JSON array of fields. Each field: {"name":"...","kind":"String|Number|DateTime|Select|Bool|User|Record|File|URL|Email","required":true/false,"label":"..."}`),
			mcp.Required(),
		),
	), handleCreateModule)

	s.AddTool(mcp.NewTool("update_module",
		mcp.WithDescription("Update a module's name, handle, or fields"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("moduleID", mcp.Description("Module ID"), mcp.Required()),
		mcp.WithString("name", mcp.Description("Module name")),
		mcp.WithString("handle", mcp.Description("URL-safe handle")),
		mcp.WithString("fields", mcp.Description("JSON array of field objects (optional)")),
	), handleUpdateModule)

	s.AddTool(mcp.NewTool("delete_module",
		mcp.WithDescription("Delete a module by ID"),
		mcp.WithString("namespaceID", mcp.Description("Namespace ID"), mcp.Required()),
		mcp.WithString("moduleID", mcp.Description("Module ID"), mcp.Required()),
	), handleDeleteModule)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://module/{nsID}/{id}", "Module by ID",
		mcp.WithTemplateDescription("Read a module by namespace and module ID"),
	), handleModuleResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://module/{nsID}/list", "Modules in namespace",
		mcp.WithTemplateDescription("List all modules in a namespace"),
	), handleModuleListResource)
}

func handleReadModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	mID := parseUint64(args, "moduleID")
	mod, err := service.DefaultModule.FindByID(ctx, nsID, mID)
	if err != nil {
		return errorResult(fmt.Errorf("module not found: %w", err)), nil
	}
	return jsonResult(mod), nil
}

func handleListModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	set, _, err := service.DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: nsID})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleSearchModules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	query := getString(args, "query")
	set, _, err := service.DefaultModule.Find(ctx, types.ModuleFilter{
		NamespaceID: nsID,
		Query:       query,
	})
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(set), nil
}

func handleCreateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	namespaceID := parseUint64(args, "namespaceID")
	name := getString(args, "name")
	handle := getString(args, "handle")
	fieldsJSON := getString(args, "fields")

	var fields []map[string]interface{}
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return textResult(fmt.Sprintf("Invalid fields JSON: %v", err)), nil
	}

	fieldSet := make(types.ModuleFieldSet, 0, len(fields))
	for _, f := range fields {
		fieldSet = append(fieldSet, &types.ModuleField{
			Name:     getFieldString(f, "name"),
			Kind:     getFieldString(f, "kind"),
			Label:    getFieldString(f, "label"),
			Required: getFieldBool(f, "required"),
		})
	}

	mod := &types.Module{
		NamespaceID: namespaceID,
		Name:        name,
		Handle:      handle,
		Fields:      fieldSet,
		Config:      types.ModuleConfig{},
	}

	created, err := service.DefaultModule.Create(ctx, mod)
	if err != nil {
		return errorResult(err), nil
	}

	return jsonResult(map[string]interface{}{
		"moduleID": fmt.Sprintf("%d", created.ID),
		"name":     created.Name,
		"handle":   created.Handle,
		"fields":   len(created.Fields),
	}), nil
}

func handleUpdateModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	mID := parseUint64(args, "moduleID")

	mod, err := service.DefaultModule.FindByID(ctx, nsID, mID)
	if err != nil {
		return errorResult(fmt.Errorf("module not found: %w", err)), nil
	}

	if v := getString(args, "name"); v != "" {
		mod.Name = v
	}
	if v := getString(args, "handle"); v != "" {
		mod.Handle = v
	}

	if v := getString(args, "fields"); v != "" {
		var fields []*types.ModuleField
		if err := json.Unmarshal([]byte(v), &fields); err != nil {
			return textResult(fmt.Sprintf("Invalid fields JSON: %v", err)), nil
		}
		mod.Fields = fields
	}

	updated, err := service.DefaultModule.Update(ctx, mod)
	if err != nil {
		return errorResult(err), nil
	}
	return jsonResult(updated), nil
}

func handleDeleteModule(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	nsID := parseUint64(args, "namespaceID")
	mID := parseUint64(args, "moduleID")
	if err := service.DefaultModule.DeleteByID(ctx, nsID, mID); err != nil {
		return errorResult(err), nil
	}
	return textResult("Module deleted"), nil
}

func handleModuleResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	idStr, _ := request.Params.Arguments["id"].(string)
	var nsID, mID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(idStr, "%d", &mID)
	mod, err := service.DefaultModule.FindByID(ctx, nsID, mID)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://module/" + nsIDStr + "/" + idStr, Text: toJSON(mod)},
	}, nil
}

func handleModuleListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	var nsID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	set, _, err := service.DefaultModule.Find(ctx, types.ModuleFilter{NamespaceID: nsID})
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{URI: "mcp://module/" + nsIDStr + "/list", Text: toJSON(set)},
	}, nil
}
