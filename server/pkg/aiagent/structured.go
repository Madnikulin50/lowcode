// Structured ("AI as a function") calls: named inputs in, a validated JSON
// object out. Shared by every non-interactive surface - rulesgo's
// ai.operation node and the automation workflow ai* functions - so they all
// agree on the prompt contract, the JSON extraction and the retry behaviour.
package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type (
	// Runner executes a single agent turn. Registry.RunAgentConfirmed fits it
	// via RegistryRunner; tests substitute a fake.
	Runner func(ctx context.Context, agent, prompt string, allowMutating bool) (*AgentResult, error)

	StructuredRequest struct {
		Agent  string
		Prompt string
		// PromptRef names the library prompt Prompt came from, for the trace
		PromptRef string
		Model     string // optional per-call model override, see ContextWithModel
		Skill     string // optional skill to follow, see ContextWithSkill

		Inputs       map[string]interface{}
		OutputSchema map[string]string // field -> json type

		AllowMutating bool
		MaxRetries    int // re-ask on invalid JSON; 0 = try once
	}

	StructuredResult struct {
		Result map[string]interface{}
		Raw    string
		// Attempts is the number of LLM turns spent, including retries.
		Attempts int

		// Trace is filled in even when RunStructured fails, so a failing step
		// can still show what was asked and what came back.
		Trace CallTrace
	}
)

// RegistryRunner adapts a Registry to Runner. A nil registry yields an error
// at call time rather than a nil dereference.
func RegistryRunner(r *Registry) Runner {
	return func(ctx context.Context, agent, prompt string, allowMutating bool) (*AgentResult, error) {
		if r == nil {
			return nil, fmt.Errorf("agent registry not available")
		}
		return r.RunAgentConfirmed(ctx, agent, prompt, nil, allowMutating)
	}
}

type modelCtxKey struct{}

// ContextWithModel asks the agents run under ctx to use model (a literal
// model name or a role such as "rulesgo.ai") for this call only, instead of
// the agent's configured one. Empty model leaves ctx unchanged.
func ContextWithModel(ctx context.Context, model string) context.Context {
	if model = strings.TrimSpace(model); model == "" {
		return ctx
	}
	return context.WithValue(ctx, modelCtxKey{}, model)
}

// ModelFromContext returns the per-call model override set by
// ContextWithModel, or "".
func ModelFromContext(ctx context.Context) string { return modelFromContext(ctx) }

func modelFromContext(ctx context.Context) string {
	m, _ := ctx.Value(modelCtxKey{}).(string)
	return m
}

// ConfirmError describes a mutating call an agent attempted without the
// caller having granted allowMutating.
func ConfirmError(res *AgentResult) error {
	names := make([]string, 0, len(res.ConfirmCalls))
	for _, c := range res.ConfirmCalls {
		names = append(names, c.Name)
	}
	return fmt.Errorf("blocked - agent attempted a mutating action requiring confirmation (%s) - enable allowMutating to permit it", strings.Join(names, ", "))
}

// RunStructured asks the agent for a JSON object matching req.OutputSchema,
// re-asking up to req.MaxRetries times when the answer is not valid.
//
// The returned result is never nil: on failure it carries only Trace.
func RunStructured(ctx context.Context, run Runner, req StructuredRequest) (*StructuredResult, error) {
	var (
		started = time.Now()
		out     = &StructuredResult{Trace: CallTrace{Agent: req.Agent}}
		err     error
	)
	defer func() { out.Trace.Finish(started, err) }()

	if req.Agent == "" {
		err = fmt.Errorf("agent is required")
		return out, err
	}
	if req.Prompt == "" {
		err = fmt.Errorf("prompt is required")
		return out, err
	}

	maxRetries := req.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	base := BuildOperationPrompt(req.Prompt, req.Inputs, req.OutputSchema)
	prompt := base
	out.Trace.PromptRef = req.PromptRef
	out.Trace.SetPrompt(base)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		var res *AgentResult
		ReportProgress(ctx, Progress{Attempt: attempt + 1})
		res, err = run(ContextWithSkill(ContextWithModel(ctx, req.Model), req.Skill), req.Agent, prompt, req.AllowMutating)
		if err != nil {
			return out, err
		}
		out.Trace.Add(res)

		if res.ConfirmNeeded {
			err = ConfirmError(res)
			return out, err
		}
		if res.Error != "" && !res.Success {
			err = fmt.Errorf("agent failed: %s", res.Error)
			return out, err
		}

		jsonStr, ok := ExtractJSONObject(res.Output)
		if !ok {
			lastErr = fmt.Errorf("response did not contain a JSON object: %s", truncate(res.Output, 200))
			out.Trace.Reject(lastErr)
			prompt = RetryPrompt(base, lastErr)
			continue
		}

		var parsed map[string]interface{}
		if jerr := json.Unmarshal([]byte(jsonStr), &parsed); jerr != nil {
			lastErr = fmt.Errorf("invalid JSON: %w", jerr)
			out.Trace.Reject(lastErr)
			prompt = RetryPrompt(base, lastErr)
			continue
		}

		if len(req.OutputSchema) > 0 {
			if verr := ValidateOutputSchema(parsed, req.OutputSchema); verr != nil {
				lastErr = verr
				out.Trace.Reject(lastErr)
				prompt = RetryPrompt(base, lastErr)
				continue
			}
		}

		out.Result, out.Raw, out.Attempts = parsed, res.Output, attempt+1
		return out, nil
	}

	err = fmt.Errorf("agent did not produce a valid response after %d attempt(s): %w", maxRetries+1, lastErr)
	return out, err
}

func RetryPrompt(basePrompt string, lastErr error) string {
	return basePrompt + "\n\nYour previous answer was rejected: " + lastErr.Error() +
		"\nRespond again with ONLY the JSON object, no prose, no markdown fences, matching every required field exactly."
}

// BuildOperationPrompt appends a rendered input-parameters block and a
// strict output-contract instruction to the base prompt.
func BuildOperationPrompt(base string, inputs map[string]interface{}, outputSchema map[string]string) string {
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

// ExtractJSONObject finds the first balanced {...} object in s, tolerating
// surrounding prose or ```json fences (LLMs add these despite instructions
// not to).
func ExtractJSONObject(s string) (string, bool) {
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

func ValidateOutputSchema(result map[string]interface{}, schema map[string]string) error {
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
