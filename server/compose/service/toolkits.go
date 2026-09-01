package service

import (
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func RegisterComposeToolKits(cat *aiagent.Catalog) {
	if cat == nil {
		cat = aiagent.DefaultCatalog()
	}
	schemaTools := make([]chat.ToolDef, 0, 24)
	schemaTools = append(schemaTools, chatModuleToolDefs()...)
	schemaTools = append(schemaTools, chatChartToolDefs()...)
	schemaTools = append(schemaTools, chatPageToolDefs()...)
	cat.Register(aiagent.ToolKit{
		Name:        "compose.schema",
		Description: "Modules, pages, and charts",
		Tools:       schemaTools,
	})
	cat.Register(aiagent.ToolKit{
		Name:        "compose.records",
		Description: "Search records across modules",
		Tools:       []chat.ToolDef{chatRecordSearchToolDef(), chatExtractAttachmentToolDef()},
	})
	cat.Register(aiagent.ToolKit{
		Name:        "compose.mail",
		Description: "Email notifications",
		Tools:       chatMailToolDefs(),
	})
	cat.Register(aiagent.ToolKit{
		Name:        "compose.visualize",
		Description: "Charts and reports from live data",
		Tools:       chatVisualizeTools(),
	})
}
