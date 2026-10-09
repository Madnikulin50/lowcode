package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/server/automation/automation"
	"github.com/madnikulin50/lowcode/server/automation/types"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/madnikulin50/lowcode/server/pkg/options"
	"github.com/madnikulin50/lowcode/server/pkg/wfevent"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var templatesBoot sync.Once

// templateWorkflowService wires just enough of the automation service to
// convert and run template workflows: the type registry, the ai* and log
// functions (the former backed by a scripted "LLM"), and a parser.
func templateWorkflowService(t *testing.T, llm aiagent.Runner) *workflow {
	templatesBoot.Do(func() {
		Registry().AddTypes(
			&expr.Any{}, &expr.Array{}, &expr.Boolean{}, &expr.ID{}, &expr.Integer{}, &expr.UnsignedInteger{},
			&expr.Float{}, &expr.String{}, &expr.Handle{}, &expr.DateTime{}, &expr.Duration{}, &expr.KV{},
			&expr.KVV{}, &expr.Reader{}, &expr.Vars{}, &expr.Bytes{},
		)
		automation.LogHandler(Registry())
	})

	// the ai* functions are re-registered per test so each gets its own LLM
	newAI := func() { automationAIForTest(llm) }
	newAI()

	svc := &workflow{
		reg:      Registry(),
		parser:   expr.NewParser(),
		log:      zap.NewNop(),
		muxCache: &sync.RWMutex{},
		cache:    map[uint64]*wfCacheItem{},
	}

	// the converter validates steps through the package-level service
	prev := DefaultWorkflow
	DefaultWorkflow = svc
	t.Cleanup(func() { DefaultWorkflow = prev })
	return svc
}

func automationAIForTest(llm aiagent.Runner) {
	automation.NewAiHandlerForTest(Registry(), llm)
}

func TestAITemplates_Catalogue(t *testing.T) {
	infos := AITemplates()
	require.Len(t, infos, 2)
	for _, i := range infos {
		require.NotEmpty(t, i.Name)
		require.NotEmpty(t, i.Description)
		require.NotEmpty(t, i.Trigger)
	}
}

func TestAITemplate_AnomalyExplain_RunsEndToEnd(t *testing.T) {
	var (
		req     = require.New(t)
		ctx     = context.Background()
		prompts []string
	)

	svc := templateWorkflowService(t, func(_ context.Context, agent, prompt string, _ bool) (*aiagent.AgentResult, error) {
		prompts = append(prompts, prompt)
		return &aiagent.AgentResult{Success: true, Output: "The value is 12 standard deviations above normal; check the import job."}, nil
	})

	wf, trigger := buildAnomalyExplain(9)
	wf.ID = 100

	// the trigger a user would enable: high severity only
	req.Equal(wfevent.ResourceAnomalyFinding, trigger.ResourceType)
	req.Equal(wfevent.OnCreate, trigger.EventType)
	req.Equal("finding.severity", trigger.Constraints[0].Name)

	g, issues := Convert(svc, wf)
	for _, is := range issues {
		t.Log(is.String())
	}
	req.Empty(issues, "template must convert cleanly")

	st, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), st))

	ses := newPersistTestService(t, st)
	ses.opt = options.WorkflowOpt{}
	wctx, stop := context.WithCancel(ctx)
	defer stop()
	ses.Watch(wctx)

	// feed it the same variables a real event would
	event := wfevent.New(wfevent.ResourceAnomalyFinding, wfevent.OnCreate, map[string]map[string]interface{}{
		wfevent.PropFinding: {"recordID": "77", "severity": "high", "score": 12.0},
		wfevent.PropRule:    {"detector": "zscore"},
	})
	input, err := event.EncodeVars()
	req.NoError(err)

	wait, _, err := ses.Start(ctx, g, types.SessionStartParams{
		Invoker:      auth.Authenticated(9),
		WorkflowID:   wf.ID,
		EventType:    trigger.EventType,
		ResourceType: trigger.ResourceType,
		Input:        input,
	})
	req.NoError(err)

	wctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	scope, _, _, _, err := wait(wctx2)
	req.NoError(err)

	req.Len(prompts, 1)
	req.Contains(prompts[0], `"severity":"high"`, "the finding is part of the prompt")

	explanation := expr.Must(expr.Select(scope, "explanation")).Get()
	req.Contains(explanation, "import job")
}

func TestAITemplate_RiskEscalationReview_AsksForApproval(t *testing.T) {
	var (
		req = require.New(t)
		ctx = context.Background()
	)

	svc := templateWorkflowService(t, func(_ context.Context, _, prompt string, _ bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true, Output: "Escalate to the risk owner today."}, nil
	})

	wf, trigger := buildRiskEscalationReview(9)
	wf.ID = 101
	req.Equal(wfevent.OnEscalated, trigger.EventType)

	g, issues := Convert(svc, wf)
	for _, is := range issues {
		t.Log(is.String())
	}
	req.Empty(issues, "template must convert cleanly")

	st, err := sqlite.ConnectInMemory(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), st))

	ses := newPersistTestService(t, st)
	wctx, stop := context.WithCancel(ctx)
	defer stop()
	ses.Watch(wctx)

	event := wfevent.New(wfevent.ResourceRiskAssessment, wfevent.OnEscalated, map[string]map[string]interface{}{
		wfevent.PropAssessment: {"level": "high", "previousLevel": "low"},
	})
	input, err := event.EncodeVars()
	req.NoError(err)

	_, sessionID, err := ses.Start(ctx, g, types.SessionStartParams{
		Invoker:    auth.Authenticated(9),
		WorkflowID: wf.ID,
		Input:      input,
	})
	req.NoError(err)

	// it stops at the approval prompt, addressed to the installer
	waitFor(t, "approval prompt", func() bool { return len(ses.states.ids(sessionID)) == 1 })
	pending := ses.PendingPrompts(auth.SetIdentityToContext(ctx, auth.Authenticated(9)))
	req.Len(pending, 1)
	req.Equal("choice", pending[0].Ref)
	msg := expr.Must(expr.Select(pending[0].Payload, "message")).Get()
	req.Contains(msg, "Escalate to the risk owner today.")
	req.Contains(msg, "high")

	// approving takes the "approved" branch to completion
	answer, _ := expr.NewVars(map[string]interface{}{"value": true})
	req.NoError(ses.Resume(sessionID, pending[0].StateID, auth.Authenticated(9), answer))
	waitFor(t, "session to complete", func() bool {
		row, err := store.LookupAutomationSessionByID(ctx, st, sessionID)
		return err == nil && row.Status == types.SessionCompleted
	})
}

// sessionService starts a session service over an in-memory database with its
// spawn loop running, for tests that run a converted workflow.
func sessionService(t *testing.T) (*session, func()) {
	t.Helper()
	ctx := context.Background()

	st, err := sqlite.ConnectInMemory(ctx)
	require.NoError(t, err)
	require.NoError(t, store.Upgrade(ctx, zap.NewNop(), st))
	require.NoError(t, st.TruncateAutomationStates(ctx))

	ses := newPersistTestService(t, st)
	wctx, stop := context.WithCancel(ctx)
	ses.Watch(wctx)
	return ses, stop
}
