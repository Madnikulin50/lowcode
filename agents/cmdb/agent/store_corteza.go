package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CortezaStore struct {
	baseURL     string
	token       string
	httpClient  *http.Client
	namespaceID uint64
}

func NewCortezaStore(baseURL, token string, namespaceID uint64) *CortezaStore {
	return &CortezaStore{
		baseURL:     strings.TrimRight(baseURL, "/"),
		token:       token,
		httpClient:  &http.Client{},
		namespaceID: namespaceID,
	}
}

type composeModule struct {
	ID     uint64 `json:"moduleID,string"`
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

type composeNamespace struct {
	ID   uint64 `json:"namespaceID,string"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type composeRecord struct {
	ID     uint64               `json:"recordID,string"`
	Values []composeRecordValue `json:"values"`
}

type composeRecordValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (c *CortezaStore) request(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var b io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		b = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, apiStatusError(resp.StatusCode, raw)
	}
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Response) > 0 {
		return envelope.Response, nil
	}
	return raw, nil
}

func unwrapSet[T any](raw []byte) ([]T, error) {
	var wrapped struct {
		Set []T `json:"set"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && wrapped.Set != nil {
		return wrapped.Set, nil
	}
	var flat []T
	if err := json.Unmarshal(raw, &flat); err != nil {
		return nil, err
	}
	return flat, nil
}

func (c *CortezaStore) resolveNamespace(ctx context.Context) (uint64, error) {
	if c.namespaceID != 0 {
		return c.namespaceID, nil
	}
	raw, err := c.request(ctx, "GET", "/compose/namespace/?slug=cmdb&limit=50", nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeNamespace](raw)
	if err != nil {
		return 0, err
	}
	for _, ns := range set {
		if strings.EqualFold(ns.Slug, "cmdb") {
			c.namespaceID = ns.ID
			return ns.ID, nil
		}
	}
	return 0, fmt.Errorf("namespace slug cmdb not found; create it with agents/cmdb/compose/apply.mjs")
}

func (c *CortezaStore) EnsureModule(ctx context.Context) (uint64, error) {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	raw, err := c.request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/?handle=devices&limit=50", nsID), nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeModule](raw)
	if err == nil {
		for _, m := range set {
			if m.Handle == "devices" || strings.EqualFold(m.Name, "Devices") {
				return m.ID, nil
			}
		}
	}
	fields := []map[string]interface{}{
		{"name": "ip_address", "kind": "String", "label": "IP Address", "isRequired": true},
		{"name": "mac_address", "kind": "String", "label": "MAC Address"},
		{"name": "hostname", "kind": "String", "label": "Hostname"},
		{"name": "vendor", "kind": "String", "label": "Vendor"},
		{"name": "model", "kind": "String", "label": "Model"},
		{"name": "device_type", "kind": "Select", "label": "Device Type",
			"options": map[string]interface{}{
				"options": []map[string]string{
					{"value": "router", "text": "Router"},
					{"value": "switch", "text": "Switch"},
					{"value": "server", "text": "Server"},
					{"value": "workstation", "text": "Workstation"},
					{"value": "printer", "text": "Printer"},
					{"value": "camera", "text": "Camera"},
					{"value": "firewall", "text": "Firewall"},
					{"value": "phone", "text": "Phone"},
					{"value": "tablet", "text": "Tablet"},
					{"value": "iot", "text": "IoT"},
					{"value": "unknown", "text": "Unknown"},
				},
			},
		},
		{"name": "os", "kind": "String", "label": "Operating System"},
		{"name": "domain", "kind": "String", "label": "Domain"},
		{"name": "open_ports", "kind": "String", "label": "Open Ports",
			"options": map[string]interface{}{
				"displayType":      "json",
				"jsonLayout":       "chips",
				"jsonTemplate":     "{{port}}/{{proto}} {{service}}",
				"jsonFields":       "port,proto,service",
				"jsonVariantField": "port",
			},
		},
		{"name": "services", "kind": "String", "label": "Services (JSON)"},
		{"name": "shares", "kind": "String", "label": "Shares"},
		{"name": "vulnerabilities", "kind": "String", "label": "Vulnerabilities (JSON)"},
		{"name": "last_seen", "kind": "DateTime", "label": "Last Seen"},
		{"name": "status", "kind": "Select", "label": "Status",
			"options": map[string]interface{}{
				"options": []map[string]string{
					{"value": "online", "text": "Online"},
					{"value": "offline", "text": "Offline"},
					{"value": "unknown", "text": "Unknown"},
				},
			},
		},
		{"name": "criticality", "kind": "Select", "label": "Criticality",
			"options": map[string]interface{}{
				"options": []map[string]string{
					{"value": "low", "text": "Low"},
					{"value": "medium", "text": "Medium"},
					{"value": "high", "text": "High"},
					{"value": "critical", "text": "Critical"},
				},
			},
		},
		{"name": "notes", "kind": "String", "label": "Notes"},
	}
	body := map[string]interface{}{
		"name": "Devices", "handle": "devices", "fields": fields, "meta": map[string]interface{}{},
	}
	raw, err = c.request(ctx, "POST", fmt.Sprintf("/compose/namespace/%d/module/", nsID), body)
	if err != nil {
		return 0, fmt.Errorf("cannot create module: %w", err)
	}
	var mod composeModule
	if err := json.Unmarshal(raw, &mod); err != nil {
		return 0, fmt.Errorf("cannot parse created module: %w", err)
	}
	return mod.ID, nil
}

func (c *CortezaStore) FindDevice(ctx context.Context, modID uint64, d Device) (uint64, error) {
	if mac := normalizeMAC(d.MAC); mac != "" {
		id, err := c.findByField(ctx, modID, "mac_address", mac)
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
		if raw := strings.TrimSpace(d.MAC); raw != "" && raw != mac {
			id, err = c.findByField(ctx, modID, "mac_address", raw)
			if err != nil {
				return 0, err
			}
			if id > 0 {
				return id, nil
			}
		}
	}
	if ip := normalizeIP(d.IP); ip != "" {
		id, err := c.findByField(ctx, modID, "ip_address", ip)
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}
	if host := stableHostname(d.Hostname); host != "" {
		id, err := c.findByField(ctx, modID, "hostname", d.Hostname)
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}
	return 0, errDeviceNotFound
}

func (c *CortezaStore) FindDeviceByIP(ctx context.Context, modID uint64, ip string) (uint64, error) {
	return c.FindDevice(ctx, modID, Device{IP: ip})
}

func (c *CortezaStore) findByField(ctx context.Context, modID uint64, field, value string) (uint64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	q := url.QueryEscape(field + " = " + qlLiteral(value))
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/?query=%s&limit=5", nsID, modID, q)
	raw, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeRecord](raw)
	if err != nil {
		return 0, err
	}
	wantMAC := field == "mac_address"
	wantIP := field == "ip_address"
	wantHost := field == "hostname"
	for _, rec := range set {
		got := recordField(rec, field)
		if wantMAC && normalizeMAC(got) == normalizeMAC(value) {
			return rec.ID, nil
		}
		if wantIP && normalizeIP(got) == normalizeIP(value) {
			return rec.ID, nil
		}
		if wantHost && strings.EqualFold(strings.TrimSpace(got), strings.TrimSpace(value)) {
			return rec.ID, nil
		}
		if !wantMAC && !wantIP && !wantHost && got == value {
			return rec.ID, nil
		}
	}
	return 0, nil
}

func recordField(rec composeRecord, name string) string {
	for _, v := range rec.Values {
		if v.Name == name {
			return v.Value
		}
	}
	return ""
}

func (c *CortezaStore) CreateDevice(ctx context.Context, modID uint64, d Device) (uint64, error) {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/", nsID, modID)
	raw, err := c.request(ctx, "POST", path, map[string]interface{}{
		"values": deviceToValues(d),
	})
	if err != nil {
		return 0, err
	}
	var rec composeRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return 0, fmt.Errorf("cannot parse created record: %w", err)
	}
	return rec.ID, nil
}

func (c *CortezaStore) UpdateDevice(ctx context.Context, modID, recordID uint64, d Device) error {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	_, err = c.request(ctx, "POST", path, map[string]interface{}{
		"values": deviceToValues(d),
	})
	return err
}

func (c *CortezaStore) ListDevices(ctx context.Context, modID uint64) ([]Device, error) {
	if modID == 0 {
		var err error
		modID, err = c.EnsureModule(ctx)
		if err != nil {
			return nil, err
		}
	}
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/?limit=500", nsID, modID)
	raw, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	records, err := unwrapSet[composeRecord](raw)
	if err != nil {
		return nil, err
	}
	devices := make([]Device, 0, len(records))
	for _, rec := range records {
		devices = append(devices, recordToDevice(rec))
	}
	return devices, nil
}

func (c *CortezaStore) GetDevice(ctx context.Context, modID, recordID uint64) (*Device, error) {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return nil, err
	}
	if modID == 0 {
		modID, err = c.EnsureModule(ctx)
		if err != nil {
			return nil, err
		}
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	raw, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	var rec composeRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return nil, err
	}
	d := recordToDevice(rec)
	return &d, nil
}

func (c *CortezaStore) DeleteDevice(ctx context.Context, modID, recordID uint64) error {
	if modID == 0 {
		var err error
		modID, err = c.EnsureModule(ctx)
		if err != nil {
			return err
		}
	}
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	_, err = c.request(ctx, "DELETE", path, nil)
	return err
}

func (c *CortezaStore) moduleByHandle(ctx context.Context, handle string) (uint64, error) {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	raw, err := c.request(ctx, "GET", fmt.Sprintf("/compose/namespace/%d/module/?handle=%s&limit=50", nsID, url.QueryEscape(handle)), nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeModule](raw)
	if err != nil {
		return 0, err
	}
	for _, m := range set {
		if m.Handle == handle {
			return m.ID, nil
		}
	}
	return 0, fmt.Errorf("module %s not found", handle)
}

func (c *CortezaStore) UpdateScan(ctx context.Context, recordID uint64, s *ScanStatus) error {
	if recordID == 0 || s == nil {
		return nil
	}
	modID, err := c.moduleByHandle(ctx, "scans")
	if err != nil {
		return err
	}
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return err
	}
	status := s.Status
	switch status {
	case "done":
		status = "completed"
	case "error":
		status = "failed"
	}
	vals := []composeRecordValue{
		{Name: "status", Value: status},
		{Name: "progress", Value: fmt.Sprintf("%.0f", s.Progress)},
		{Name: "found", Value: fmt.Sprintf("%d", s.Found)},
		{Name: "scanning_ip", Value: s.ScanningIP},
		{Name: "error", Value: firstNonEmpty(s.Error, s.Message)},
	}
	if !s.StartedAt.IsZero() {
		vals = append(vals, composeRecordValue{Name: "started_at", Value: s.StartedAt.Format(time.RFC3339)})
	}
	if s.FinishedAt != nil {
		vals = append(vals, composeRecordValue{Name: "finished_at", Value: s.FinishedAt.Format(time.RFC3339)})
	}
	if s.Target != "" {
		vals = append(vals, composeRecordValue{Name: "target", Value: s.Target})
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	_, err = c.request(ctx, "POST", path, map[string]interface{}{"values": vals})
	return err
}

func (c *CortezaStore) syncRelated(ctx context.Context, _ uint64, devices []Device) {
	svcMod, svcErr := c.moduleByHandle(ctx, "services")
	if svcErr != nil {
		log.Printf("syncRelated services: %v", svcErr)
		svcMod = 0
	}
	vulnMod, vulnErr := c.moduleByHandle(ctx, "vulnerabilities")
	if vulnErr != nil {
		log.Printf("syncRelated vulnerabilities: %v", vulnErr)
		vulnMod = 0
	}
	if svcMod == 0 && vulnMod == 0 {
		return
	}
	for _, d := range devices {
		devID := d.RecordID
		if devID == 0 {
			id, err := c.FindDevice(ctx, 0, d)
			if err != nil || id == 0 {
				continue
			}
			devID = id
		}
		if svcMod != 0 {
			for _, p := range d.OpenPorts {
				if err := c.upsertService(ctx, svcMod, devID, p); err != nil {
					log.Printf("sync service %s:%d: %v", d.IP, p.Port, err)
				}
			}
		}
		if vulnMod != 0 {
			for _, v := range d.Vulnerabilities {
				if strings.TrimSpace(v.Name) == "" {
					continue
				}
				if err := c.upsertVuln(ctx, vulnMod, devID, v, d.LastSeen); err != nil {
					log.Printf("sync vuln %s %s: %v", d.IP, v.Name, err)
				}
			}
		}
	}
}

func (c *CortezaStore) upsertService(ctx context.Context, modID, deviceID uint64, p Port) error {
	if p.Port <= 0 {
		return nil
	}
	proto := strings.TrimSpace(p.Proto)
	if proto == "" {
		proto = "tcp"
	}
	vals := []composeRecordValue{
		{Name: "device", Value: strconv.FormatUint(deviceID, 10)},
		{Name: "port", Value: strconv.Itoa(p.Port)},
		{Name: "proto", Value: proto},
		{Name: "service", Value: p.Service},
		{Name: "version", Value: p.Version},
		{Name: "banner", Value: p.Banner},
	}
	existing, err := c.findRelated(ctx, modID, deviceID, map[string]string{
		"port":  strconv.Itoa(p.Port),
		"proto": proto,
	})
	if err != nil {
		return err
	}
	return c.writeRecord(ctx, modID, existing, vals)
}

func (c *CortezaStore) upsertVuln(ctx context.Context, modID, deviceID uint64, v Vulnerability, detectedAt string) error {
	vals := []composeRecordValue{
		{Name: "device", Value: strconv.FormatUint(deviceID, 10)},
		{Name: "name", Value: v.Name},
		{Name: "severity", Value: v.Severity},
		{Name: "cve", Value: v.CVE},
		{Name: "description", Value: v.Description},
		{Name: "remediation", Value: v.Remediation},
		{Name: "status", Value: "open"},
	}
	if strings.TrimSpace(detectedAt) != "" {
		vals = append(vals, composeRecordValue{Name: "detected_at", Value: detectedAt})
	}
	existing, err := c.findRelated(ctx, modID, deviceID, map[string]string{"name": v.Name})
	if err != nil {
		return err
	}
	return c.writeRecord(ctx, modID, existing, vals)
}

func (c *CortezaStore) writeRecord(ctx context.Context, modID, recordID uint64, vals []composeRecordValue) error {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return err
	}
	body := map[string]interface{}{"values": vals}
	if recordID == 0 {
		path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/", nsID, modID)
		_, err = c.request(ctx, "POST", path, body)
		return err
	}
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/%d", nsID, modID, recordID)
	_, err = c.request(ctx, "POST", path, body)
	return err
}

func (c *CortezaStore) findRelated(ctx context.Context, modID, deviceID uint64, extra map[string]string) (uint64, error) {
	nsID, err := c.resolveNamespace(ctx)
	if err != nil {
		return 0, err
	}
	q := url.QueryEscape("device = " + qlLiteral(strconv.FormatUint(deviceID, 10)))
	path := fmt.Sprintf("/compose/namespace/%d/module/%d/record/?query=%s&limit=200", nsID, modID, q)
	raw, err := c.request(ctx, "GET", path, nil)
	if err != nil {
		return 0, err
	}
	set, err := unwrapSet[composeRecord](raw)
	if err != nil {
		return 0, err
	}
	for _, rec := range set {
		match := true
		for k, want := range extra {
			if !relatedFieldEqual(recordField(rec, k), want) {
				match = false
				break
			}
		}
		if match {
			return rec.ID, nil
		}
	}
	return 0, nil
}

func relatedFieldEqual(got, want string) bool {
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)
	if strings.EqualFold(got, want) {
		return true
	}
	gf, gErr := strconv.ParseFloat(got, 64)
	wf, wErr := strconv.ParseFloat(want, 64)
	return gErr == nil && wErr == nil && gf == wf
}

func recordToDevice(rec composeRecord) Device {
	d := Device{RecordID: rec.ID}
	for _, v := range rec.Values {
		switch v.Name {
		case "ip_address":
			d.IP = v.Value
		case "mac_address":
			d.MAC = v.Value
		case "hostname":
			d.Hostname = v.Value
		case "vendor":
			d.Vendor = v.Value
		case "device_type":
			d.DeviceType = v.Value
		case "os":
			d.OS = v.Value
		case "domain":
			d.Domain = v.Value
		case "last_seen":
			d.LastSeen = v.Value
		case "status":
			d.Status = v.Value
		case "open_ports":
			_ = json.Unmarshal([]byte(v.Value), &d.OpenPorts)
		case "services":
			_ = json.Unmarshal([]byte(v.Value), &d.Services)
		case "shares":
			_ = json.Unmarshal([]byte(v.Value), &d.Shares)
		case "vulnerabilities":
			_ = json.Unmarshal([]byte(v.Value), &d.Vulnerabilities)
		}
	}
	if d.Status == "" {
		d.Status = "unknown"
	}
	return d
}

func deviceToValues(d Device) []composeRecordValue {
	portsJSON, _ := json.Marshal(d.OpenPorts)
	svcJSON, _ := json.Marshal(d.Services)
	sharesJSON, _ := json.Marshal(d.Shares)
	vulnJSON, _ := json.Marshal(d.Vulnerabilities)
	kv := [][2]string{
		{"ip_address", d.IP},
		{"mac_address", d.MAC},
		{"hostname", d.Hostname},
		{"vendor", d.Vendor},
		{"model", d.Model},
		{"device_type", d.DeviceType},
		{"os", d.OS},
		{"domain", d.Domain},
		{"open_ports", string(portsJSON)},
		{"services", string(svcJSON)},
		{"shares", string(sharesJSON)},
		{"vulnerabilities", string(vulnJSON)},
		{"last_seen", d.LastSeen},
		{"status", d.Status},
	}
	out := make([]composeRecordValue, 0, len(kv))
	for _, p := range kv {
		out = append(out, composeRecordValue{Name: p[0], Value: p[1]})
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func apiStatusError(code int, raw []byte) error {
	msg := strings.TrimSpace(string(raw))
	var envelope struct {
		Error json.RawMessage `json:"error"`
	}
	if json.Unmarshal(raw, &envelope) == nil && len(envelope.Error) > 0 {
		var obj map[string]interface{}
		if json.Unmarshal(envelope.Error, &obj) == nil {
			if m, _ := obj["message"].(string); strings.TrimSpace(m) != "" {
				msg = m
			}
		} else {
			var s string
			if json.Unmarshal(envelope.Error, &s) == nil && strings.TrimSpace(s) != "" {
				msg = s
			}
		}
	}
	if i := strings.Index(msg, "\n"); i > 0 {
		msg = strings.TrimSpace(msg[:i])
	}
	if len(msg) > 300 {
		msg = msg[:300] + "..."
	}
	err := fmt.Errorf("API error %d: %s", code, msg)
	if code == http.StatusUnauthorized || code == http.StatusForbidden {
		return fmt.Errorf("%w: %s", errUnauthorizedAPI, err.Error())
	}
	return err
}
