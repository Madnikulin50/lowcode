package rest

import (
	"encoding/json"
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
