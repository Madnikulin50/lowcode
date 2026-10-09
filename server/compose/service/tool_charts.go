package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

// ChartToolDefs is the single definition of the chart CRUD tools, shared by
// chat (a curated subset, see Chat() in chat.go) and MCP (all of them,
// registered verbatim — see mcp/handlers/charts.go).
func ChartToolDefs() []Def {
	return []Def{
		{
			Name:        "read_chart",
			Description: "Read a chart by ID, including its config and reports",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				cID := parseUint64(args["chartID"])
				ch, err := DefaultChart.FindByID(ctx, ns, cID)
				if err != nil {
					return nil, fmt.Errorf("chart not found: %w", err)
				}
				return ch, nil
			},
		},
		{
			Name:        "list_charts",
			Description: "List all charts in the current namespace",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: ns})
				if err != nil {
					return nil, fmt.Errorf("failed to list charts: %w", err)
				}
				return set, nil
			},
			Render: renderChartList,
		},
		{
			Name:        "search_charts",
			Description: "Search charts by name",
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "query", Type: "string", Required: true, Description: "Search query"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				set, _, err := DefaultChart.Find(ctx, types.ChartFilter{NamespaceID: ns, Query: args["query"]})
				if err != nil {
					return nil, fmt.Errorf("failed to search charts: %w", err)
				}
				return set, nil
			},
		},
		{
			Name:        "create_chart",
			Description: "Create a new chart with reports, dimensions, and metrics",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "name", Type: "string", Required: true, Description: "Chart display name"},
				{Name: "handle", Type: "string", Required: false, Description: "URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "Short chart description"},
				{Name: "help", Type: "string", Required: false, Description: "Optional Markdown help shown to users"},
				{Name: "config", Type: "json", Required: true, Description: `JSON chart config. Example: {"reports":[{"moduleID":"...","dimensions":[{"field":"createdAt","modifier":"MONTH"}],"metrics":[{"field":"count","type":"bar","label":"Count"}]}]}`},
			},
			Handler: createChartTool,
			Render:  renderChartCreated,
		},
		{
			Name:        "update_chart",
			Description: "Update a chart's name, handle, description, help, or config",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
				{Name: "name", Type: "string", Required: false, Description: "New chart name"},
				{Name: "handle", Type: "string", Required: false, Description: "New URL-safe handle"},
				{Name: "description", Type: "string", Required: false, Description: "Short chart description"},
				{Name: "help", Type: "string", Required: false, Description: "Optional Markdown help shown to users"},
				{Name: "config", Type: "json", Required: false, Description: "JSON chart config"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				cID := parseUint64(args["chartID"])
				ch, err := DefaultChart.FindByID(ctx, ns, cID)
				if err != nil {
					return nil, fmt.Errorf("chart not found: %w", err)
				}
				if v := args["name"]; v != "" {
					ch.Name = v
				}
				if v := args["handle"]; v != "" {
					ch.Handle = v
				}
				if v := args["config"]; v != "" {
					var cfg types.ChartConfig
					if err := json.Unmarshal([]byte(v), &cfg); err != nil {
						return nil, fmt.Errorf("invalid config JSON: %w", err)
					}
					ch.Config = cfg
				}
				if v := args["description"]; v != "" {
					ch.Config.Description = v
				}
				if v := args["help"]; v != "" {
					ch.Config.Help = v
				}
				updated, err := DefaultChart.Update(ctx, ch)
				if err != nil {
					return nil, fmt.Errorf("failed to update chart: %w", err)
				}
				return updated, nil
			},
		},
		{
			Name:        "delete_chart",
			Description: "Delete a chart by ID",
			Mutating:    true,
			Params: []Param{
				{Name: "namespaceID", Type: "string", Required: true, Description: "Namespace ID"},
				{Name: "chartID", Type: "string", Required: true, Description: "Chart ID"},
			},
			Handler: func(ctx context.Context, args map[string]string) (any, error) {
				ns := parseUint64(args["namespaceID"])
				cID := parseUint64(args["chartID"])
				if err := DefaultChart.DeleteByID(ctx, ns, cID); err != nil {
					return nil, fmt.Errorf("failed to delete chart: %w", err)
				}
				return "Chart deleted", nil
			},
		},
	}
}

func createChartTool(ctx context.Context, args map[string]string) (any, error) {
	ns := parseUint64(args["namespaceID"])
	name := args["name"]
	configJSON := args["config"]
	if ns == 0 || name == "" || configJSON == "" {
		return nil, fmt.Errorf("missing required parameters (namespaceID='%s', name='%s')", args["namespaceID"], name)
	}

	var config types.ChartConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("invalid config JSON for chart '%s': %w", name, err)
	}
	if v := args["description"]; v != "" {
		config.Description = v
	}
	if v := args["help"]; v != "" {
		config.Help = v
	}

	ch := &types.Chart{
		NamespaceID: ns,
		Name:        name,
		Handle:      args["handle"],
		Config:      config,
	}

	created, err := DefaultChart.Create(ctx, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to create chart '%s': %w", name, err)
	}
	return created, nil
}

func renderChartCreated(data any) string {
	c := data.(*types.Chart)
	return fmt.Sprintf("✅ Chart '%s' created! ID: %d, Handle: %s", c.Name, c.ID, c.Handle)
}

func renderChartList(data any) string {
	set := data.(types.ChartSet)
	if len(set) == 0 {
		return "No charts found in this namespace."
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📊 **Charts (%d):**\n\n", len(set))
	for _, c := range set {
		fmt.Fprintf(&b, "• **%s** (ID: %d, Handle: %s, Reports: %d)\n", c.Name, c.ID, c.Handle, len(c.Config.Reports))
	}
	return b.String()
}
