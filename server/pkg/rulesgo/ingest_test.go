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

func (m *memCRUD) Create(_ context.Context, _, moduleID uint64, values map[string]interface{}) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.next++
	id := fmt.Sprintf("%d", m.next)
	rec := map[string]interface{}{"recordID": id, "moduleID": fmt.Sprintf("%d", moduleID)}
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
	wrapped := map[string]interface{}{"devices": []interface{}{map[string]interface{}{"ip": "10.0.0.1"}}}
	if n := len(CollectItems(wrapped)); n != 1 {
		t.Fatalf("devices wrap %d", n)
	}
	camel := []interface{}{map[string]interface{}{"ip": "10.1.1.1", "mac": "aa:bb:cc:dd:ee:ff", "deviceType": "router"}}
	if n := len(CollectItems(camel)); n != 1 {
		t.Fatalf("camelCase slice %d", n)
	}
	if n := len(jsonToItems(camel)); n != 1 {
		t.Fatalf("jsonToItems slice %d", n)
	}
}

func TestNormalizeIngestEnvelope(t *testing.T) {
	env := NormalizeIngestEnvelope(map[string]interface{}{
		"namespaceID":     json.Number("509463708777775105"),
		"createdRecordID": "99",
		"items":           []interface{}{map[string]interface{}{"ip": "10.0.0.1", "deviceType": "pc"}},
	})
	if env["namespaceID"] != "509463708777775105" {
		t.Fatalf("ns %v", env["namespaceID"])
	}
	if env["scanRecordID"] != "99" || env["createdRecordID"] != "99" {
		t.Fatalf("aliases %+v", env)
	}
	items := CollectItems(env["items"])
	if len(items) != 1 {
		t.Fatalf("items %d", len(items))
	}

	var raw interface{}
	if err := json.Unmarshal([]byte(`{"namespaceID":509463708777775105,"items":[{"ip":"10.0.0.8"}]}`), &raw); err != nil {
		t.Fatal(err)
	}
	env = NormalizeIngestEnvelope(raw.(map[string]interface{}))
	if _, ok := env["namespaceID"]; ok {
		t.Fatalf("imprecise float namespaceID should be dropped, got %v", env["namespaceID"])
	}
	if len(CollectItems(env["items"])) != 1 {
		t.Fatalf("items from webhook JSON")
	}
}

func TestHasUpsertIdentityKeepsRealHosts(t *testing.T) {
	if !hasUpsertIdentity([]string{"mac_address", "ip_address", "hostname"}, map[string]interface{}{
		"ip_address": "10.0.0.1",
		"hostname":   "gateway",
	}) {
		t.Fatal("IP should be enough identity")
	}
	if !hasUpsertIdentity([]string{"mac_address", "ip_address", "hostname"}, map[string]interface{}{
		"hostname": "10.0.0.8",
	}) {
		t.Fatal("IP-like hostname is a real host, must not skip")
	}
	if hasUpsertIdentity([]string{"hostname"}, map[string]interface{}{"hostname": "localhost"}) {
		t.Fatal("generic hostname should skip")
	}
	if hasUpsertIdentity([]string{"ip_address"}, map[string]interface{}{"ip_address": "{{item.ip}}"}) {
		t.Fatal("leftover template is not identity")
	}
}

func TestIngestShouldStop(t *testing.T) {
	ok := &ChainResult{Success: true}
	env := map[string]interface{}{"found": 3}
	if ingestShouldStop("complete", env, ok, nil) {
		t.Fatal("found>0 and no items must keep polling")
	}
	env["items"] = []interface{}{map[string]interface{}{"ip": "1.1.1.1"}}
	if !ingestShouldStop("complete", env, ok, nil) {
		t.Fatal("items present should stop")
	}
	if ingestShouldStop("complete", env, ok, fmt.Errorf("boom")) {
		t.Fatal("ingest error must keep polling")
	}
}

func TestBaseEnvelopeCopiesScanItems(t *testing.T) {
	spec := pollSpec{jobID: "j1", base: map[string]interface{}{"createdRecordID": "77", "namespaceID": "1"}}
	st := map[string]interface{}{
		"status": "done",
		"found":  1,
		"items":  []interface{}{map[string]interface{}{"ip": "10.0.0.1", "mac": "aa:bb:cc:dd:ee:ff"}},
	}
	env := baseEnvelope(spec, st)
	if env["scanRecordID"] != "77" {
		t.Fatalf("scanRecordID %v", env["scanRecordID"])
	}
	if len(CollectItems(env["items"])) != 1 {
		t.Fatalf("status items %v", env["items"])
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

func TestForeachNestedServicesAndVulns(t *testing.T) {
	crud := &memCRUD{}
	reg := DefaultRegistry(&DefaultConfig{CRUD: crud})
	engine := NewEngine(reg)
	engine.RegisterChain(&Chain{
		ID:        "ingest",
		EntryNode: "foreach_items",
		Nodes: []ChainNode{
			{ID: "foreach_items", Type: "foreach", Config: json.RawMessage(`{"items":"items","itemVar":"item"}`)},
			{ID: "upsert_device", Type: "crud.upsert", Config: json.RawMessage(`{
				"namespaceID":"1","moduleID":"2","resultVar":"deviceRecordID",
				"matchBy":["ip_address"],"omitEmpty":true,
				"fields":{"ip_address":"{{item.ip}}","hostname":"{{item.hostname}}"}
			}`)},
			{ID: "foreach_ports", Type: "foreach", Config: json.RawMessage(`{"items":"item.openPorts","itemVar":"port"}`)},
			{ID: "upsert_service", Type: "crud.upsert", Config: json.RawMessage(`{
				"namespaceID":"1","moduleID":"3","matchAll":true,"omitEmpty":true,
				"matchBy":["device","port","proto"],
				"fields":{"device":"{{deviceRecordID}}","port":"{{port.port}}","proto":"{{port.proto}}","service":"{{port.service}}"}
			}`)},
			{ID: "foreach_vulns", Type: "foreach", Config: json.RawMessage(`{"items":"item.vulnerabilities","itemVar":"vuln"}`)},
			{ID: "upsert_vuln", Type: "crud.upsert", Config: json.RawMessage(`{
				"namespaceID":"1","moduleID":"4","matchAll":true,"omitEmpty":true,
				"matchBy":["device","name"],
				"fields":{"device":"{{deviceRecordID}}","name":"{{vuln.name}}","severity":"{{vuln.severity}}","status":"open"}
			}`)},
		},
		Edges: []ChainEdge{
			{From: "foreach_items", To: "upsert_device"},
			{From: "foreach_items", To: "foreach_ports"},
			{From: "foreach_ports", To: "upsert_service"},
			{From: "foreach_items", To: "foreach_vulns"},
			{From: "foreach_vulns", To: "upsert_vuln"},
		},
	})

	items := []interface{}{
		map[string]interface{}{
			"ip":       "10.0.0.1",
			"hostname": "host-a",
			"openPorts": []interface{}{
				map[string]interface{}{"port": 22, "proto": "tcp", "service": "ssh"},
				map[string]interface{}{"port": 80, "proto": "tcp", "service": "http"},
			},
			"vulnerabilities": []interface{}{
				map[string]interface{}{"name": "Expired TLS", "severity": "HIGH"},
			},
		},
	}
	res, err := engine.Run(context.Background(), "ingest", map[string]interface{}{"items": items})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("fail: %+v", res)
	}

	devices, services, vulns := partitionByModule(crud.records)
	if len(devices) != 1 || fmt.Sprintf("%v", devices[0]["ip_address"]) != "10.0.0.1" {
		t.Fatalf("devices %+v", devices)
	}
	devID := fmt.Sprintf("%v", devices[0]["recordID"])
	if len(services) != 2 {
		t.Fatalf("services %d %+v", len(services), services)
	}
	for _, s := range services {
		if fmt.Sprintf("%v", s["device"]) != devID {
			t.Fatalf("service device %v want %s", s["device"], devID)
		}
	}
	if len(vulns) != 1 || fmt.Sprintf("%v", vulns[0]["name"]) != "Expired TLS" {
		t.Fatalf("vulns %+v", vulns)
	}
	if fmt.Sprintf("%v", vulns[0]["device"]) != devID {
		t.Fatalf("vuln device %v want %s", vulns[0]["device"], devID)
	}

	res, err = engine.Run(context.Background(), "ingest", map[string]interface{}{"items": items})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("second fail: %+v", res)
	}
	devices, services, vulns = partitionByModule(crud.records)
	if len(devices) != 1 || len(services) != 2 || len(vulns) != 1 {
		t.Fatalf("upsert related should update, got devices=%d services=%d vulns=%d", len(devices), len(services), len(vulns))
	}
}

func partitionByModule(recs []map[string]interface{}) (devices, services, vulns []map[string]interface{}) {
	for _, rec := range recs {
		switch fmt.Sprintf("%v", rec["moduleID"]) {
		case "2":
			devices = append(devices, rec)
		case "3":
			services = append(services, rec)
		case "4":
			vulns = append(vulns, rec)
		}
	}
	return
}

func TestMatchAllRequiresEveryField(t *testing.T) {
	if hasAllUpsertIdentity([]string{"device", "port"}, map[string]interface{}{"device": "1"}) {
		t.Fatal("missing port must skip")
	}
	if !hasAllUpsertIdentity([]string{"device", "port"}, map[string]interface{}{"device": "1", "port": "22"}) {
		t.Fatal("device+port should match")
	}
}

func TestIngestContinuesAfterScanUpdateFailure(t *testing.T) {
	crud := &memCRUD{}
	reg := DefaultRegistry(&DefaultConfig{CRUD: crud})
	engine := NewEngine(reg)
	engine.RegisterChain(&Chain{
		ID:        "ingest",
		EntryNode: "update_scan",
		Nodes: []ChainNode{
			{ID: "update_scan", Type: "crud", Config: json.RawMessage(`{
				"operation":"update","namespaceID":"1","moduleID":"1",
				"recordID":"{{scanRecordID}}","omitEmpty":true,"continueOnError":true,
				"fields":{"status":"{{status}}","finished_at":"{{finishedAt}}"}
			}`)},
			{ID: "foreach_items", Type: "foreach", Config: json.RawMessage(`{"items":"items","itemVar":"item"}`)},
			{ID: "upsert_device", Type: "crud.upsert", Config: json.RawMessage(`{
				"namespaceID":"1","moduleID":"2","omitEmpty":true,
				"matchBy":["ip_address"],
				"fields":{"ip_address":"{{item.ip}}","hostname":"{{item.hostname}}"}
			}`)},
		},
		Edges: []ChainEdge{
			{From: "update_scan", To: "foreach_items"},
			{From: "foreach_items", To: "upsert_device"},
		},
	})

	items := []interface{}{map[string]interface{}{"ip": "10.0.0.8", "hostname": "host-x"}}
	res, err := engine.Run(context.Background(), "ingest", map[string]interface{}{
		"items":  items,
		"status": "completed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("scan update must not abort ingest: %+v", res)
	}
	if len(crud.records) != 1 {
		t.Fatalf("expected device upsert, got %d records: %+v", len(crud.records), crud.records)
	}
	if fmt.Sprintf("%v", crud.records[0]["ip_address"]) != "10.0.0.8" {
		t.Fatalf("device %+v", crud.records[0])
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
