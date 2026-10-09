package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/compose/service"
	"github.com/madnikulin50/lowcode/server/pkg/id"
	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
	"github.com/madnikulin50/lowcode/server/store"
	"github.com/madnikulin50/lowcode/server/store/adapters/rdbms/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	id.Init(context.Background())
}

// withChiParams injects chi URL params into a request, the same way the
// router would after matching a pattern like /{chainID}/runs/{runID}.
func withChiParams(r *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

// unwrapResponse pulls the "response" envelope api.Send wraps every payload in.
func unwrapResponse(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var env struct {
		Response map[string]interface{} `json:"response"`
	}
	require.NoError(t, json.Unmarshal(body, &env))
	return env.Response
}

func TestRuleChainAdmin_ListAndGetRuns(t *testing.T) {
	req := require.New(t)
	ctx := context.Background()

	s, err := sqlite.ConnectInMemoryWithDebug(ctx)
	req.NoError(err)
	req.NoError(store.Upgrade(ctx, zap.NewNop(), s))

	prevStore := service.DefaultStore
	service.DefaultStore = s
	defer func() { service.DefaultStore = prevStore }()

	persist := service.NewRuleChainPersistence()
	req.NoError(persist.SaveChain(ctx, &rulesgo.Chain{ID: "runs_test_chain", Name: "Runs Test", NamespaceID: 7}))

	saver, ok := persist.(rulesgo.RunLogPersistence)
	req.True(ok)

	started := time.Now().Add(-time.Second)
	finished := started.Add(120 * time.Millisecond)
	req.NoError(saver.SaveRun(ctx, rulesgo.ExecRecord{
		ChainID: "runs_test_chain",
		Input:   map[string]interface{}{"a": 1},
		Result: &rulesgo.ChainResult{
			ChainID: "runs_test_chain",
			Success: true,
			Output:  map[string]interface{}{"b": 2},
			Nodes:   []rulesgo.NodeResult{{NodeID: "n1", Type: "condition"}},
		},
		StartedAt:   started,
		FinishedAt:  finished,
		TriggerType: "manual-test",
	}))

	admin := RuleChainAdmin{}

	// --- ListRuns ---
	listReq := httptest.NewRequest(http.MethodGet, "/admin/rulechain/runs_test_chain/runs?limit=10", nil)
	listReq = withChiParams(listReq, map[string]string{"chainID": "runs_test_chain"})
	listRec := httptest.NewRecorder()
	admin.ListRuns(listRec, listReq)

	req.Equal(http.StatusOK, listRec.Code)
	listResp := unwrapResponse(t, listRec.Body.Bytes())
	runs, ok := listResp["runs"].([]interface{})
	req.True(ok, "expected runs array, got %#v", listResp["runs"])
	req.Len(runs, 1)

	first, ok := runs[0].(map[string]interface{})
	req.True(ok)
	req.Equal("runs_test_chain", first["chainID"])
	req.Equal("manual-test", first["triggerType"])
	req.Equal(true, first["success"])
	req.EqualValues(120, first["durationMs"])
	runID, _ := first["runID"].(string)
	req.NotEmpty(runID)

	// list view must not leak the full trace (kept light)
	req.NotContains(first, "nodes")
	req.NotContains(first, "input")
	req.NotContains(first, "output")

	// --- GetRun ---
	getReq := httptest.NewRequest(http.MethodGet, "/admin/rulechain/runs_test_chain/runs/"+runID, nil)
	getReq = withChiParams(getReq, map[string]string{"chainID": "runs_test_chain", "runID": runID})
	getRec := httptest.NewRecorder()
	admin.GetRun(getRec, getReq)

	req.Equal(http.StatusOK, getRec.Code)
	getResp := unwrapResponse(t, getRec.Body.Bytes())
	run, ok := getResp["run"].(map[string]interface{})
	req.True(ok, "expected run object, got %#v", getResp["run"])

	nodesJSON, err := json.Marshal(run["nodes"])
	req.NoError(err)
	req.Contains(string(nodesJSON), "n1")

	outputJSON, err := json.Marshal(run["output"])
	req.NoError(err)
	req.Contains(string(outputJSON), `"b":2`)

	// --- GetRun for a run belonging to another chain must 404/error ---
	mismatchReq := httptest.NewRequest(http.MethodGet, "/admin/rulechain/other_chain/runs/"+runID, nil)
	mismatchReq = withChiParams(mismatchReq, map[string]string{"chainID": "other_chain", "runID": runID})
	mismatchRec := httptest.NewRecorder()
	admin.GetRun(mismatchRec, mismatchReq)

	var mismatchBody map[string]interface{}
	req.NoError(json.Unmarshal(mismatchRec.Body.Bytes(), &mismatchBody))
	req.NotNil(mismatchBody["error"], "expected an error for mismatched chainID")
}
