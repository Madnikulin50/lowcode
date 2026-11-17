package agent

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Пагинация Bitrix24: два захода, второй с start=2 возвращает пусто.
func TestBitrix24Pagination(t *testing.T) {
	var starts []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/rest/tok/crm.lead.list.json") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		start := r.URL.Query().Get("start")
		starts = append(starts, start)
		w.Header().Set("Content-Type", "application/json")
		if start == "0" || start == "" {
			_, _ = w.Write([]byte(`{"result":[{"ID":"1"},{"ID":"2"}],"next":2}`))
			return
		}
		_, _ = w.Write([]byte(`{"result":[]}`))
	}))
	t.Cleanup(srv.Close)

	rows, err := bitrixListAll(context.Background(), srv.URL, "tok", "crm.lead.list")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows %d", len(rows))
	}
	if len(starts) != 2 || starts[0] != "0" || starts[1] != "2" {
		t.Fatalf("starts %v", starts)
	}
}

func TestWriteSaasTarGzRoundtrip(t *testing.T) {
	files := []saasExportFile{{
		Name: "bitrix24/crm_lead_list.json",
		Mod:  time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC),
		Body: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(`[{"ID":"1"}]`)), nil
		},
	}}
	var buf bytes.Buffer
	n, err := writeSaasTarGz(context.Background(), &buf, files)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("n %d", n)
	}
	gz, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(gz)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Name != "bitrix24/crm_lead_list.json" || hdr.Size != int64(len(`[{"ID":"1"}]`)) {
		t.Fatalf("hdr %+v", hdr)
	}
	body, _ := io.ReadAll(tr)
	if string(body) != `[{"ID":"1"}]` {
		t.Fatalf("body %s", body)
	}
}

func TestSaasTokenExtraction(t *testing.T) {
	cred := &Credential{Extra: `{"token":"abc"}`}
	if got := saasToken(cred); got != "abc" {
		t.Fatalf("json token %q", got)
	}
	cred = &Credential{Extra: "raw-token"}
	if got := saasToken(cred); got != "raw-token" {
		t.Fatalf("raw %q", got)
	}
	t.Setenv("BACKUP_SECRET_x", "env-secret")
	cred = &Credential{Handle: "x", Extra: ""}
	if got := saasToken(cred); got != "env-secret" {
		t.Fatalf("env %q", got)
	}
}

func TestK8sGVRFor(t *testing.T) {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("apps/v1")
	obj.SetKind("Deployment")
	gvr, err := k8sGVRFor(obj)
	if err != nil {
		t.Fatal(err)
	}
	if gvr.Group != "apps" || gvr.Version != "v1" || gvr.Resource != "deployments" {
		t.Fatalf("gvr %s", gvr)
	}
	obj.SetKind("UnknownKind")
	if _, err := k8sGVRFor(obj); err != nil {
		t.Fatal(err) // дефолтное plural-правило тоже работает
	}
}

func TestCleanPointSkipsInfected(t *testing.T) {
	// Compose-mock: namespace -> module -> records.
	mux := http.NewServeMux()
	mux.HandleFunc("/compose/namespace/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"set":[{"namespaceID":"1","slug":"backup"}]}}`))
	})
	mux.HandleFunc("/compose/namespace/1/module/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.RawQuery, "handle=snapshots") {
			_, _ = w.Write([]byte(`{"response":{"set":[{"moduleID":"9","handle":"snapshots"}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"response":{"set":[]}}`))
	})
	mux.HandleFunc("/compose/namespace/1/module/9/record/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":{"set":[
			{"recordID":"101","values":[
				{"name":"source","value":"1"},{"name":"scan_status","value":"infected"},
				{"name":"restorable","value":"1"},{"name":"s3_key","value":"bad"}]},
			{"recordID":"102","values":[
				{"name":"source","value":"1"},{"name":"scan_status","value":"clean"},
				{"name":"restorable","value":"1"},{"name":"s3_key","value":"good"},
				{"name":"size_bytes","value":"42"},{"name":"engine","value":"archive"}]}
		]}}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	a := New(Config{}, nil, NewCorteza(srv.URL, "tok", 0))
	cp, err := a.CleanPoint(context.Background(), JobRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if cp["snapshotID"] != "102" || cp["s3Key"] != "good" {
		t.Fatalf("clean point %+v", cp)
	}
	_ = json.Valid
}
