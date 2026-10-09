package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

var (
	ruleChainCatalogMu sync.RWMutex
	ruleChainCatalog   func() []rulesgo.NodeSchema
)

// SetRuleChainCatalog supplies the node catalog the chain tools check against.
// It is injected: the catalog lives in compose/rest, which depends on this
// package.
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

func promptRefCheck(ctx context.Context) func(string) error {
	return func(ref string) error {
		_, _, err := aiagent.ResolvePrompt(ctx, ref)
		return err
	}
}

// chatRuleChainToolDefs let the assistant build rule chains safely: see what
// node types exist, check a draft, and save it only if it holds up.
func chatRuleChainToolDefs() []chat.ToolDef {
	chainParams := []chat.ParamDef{
		{Name: "name", Type: "string", Required: true, Description: "Rule chain name"},
		{Name: "nodes", Type: "objects", Required: true, Description: `JSON array: [{"id":"check","type":"condition","config":{...}}] - use list_rule_chain_node_types for types and settings`},
		{Name: "edges", Type: "objects", Description: `JSON array: [{"from":"check","to":"notify","condition":"passed"}]; a condition is the NAME of a variable`},
		{Name: "entryNode", Type: "string", Description: "Id of the first node (default: the first in the list)"},
	}

	return []chat.ToolDef{
		{
			Name:        "list_rule_chain_node_types",
			Description: "List the node types a rule chain can use with each one's settings. Call this before writing a chain; do not invent node types or settings.",
			Handler: func(context.Context, map[string]string) string {
				catalog := currentRuleChainCatalog()
				if len(catalog) == 0 {
					return "Node catalog not available"
				}
				return rulesgo.DescribeNodeTypes(catalog)
			},
		},
		{
			Name:        "validate_rule_chain",
			Description: "Check a rule chain draft against the node catalog without saving it. Returns the problems to fix.",
			Params:      chainParams,
			Handler: func(ctx context.Context, p map[string]string) string {
				chain, issues, errText := checkChainDraft(ctx, p)
				if errText != "" {
					return errText
				}
				if len(issues) == 0 {
					return fmt.Sprintf("OK: %d nodes, %d edges. Nothing was saved.", len(chain.Nodes), len(chain.Edges))
				}
				verdict := "Fix the errors, then validate again"
				if !rulesgo.HasErrors(issues) {
					verdict = "No errors, only warnings"
				}
				return verdict + ":\n" + rulesgo.FormatIssues(issues)
			},
		},
		{
			Name:        "create_rule_chain",
			Description: "Save a rule chain draft that passed validate_rule_chain. A chain with errors is not saved.",
			Mutating:    true,
			Params:      chainParams,
			Handler: func(ctx context.Context, p map[string]string) string {
				chain, issues, errText := checkChainDraft(ctx, p)
				if errText != "" {
					return errText
				}
				if rulesgo.HasErrors(issues) {
					return "Not saved - fix the errors:\n" + rulesgo.FormatIssues(issues)
				}
				if DefaultRuleEngine == nil {
					return "Rule engine not initialized"
				}

				DefaultRuleEngine.RegisterChain(chain)

				out := fmt.Sprintf("Saved rule chain %q (%d nodes, %d edges).", chain.ID, len(chain.Nodes), len(chain.Edges))
				if len(issues) > 0 {
					out += "\nWarnings:\n" + rulesgo.FormatIssues(issues)
				}
				return out
			},
		},
	}
}

// checkChainDraft assembles and checks a draft from tool parameters. errText
// is set when the draft cannot even be read.
func checkChainDraft(ctx context.Context, p map[string]string) (*rulesgo.Chain, []rulesgo.ChainIssue, string) {
	catalog := currentRuleChainCatalog()
	if len(catalog) == 0 {
		return nil, nil, "Node catalog not available, so a chain cannot be checked"
	}

	chain, err := rulesgo.ChainFromParts(strings.TrimSpace(p["name"]), "", p["nodes"], p["edges"], strings.TrimSpace(p["entryNode"]))
	if err != nil {
		return nil, nil, err.Error()
	}
	return chain, rulesgo.ValidateChain(chain, catalog, rulesgo.ValidateOptions{PromptExists: promptRefCheck(ctx)}), ""
}
