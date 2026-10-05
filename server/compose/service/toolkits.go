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
	schemaTools = append(schemaTools, ToChatToolDefs(ModuleToolDefs()...)...)
	schemaTools = append(schemaTools, ToChatToolDefs(ChartToolDefs()...)...)
	schemaTools = append(schemaTools, ToChatToolDefs(PageToolDefs()...)...)
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
	cat.Register(aiagent.ToolKit{
		Name:        "workflows",
		Description: "List, run and follow automation workflows (AI, approval and integration processes)",
		Tools:       chatWorkflowToolDefs(),
	})
	cat.Register(aiagent.ToolKit{
		Name:        "rulechains",
		Description: "Build rule chains: node types, checking a draft, saving it",
		Tools:       chatRuleChainToolDefs(),
	})
	cat.Register(aiagent.ToolKit{
		Name:        "prompts",
		Description: "The prompt library: versioned prompts for AI steps, with test cases",
		Tools:       chatPromptToolDefs(),
	})
	cat.Register(aiagent.ToolKit{
		Name:        "skills",
		Description: "The skill library: versioned instructions that agents load on demand",
		Tools:       chatSkillToolDefs(),
	})
	cat.Register(aiagent.ToolKit{
		Name:        "risk",
		Description: "Suggest risk factors and explain risk model assessments",
		Tools:       chatRiskToolDefs(),
	})
}
