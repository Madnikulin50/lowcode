package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (a *Agent) MCPToolDefs() []mcp.Tool {
	return []mcp.Tool{
		mcp.NewTool("scan_network",
			mcp.WithDescription("Scan a network CIDR range for devices and save them to CMDB"),
			mcp.WithString("cidr", mcp.Description("Network CIDR, e.g. 192.168.1.0/24"), mcp.Required()),
			mcp.WithString("namespaceID", mcp.Description("Namespace ID for CMDB storage")),
		),
		mcp.NewTool("list_devices",
			mcp.WithDescription("List all discovered network devices from CMDB"),
			mcp.WithString("moduleID", mcp.Description("Module ID (auto-created if empty)")),
		),
		mcp.NewTool("get_device",
			mcp.WithDescription("Get details of a specific device by record ID"),
			mcp.WithString("recordID", mcp.Description("Record ID"), mcp.Required()),
			mcp.WithString("moduleID", mcp.Description("Module ID")),
		),
		mcp.NewTool("delete_device",
			mcp.WithDescription("Delete a device record from CMDB"),
			mcp.WithString("recordID", mcp.Description("Record ID"), mcp.Required()),
			mcp.WithString("moduleID", mcp.Description("Module ID")),
		),
		mcp.NewTool("get_scan_status",
			mcp.WithDescription("Get the status of a running or completed network scan"),
			mcp.WithString("scanID", mcp.Description("Scan ID"), mcp.Required()),
		),
		mcp.NewTool("list_scans",
			mcp.WithDescription("List all network scans"),
		),
		mcp.NewTool("ensure_module",
			mcp.WithDescription("Create or verify the Devices module in CMDB"),
		),
	}
}

func (a *Agent) MCPHandlers() map[string]server.ToolHandlerFunc {
	return map[string]server.ToolHandlerFunc{
		"scan_network": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			cidr := req.GetString("cidr", "")
			if cidr == "" {
				return mcp.NewToolResultText("Error: cidr is required"), nil
			}

			nsID := a.cfg.NamespaceID
			if v := req.GetString("namespaceID", ""); v != "" {
				fmt.Sscanf(v, "%d", &nsID)
			}

			status, err := a.StartScan(ctx, ScanTarget{CIDR: cidr, NamespaceID: FlexUint64(nsID)})
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
			}

			b, _ := json.Marshal(status)
			return mcp.NewToolResultText(string(b)), nil
		},
		"list_devices": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			devices, err := a.ListDevices(ctx, 0)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
			}
			if len(devices) == 0 {
				return mcp.NewToolResultText("No devices found"), nil
			}
			b, _ := json.Marshal(devices)
			return mcp.NewToolResultText(string(b)), nil
		},
		"get_device": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			recordID := req.GetInt("recordID", 0)
			device, err := a.GetDevice(ctx, 0, uint64(recordID))
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
			}
			b, _ := json.Marshal(device)
			return mcp.NewToolResultText(string(b)), nil
		},
		"delete_device": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			recordID := req.GetInt("recordID", 0)
			if err := a.DeleteDevice(ctx, 0, uint64(recordID)); err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
			}
			return mcp.NewToolResultText("Device deleted"), nil
		},
		"get_scan_status": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			scanID := req.GetString("scanID", "")
			status := a.GetStatus(scanID)
			if status == nil {
				return mcp.NewToolResultText("Scan not found"), nil
			}
			b, _ := json.Marshal(status)
			return mcp.NewToolResultText(string(b)), nil
		},
		"list_scans": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			scans := a.ListScans()
			if len(scans) == 0 {
				return mcp.NewToolResultText("No scans found"), nil
			}
			b, _ := json.Marshal(scans)
			return mcp.NewToolResultText(string(b)), nil
		},
		"ensure_module": func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			modID, err := a.EnsureModule(ctx)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Module ready, ID: %d", modID)), nil
		},
	}
}

func StartMCPServer(ctx context.Context, agent *Agent, addr string) error {
	s := server.NewMCPServer(
		"CMDB Discovery Agent",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	for _, tool := range agent.MCPToolDefs() {
		s.AddTool(tool, agent.MCPHandlers()[tool.Name])
	}

	if addr == "stdio" {
		return server.ServeStdio(s)
	}

	sseServer := server.NewSSEServer(s)
	return sseServer.Start(addr)
}
