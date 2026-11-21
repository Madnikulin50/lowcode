package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/automation/service"
	"github.com/madnikulin50/lowcode/server/pkg/aiagent"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type allowAll struct{ read, write bool }

func (a allowAll) CanSearchWorkflows(context.Context) bool { return a.read }
func (a allowAll) CanCreateWorkflow(context.Context) bool  { return a.write }

type promptAPI struct {
	t      *testing.T
	router chi.Router
}

func newPromptAPI(t *testing.T, ac allowAll) *promptAPI {
	t.Helper()
	ctx := context.Background()

	st, err := sqlite.ConnectInMemory(ctx)
	require.NoError(t, err)
	require.NoError(t, store.Upgrade(ctx, zap.NewNop(), st))
	require.NoError(t, st.TruncateAutomationPromptVersions(ctx))

	prev := service.DefaultPrompts
	service.DefaultPrompts = service.PromptLibrary(st, ac)
	t.Cleanup(func() { service.DefaultPrompts = prev })

	router := chi.NewRouter()
	// requests carry a signed-in user, as the token validator would arrange
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(auth.SetIdentityToContext(r.Context(), auth.Authenticated(5))))
		})
	})
	MountPromptRoutes(router)
	return &promptAPI{t: t, router: router}
}

func (p *promptAPI) do(method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	p.router.ServeHTTP(rec, httptest.NewRequest(method, path, bytes.NewBufferString(body)))
	return rec
}

func (p *promptAPI) json(method, path, body string) map[string]interface{} {
	rec := p.do(method, path, body)
	var out struct {
		Response map[string]interface{} `json:"response"`
		Error    map[string]interface{} `json:"error"`
	}
	require.NoError(p.t, json.Unmarshal(rec.Body.Bytes(), &out), rec.Body.String())
	if out.Error != nil {
		return map[string]interface{}{"error": out.Error["message"]}
	}
	return out.Response
}

func TestPromptsAPI_Lifecycle(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	req := require.New(t)

	// nothing yet
	req.Empty(api.json("GET", "/prompts/", "")["set"])

	// save: the first version is live
	saved := api.json("POST", "/prompts/triage", `{"text":"Rate urgency.","description":"Ticket triage","note":"first","cases":[{"name":"outage","inputs":{"ticket":"server down"},"expect":{"priority":"high"}}]}`)
	req.EqualValues(1, saved["version"])
	req.Equal(true, saved["active"])

	// a second version, not activated
	saved = api.json("POST", "/prompts/triage", `{"text":"Rate urgency from 1 to 5.","note":"numeric"}`)
	req.EqualValues(2, saved["version"])
	req.Equal(false, saved["active"])

	list := api.json("GET", "/prompts/", "")["set"].([]interface{})
	req.Len(list, 1)
	row := list[0].(map[string]interface{})
	req.Equal("triage", row["handle"])
	req.EqualValues(1, row["activeVersion"])
	req.EqualValues(2, row["versions"])
	req.EqualValues(1, row["cases"])

	hist := api.json("GET", "/prompts/triage", "")["set"].([]interface{})
	req.Len(hist, 2)

	// one version with its text and cases; 0 is the active one
	v := api.json("GET", "/prompts/triage/versions/2", "")
	req.Equal("Rate urgency from 1 to 5.", v["prompt"].(map[string]interface{})["text"])
	req.Len(v["cases"], 1, "cases carry forward")

	active := api.json("GET", "/prompts/triage/versions/0", "")
	req.EqualValues(1, active["prompt"].(map[string]interface{})["version"])

	// roll forward, then back
	req.Equal(true, api.json("POST", "/prompts/triage/activate", `{"version":2}`)["active"])
	req.EqualValues(2, api.json("GET", "/prompts/triage/versions/0", "")["prompt"].(map[string]interface{})["version"])
	api.json("POST", "/prompts/triage/activate", `{"version":1}`)

	// cases can be replaced or cleared
	api.json("POST", "/prompts/triage", `{"text":"v3","cases":[]}`)
	req.Empty(api.json("GET", "/prompts/triage/versions/3", "")["cases"])

	// delete
	req.Equal("triage", api.json("DELETE", "/prompts/triage", "")["deleted"])
	req.Empty(api.json("GET", "/prompts/", "")["set"])
}

func TestPromptsAPI_Errors(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	req := require.New(t)

	req.Contains(api.json("GET", "/prompts/nope", "")["error"], "does not exist")
	req.Contains(api.json("GET", "/prompts/nope/versions/abc", "")["error"], "version must be a number")
	req.Contains(api.json("POST", "/prompts/Bad_Handle", `{"text":"x"}`)["error"], "not valid")
	req.Contains(api.json("POST", "/prompts/ok", `{"text":"  "}`)["error"], "empty")
	req.Contains(api.json("POST", "/prompts/ok", `{not json`)["error"], "invalid JSON")
	req.Contains(api.json("POST", "/prompts/nope/activate", `{"version":1}`)["error"], "")
	req.Contains(api.json("DELETE", "/prompts/nope", "")["error"], "does not exist")
}

func TestPromptsAPI_Permissions(t *testing.T) {
	reader := newPromptAPI(t, allowAll{read: true})
	req := require.New(t)

	req.NotNil(reader.json("GET", "/prompts/", "")["set"], "reading is allowed")
	req.NotNil(reader.json("POST", "/prompts/p", `{"text":"x"}`)["error"], "writing is not")
	req.NotNil(reader.json("DELETE", "/prompts/p", "")["error"])

	nobody := newPromptAPI(t, allowAll{})
	req.NotNil(nobody.json("GET", "/prompts/", "")["error"])
	req.NotNil(nobody.json("GET", "/prompts/p", "")["error"])
}

func TestPromptsAPI_EvalStreamsEachCase(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	req := require.New(t)

	api.json("POST", "/prompts/echo", `{"text":"Say hi.","cases":[{"name":"a","contains":["hi"]},{"name":"b","contains":["bye"]}]}`)

	var model string
	service.SetPromptEvalRunner(func(ctx context.Context, _, _ string, _ bool) (*aiagent.AgentResult, error) {
		model = aiagent.ModelFromContext(ctx)
		return &aiagent.AgentResult{Success: true, Output: "hi there", PromptTokens: 3}, nil
	})
	defer service.SetPromptEvalRunner(nil)

	rec := api.do("POST", "/prompts/echo/eval?stream=1", `{"versions":[1],"model":"qwen3:8b"}`)
	req.Equal("text/event-stream", rec.Header().Get("Content-Type"))
	req.Equal("qwen3:8b", model, "the chosen model reaches the call")

	events := sseEvents(t, rec.Body.String())
	req.Len(events, 3, "one event per case, then the reports")

	a, b := events[0], events[1]
	req.EqualValues(1, a["version"])
	req.Equal("a", a["case"].(map[string]interface{})["case"])
	req.Equal(true, a["case"].(map[string]interface{})["pass"])
	req.Equal(false, a["done"])
	req.Equal("b", b["case"].(map[string]interface{})["case"])
	req.Equal(false, b["case"].(map[string]interface{})["pass"], "the answer does not say bye")

	last := events[2]
	req.Equal(true, last["done"])
	report := last["reports"].(map[string]interface{})["v1"].(map[string]interface{})
	req.EqualValues(2, report["total"])
	req.EqualValues(1, report["passed"])
	req.EqualValues(0.5, report["passRate"])
}

func TestPromptsAPI_EvalWithoutStream(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	req := require.New(t)

	api.json("POST", "/prompts/echo", `{"text":"Say hi.","cases":[{"name":"a","contains":["hi"]}]}`)
	service.SetPromptEvalRunner(func(context.Context, string, string, bool) (*aiagent.AgentResult, error) {
		return &aiagent.AgentResult{Success: true, Output: "hi"}, nil
	})
	defer service.SetPromptEvalRunner(nil)

	out := api.json("POST", "/prompts/echo/eval", `{}`)
	req.EqualValues(1, out["reports"].(map[string]interface{})["v1"].(map[string]interface{})["passed"], "no versions named: the active one")

	// cases sent with the request replace the stored ones for this run
	out = api.json("POST", "/prompts/echo/eval", `{"cases":[{"name":"x","contains":["nope"]},{"name":"y","contains":["hi"]}]}`)
	req.EqualValues(2, out["reports"].(map[string]interface{})["v1"].(map[string]interface{})["total"])
}

func TestPromptsAPI_EvalErrors(t *testing.T) {
	api := newPromptAPI(t, allowAll{true, true})
	req := require.New(t)

	api.json("POST", "/prompts/bare", `{"text":"x"}`)
	req.Contains(api.json("POST", "/prompts/bare/eval", `{}`)["error"], "no test cases")
	req.Contains(api.json("POST", "/prompts/nope/eval", `{}`)["error"], "does not exist")

	// in stream mode an error is an event, since the stream has begun
	rec := api.do("POST", "/prompts/bare/eval?stream=1", `{}`)
	events := sseEvents(t, rec.Body.String())
	req.Contains(events[len(events)-1]["error"], "no test cases")
}
