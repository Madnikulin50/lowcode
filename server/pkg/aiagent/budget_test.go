package aiagent

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBudget_NilAndZeroAreUnlimited(t *testing.T) {
	var none *Budget
	require.NoError(t, none.Check())
	none.Charge(5, 5000)
	calls, tokens := none.Spent()
	require.Zero(t, calls+tokens)

	require.Nil(t, NewBudget(0, 0), "no limits, no budget")
	require.Nil(t, NewBudget(-1, -1))
}

func TestBudget_CallLimit(t *testing.T) {
	b := NewBudget(0, 3)
	for i := 0; i < 3; i++ {
		require.NoError(t, b.Check(), "call %d", i+1)
		b.Charge(1, 0)
	}

	err := b.Check()
	var exceeded *ErrBudgetExceeded
	require.True(t, errors.As(err, &exceeded))
	require.Equal(t, "LLM calls", exceeded.What)
	require.Equal(t, 3, exceeded.Used)
	require.Equal(t, 3, exceeded.Limit)
	require.Contains(t, err.Error(), "3 of 3 LLM calls")
}

func TestBudget_TokenLimitAllowsTheCallThatCrossesIt(t *testing.T) {
	b := NewBudget(1000, 0)
	require.NoError(t, b.Check())
	b.Charge(1, 900)
	require.NoError(t, b.Check(), "900 of 1000 is still within budget")

	b.Charge(1, 400) // this call took it over: that is allowed, the next is not
	err := b.Check()
	var exceeded *ErrBudgetExceeded
	require.True(t, errors.As(err, &exceeded))
	require.Equal(t, "tokens", exceeded.What)
	require.Equal(t, 1300, exceeded.Used)
}

func TestBudget_ACallThatReportsNoLLMCallsStillCounts(t *testing.T) {
	b := NewBudget(0, 2)
	b.Charge(0, 0)
	b.Charge(0, 0)
	calls, _ := b.Spent()
	require.Equal(t, 2, calls)
	require.Error(t, b.Check())
}

func TestBudget_ConcurrentSteps(t *testing.T) {
	b := NewBudget(0, 0+1000)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Check()
			b.Charge(2, 10)
		}()
	}
	wg.Wait()
	calls, tokens := b.Spent()
	require.Equal(t, 100, calls)
	require.Equal(t, 500, tokens)
}

func TestBudgetedRunner(t *testing.T) {
	calls := 0
	run := BudgetedRunner(func(context.Context, string, string, bool) (*AgentResult, error) {
		calls++
		return &AgentResult{Success: true, LLMCalls: 2, PromptTokens: 60, CompletionTokens: 15}, nil
	})

	// no budget on the context: nothing is limited, nothing is counted
	_, err := run(context.Background(), "a", "p", false)
	require.NoError(t, err)

	b := NewBudget(100, 0)
	ctx := ContextWithBudget(context.Background(), b)

	_, err = run(ctx, "a", "p", false)
	require.NoError(t, err)
	_, tokens := b.Spent()
	require.Equal(t, 75, tokens)

	_, err = run(ctx, "a", "p", false) // 75 < 100: allowed, ends at 150
	require.NoError(t, err)

	before := calls
	res, err := run(ctx, "a", "p", false) // 150 >= 100: refused, model not called
	require.Nil(t, res)
	var exceeded *ErrBudgetExceeded
	require.True(t, errors.As(err, &exceeded))
	require.Equal(t, before, calls, "the model must not be called once the budget is gone")
}

func TestBudgetedRunner_FailedCallsDoNotCharge(t *testing.T) {
	run := BudgetedRunner(func(context.Context, string, string, bool) (*AgentResult, error) {
		return nil, errors.New("model down")
	})
	b := NewBudget(0, 5)
	_, err := run(ContextWithBudget(context.Background(), b), "a", "p", false)
	require.Error(t, err)
	calls, _ := b.Spent()
	require.Zero(t, calls, "a call that never reached a result costs nothing")
}

func TestBudgetFromEnv(t *testing.T) {
	t.Setenv("AI_MAX_TOKENS_PER_RUN", "")
	t.Setenv("AI_MAX_LLM_CALLS_PER_RUN", "")
	require.Nil(t, BudgetFromEnv())

	t.Setenv("AI_MAX_LLM_CALLS_PER_RUN", "7")
	b := BudgetFromEnv()
	require.NotNil(t, b)
	for i := 0; i < 7; i++ {
		b.Charge(1, 0)
	}
	require.Error(t, b.Check())

	t.Setenv("AI_MAX_TOKENS_PER_RUN", "abc") // junk means no limit, not a crash
	t.Setenv("AI_MAX_LLM_CALLS_PER_RUN", "")
	require.Nil(t, BudgetFromEnv())
}
