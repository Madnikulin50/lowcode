package aiagent

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"
)

// CallTrace describes one AI step execution well enough to answer "what did
// the model see, what did it say, and what did it cost" after the fact. It is
// returned by the workflow ai* functions and the rule chain ai nodes, so it
// lands in the session stacktrace / chain run record with no extra plumbing.
//
// Prompt and Response are truncated (TracePreviewLimit): the full text can be
// large and may hold personal data that has no business sitting in every
// execution log.
type CallTrace struct {
	Agent string `json:"agent"`
	Model string `json:"model,omitempty"`

	// Attempts is how many times the step asked the agent (retries after an
	// invalid answer count); LLMCalls is how many model turns that took in
	// total, tool round-trips included.
	Attempts int `json:"attempts"`
	LLMCalls int `json:"llmCalls,omitempty"`

	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`

	DurationMs int64 `json:"durationMs"`

	// Tools the agent called, in order
	Tools []string `json:"tools,omitempty"`

	// Skills the agent loaded, as "handle@version", in order
	Skills []string `json:"skills,omitempty"`

	// PromptRef is the library prompt used ("handle@version"), when the step
	// pointed at one instead of carrying its instruction inline
	PromptRef string `json:"promptRef,omitempty"`

	Prompt   string `json:"prompt,omitempty"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`

	// Rejected lists why earlier attempts were thrown away (invalid JSON,
	// missing field, ...), oldest first.
	Rejected []string `json:"rejected,omitempty"`
}

// TracePreviewLimit bounds Prompt and Response in a trace, in characters.
const TracePreviewLimit = 2000

// Add folds one agent turn into the trace.
func (t *CallTrace) Add(res *AgentResult) {
	if res == nil {
		return
	}
	tools := make([]string, 0, 2)
	for _, s := range res.Steps {
		for _, tc := range s.Tools {
			tools = append(tools, tc.Name)
		}
	}
	t.AddUsage(res.Model, res.LLMCalls, res.PromptTokens, res.CompletionTokens, tools, res.Output)
	for _, ref := range res.Skills {
		t.AddSkill(ref)
	}
}

// AddSkill notes a skill the run loaded (once per attempt chain)
func (t *CallTrace) AddSkill(ref string) {
	for _, have := range t.Skills {
		if have == ref {
			return
		}
	}
	t.Skills = append(t.Skills, ref)
}

// AddUsage folds one agent turn into the trace from its raw figures, for
// callers that do not hold an AgentResult (the rule chain AI bridge).
func (t *CallTrace) AddUsage(model string, llmCalls, promptTokens, completionTokens int, tools []string, output string) {
	t.Attempts++
	if model != "" {
		t.Model = model
	}
	t.LLMCalls += llmCalls
	t.PromptTokens += promptTokens
	t.CompletionTokens += completionTokens
	t.Tools = append(t.Tools, tools...)
	t.Response = Preview(output)
}

// Reject records why an attempt's answer was discarded.
func (t *CallTrace) Reject(err error) {
	if err != nil {
		t.Rejected = append(t.Rejected, Preview(err.Error()))
	}
}

// SetPrompt records what was sent (the first attempt's prompt is the
// meaningful one; retries only append a rejection note).
func (t *CallTrace) SetPrompt(prompt string) {
	if t.Prompt == "" {
		t.Prompt = Preview(prompt)
	}
}

// Finish stamps the elapsed time and, for a failed step, its error.
func (t *CallTrace) Finish(started time.Time, err error) {
	t.DurationMs = time.Since(started).Milliseconds()
	if err != nil {
		t.Error = err.Error()
	}
}

// Map is the trace as plain JSON-like data, for step results and node outputs.
func (t CallTrace) Map() map[string]interface{} {
	raw, err := json.Marshal(t)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	out := map[string]interface{}{}
	_ = json.Unmarshal(raw, &out)
	return out
}

// Preview cuts s to TracePreviewLimit characters (not bytes), marking the cut.
func Preview(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= TracePreviewLimit {
		return s
	}
	r := []rune(s)
	return string(r[:TracePreviewLimit]) + "…"
}
