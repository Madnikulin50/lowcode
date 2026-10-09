package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/errors"
)

// The prompt library as the chat and MCP tools use it: parameters arrive as
// strings, results leave as plain data.

// ParsePromptVersions reads "1,3" into [1 3]; empty means "the active one".
func ParsePromptVersions(s string) ([]int, error) {
	var out []int
	for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 {
			return nil, errors.InvalidData("%q is not a version number", part)
		}
		out = append(out, n)
	}
	return out, nil
}

// ParsePromptCases reads a JSON array of golden examples. Empty gives nil, not
// an empty set - "no cases given" and "clear the cases" are different things.
func ParsePromptCases(s string) (*types.PromptCases, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	var cases types.PromptCases
	if err := json.Unmarshal([]byte(s), &cases); err != nil {
		return nil, errors.InvalidData(`cases must be a JSON array like [{"name":"outage","inputs":{"ticket":"server down"},"expect":{"priority":"high"}}]: %v`, err)
	}
	return &cases, nil
}

// PromptBrief is a version without its golden examples' bulk
type PromptBrief struct {
	Handle      string `json:"handle"`
	Version     int    `json:"version"`
	Active      bool   `json:"active"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	Text        string `json:"text,omitempty"`
	Cases       int    `json:"cases"`
	CreatedBy   string `json:"createdBy,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

func briefOf(v *types.PromptVersion, withText bool) PromptBrief {
	b := PromptBrief{
		Handle: v.Handle, Version: v.Version, Active: v.Active,
		Description: v.Description, Note: v.Note, Cases: len(v.Cases),
		CreatedAt: v.CreatedAt.Format("2006-01-02 15:04"),
	}
	if v.CreatedBy != 0 {
		b.CreatedBy = strconv.FormatUint(v.CreatedBy, 10)
	}
	if withText {
		b.Text = v.Text
	}
	return b
}

func promptLib() (*promptLibrary, error) {
	if DefaultPrompts == nil {
		return nil, fmt.Errorf("prompt library not available")
	}
	return DefaultPrompts, nil
}

func ListPrompts(ctx context.Context) ([]PromptSummary, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	return lib.List(ctx)
}

// GetPrompt returns one version (0: the active one), with its text.
func GetPrompt(ctx context.Context, handle string, version int) (*PromptBrief, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Get(ctx, handle, version)
	if err != nil {
		return nil, err
	}
	b := briefOf(v, true)
	return &b, nil
}

// GetPromptWithCases is GetPrompt plus the version's golden examples, for an
// editor that has to show and change them.
func GetPromptWithCases(ctx context.Context, handle string, version int) (*PromptBrief, types.PromptCases, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, nil, err
	}
	v, err := lib.Get(ctx, handle, version)
	if err != nil {
		return nil, nil, err
	}
	b := briefOf(v, true)
	cases := v.Cases
	if cases == nil {
		cases = types.PromptCases{}
	}
	return &b, cases, nil
}

func PromptHistory(ctx context.Context, handle string) ([]PromptBrief, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	vv, err := lib.History(ctx, handle)
	if err != nil {
		return nil, err
	}
	out := make([]PromptBrief, 0, len(vv))
	for _, v := range vv {
		out = append(out, briefOf(v, false))
	}
	return out, nil
}

func SavePromptVersion(ctx context.Context, in SavePrompt) (*PromptBrief, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Save(ctx, in)
	if err != nil {
		return nil, err
	}
	b := briefOf(v, false)
	return &b, nil
}

func ActivatePrompt(ctx context.Context, handle string, version int) (*PromptBrief, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	v, err := lib.Activate(ctx, handle, version)
	if err != nil {
		return nil, err
	}
	b := briefOf(v, false)
	return &b, nil
}

func DeletePrompt(ctx context.Context, handle string) error {
	lib, err := promptLib()
	if err != nil {
		return err
	}
	return lib.Delete(ctx, handle)
}

var (
	evalRunnerMu sync.RWMutex
	evalRunner   aiagent.Runner
)

// SetPromptEvalRunner changes how prompt evaluations call the model; nil
// restores the default (the platform's agent registry).
func SetPromptEvalRunner(run aiagent.Runner) {
	evalRunnerMu.Lock()
	evalRunner = run
	evalRunnerMu.Unlock()
}

func promptEvalRunner() aiagent.Runner {
	evalRunnerMu.RLock()
	defer evalRunnerMu.RUnlock()
	if evalRunner != nil {
		return evalRunner
	}
	return aiagent.RegistryRunner(aiagent.DefaultRegistry())
}

// EvaluatePrompt runs a prompt's golden examples on the real model. versions
// empty: the active one; several: one report each, to compare them.
func EvaluatePrompt(ctx context.Context, handle, versions, agent, model, cases string) (map[string]*aiagent.EvalReport, error) {
	return EvaluatePromptWithProgress(ctx, handle, versions, agent, model, cases, nil)
}

// EvaluatePromptWithProgress is EvaluatePrompt that also reports each case as
// it is judged.
func EvaluatePromptWithProgress(ctx context.Context, handle, versions, agent, model, cases string, onCase func(version int, res aiagent.EvalResult)) (map[string]*aiagent.EvalReport, error) {
	lib, err := promptLib()
	if err != nil {
		return nil, err
	}
	vers, err := ParsePromptVersions(versions)
	if err != nil {
		return nil, err
	}
	parsed, err := ParsePromptCases(cases)
	if err != nil {
		return nil, err
	}
	var override types.PromptCases
	if parsed != nil {
		override = *parsed
	}

	reports, err := lib.Eval(ctx, promptEvalRunner(), handle, vers, agent, model, override, onCase)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*aiagent.EvalReport, len(reports))
	for n, r := range reports {
		out[fmt.Sprintf("v%d", n)] = r
	}
	return out, nil
}
