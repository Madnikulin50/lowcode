package rest

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJSONIDUnmarshal(t *testing.T) {
	type wrap struct {
		PageID      jsonID `json:"pageID"`
		ModuleID    jsonID `json:"moduleID"`
		NamespaceID jsonID `json:"namespaceID"`
	}

	var w wrap
	raw := []byte(`{"pageID":"496258658610774017","moduleID":"495727984904044545","namespaceID":"495727984893558785"}`)
	if err := json.Unmarshal(raw, &w); err != nil {
		t.Fatal(err)
	}
	if uint64(w.PageID) != 496258658610774017 {
		t.Fatalf("pageID=%d", w.PageID)
	}

	if err := json.Unmarshal([]byte(`{"pageID":42}`), &w); err != nil {
		t.Fatal(err)
	}
	if w.PageID != 42 {
		t.Fatalf("numeric pageID=%d", w.PageID)
	}

	if err := json.Unmarshal([]byte(`{"pageID":""}`), &w); err != nil {
		t.Fatal(err)
	}
	if w.PageID != 0 {
		t.Fatalf("empty pageID=%d", w.PageID)
	}
}

func TestFlattenValuesRawArray(t *testing.T) {
	ctx := map[string]interface{}{}
	flattenValues(ctx, []interface{}{
		map[string]interface{}{"name": "store_id", "value": "34"},
		map[string]interface{}{"name": "store_name", "value": "МСК-01"},
	})
	if ctx["store_id"] != "34" {
		t.Fatalf("store_id=%v", ctx["store_id"])
	}
	if ctx["store_name"] != "МСК-01" {
		t.Fatalf("store_name=%v", ctx["store_name"])
	}

	ctx = map[string]interface{}{}
	flattenValues(ctx, map[string]interface{}{"store_id": "7", "store_name": "X"})
	if ctx["store_id"] != "7" {
		t.Fatalf("map store_id=%v", ctx["store_id"])
	}
}

func TestInjectAgentCallbackURL(t *testing.T) {
	t.Setenv("CORTEZA_API", "")
	t.Setenv("HTTP_API_BASE_URL", "")
	t.Setenv("HTTP_BASE_URL", "")
	t.Setenv("HTTP_ADDR", "")

	bag := map[string]interface{}{}
	injectAgentCallback(nil, bag)
	got := fmt.Sprintf("%v", bag["callbackUrl"])
	want := "http://localhost:3333/compose/rulechain/cmdb-ingest-scan/run"
	if got != want {
		t.Fatalf("default callback %q want %q", got, want)
	}

	t.Setenv("CORTEZA_API", "http://localhost:3333/api")
	bag = map[string]interface{}{}
	injectAgentCallback(nil, bag)
	got = fmt.Sprintf("%v", bag["callbackUrl"])
	if got != want {
		t.Fatalf("stripped /api callback %q want %q", got, want)
	}

	t.Setenv("CORTEZA_API", "")
	t.Setenv("HTTP_API_BASE_URL", "/api")
	t.Setenv("HTTP_ADDR", ":3333")
	bag = map[string]interface{}{}
	injectAgentCallback(nil, bag)
	got = fmt.Sprintf("%v", bag["callbackUrl"])
	wantAPI := "http://localhost:3333/api/compose/rulechain/cmdb-ingest-scan/run"
	if got != wantAPI {
		t.Fatalf("HTTP_API_BASE_URL=/api callback %q want %q", got, wantAPI)
	}

	bag = map[string]interface{}{"callbackUrl": "http://example/custom"}
	injectAgentCallback(nil, bag)
	if bag["callbackUrl"] != "http://example/custom" {
		t.Fatalf("explicit callback overwritten: %v", bag["callbackUrl"])
	}

	t.Setenv("CORTEZA_API", "")
	t.Setenv("HTTP_API_BASE_URL", "/")
	t.Setenv("HTTP_BASE_URL", "")
	t.Setenv("HTTP_ADDR", ":3333")
	bag = map[string]interface{}{}
	injectAgentCallback(nil, bag)
	got = fmt.Sprintf("%v", bag["callbackUrl"])
	if got != want {
		t.Fatalf("HTTP_API_BASE_URL=/ callback %q want %q", got, want)
	}
}

func TestAliasTriggerRecordIDs(t *testing.T) {
	bag := map[string]interface{}{"sourceID": "${recordID}"}
	flattenTriggerContext(bag, &triggerRequest{RecordID: "510291663494250497"})
	if bag["recordID"] != "510291663494250497" {
		t.Fatalf("recordID=%v", bag["recordID"])
	}
	if bag["sourceID"] != "510291663494250497" {
		t.Fatalf("sourceID=%v (placeholder should yield to recordID)", bag["sourceID"])
	}

	bag = map[string]interface{}{"projectID": "${recordID}"}
	flattenTriggerContext(bag, &triggerRequest{RecordID: "510291663494250497"})
	if bag["recordID"] != "510291663494250497" {
		t.Fatalf("recordID=%v", bag["recordID"])
	}
	if bag["projectID"] != "510291663494250497" {
		t.Fatalf("projectID=%v (placeholder should yield to recordID)", bag["projectID"])
	}

	bag = map[string]interface{}{"policyID": "99"}
	flattenTriggerContext(bag, &triggerRequest{})
	if bag["recordID"] != "99" {
		t.Fatalf("policy alias recordID=%v", bag["recordID"])
	}
	if bag["sourceID"] != nil && bag["sourceID"] != "" {
		t.Fatalf("must not copy policy id into sourceID: %v", bag["sourceID"])
	}
}
