package mcp

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/compose/mcp/handlers"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

// ruleChainAICall backs the rule chain ai and ai.operation nodes: it runs the
// named agent and hands back the answer together with what the call cost, so
// the node can show it in its trace.
//
// model, when set, is used for this call only (see aiagent.ContextWithModel);
// before this was honoured, a node's model field was silently ignored.
//
// allowMutating stands in for the interactive "да" a chat user would type
// before a data-changing tool call runs: false reports such a call back as
// ConfirmNeeded instead of executing it.
func ruleChainAICall(ctx context.Context, agent, prompt, model string, allowMutating bool) (*rulesgo.AIOperationResult, error) {
	if handlers.AgentRegistry == nil {
		return nil, fmt.Errorf("agent registry not available")
	}

	res, err := handlers.AgentRegistry.RunAgentConfirmed(aiagent.ContextWithModel(ctx, model), agent, prompt, nil, allowMutating)
	if err != nil {
		return nil, err
	}

	confirmCalls := make([]string, 0, len(res.ConfirmCalls))
	for _, c := range res.ConfirmCalls {
		confirmCalls = append(confirmCalls, c.Name)
	}

	var tools []string
	for _, s := range res.Steps {
		for _, tc := range s.Tools {
			tools = append(tools, tc.Name)
		}
	}

	return &rulesgo.AIOperationResult{
		Output:           res.Output,
		Success:          res.Success,
		Error:            res.Error,
		ConfirmNeeded:    res.ConfirmNeeded,
		ConfirmCalls:     confirmCalls,
		Model:            res.Model,
		LLMCalls:         res.LLMCalls,
		PromptTokens:     res.PromptTokens,
		CompletionTokens: res.CompletionTokens,
		Tools:            tools,
		Skills:           res.Skills,
	}, nil
}
