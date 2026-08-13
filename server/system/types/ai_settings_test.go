package types

import (
	"testing"

	sqlTypes "github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/require"
)

func TestAISettingsDecodeKV(t *testing.T) {
	kv := SettingsKV{
		"ai.enabled":               sqlTypes.JSONText(`true`),
		"ai.ollama-url":            sqlTypes.JSONText(`"http://ollama:11434"`),
		"ai.catalog":               sqlTypes.JSONText(`[{"name":"qwen3:8b","enabled":true},{"name":"deepseek-r1","enabled":false}]`),
		"ai.roles.compose-chat":    sqlTypes.JSONText(`"qwen3:8b"`),
		"ai.roles.mcp-agent":       sqlTypes.JSONText(`"qwen3:8b"`),
		"ai.roles.automation-chat": sqlTypes.JSONText(`""`),
		"ai.roles.rulesgo-ai":      sqlTypes.JSONText(`"qwen3:8b"`),
	}

	var aux AppSettings
	require.NoError(t, DecodeKV(kv, &aux))
	require.True(t, aux.AI.Enabled)
	require.Equal(t, "http://ollama:11434", aux.AI.OllamaURL)
	require.Len(t, aux.AI.Catalog, 2)
	require.Equal(t, "qwen3:8b", aux.AI.Catalog[0].Name)
	require.True(t, aux.AI.Catalog[0].Enabled)
	require.False(t, aux.AI.Catalog[1].Enabled)
	require.Equal(t, "qwen3:8b", aux.AI.Roles.ComposeChat)
	require.Equal(t, "qwen3:8b", aux.AI.Roles.MCPAgent)
	require.Equal(t, "", aux.AI.Roles.AutomationChat)
	require.Equal(t, "qwen3:8b", aux.AI.Roles.RulesgoAI)
}

func TestUIMapSettingsDecodeKV(t *testing.T) {
	kv := SettingsKV{
		"ui.map": sqlTypes.JSONText(`{"tileSource":"local","tileURL":"http://localhost:8081/{z}/{x}/{y}.png","minZoom":0,"maxZoom":16,"attribution":"© OSM"}`),
	}
	var aux AppSettings
	require.NoError(t, DecodeKV(kv, &aux))
	require.Equal(t, "local", aux.UI.Map.TileSource)
	require.Equal(t, "http://localhost:8081/{z}/{x}/{y}.png", aux.UI.Map.TileURL)
	require.Equal(t, uint(16), aux.UI.Map.MaxZoom)
	require.Equal(t, "© OSM", aux.UI.Map.Attribution)
}
