package aiagent

import (
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

func NewCRUDAgent(client *chat.Client, tools []chat.ToolDef) *Agent {
	cfg := AgentConfig{
		Name:        "crud-agent",
		Description: "Autonomous data operations. Creates, updates, deletes, and searches records.",
		Model:       "deepseek-v2",
		SystemPrompt: `You are a CRUD agent for the Corteza lowcode platform.
You can:
- Search records with search_records(module, query, limit)
- Create records with create_record(module, values_json)
- Update records with update_record(module, recordID, values_json)
- Delete records with delete_record(module, recordID)

Always present search results as a table. Confirm before deleting.
Start final answers with "FINAL:".`,
		Tools:    tools,
		MaxSteps: 8,
	}
	return New(client, cfg)
}

func NewAssistantAgent(client *chat.Client, tools []chat.ToolDef) *Agent {
	cfg := AgentConfig{
		Name:        "assistant",
		Description: "General-purpose assistant for the Corteza platform. Can search data, answer questions, create content.",
		Model:       "deepseek-v2",
		SystemPrompt: `You are an AI assistant for the Corteza lowcode platform.
You can help users with:
- Finding information across modules (search_records, list_modules)
- Creating and editing records
- Answering questions about data
- Generating reports and summaries

Be helpful and concise. When presenting data, use clear formatting.
Start final answers with "FINAL:".`,
		Tools:    tools,
		MaxSteps: 6,
	}
	return New(client, cfg)
}
