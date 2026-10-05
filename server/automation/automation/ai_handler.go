package automation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
)

const (
	// aiDefaultAgent is the tool-less built-in agent (pkg/aiagent/defs)
	// used when a step does not name one.
	aiDefaultAgent = "workflow-llm"

	aiDefaultRetries = 1
	aiDefaultTimeout = 5 * time.Minute
)

type (
	aiHandler struct {
		reg aiHandlerRegistry

		// run executes one agent turn. Defaults to the process-wide agent
		// registry, resolved at call time since it is created after the
		// automation service during boot.
		run aiagent.Runner

		// execApproved runs tool calls a human approved (aiRunCalls).
		execApproved aiExecApprovedFunc
	}

	aiExecApprovedFunc func(ctx context.Context, agent string, calls []aiagent.Call) (string, error)
)

func AiHandler(reg aiHandlerRegistry) *aiHandler {
	return newAiHandler(reg,
		func(ctx context.Context, agent, prompt string, allowMutating bool) (*aiagent.AgentResult, error) {
			return aiagent.RegistryRunner(aiagent.DefaultRegistry())(ctx, agent, prompt, allowMutating)
		},
		func(ctx context.Context, agent string, calls []aiagent.Call) (string, error) {
			return aiagent.DefaultRegistry().ExecApproved(ctx, agent, calls)
		},
	)
}

func newAiHandler(reg aiHandlerRegistry, run aiagent.Runner, execApproved aiExecApprovedFunc) *aiHandler {
	// every model call of a step goes through the run's budget, if it has one
	h := &aiHandler{reg: reg, run: aiagent.BudgetedRunner(run), execApproved: execApproved}
	h.register()
	return h
}

// aiStepContext bounds one LLM step. Local models can hang or crawl, and an
// unbounded step would pin the whole workflow session, so every ai* step
// gets a deadline: the author's timeoutSec, or aiDefaultTimeout.
func aiStepContext(ctx context.Context, timeoutSec int64, set bool) (context.Context, context.CancelFunc) {
	d := aiDefaultTimeout
	if set && timeoutSec > 0 {
		d = time.Duration(timeoutSec) * time.Second
	}
	return context.WithTimeout(ctx, d)
}

func aiAgentName(a string) string {
	if a = strings.TrimSpace(a); a != "" {
		return a
	}
	return aiDefaultAgent
}

func aiInputs(in map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func (h aiHandler) ask(ctx context.Context, args *aiAskArgs) (*aiAskResults, error) {
	ctx, cancel := aiStepContext(ctx, args.TimeoutSec, args.hasTimeoutSec)
	defer cancel()

	instruction, promptRef, err := aiagent.ResolvePrompt(ctx, args.Prompt)
	if err != nil {
		return nil, fmt.Errorf("aiAsk: %w", err)
	}

	var (
		started = time.Now()
		agent   = aiAgentName(args.Agent)
		trace   = aiagent.CallTrace{Agent: agent, PromptRef: promptRef}
	)

	prompt := aiagent.BuildOperationPrompt(instruction, aiInputs(args.Inputs), nil)
	trace.SetPrompt(prompt)

	res, err := h.run(aiagent.ContextWithSkill(aiagent.ContextWithModel(ctx, args.Model), args.Skill), agent, prompt, args.AllowMutating)
	if err != nil {
		return nil, fmt.Errorf("aiAsk: %w", err)
	}
	trace.Add(res)
	trace.Finish(started, nil)
	traceVars, err := traceToVars(trace)
	if err != nil {
		return nil, fmt.Errorf("aiAsk: %w", err)
	}

	if res.ConfirmNeeded {
		if !args.DeferConfirm {
			return nil, fmt.Errorf("aiAsk: %w", aiagent.ConfirmError(res))
		}

		// Hand the proposed actions to the workflow instead of failing: the
		// author routes them through an approval prompt and, once a human
		// agrees, executes them with aiRunCalls.
		pending, err := json.Marshal(res.ConfirmCalls)
		if err != nil {
			return nil, fmt.Errorf("aiAsk: %w", err)
		}
		return &aiAskResults{
			NeedsApproval: true,
			PendingCalls:  string(pending),
			Summary:       summarizeCalls(res.ConfirmCalls),
			Trace:         traceVars,
		}, nil
	}
	if !res.Success && res.Error != "" {
		return nil, fmt.Errorf("aiAsk: agent failed: %s", res.Error)
	}

	return &aiAskResults{Text: res.Output, Trace: traceVars}, nil
}

func (h aiHandler) runCalls(ctx context.Context, args *aiRunCallsArgs) (*aiRunCallsResults, error) {
	var calls []aiagent.Call
	if err := json.Unmarshal([]byte(args.Calls), &calls); err != nil {
		return nil, fmt.Errorf("aiRunCalls: calls must be the pendingCalls JSON produced by aiAsk: %w", err)
	}
	if len(calls) == 0 {
		return nil, fmt.Errorf("aiRunCalls: no calls to run")
	}

	out, err := h.execApproved(ctx, aiAgentName(args.Agent), calls)
	if err != nil {
		return nil, fmt.Errorf("aiRunCalls: %w", err)
	}
	return &aiRunCallsResults{Result: out}, nil
}

func traceToVars(t aiagent.CallTrace) (*expr.Vars, error) {
	return expr.NewVars(t.Map())
}

// failedWith adds what the model last said to a failed step's error, which is
// usually the quickest way to see why an answer was rejected.
func failedWith(err error, t aiagent.CallTrace) error {
	if t.Response == "" {
		return err
	}
	last := t.Response
	if r := []rune(last); len(r) > 300 {
		last = string(r[:300]) + "…"
	}
	return fmt.Errorf("%w (last response: %q)", err, last)
}

// summarizeCalls renders proposed calls as one line per call, for the
// approval prompt a human reads.
func summarizeCalls(calls []aiagent.Call) string {
	lines := make([]string, 0, len(calls))
	for _, c := range calls {
		lines = append(lines, c.Name+" "+c.Params)
	}
	return strings.Join(lines, "\n")
}

func (h aiHandler) extract(ctx context.Context, args *aiExtractArgs) (*aiExtractResults, error) {
	ctx, cancel := aiStepContext(ctx, args.TimeoutSec, args.hasTimeoutSec)
	defer cancel()

	retries := aiDefaultRetries
	if args.hasMaxRetries {
		retries = int(args.MaxRetries)
	}

	instruction, promptRef, err := aiagent.ResolvePrompt(ctx, args.Prompt)
	if err != nil {
		return nil, fmt.Errorf("aiExtract: %w", err)
	}

	sr, err := aiagent.RunStructured(ctx, h.run, aiagent.StructuredRequest{
		Agent:         aiAgentName(args.Agent),
		Prompt:        instruction,
		PromptRef:     promptRef,
		Model:         args.Model,
		Skill:         args.Skill,
		Inputs:        aiInputs(args.Inputs),
		OutputSchema:  args.OutputSchema,
		AllowMutating: args.AllowMutating,
		MaxRetries:    retries,
	})
	if err != nil {
		return nil, fmt.Errorf("aiExtract: %w", failedWith(err, sr.Trace))
	}

	vars, err := expr.NewVars(sr.Result)
	if err != nil {
		return nil, fmt.Errorf("aiExtract: %w", err)
	}
	traceVars, err := traceToVars(sr.Trace)
	if err != nil {
		return nil, fmt.Errorf("aiExtract: %w", err)
	}
	return &aiExtractResults{Result: vars, Trace: traceVars}, nil
}

func (h aiHandler) classify(ctx context.Context, args *aiClassifyArgs) (*aiClassifyResults, error) {
	labels := splitLabels(args.Labels)
	if len(labels) < 2 {
		return nil, fmt.Errorf("aiClassify: at least two labels required (comma or newline separated)")
	}

	instruction := strings.TrimSpace(args.Instruction)
	if instruction == "" {
		instruction = "Classify the text."
	}
	instruction, promptRef, err := aiagent.ResolvePrompt(ctx, instruction)
	if err != nil {
		return nil, fmt.Errorf("aiClassify: %w", err)
	}
	instruction += "\nChoose exactly one label from: " + strings.Join(labels, ", ") +
		".\n\"confidence\" is your certainty between 0 and 1; \"reason\" is one short sentence."

	ctx, cancel := aiStepContext(ctx, args.TimeoutSec, args.hasTimeoutSec)
	defer cancel()

	retries := aiDefaultRetries
	if args.hasMaxRetries {
		retries = int(args.MaxRetries)
	}

	sr, err := aiagent.RunStructured(ctx, h.run, aiagent.StructuredRequest{
		Agent:     aiDefaultAgent,
		Prompt:    instruction,
		PromptRef: promptRef,
		Model:     args.Model,
		Skill:     args.Skill,
		Inputs:    map[string]interface{}{"text": args.Text},
		OutputSchema: map[string]string{
			"label":      "string",
			"confidence": "number",
			"reason":     "string",
		},
		MaxRetries: retries,
	})
	if err != nil {
		return nil, fmt.Errorf("aiClassify: %w", failedWith(err, sr.Trace))
	}

	label, _ := sr.Result["label"].(string)
	label = matchLabel(label, labels)
	if label == "" {
		return nil, fmt.Errorf("aiClassify: model answered %q, which is not one of %v", sr.Result["label"], labels)
	}

	conf, _ := sr.Result["confidence"].(float64)
	reason, _ := sr.Result["reason"].(string)
	traceVars, err := traceToVars(sr.Trace)
	if err != nil {
		return nil, fmt.Errorf("aiClassify: %w", err)
	}
	return &aiClassifyResults{Label: label, Confidence: conf, Reason: reason, Trace: traceVars}, nil
}

func splitLabels(s string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == '\n' || r == ';' }) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// matchLabel returns the canonical spelling of got from labels
// (case-insensitive), or "" when it is not a member.
func matchLabel(got string, labels []string) string {
	got = strings.TrimSpace(got)
	for _, l := range labels {
		if strings.EqualFold(l, got) {
			return l
		}
	}
	return ""
}

type (
	// RAGHit is one knowledge-base passage returned by the RAG search hook.
	RAGHit struct {
		Text  string
		Score float64
	}

	// RAGSearchFunc is injected at boot (see compose/mcp/bridge.go): the RAG
	// service lives in compose/service, which automation must not import.
	RAGSearchFunc func(ctx context.Context, namespaceID, query string, topK int) ([]RAGHit, error)
)

var (
	ragSearchMu sync.RWMutex
	ragSearchFn RAGSearchFunc
)

func SetRAGSearch(fn RAGSearchFunc) {
	ragSearchMu.Lock()
	ragSearchFn = fn
	ragSearchMu.Unlock()
}

func (h aiHandler) tool(ctx context.Context, args *aiToolArgs) (*aiToolResults, error) {
	name := strings.TrimSpace(args.Tool)
	for _, t := range aiagent.DefaultCatalog().Resolve("*") {
		if t.Name != name {
			continue
		}
		// Same guardrail as the agent-driven steps: a data-changing tool
		// needs the workflow author's explicit consent.
		if (t.Mutating || aiagent.DefaultNeedsConfirm([]aiagent.Call{{Name: name}})) && !args.AllowMutating {
			return nil, fmt.Errorf("aiTool: %q changes data - enable allowMutating to permit it", name)
		}
		if t.Handler == nil {
			return nil, fmt.Errorf("aiTool: tool %q has no handler", name)
		}
		return &aiToolResults{Result: t.Handler(ctx, args.Params)}, nil
	}
	return nil, fmt.Errorf("aiTool: unknown tool %q", name)
}

func (h aiHandler) ragSearch(ctx context.Context, args *aiRagSearchArgs) (*aiRagSearchResults, error) {
	ragSearchMu.RLock()
	fn := ragSearchFn
	ragSearchMu.RUnlock()
	if fn == nil {
		return nil, fmt.Errorf("aiRagSearch: knowledge base is not available")
	}

	topK := 5
	if args.hasTopK && args.TopK > 0 {
		topK = int(args.TopK)
	}
	hits, err := fn(ctx, args.Namespace, args.Query, topK)
	if err != nil {
		return nil, fmt.Errorf("aiRagSearch: %w", err)
	}

	var b strings.Builder
	for i, hit := range hits {
		fmt.Fprintf(&b, "[Document %d]: %s\n", i+1, hit.Text)
	}
	return &aiRagSearchResults{Text: b.String(), Count: int64(len(hits))}, nil
}

// NewAiHandlerForTest registers the ai* functions backed by a scripted LLM,
// for tests of workflows that use them (see automation/service/ai_templates_test.go).
func NewAiHandlerForTest(reg aiHandlerRegistry, run aiagent.Runner) {
	newAiHandler(reg, run, nil)
}
