package aiagent

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// A budget caps what one run - a workflow session, a rule chain run - may
// spend on the model. Every AI step in the run draws on the same budget, so a
// loop that asks the model on each pass, or an agent that keeps calling tools,
// stops at the limit instead of at the end of the month.
//
// Two things are counted: LLM calls, which every model supports, and tokens,
// which only models that report their usage can be held to. Zero means no
// limit on that count.
//
// The check happens before a call and the charge after it, so a run can end
// slightly over the limit by the cost of its last call - the price of not
// guessing a call's size in advance.

// ErrBudgetExceeded is returned by a step that cannot run because its run has
// used up its budget.
type ErrBudgetExceeded struct {
	What        string // "tokens" or "LLM calls"
	Used, Limit int
}

func (e *ErrBudgetExceeded) Error() string {
	return fmt.Sprintf("AI budget of this run is used up: %d of %d %s spent", e.Used, e.Limit, e.What)
}

type Budget struct {
	mu                sync.Mutex
	maxTokens, maxCal int
	tokens, calls     int
}

// NewBudget returns a budget, or nil (unlimited) when both limits are zero.
// A nil *Budget is valid and never limits anything.
func NewBudget(maxTokens, maxCalls int) *Budget {
	if maxTokens <= 0 && maxCalls <= 0 {
		return nil
	}
	return &Budget{maxTokens: maxTokens, maxCal: maxCalls}
}

// BudgetFromEnv is the platform-wide default, for runs that set none:
// AI_MAX_TOKENS_PER_RUN and AI_MAX_LLM_CALLS_PER_RUN.
func BudgetFromEnv() *Budget {
	return NewBudget(envInt("AI_MAX_TOKENS_PER_RUN"), envInt("AI_MAX_LLM_CALLS_PER_RUN"))
}

func envInt(name string) int {
	n, _ := strconv.Atoi(os.Getenv(name))
	return n
}

// Check says whether another call may start.
func (b *Budget) Check() error {
	if b == nil {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.maxCal > 0 && b.calls >= b.maxCal {
		return &ErrBudgetExceeded{What: "LLM calls", Used: b.calls, Limit: b.maxCal}
	}
	if b.maxTokens > 0 && b.tokens >= b.maxTokens {
		return &ErrBudgetExceeded{What: "tokens", Used: b.tokens, Limit: b.maxTokens}
	}
	return nil
}

// Charge records a finished call. A call that reports no LLM calls of its own
// still counts as one: it reached the model.
func (b *Budget) Charge(llmCalls, tokens int) {
	if b == nil {
		return
	}
	if llmCalls < 1 {
		llmCalls = 1
	}
	b.mu.Lock()
	b.calls += llmCalls
	b.tokens += tokens
	b.mu.Unlock()
}

// Spent returns what the run has used so far.
func (b *Budget) Spent() (llmCalls, tokens int) {
	if b == nil {
		return 0, 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls, b.tokens
}

type budgetCtxKey struct{}

// ContextWithBudget makes b the budget of every AI step run under ctx. Use
// the context of the whole run, so all its steps share one budget.
func ContextWithBudget(ctx context.Context, b *Budget) context.Context {
	if b == nil {
		return ctx
	}
	return context.WithValue(ctx, budgetCtxKey{}, b)
}

func BudgetFromContext(ctx context.Context) *Budget {
	b, _ := ctx.Value(budgetCtxKey{}).(*Budget)
	return b
}

// BudgetedRunner makes a Runner respect the budget on its context: it refuses
// to call the model once the budget is used up, and charges what each call
// cost. Without a budget on the context it changes nothing.
func BudgetedRunner(run Runner) Runner {
	return func(ctx context.Context, agent, prompt string, allowMutating bool) (*AgentResult, error) {
		b := BudgetFromContext(ctx)
		if err := b.Check(); err != nil {
			return nil, err
		}

		res, err := run(ctx, agent, prompt, allowMutating)
		if res != nil {
			b.Charge(res.LLMCalls, res.PromptTokens+res.CompletionTokens)
		} else if err == nil {
			b.Charge(1, 0)
		}
		return res, err
	}
}
