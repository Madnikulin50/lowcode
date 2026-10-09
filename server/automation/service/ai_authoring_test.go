package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
	"github.com/stretchr/testify/require"
)

func compile(t *testing.T, svc *workflow, raw string) *AIFlowCompiled {
	t.Helper()
	spec, err := ParseAIFlow(raw)
	require.NoError(t, err)
	return CompileAIFlow(svc, spec, 9)
}

func issuesText(c *AIFlowCompiled) string { return strings.Join(c.Issues, "\n") }

const anomalyFlow = `{
  "name": "Explain anomalies",
  "description": "Explains high-severity findings",
  "trigger": {"resourceType": "anomaly:finding", "eventType": "onCreate",
              "constraints": [{"name": "finding.severity", "op": "=", "values": ["high"]}]},
  "steps": [
    {"id": "explain", "function": "aiAsk",
     "args": {"prompt": "=\"Explain this finding: \" + toPlainJSON(finding)", "timeoutSec": 30},
     "results": {"explanation": "text"}},
    {"id": "log", "function": "logInfo", "args": {"message": "=explanation"}}
  ],
  "flow": [{"from": "explain", "to": "log"}]
}`

func TestCompileAIFlow_ValidFlowConvertsAndRuns(t *testing.T) {
	var prompts []string
	svc := templateWorkflowService(t, func(_ context.Context, _, prompt string, _ bool) (*aiagent.AgentResult, error) {
		prompts = append(prompts, prompt)
		return &aiagent.AgentResult{Success: true, Output: "Check the import job."}, nil
	})
	req := require.New(t)

	c := compile(t, svc, anomalyFlow)
	req.Empty(c.Issues, issuesText(c))

	wf := c.Workflow
	req.Equal("explain_anomalies", wf.Handle, "handle derived from the name")
	req.Equal(uint64(9), wf.RunAs)
	req.False(wf.Enabled, "nothing starts reacting until a person enables it")
	req.Len(wf.Steps, 2)
	req.Equal(uint64(1), wf.Steps[0].ID)
	req.Len(wf.Paths, 1)
	req.Equal(wfevent.ResourceAnomalyFinding, c.Trigger.ResourceType)
	req.Equal("finding.severity", c.Trigger.Constraints[0].Name)

	// and it really runs
	wf.ID = 300
	g, issues := Convert(svc, wf)
	req.Empty(issues)

	ctx := context.Background()
	ses, stop := sessionService(t)
	defer stop()

	input, err := wfevent.New(wfevent.ResourceAnomalyFinding, wfevent.OnCreate, map[string]map[string]interface{}{
		wfevent.PropFinding: {"severity": "high", "recordID": "42"},
	}).EncodeVars()
	req.NoError(err)

	wait, _, err := ses.Start(ctx, g, types.SessionStartParams{Invoker: auth.Authenticated(9), WorkflowID: wf.ID, Input: input})
	req.NoError(err)

	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	scope, _, _, _, err := wait(wctx)
	req.NoError(err)

	req.Len(prompts, 1)
	req.Contains(prompts[0], `"recordID":"42"`, "the expression saw the event's variables")
	req.Equal("Check the import job.", expr.Must(expr.Select(scope, "explanation")).Get())
}

func TestCompileAIFlow_ApprovalBranchRuns(t *testing.T) {
	svc := templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true, Output: "Block the account."}, nil
	})
	req := require.New(t)

	c := compile(t, svc, `{
	  "name": "Review with approval",
	  "steps": [
	    {"id": "propose", "function": "aiAsk", "args": {"prompt": "Propose a response"}, "results": {"proposal": "text"}},
	    {"id": "ask", "prompt": "choice",
	     "args": {"owner": 9, "message": "=\"Proposal: \" + proposal", "confirmButtonLabel": "Do it", "rejectButtonLabel": "Skip"},
	     "results": {"approved": "value"}},
	    {"id": "decide", "gateway": "excl"},
	    {"id": "yes", "function": "logInfo", "args": {"message": "approved"}},
	    {"id": "no", "function": "logInfo", "args": {"message": "declined"}}
	  ],
	  "flow": [
	    {"from": "propose", "to": "ask"},
	    {"from": "ask", "to": "decide"},
	    {"from": "decide", "to": "yes", "when": "=approved == true"},
	    {"from": "decide", "to": "no", "when": "approved != true"}
	  ]
	}`)
	req.Empty(c.Issues, issuesText(c))
	req.Nil(c.Trigger, "no trigger described, none created")

	c.Workflow.ID = 301
	g, issues := Convert(svc, c.Workflow)
	req.Empty(issues)

	ses, stop := sessionService(t)
	defer stop()
	ctx := context.Background()

	_, sessionID, err := ses.Start(ctx, g, types.SessionStartParams{Invoker: auth.Authenticated(9), WorkflowID: 301, Input: &expr.Vars{}})
	req.NoError(err)

	waitFor(t, "approval prompt", func() bool { return len(ses.states.ids(sessionID)) == 1 })
	pending := ses.PendingPrompts(auth.SetIdentityToContext(ctx, auth.Authenticated(9)))
	req.Len(pending, 1, "the prompt is addressed to the owner given in args")
	req.Equal("choice", pending[0].Ref)
	req.Contains(expr.Must(expr.Select(pending[0].Payload, "message")).Get(), "Block the account.")

	answer, _ := expr.NewVars(map[string]interface{}{"value": true})
	req.NoError(ses.Resume(sessionID, pending[0].StateID, auth.Authenticated(9), answer))
	waitFor(t, "the session to complete", func() bool { return len(ses.states.ids(sessionID)) == 0 })
}

func TestCompileAIFlow_ReportsWhatIsWrong(t *testing.T) {
	svc := templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true}, nil
	})

	cases := map[string]struct {
		flow string
		want []string // every phrase must appear in the issues
	}{
		"no name or steps": {`{}`, []string{"name is required", "steps is empty"}},
		"unknown function": {
			`{"name":"x","steps":[{"id":"a","function":"aiThink"}],"flow":[]}`,
			[]string{`unknown function "aiThink"`, "aiAsk", "logInfo"},
		},
		"unknown argument": {
			`{"name":"x","steps":[{"id":"a","function":"aiAsk","args":{"prompt":"p","temperature":1}}],"flow":[]}`,
			[]string{`no argument "temperature"`, "Its arguments:", "prompt"},
		},
		"missing required argument": {
			`{"name":"x","steps":[{"id":"a","function":"aiExtract","args":{"prompt":"p"}}],"flow":[]}`,
			[]string{`needs the argument "outputSchema"`},
		},
		"unknown result": {
			`{"name":"x","steps":[{"id":"a","function":"aiAsk","args":{"prompt":"p"},"results":{"v":"answer"}}],"flow":[]}`,
			[]string{`no result "answer"`, "text"},
		},
		"step kinds mixed": {
			`{"name":"x","steps":[{"id":"a","function":"aiAsk","gateway":"excl","args":{"prompt":"p"}}],"flow":[]}`,
			[]string{"exactly one of function, prompt or gateway"},
		},
		"no kind": {
			`{"name":"x","steps":[{"id":"a"}],"flow":[]}`,
			[]string{"exactly one of function, prompt or gateway (it sets 0)"},
		},
		"duplicate and missing ids": {
			`{"name":"x","steps":[{"id":"a","gateway":"fork"},{"id":"a","gateway":"join"},{"gateway":"fork"}],"flow":[]}`,
			[]string{`step id "a" is used twice`, "has no id"},
		},
		"bad flow links": {
			`{"name":"x","steps":[{"id":"a","gateway":"fork"}],"flow":[{"from":"a","to":"nope"},{"from":"ghost","to":"a"},{"from":"a","to":"a"}]}`,
			[]string{`unknown step "nope"`, `unknown step "ghost"`, "links to itself", "steps: a"},
		},
		"unknown gateway": {
			`{"name":"x","steps":[{"id":"a","gateway":"maybe"}],"flow":[]}`,
			[]string{"gateway must be one of excl, incl, fork, join"},
		},
		"unknown prompt": {
			`{"name":"x","steps":[{"id":"a","prompt":"dance","args":{"message":"m"}}],"flow":[]}`,
			[]string{`unknown prompt "dance"`, "choice"},
		},
		"prompt without message": {
			`{"name":"x","steps":[{"id":"a","prompt":"choice"}],"flow":[]}`,
			[]string{"needs a message"},
		},
		"trigger half given": {
			`{"name":"x","steps":[{"id":"a","gateway":"fork"}],"flow":[],"trigger":{"resourceType":"anomaly:finding"}}`,
			[]string{"both resourceType and eventType"},
		},
		"gateway with args": {
			`{"name":"x","steps":[{"id":"a","gateway":"excl","args":{"k":"v"}}],"flow":[]}`,
			[]string{"gateway takes no args"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := issuesText(compile(t, svc, c.flow))
			require.NotEmpty(t, got)
			for _, phrase := range c.want {
				require.Contains(t, got, phrase)
			}
		})
	}
}

func TestCompileAIFlow_ConverterProblemsAreReported(t *testing.T) {
	svc := templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true}, nil
	})

	// every name checks out, but the expression is not valid: only the
	// platform's own converter can tell
	c := compile(t, svc, `{"name":"x","steps":[{"id":"a","function":"aiAsk","args":{"prompt":"=toPlainJSON("}}],"flow":[]}`)
	require.NotEmpty(t, c.Issues, "a broken expression must not pass as valid")
	require.Contains(t, issuesText(c), `step "a":`, "issues name the step of the description, not a number")
	require.Contains(t, issuesText(c), "check the expression syntax")

	// a path condition that does not parse
	c = compile(t, svc, `{"name":"x","steps":[
	  {"id":"g","gateway":"excl"},
	  {"id":"a","function":"logInfo","args":{"message":"m"}}],
	  "flow":[{"from":"g","to":"a","when":"=approved =="}]}`)
	require.NotEmpty(t, c.Issues, "a broken condition must not pass as valid")
	t.Log(issuesText(c))
}

func TestParseAIFlow_RejectsMisspeltFields(t *testing.T) {
	_, err := ParseAIFlow(`{"name":"x","step":[]}`)
	require.ErrorContains(t, err, "unknown field")
	require.ErrorContains(t, err, "expected fields")

	_, err = ParseAIFlow(`not json`)
	require.Error(t, err)

	spec, err := ParseAIFlow(`{"name":"x","steps":[{"id":"a","function":"logInfo"}],"flow":[]}`)
	require.NoError(t, err)
	require.Equal(t, "x", spec.Name)
}

func TestHandleFromName(t *testing.T) {
	for in, want := range map[string]string{
		"Explain anomalies":   "explain_anomalies",
		"  Risk: review (v2)": "risk_review_v2",
		"123 go":              "flow_123_go",
		"":                    "flow_",
		"!!!":                 "flow_",
		"Ünï":                 "n",
	} {
		require.Equal(t, want, handleFromName(in), in)
	}
	require.LessOrEqual(t, len(handleFromName(strings.Repeat("a", 200))), 64)
}

func TestAIFunctionCatalog(t *testing.T) {
	templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) { return nil, nil })

	cat := AIFunctionCatalog()
	for _, want := range []string{"aiAsk", "aiExtract", "aiClassify", "aiRunCalls", "logInfo", "arg prompt (String, required)", "results: text", "choice", "excl"} {
		require.Contains(t, cat, want)
	}
	require.NotContains(t, cat, "corredorExec", "only what an assistant should offer")
	require.NotContains(t, cat, "httpRequestSend")
}

const threeAsks = `{
  "name": "Three questions",
  "steps": [
    {"id": "a", "function": "aiAsk", "args": {"prompt": "one"}, "results": {"one": "text"}},
    {"id": "b", "function": "aiAsk", "args": {"prompt": "two"}, "results": {"two": "text"}},
    {"id": "c", "function": "aiAsk", "args": {"prompt": "three"}, "results": {"three": "text"}}
  ],
  "flow": [{"from": "a", "to": "b"}, {"from": "b", "to": "c"}]
}`

func runThreeAsks(t *testing.T, budget *types.AIBudget) (calls int, err error) {
	t.Helper()
	svc := templateWorkflowService(t, func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		calls++
		return &aiagent.AgentResult{Success: true, Output: "ok", LLMCalls: 1}, nil
	})

	c := compile(t, svc, threeAsks)
	require.Empty(t, c.Issues, issuesText(c))
	wf := c.Workflow
	wf.ID = 400
	wf.Meta.AIBudget = budget
	svc.cache[wf.ID] = &wfCacheItem{wf: wf}

	g, issues := Convert(svc, wf)
	require.Empty(t, issues)

	ses, stop := sessionService(t)
	defer stop()

	// the session service builds the budget the way it does for a real start
	wait, _, err := ses.Start(context.Background(), g, types.SessionStartParams{Invoker: auth.Authenticated(9), WorkflowID: wf.ID, Input: &expr.Vars{}})
	require.NoError(t, err)

	wctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, _, _, err = wait(wctx)
	return calls, err
}

func TestWorkflowAIBudget_StopsTheSession(t *testing.T) {
	calls, err := runThreeAsks(t, &types.AIBudget{MaxLLMCalls: 2})

	require.Error(t, err, "the third AI step must not run")
	require.Contains(t, err.Error(), "budget")
	require.Contains(t, err.Error(), "2 of 2 LLM calls")
	require.Equal(t, 2, calls, "the model is not called once the budget is gone")
}

func TestWorkflowAIBudget_UnlimitedByDefault(t *testing.T) {
	calls, err := runThreeAsks(t, nil)
	require.NoError(t, err)
	require.Equal(t, 3, calls)
}

func TestWorkflowAIBudget_EnvDefaultAndOwnLimitsWin(t *testing.T) {
	t.Setenv("AI_MAX_LLM_CALLS_PER_RUN", "1")

	calls, err := runThreeAsks(t, nil)
	require.Error(t, err, "a workflow with no limits of its own gets the platform default")
	require.Equal(t, 1, calls)

	calls, err = runThreeAsks(t, &types.AIBudget{MaxLLMCalls: 3})
	require.NoError(t, err, "its own limits replace the default")
	require.Equal(t, 3, calls)
}

func TestWorkflowAIBudget_EachSessionStartsFresh(t *testing.T) {
	svc := &workflow{muxCache: &sync.RWMutex{}, cache: map[uint64]*wfCacheItem{
		1: {wf: &types.Workflow{Meta: &types.WorkflowMeta{AIBudget: &types.AIBudget{MaxLLMCalls: 1}}}},
	}}

	a, b := svc.aiBudgetFor(1), svc.aiBudgetFor(1)
	require.NotSame(t, a, b, "one run's spending must not eat into the next run's budget")
	a.Charge(1, 0)
	require.Error(t, a.Check())
	require.NoError(t, b.Check())

	require.Nil(t, svc.aiBudgetFor(99), "unknown workflow, no default: unlimited")
}
