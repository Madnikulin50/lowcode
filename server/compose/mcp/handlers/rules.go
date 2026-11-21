package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	RuleEngine     *rulesgo.EngineWithPersistence
	OnChainMissing func(ctx context.Context, chainID string)
)

func SetRuleEngine(engine *rulesgo.EngineWithPersistence) {
	RuleEngine = engine
}

var (
	ruleChainCatalogMu sync.RWMutex
	ruleChainCatalog   func() []rulesgo.NodeSchema
)

// SetRuleChainCatalog supplies the node catalog the chain tools validate
// against. It is injected because the catalog lives in compose/rest, which
// depends on this package.
func SetRuleChainCatalog(fn func() []rulesgo.NodeSchema) {
	ruleChainCatalogMu.Lock()
	ruleChainCatalog = fn
	ruleChainCatalogMu.Unlock()
}

func currentRuleChainCatalog() []rulesgo.NodeSchema {
	ruleChainCatalogMu.RLock()
	fn := ruleChainCatalog
	ruleChainCatalogMu.RUnlock()
	if fn == nil {
		return nil
	}
	return fn()
}

// chainPromptCheck lets the validator confirm that "@prompt:..." references
// in a chain point at real prompts.
func chainPromptCheck(ctx context.Context) func(string) error {
	return func(ref string) error {
		_, _, err := aiagent.ResolvePrompt(ctx, ref)
		return err
	}
}

func SetOnChainMissing(fn func(ctx context.Context, chainID string)) {
	OnChainMissing = fn
}

func initRules(ctx context.Context, s *server.MCPServer) {
	s.AddTool(mcp.NewTool("run_rule_chain",
		mcp.WithDescription("Execute a rule chain with the given context data. Rule chains are visual workflows of conditions, actions, and AI nodes."),
		mcp.WithString("chainID", mcp.Description("Rule chain ID"), mcp.Required()),
		mcp.WithString("context", mcp.Description("JSON object with input data for the rule chain"), mcp.Required()),
	), handleRunRuleChain)

	s.AddTool(mcp.NewTool("get_rule_chain",
		mcp.WithDescription("Get details about a specific rule chain"),
		mcp.WithString("chainID", mcp.Description("Rule chain ID"), mcp.Required()),
	), handleGetRuleChain)

	s.AddTool(mcp.NewTool("list_rule_chains",
		mcp.WithDescription("List all available rule chains"),
	), handleListRuleChains)

	s.AddTool(mcp.NewTool("rulechain_node_types",
		mcp.WithDescription("List the node types a rule chain can use, with each one's settings (which are required, allowed values). Call this before writing a chain: do not invent node types or settings."),
	), handleRuleChainNodeTypes)

	s.AddTool(mcp.NewTool("validate_rule_chain",
		mcp.WithDescription("Check a rule chain against the node catalog without saving it. Returns the problems to fix (errors block saving) and warnings."),
		mcp.WithString("name", mcp.Description("Rule chain name")),
		mcp.WithString("nodes", mcp.Description("JSON array of nodes: {id, type, config}"), mcp.Required()),
		mcp.WithString("edges", mcp.Description("JSON array of edges: {from, to, condition}")),
		mcp.WithString("entryNode", mcp.Description("Entry point node ID (default: the first node)")),
	), handleValidateRuleChain)

	s.AddTool(mcp.NewTool("create_rule_chain",
		mcp.WithDescription("Create a new rule chain. It is checked first (see validate_rule_chain): a chain with errors is not saved, and the problems are returned."),
		mcp.WithString("name", mcp.Description("Rule chain name"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Rule chain description")),
		mcp.WithString("nodes", mcp.Description("JSON array of nodes with id, type, config")),
		mcp.WithString("edges", mcp.Description("JSON array of edges with from, to, condition")),
		mcp.WithString("entryNode", mcp.Description("Entry point node ID")),
	), handleCreateRuleChain)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://rulechain/{id}", "Rule chain by ID",
		mcp.WithTemplateDescription("Read a rule chain definition"),
	), handleRuleChainResource)

	s.AddResourceTemplate(mcp.NewResourceTemplate("mcp://rulechain/list", "All rule chains",
		mcp.WithTemplateDescription("List all available rule chains"),
	), handleRuleChainListResource)
}

func handleRunRuleChain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	chainID := getString(args, "chainID")
	contextJSON := getString(args, "context")

	var input map[string]interface{}
	if contextJSON != "" {
		if err := json.Unmarshal([]byte(contextJSON), &input); err != nil {
			return textResult(fmt.Sprintf("Invalid context JSON: %v", err)), nil
		}
	}

	if RuleEngine == nil {
		return textResult("Rule engine not initialized"), nil
	}

	result, err := RuleEngine.RunWithLog(ctx, chainID, input, "mcp")
	if err != nil {
		return errorResult(fmt.Errorf("rule chain execution failed: %w", err)), nil
	}

	return jsonResult(result), nil
}

func handleGetRuleChain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	chainID := getString(args, "chainID")

	if RuleEngine == nil {
		return textResult("Rule engine not initialized"), nil
	}

	chain := RuleEngine.Chain(chainID)
	if chain == nil {
		return textResult(fmt.Sprintf("Chain not found: %s", chainID)), nil
	}

	// Return without raw config for security
	type safeNode struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Label string `json:"label"`
	}
	safe := struct {
		ID          string              `json:"id"`
		Name        string              `json:"name"`
		Description string              `json:"description"`
		Nodes       []safeNode          `json:"nodes"`
		Edges       []rulesgo.ChainEdge `json:"edges"`
		EntryNode   string              `json:"entryNode"`
	}{
		ID:          chain.ID,
		Name:        chain.Name,
		Description: chain.Description,
		Nodes:       make([]safeNode, len(chain.Nodes)),
		Edges:       chain.Edges,
		EntryNode:   chain.EntryNode,
	}
	for i, n := range chain.Nodes {
		safe.Nodes[i] = safeNode{ID: n.ID, Type: n.Type, Label: n.Label}
	}

	return jsonResult(safe), nil
}

func handleListRuleChains(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)

	if RuleEngine == nil {
		return textResult("Rule engine not initialized"), nil
	}

	chains := RuleEngine.Chains()
	result := make([]map[string]interface{}, 0, len(chains))
	for _, c := range chains {
		result = append(result, map[string]interface{}{
			"id":          c.ID,
			"name":        c.Name,
			"description": c.Description,
			"nodeCount":   len(c.Nodes),
		})
	}

	return jsonResult(map[string]interface{}{
		"chains": result,
		"total":  len(result),
	}), nil
}

func handleCreateRuleChain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)
	name := getString(args, "name")
	description := getString(args, "description")
	nodesJSON := getString(args, "nodes")
	edgesJSON := getString(args, "edges")
	entryNode := getString(args, "entryNode")

	if RuleEngine == nil {
		return textResult("Rule engine not initialized"), nil
	}

	var nodes []rulesgo.ChainNode
	if nodesJSON != "" {
		if err := json.Unmarshal([]byte(nodesJSON), &nodes); err != nil {
			return textResult(fmt.Sprintf("Invalid nodes JSON: %v", err)), nil
		}
	}

	var edges []rulesgo.ChainEdge
	if edgesJSON != "" {
		if err := json.Unmarshal([]byte(edgesJSON), &edges); err != nil {
			return textResult(fmt.Sprintf("Invalid edges JSON: %v", err)), nil
		}
	}

	chainID := fmt.Sprintf("rc_%s", name)
	chain := &rulesgo.Chain{
		ID:          chainID,
		Name:        name,
		Description: description,
		Nodes:       nodes,
		Edges:       edges,
		EntryNode:   entryNode,
	}
	if chain.EntryNode == "" && len(nodes) > 0 {
		chain.EntryNode = nodes[0].ID
	}

	// Without the node catalog nothing can be vouched for; keep working as
	// before rather than refusing every chain.
	var issues []rulesgo.ChainIssue
	validated := false
	if catalog := currentRuleChainCatalog(); len(catalog) > 0 {
		validated = true
		issues = rulesgo.ValidateChain(chain, catalog, rulesgo.ValidateOptions{PromptExists: chainPromptCheck(ctx)})
	}
	if rulesgo.HasErrors(issues) {
		return jsonResult(map[string]interface{}{
			"created": false,
			"reason":  "the chain has errors - fix them and try again (nothing was saved)",
			"issues":  issues,
		}), nil
	}

	RuleEngine.RegisterChain(chain)

	result := map[string]interface{}{
		"chainID":   chainID,
		"name":      name,
		"nodeCount": len(nodes),
		"edgeCount": len(edges),
		"created":   true,
	}
	if len(issues) > 0 {
		result["warnings"] = issues
	}
	if !validated {
		result["note"] = "the node catalog is not available, so the chain was not checked"
	}
	return jsonResult(result), nil
}

func handleRuleChainNodeTypes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	catalog := currentRuleChainCatalog()
	if len(catalog) == 0 {
		return textResult("Node catalog not available"), nil
	}
	return textResult(rulesgo.DescribeNodeTypes(catalog)), nil
}

func handleValidateRuleChain(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx = withAuth(ctx)
	args := argsMap(request)

	name := getString(args, "name")
	if name == "" {
		name = "validation"
	}
	chain, err := rulesgo.ChainFromParts(name, "", getString(args, "nodes"), getString(args, "edges"), getString(args, "entryNode"))
	if err != nil {
		return textResult(err.Error()), nil
	}

	catalog := currentRuleChainCatalog()
	if len(catalog) == 0 {
		return textResult("Node catalog not available"), nil
	}

	issues := rulesgo.ValidateChain(chain, catalog, rulesgo.ValidateOptions{PromptExists: chainPromptCheck(ctx)})
	if len(issues) == 0 {
		return jsonResult(map[string]interface{}{"valid": true, "issues": []rulesgo.ChainIssue{}, "note": "nothing was saved"}), nil
	}
	return jsonResult(map[string]interface{}{"valid": !rulesgo.HasErrors(issues), "issues": issues, "note": "nothing was saved"}), nil
}

func handleRuleChainResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	idStr, _ := request.Params.Arguments["id"].(string)

	if RuleEngine == nil {
		return nil, fmt.Errorf("rule engine not initialized")
	}

	chain := RuleEngine.Chain(idStr)
	if chain == nil {
		return nil, fmt.Errorf("chain not found: %s", idStr)
	}

	data, _ := json.Marshal(chain)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "mcp://rulechain/" + idStr,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

func handleRuleChainListResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	if RuleEngine == nil {
		return nil, fmt.Errorf("rule engine not initialized")
	}

	chains := RuleEngine.Chains()
	data, _ := json.Marshal(chains)
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "mcp://rulechain/list",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
