package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/id"
)

func init() {
	id.Init(context.Background())
}

const bpmnRESTFixture = `<?xml version="1.0"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <bpmn:process id="P">
    <bpmn:startEvent id="S" />
    <bpmn:sequenceFlow id="f1" sourceRef="S" targetRef="T" />
    <bpmn:serviceTask id="T" name="Sync">
      <bpmn:extensionElements><properties><property name="chainID" value="sync_x" /></properties></bpmn:extensionElements>
    </bpmn:serviceTask>
    <bpmn:sequenceFlow id="f2" sourceRef="T" targetRef="E" />
    <bpmn:endEvent id="E" name="End" />
  </bpmn:process>
</bpmn:definitions>`

func unwrapAPIResponse(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var env struct {
		Response map[string]interface{} `json:"response"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("invalid response envelope: %v (body: %s)", err, body)
	}
	return env.Response
}

func TestBPMNCompile_Success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/bpmn/compile", strings.NewReader(bpmnRESTFixture))
	rec := httptest.NewRecorder()

	BPMN{}.Compile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	resp := unwrapAPIResponse(t, rec.Body.Bytes())
	steps, ok := resp["steps"].([]interface{})
	if !ok || len(steps) != 2 { // Sync (function) + End (termination); startEvent is skipped
		t.Fatalf("expected 2 compiled steps, got %#v", resp["steps"])
	}
	paths, ok := resp["paths"].([]interface{})
	if !ok || len(paths) != 1 { // startEvent's flow is dropped, only Sync->End remains
		t.Fatalf("expected 1 path, got %#v", resp["paths"])
	}
}

func TestBPMNCompile_InvalidXMLReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/bpmn/compile", strings.NewReader("not xml"))
	rec := httptest.NewRecorder()

	BPMN{}.Compile(rec, req)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if body["error"] == nil {
		t.Fatalf("expected an error in the response, got: %s", rec.Body.String())
	}
}

func TestBPMNCompile_EmptyBodyReturnsError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/bpmn/compile", strings.NewReader(""))
	rec := httptest.NewRecorder()

	BPMN{}.Compile(rec, req)

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if body["error"] == nil {
		t.Fatalf("expected an error for empty body, got: %s", rec.Body.String())
	}
}
