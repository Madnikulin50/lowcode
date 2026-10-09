package aiagent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Evaluating a prompt: run it over a set of golden examples and report how
// many came out right. This is how a prompt change gets judged by evidence
// rather than by feel - run the old and the new version over the same cases
// and compare the pass rates.

type (
	EvalCase struct {
		Name   string
		Inputs map[string]interface{}

		// Expect: output field -> required value. A prompt with expectations
		// is asked for a JSON object with exactly these fields.
		Expect map[string]interface{}

		// Contains: phrases a prose answer must include (case-insensitive)
		Contains []string
	}

	EvalSpec struct {
		Agent string
		Model string
		// Prompt is the instruction text under test
		Prompt    string
		PromptRef string
		Cases     []EvalCase
		// MaxRetries for invalid JSON; default 0, so an evaluation shows the
		// prompt's own reliability rather than the retry loop's
		MaxRetries int

		// OnCase, when set, is told about each case as soon as it is judged -
		// an evaluation of many cases on a local model takes minutes, and a
		// screen wants to show them arriving
		OnCase func(EvalResult)
	}

	EvalResult struct {
		Case     string                 `json:"case"`
		Pass     bool                   `json:"pass"`
		Failures []string               `json:"failures,omitempty"`
		Got      map[string]interface{} `json:"got,omitempty"`
		Response string                 `json:"response,omitempty"`
		Error    string                 `json:"error,omitempty"`

		DurationMs       int64 `json:"durationMs"`
		PromptTokens     int   `json:"promptTokens,omitempty"`
		CompletionTokens int   `json:"completionTokens,omitempty"`
	}

	EvalReport struct {
		PromptRef string       `json:"promptRef,omitempty"`
		Total     int          `json:"total"`
		Passed    int          `json:"passed"`
		PassRate  float64      `json:"passRate"`
		Cases     []EvalResult `json:"cases"`

		DurationMs       int64 `json:"durationMs"`
		PromptTokens     int   `json:"promptTokens,omitempty"`
		CompletionTokens int   `json:"completionTokens,omitempty"`
	}
)

var placeholder = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.]+)\s*\}\}`)

// RenderPrompt fills {{name}} placeholders from inputs. When the text has no
// placeholders the inputs are not lost: the caller appends them as a JSON
// block (see BuildOperationPrompt), the way the workflow steps do.
func RenderPrompt(text string, inputs map[string]interface{}) (rendered string, usedPlaceholders bool) {
	if !placeholder.MatchString(text) {
		return text, false
	}
	return placeholder.ReplaceAllStringFunc(text, func(m string) string {
		key := placeholder.FindStringSubmatch(m)[1]
		if v, ok := inputs[key]; ok {
			if s, isStr := v.(string); isStr {
				return s
			}
			b, _ := json.Marshal(v)
			return string(b)
		}
		return m
	}), true
}

// schemaFromExpect derives the output schema a case's expectations imply.
func schemaFromExpect(expect map[string]interface{}) map[string]string {
	out := make(map[string]string, len(expect))
	for k, v := range expect {
		switch v.(type) {
		case string:
			out[k] = "string"
		case bool:
			out[k] = "boolean"
		case []interface{}:
			out[k] = "array"
		case map[string]interface{}:
			out[k] = "object"
		default:
			out[k] = "number"
		}
	}
	return out
}

// sameValue compares an expected and an actual value as JSON would: numbers
// by value, strings ignoring case and surrounding space (a model that answers
// "High" where "high" was expected has not got it wrong).
func sameValue(want, got interface{}) bool {
	if ws, ok := want.(string); ok {
		gs, ok := got.(string)
		return ok && strings.EqualFold(strings.TrimSpace(ws), strings.TrimSpace(gs))
	}
	wb, err1 := json.Marshal(want)
	gb, err2 := json.Marshal(got)
	return err1 == nil && err2 == nil && string(wb) == string(gb)
}

// RunEval runs every case of spec and reports. Cases run one after another:
// a local model gains nothing from parallel requests and the order keeps the
// report readable.
func RunEval(ctx context.Context, run Runner, spec EvalSpec) *EvalReport {
	started := time.Now()
	report := &EvalReport{PromptRef: spec.PromptRef, Total: len(spec.Cases), Cases: make([]EvalResult, 0, len(spec.Cases))}

	for _, c := range spec.Cases {
		res := evalCase(ctx, run, spec, c)
		if spec.OnCase != nil {
			spec.OnCase(res)
		}
		report.Cases = append(report.Cases, res)
		if res.Pass {
			report.Passed++
		}
		report.PromptTokens += res.PromptTokens
		report.CompletionTokens += res.CompletionTokens
	}

	if report.Total > 0 {
		report.PassRate = float64(report.Passed) / float64(report.Total)
	}
	report.DurationMs = time.Since(started).Milliseconds()
	return report
}

func evalCase(ctx context.Context, run Runner, spec EvalSpec, c EvalCase) EvalResult {
	res := EvalResult{Case: c.Name}
	if res.Case == "" {
		res.Case = "case"
	}
	started := time.Now()
	defer func() { res.DurationMs = time.Since(started).Milliseconds() }()

	text, rendered := RenderPrompt(spec.Prompt, c.Inputs)
	inputs := c.Inputs
	if rendered {
		inputs = nil // already in the text
	}
	agent := spec.Agent
	if agent == "" {
		agent = "workflow-llm"
	}

	switch {
	case len(c.Expect) > 0:
		sr, err := RunStructured(ctx, run, StructuredRequest{
			Agent:        agent,
			Prompt:       text,
			Model:        spec.Model,
			Inputs:       inputs,
			OutputSchema: schemaFromExpect(c.Expect),
			MaxRetries:   spec.MaxRetries,
		})
		res.PromptTokens, res.CompletionTokens = sr.Trace.PromptTokens, sr.Trace.CompletionTokens
		res.Response = sr.Trace.Response
		if err != nil {
			res.Error = err.Error()
			res.Failures = append(res.Failures, "no valid answer: "+err.Error())
			return res
		}

		res.Got = sr.Result
		keys := make([]string, 0, len(c.Expect))
		for k := range c.Expect {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if !sameValue(c.Expect[k], sr.Result[k]) {
				res.Failures = append(res.Failures, fmt.Sprintf("%s: expected %v, got %v", k, c.Expect[k], sr.Result[k]))
			}
		}

	default:
		prompt := text
		if len(inputs) > 0 {
			prompt = BuildOperationPrompt(text, inputs, nil)
		}
		r, err := run(ContextWithModel(ctx, spec.Model), agent, prompt, false)
		if err != nil {
			res.Error = err.Error()
			res.Failures = append(res.Failures, "no answer: "+err.Error())
			return res
		}
		res.PromptTokens, res.CompletionTokens = r.PromptTokens, r.CompletionTokens
		res.Response = Preview(r.Output)
		if r.Error != "" && !r.Success {
			res.Error = r.Error
			res.Failures = append(res.Failures, "agent failed: "+r.Error)
		}
	}

	for _, phrase := range c.Contains {
		if !strings.Contains(strings.ToLower(res.Response), strings.ToLower(phrase)) {
			res.Failures = append(res.Failures, fmt.Sprintf("answer does not mention %q", phrase))
		}
	}
	if len(c.Expect) == 0 && len(c.Contains) == 0 && res.Error == "" {
		res.Failures = append(res.Failures, "the case expects nothing (add expect or contains)")
	}

	res.Pass = len(res.Failures) == 0
	return res
}
