package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/automation/automation"
	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/expr"
	"github.com/stretchr/testify/require"
)

// bootAIFunctions registers the ai* functions backed by a scripted model.
func bootAIFunctions(t *testing.T, answer func(allowMutating bool, prompt string) *aiagent.AgentResult) {
	t.Helper()
	service.Registry().AddTypes(&expr.Any{}, &expr.Boolean{}, &expr.String{}, &expr.Integer{}, &expr.Float{}, &expr.KV{}, &expr.Vars{})
	automation.NewAiHandlerForTest(service.Registry(), func(_ context.Context, _, prompt string, allowMutating bool) (*aiagent.AgentResult, error) {
		return answer(allowMutating, prompt), nil
	})
}

func TestRunAIStep_Ask(t *testing.T) {
	var sawAllow = true
	bootAIFunctions(t, func(allow bool, prompt string) *aiagent.AgentResult {
		sawAllow = allow
		return &aiagent.AgentResult{Success: true, Output: "pong: " + prompt[:4], Model: "m", PromptTokens: 9}
	})

	out, err := runAIStep(context.Background(), "aiAsk", map[string]interface{}{
		"prompt":        "ping the server",
		"allowMutating": true, // asked for, but a test must refuse
	})
	require.NoError(t, err)

	require.False(t, sawAllow, "a test run must never allow data changes")
	require.Equal(t, "pong: ping", out["text"])

	trace, ok := out["trace"].(map[string]interface{})
	require.True(t, ok, "the trace is part of the results: %v", out)
	require.Equal(t, "m", trace["model"])
	require.Contains(t, trace["prompt"], "ping the server")
}

func TestRunAIStep_ProposedChangesAreReportedNotMade(t *testing.T) {
	bootAIFunctions(t, func(bool, string) *aiagent.AgentResult {
		return &aiagent.AgentResult{ConfirmNeeded: true, ConfirmCalls: []aiagent.Call{{Name: "delete_module", Params: `{"id":"1"}`}}}
	})

	out, err := runAIStep(context.Background(), "aiAsk", map[string]interface{}{"prompt": "clean up"})
	require.NoError(t, err, "the test shows what the step would do instead of failing")
	require.Equal(t, true, out["needsApproval"])
	require.Contains(t, out["summary"], "delete_module")
}

func TestRunAIStep_Extract(t *testing.T) {
	bootAIFunctions(t, func(bool, string) *aiagent.AgentResult {
		return &aiagent.AgentResult{Success: true, Output: `{"risk": 4}`}
	})

	out, err := runAIStep(context.Background(), "aiExtract", map[string]interface{}{
		"prompt":       "score",
		"outputSchema": map[string]interface{}{"risk": "number"},
	})
	require.NoError(t, err)
	require.EqualValues(t, 4, out["result"].(map[string]interface{})["risk"])
}

func TestRunAIStep_Validation(t *testing.T) {
	bootAIFunctions(t, func(bool, string) *aiagent.AgentResult { return &aiagent.AgentResult{Success: true} })

	_, err := runAIStep(context.Background(), "aiAsk", map[string]interface{}{})
	require.ErrorContains(t, err, "prompt is required")

	_, err = runAIStep(context.Background(), "aiNope", nil)
	require.ErrorContains(t, err, "not registered")
}

func TestAIStepTest_OnlyTestsHarmlessFunctions(t *testing.T) {
	router := chi.NewRouter()
	MountAIStepRoutes(router)

	for _, ref := range []string{"aiTool", "aiRunCalls", "logInfo", "corredorExec"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/"+ref+"/test", bytes.NewBufferString(`{"args":{}}`)))
		require.Contains(t, rec.Body.String(), "cannot be tested on its own", ref)
	}
}

func TestAIStepTest_RequiresAWorkflowBuilder(t *testing.T) {
	router := chi.NewRouter()
	MountAIStepRoutes(router)

	// no access control wired in (or a caller who may not create workflows)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/aiAsk/test", bytes.NewBufferString(`{"args":{"prompt":"x"}}`)))
	require.False(t, strings.Contains(rec.Body.String(), `"text"`), "an unauthorised caller must not reach the model: %s", rec.Body.String())
}

// sseEvents reads a server-sent-events body into its JSON payloads.
func sseEvents(t *testing.T, body string) []map[string]interface{} {
	t.Helper()
	var out []map[string]interface{}
	for _, block := range strings.Split(strings.TrimSpace(body), "\n\n") {
		line := strings.TrimPrefix(strings.TrimSpace(block), "data: ")
		var ev map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &ev), block)
		out = append(out, ev)
	}
	return out
}

func allowTesting(t *testing.T, allowed bool) {
	t.Helper()
	prev := canTestAIStep
	canTestAIStep = func(context.Context) bool { return allowed }
	t.Cleanup(func() { canTestAIStep = prev })
}

func TestAIStepTest_StreamsProgressThenTheResult(t *testing.T) {
	allowTesting(t, true)
	bootAIFunctions(t, func(bool, string) *aiagent.AgentResult { return nil })

	// a model that streams its answer, as the agent runtime does when watched
	automation.NewAiHandlerForTest(service.Registry(), func(ctx context.Context, _, _ string, _ bool) (*aiagent.AgentResult, error) {
		aiagent.ReportProgress(ctx, aiagent.Progress{Status: "warming"})
		aiagent.ReportProgress(ctx, aiagent.Progress{Token: "Hel"})
		aiagent.ReportProgress(ctx, aiagent.Progress{Token: "lo"})
		return &aiagent.AgentResult{Success: true, Output: "Hello", PromptTokens: 4}, nil
	})

	router := chi.NewRouter()
	MountAIStepRoutes(router)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/aiAsk/test?stream=1", bytes.NewBufferString(`{"args":{"prompt":"say hello"}}`)))

	require.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))
	events := sseEvents(t, rec.Body.String())

	require.Equal(t, "warming", events[0]["status"])
	require.Equal(t, "Hel", events[1]["token"])
	require.Equal(t, "lo", events[2]["token"])

	last := events[len(events)-1]
	require.Equal(t, true, last["done"])
	require.Nil(t, last["error"])
	result := last["result"].(map[string]interface{})
	require.Equal(t, "Hello", result["text"])
	require.NotNil(t, result["trace"])

	for _, ev := range events[:len(events)-1] {
		require.Equal(t, false, ev["done"], "only the last event is final")
	}
}

func TestAIStepTest_StreamsAnErrorAsAnEvent(t *testing.T) {
	allowTesting(t, true)
	bootAIFunctions(t, func(bool, string) *aiagent.AgentResult { return nil })
	automation.NewAiHandlerForTest(service.Registry(), func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return nil, errors.New("model down")
	})

	router := chi.NewRouter()
	MountAIStepRoutes(router)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/aiAsk/test?stream=1", bytes.NewBufferString(`{"args":{"prompt":"x"}}`)))

	events := sseEvents(t, rec.Body.String())
	last := events[len(events)-1]
	require.Contains(t, last["error"], "model down")
	require.Equal(t, true, last["done"])
}

func TestAIStepTest_StreamModeStillChecksPermissionAndFunction(t *testing.T) {
	router := chi.NewRouter()
	MountAIStepRoutes(router)

	// not allowed: an ordinary error response, never a stream
	allowTesting(t, false)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/aiAsk/test?stream=1", bytes.NewBufferString(`{"args":{"prompt":"x"}}`)))
	require.NotEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
	require.NotContains(t, rec.Body.String(), `"text"`)

	// allowed, but not a function that can be tested
	allowTesting(t, true)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/ai/functions/aiTool/test?stream=1", bytes.NewBufferString(`{"args":{}}`)))
	require.NotEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
	require.Contains(t, rec.Body.String(), "cannot be tested on its own")
}
