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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type AIOperationConfig struct {
	Agent  string `json:"agent"`
	Prompt string `json:"prompt"` // base instruction; inputs/outputSchema are appended to it
	Model  string `json:"model,omitempty"`

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
	cfg, err := ParseNodeConfig[AIOperationConfig](node.Config)
	if err != nil {
		return nil, err
	}
	cfg.Agent = resolveTemplateValue(cfg.Agent, ec)
	cfg.Prompt = resolveTemplateValue(cfg.Prompt, ec)
	cfg.Model = resolveTemplateValue(cfg.Model, ec)

	if cfg.Agent == "" {
		return nil, fmt.Errorf("ai.operation: agent is required")
	}
	if cfg.Prompt == "" {
		return nil, fmt.Errorf("ai.operation: prompt is required")
	}

	inputs := make(map[string]interface{}, len(cfg.Inputs))
	for k, v := range cfg.Inputs {
		inputs[k] = resolveTemplateValue(v, ec)
	}

	if n.call == nil {
		return map[string]interface{}{"agent": cfg.Agent, "status": "not_configured"}, nil
	}

	// MaxRetries is taken at face value (0 = try once, no retry): a plain
	// int can't tell "explicitly 0" apart from "field omitted" in JSON, so
	// clamping a zero value up to some implicit default would silently
	// override an explicit "don't retry" from the chain author.
	maxRetries := cfg.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	basePrompt := buildOperationPrompt(cfg.Prompt, inputs, cfg.OutputSchema)
	prompt := basePrompt

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err := n.call(ctx, cfg.Agent, prompt, cfg.Model, cfg.AllowMutating)
		if err != nil {
			return nil, fmt.Errorf("ai.operation: %w", err)
		}
		if res.ConfirmNeeded {
			return nil, fmt.Errorf("ai.operation: blocked - agent attempted a mutating action requiring confirmation (%s) - set allowMutating:true on this node to permit it", strings.Join(res.ConfirmCalls, ", "))
		}

		jsonStr, ok := extractJSONObject(res.Output)
		if !ok {
			lastErr = fmt.Errorf("response did not contain a JSON object: %s", truncateForError(res.Output, 200))
			prompt = retryPrompt(basePrompt, lastErr)
			continue
		}

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			lastErr = fmt.Errorf("invalid JSON: %w", err)
			prompt = retryPrompt(basePrompt, lastErr)
			continue
		}

		if len(cfg.OutputSchema) > 0 {
			if err := validateOutputSchema(result, cfg.OutputSchema); err != nil {
				lastErr = err
				prompt = retryPrompt(basePrompt, lastErr)
				continue
			}
		}

		return map[string]interface{}{
			"success": true,
			"agent":   cfg.Agent,
			"result":  result,
			"raw":     res.Output,
		}, nil
	}

	return nil, fmt.Errorf("ai.operation: agent did not produce a valid response after %d attempt(s): %w", maxRetries+1, lastErr)
}

func retryPrompt(basePrompt string, lastErr error) string {
	return basePrompt + "\n\nYour previous answer was rejected: " + lastErr.Error() +
		"\nRespond again with ONLY the JSON object, no prose, no markdown fences, matching every required field exactly."
}

// buildOperationPrompt appends a rendered input-parameters block and a
// strict output-contract instruction to the base prompt.
func buildOperationPrompt(base string, inputs map[string]interface{}, outputSchema map[string]string) string {
	var b strings.Builder
	b.WriteString(base)

	if len(inputs) > 0 {
		if raw, err := json.MarshalIndent(inputs, "", "  "); err == nil {
			b.WriteString("\n\n## Input parameters\n```json\n")
			b.Write(raw)
			b.WriteString("\n```\n")
		}
	}

	if len(outputSchema) > 0 {
		keys := make([]string, 0, len(outputSchema))
		for k := range outputSchema {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		b.WriteString("\n## Required output\nRespond with ONLY a single JSON object with exactly these fields (no prose, no markdown fences):\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "- %q: %s\n", k, outputSchema[k])
		}
	}

	return b.String()
}

// extractJSONObject finds the first balanced {...} object in s, tolerating
// surrounding prose or ```json fences (LLMs add these despite instructions
// not to).
func extractJSONObject(s string) (string, bool) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", false
	}

	depth := 0
	inString := false
	escaped := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], true
			}
		}
	}
	return "", false
}

func validateOutputSchema(result map[string]interface{}, schema map[string]string) error {
	keys := make([]string, 0, len(schema))
	for k := range schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, field := range keys {
		wantType := schema[field]
		v, ok := result[field]
		if !ok {
			return fmt.Errorf("missing required field %q", field)
		}
		if !jsonTypeMatches(v, wantType) {
			return fmt.Errorf("field %q: expected %s, got %T", field, wantType, v)
		}
	}
	return nil
}

func jsonTypeMatches(v interface{}, want string) bool {
	switch want {
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(float64)
		return ok
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "array":
		_, ok := v.([]interface{})
		return ok
	case "object":
		_, ok := v.(map[string]interface{})
		return ok
	default:
		return true // unknown declared type - don't block on something we can't check
	}
}

func truncateForError(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
