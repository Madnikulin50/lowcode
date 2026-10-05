// ai.operation is the "AI as a function" node: unlike the plain ai node
// (free-text prompt -> free-text response), it takes named input parameters
// and a declared output schema, and returns a parsed, validated JSON object
// instead of prose - a predictable contract a downstream node can rely on,
// the same idea as ELMA365's "AI-операция".
//
// It shares the same guardrail as the ai node: a mutating tool call the
// agent attempts is blocked unless allowMutating is set - see AIOperationResult
// and aiagent.Agent.RunConfirmed.
package rulesgo

import (
	"context"
	"fmt"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/chat"
)

type AIOperationConfig struct {
	Agent  string `json:"agent"`
	Prompt string `json:"prompt"` // base instruction; inputs/outputSchema are appended to it
	Model  string `json:"model,omitempty"`
	// Skill: a skill for the agent to follow, "handle" or "handle@version"
	Skill string `json:"skill,omitempty"`

	// Inputs: named parameter -> template value. Rendered into the prompt as
	// a JSON block instead of the caller having to interpolate a single
	// free-text string themselves.
	Inputs map[string]string `json:"inputs,omitempty"`

	// OutputSchema: field name -> JSON type ("string", "number", "boolean",
	// "array", "object"). The agent is instructed to answer with exactly
	// this shape; the response is parsed and validated against it before
	// the node succeeds.
	OutputSchema map[string]string `json:"outputSchema,omitempty"`

	AllowMutating bool `json:"allowMutating,omitempty"`
	MaxRetries    int  `json:"maxRetries,omitempty"` // re-ask on invalid/mismatched JSON, default 1
}

type aiOperationExecutor struct {
	call func(ctx context.Context, agent, prompt, model string, allowMutating bool) (*AIOperationResult, error)
}

func (n *aiOperationExecutor) Execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, error) {
	out, trace, err := n.execute(ctx, node, ec)
	if trace != nil {
		recordTrace(ctx, trace.Map())
	}
	return out, err
}

func (n *aiOperationExecutor) execute(ctx context.Context, node ChainNode, ec *ExecutionContext) (map[string]interface{}, *aiagent.CallTrace, error) {
	cfg, err := ParseNodeConfig[AIOperationConfig](node.Config)
	if err != nil {
		return nil, nil, err
	}
	rawPrompt, promptRef, err := aiagent.ResolvePrompt(ctx, cfg.Prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("ai.operation: %w", err)
	}
	cfg.Agent = resolveTemplateValue(cfg.Agent, ec)
	cfg.Prompt = resolveTemplateValue(rawPrompt, ec)
	cfg.Model = resolveTemplateValue(cfg.Model, ec)
	cfg.Skill = resolveTemplateValue(cfg.Skill, ec)

	if cfg.Agent == "" {
		return nil, nil, fmt.Errorf("ai.operation: agent is required")
	}
	if cfg.Prompt == "" {
		return nil, nil, fmt.Errorf("ai.operation: prompt is required")
	}

	inputs := make(map[string]interface{}, len(cfg.Inputs))
	for k, v := range cfg.Inputs {
		inputs[k] = resolveTemplateValue(v, ec)
	}

	if n.call == nil {
		return map[string]interface{}{"agent": cfg.Agent, "status": "not_configured"}, nil, nil
	}

	// MaxRetries is taken at face value (0 = try once, no retry): a plain
	// int can't tell "explicitly 0" apart from "field omitted" in JSON, so
	// clamping a zero value up to some implicit default would silently
	// override an explicit "don't retry" from the chain author.
	sr, err := aiagent.RunStructured(ctx, aiagent.BudgetedRunner(n.runner()), aiagent.StructuredRequest{
		Agent:         cfg.Agent,
		Prompt:        cfg.Prompt,
		PromptRef:     promptRef,
		Model:         cfg.Model,
		Skill:         cfg.Skill,
		Inputs:        inputs,
		OutputSchema:  cfg.OutputSchema,
		AllowMutating: cfg.AllowMutating,
		MaxRetries:    cfg.MaxRetries,
	})
	trace := &sr.Trace
	if err != nil {
		return nil, trace, fmt.Errorf("ai.operation: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"agent":   cfg.Agent,
		"result":  sr.Result,
		"raw":     sr.Raw,
	}, trace, nil
}

// runner adapts the injected AICall to the shared structured-call loop. The
// model override travels on the context (aiagent.ContextWithModel).
func (n *aiOperationExecutor) runner() aiagent.Runner {
	return func(ctx context.Context, agent, prompt string, allowMutating bool) (*aiagent.AgentResult, error) {
		res, err := n.call(ctx, agent, prompt, aiagent.ModelFromContext(ctx), allowMutating)
		if err != nil {
			return nil, err
		}

		out := &aiagent.AgentResult{
			Success:          res.Success,
			Output:           res.Output,
			Error:            res.Error,
			ConfirmNeeded:    res.ConfirmNeeded,
			Model:            res.Model,
			LLMCalls:         res.LLMCalls,
			PromptTokens:     res.PromptTokens,
			CompletionTokens: res.CompletionTokens,
			Skills:           res.Skills,
		}
		for _, name := range res.ConfirmCalls {
			out.ConfirmCalls = append(out.ConfirmCalls, aiagent.Call{Name: name})
		}
		if len(res.Tools) > 0 {
			step := aiagent.AgentStep{Type: "execute"}
			for _, name := range res.Tools {
				step.Tools = append(step.Tools, chat.ToolCall{Name: name})
			}
			out.Steps = []aiagent.AgentStep{step}
		}
		return out, nil
	}
}

// The prompt contract, JSON extraction and schema validation are shared with
// the automation workflow ai* functions - see pkg/aiagent/structured.go.
var (
	retryPrompt          = aiagent.RetryPrompt
	buildOperationPrompt = aiagent.BuildOperationPrompt
	extractJSONObject    = aiagent.ExtractJSONObject
	validateOutputSchema = aiagent.ValidateOutputSchema
)

func truncateForError(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
