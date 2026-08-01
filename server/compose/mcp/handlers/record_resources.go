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

func initRecordResources(ctx context.Context, s *server.MCPServer) {
	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://record/{nsID}/{moduleID}/{id}", "Record by ID",
		mcp.WithTemplateDescription("Read a specific record from a module"),
	), handleRecordResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://record/{nsID}/{moduleID}/list", "Records in module",
		mcp.WithTemplateDescription("List records in a module"),
	), handleRecordListResource)
}

func handleRecordResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	modIDStr, _ := request.Params.Arguments["moduleID"].(string)
	idStr, _ := request.Params.Arguments["id"].(string)

	var nsID, moduleID, recordID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(modIDStr, "%d", &moduleID)
	fmt.Sscanf(idStr, "%d", &recordID)

	mod, err := service.DefaultModule.FindByID(ctx, nsID, moduleID)
	if err != nil {
		return nil, fmt.Errorf("module not found: %w", err)
	}

	record, _, err := service.DefaultRecord.FindByID(ctx, nsID, moduleID, recordID)
	if err != nil {
		return nil, fmt.Errorf("record not found: %w", err)
	}

	_ = mod
	data, _ := json.Marshal(record)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "mcp://record/" + nsIDStr + "/" + modIDStr + "/" + idStr,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func handleRecordListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	ctx = withAuth(ctx)
	nsIDStr, _ := request.Params.Arguments["nsID"].(string)
	modIDStr, _ := request.Params.Arguments["moduleID"].(string)

	var nsID, moduleID uint64
	fmt.Sscanf(nsIDStr, "%d", &nsID)
	fmt.Sscanf(modIDStr, "%d", &moduleID)

	mod, err := service.DefaultModule.FindByID(ctx, nsID, moduleID)
	if err != nil {
		return nil, fmt.Errorf("module not found: %w", err)
	}

	ff := types.RecordFilter{
		ModuleID: moduleID,
	}

	set, _, err := service.DefaultRecord.Find(ctx, ff)
	if err != nil {
		return nil, fmt.Errorf("records not found: %w", err)
	}

	_ = mod
	data, _ := json.Marshal(set)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "mcp://record/" + nsIDStr + "/" + modIDStr + "/list",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
