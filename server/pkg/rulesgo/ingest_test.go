package rulesgo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
)

type memCRUD struct {
	mu      sync.Mutex
	next    int
	records []map[string]interface{}
}

func (m *memCRUD) Create(_ context.Context, _, _ uint64, values map[string]interface{}) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	id := fmt.Sprintf("%d", m.next)
	rec := map[string]interface{}{"recordID": id}
	for k, v := range values {
		rec[k] = v
	}
	m.records = append(m.records, rec)
	return id, "now", nil
}

func (m *memCRUD) Update(_ context.Context, _, _ uint64, recordID string, values map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, rec := range m.records {
		if fmt.Sprintf("%v", rec["recordID"]) == recordID {
			for k, v := range values {
				rec[k] = v
			}
			return "now", nil
		}
	}
	return "", fmt.Errorf("not found")
}

func (m *memCRUD) Delete(context.Context, uint64, uint64, string) error { return nil }

func (m *memCRUD) Search(_ context.Context, _, _ uint64, query string, _ int) ([]map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	field, value, ok := parseEqQuery(query)
	if !ok {
		return append([]map[string]interface{}{}, m.records...), nil
	}
	var out []map[string]interface{}
	for _, rec := range m.records {
		if strings.EqualFold(fmt.Sprintf("%v", rec[field]), value) {
			cp := map[string]interface{}{}
			for k, v := range rec {
				cp[k] = v
			}
			out = append(out, cp)
		}
	}
	return out, nil
}

func parseEqQuery(q string) (field, value string, ok bool) {
	parts := strings.SplitN(q, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	field = strings.TrimSpace(parts[0])
	value = strings.Trim(strings.TrimSpace(parts[1]), "'")
	value = strings.ReplaceAll(value, "\\'", "'")
	return field, value, field != ""
}

func TestCollectItems(t *testing.T) {
	if n := len(CollectItems([]interface{}{"a", "b"})); n != 2 {
		t.Fatalf("slice %d", n)
	}
	raw, _ := json.Marshal([]map[string]interface{}{{"ip": "1.1.1.1"}})
	if n := len(CollectItems(string(raw))); n != 1 {
		t.Fatalf("json string %d", n)
	}
	if n := len(CollectItems(map[string]interface{}{"ip": "1.1.1.1"})); n != 0 {
		t.Fatalf("map should not be items")
	}
}

func TestForeachUpsert(t *testing.T) {
	crud := &memCRUD{}
	reg := DefaultRegistry(&DefaultConfig{CRUD: crud})
	engine := NewEngine(reg)
	engine.RegisterChain(&Chain{
		ID:        "ingest",
		EntryNode: "foreach_items",
		Nodes: []ChainNode{
			{ID: "foreach_items", Type: "foreach", Config: json.RawMessage(`{"items":"items","itemVar":"item"}`)},
			{ID: "upsert_device", Type: "crud.upsert", Config: json.RawMessage(`{
				"namespaceID":"1","moduleID":"2",
				"matchBy":["mac_address","ip_address"],
				"fields":{"ip_address":"{{item.ip}}","mac_address":"{{item.mac}}","hostname":"{{item.hostname}}"}
			}`)},
		},
		Edges: []ChainEdge{{From: "foreach_items", To: "upsert_device"}},
	})

	items := []interface{}{
		map[string]interface{}{"ip": "10.0.0.1", "mac": "aa:bb:cc:dd:ee:ff", "hostname": "host-a"},
		map[string]interface{}{"ip": "10.0.0.2", "mac": "aa:bb:cc:dd:ee:00", "hostname": "host-b"},
	}
	res, err := engine.Run(context.Background(), "ingest", map[string]interface{}{"items": items})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("fail: %+v", res)
	}
	if len(crud.records) != 2 {
		t.Fatalf("created %d", len(crud.records))
	}

	res, err = engine.Run(context.Background(), "ingest", map[string]interface{}{"items": items})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("second fail: %+v", res)
	}
	if len(crud.records) != 2 {
		t.Fatalf("upsert should update, got %d records", len(crud.records))
	}
}

func TestComposeJobStatusAndCancel(t *testing.T) {
	if composeJobStatus("done") != "completed" {
		t.Fatal("done")
	}
	p := NewAgentPoller()
	SetDefaultPoller(p)
	ctx, cancel := context.WithCancel(context.Background())
	p.mu.Lock()
	p.cancels["job-1"] = cancel
	p.mu.Unlock()
	CancelPollIfTerminal(map[string]interface{}{"kind": "complete", "jobID": "job-1"})
	<-ctx.Done()
}

func TestFlattenItemJSON(t *testing.T) {
	ec := &ExecutionContext{Variables: map[string]interface{}{}}
	FlattenItem(ec, "item", map[string]interface{}{
		"ip":        "10.0.0.1",
		"openPorts": []interface{}{map[string]interface{}{"port": 22}},
	})
	if fmt.Sprintf("%v", ec.Get("item.ip")) != "10.0.0.1" {
		t.Fatalf("ip %v", ec.Get("item.ip"))
	}
	s, _ := ec.Get("item.openPorts").(string)
	if !strings.Contains(s, `"port":22`) && !strings.Contains(s, `"port": 22`) {
		t.Fatalf("ports %q", s)
	}
}
