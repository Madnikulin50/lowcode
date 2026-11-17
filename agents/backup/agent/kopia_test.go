package agent

import (
	"strings"
	"testing"
)

func TestKopiaConnectArgs(t *testing.T) {
	cfg := Config{
		Minio: MinioConfig{
			Endpoint: "minio:9000",
			Access:   "ak",
			Secret:   "sk",
			Bucket:   "backups",
			Region:   "ru-1",
			Secure:   false,
		},
	}
	kargs, err := kopiaConnectArgs(t.Context(), cfg, nil, "demo fs", false)
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(kargs, " ")
	for _, want := range []string{
		"--bucket=backups",
		"--prefix=kopia/demo-fs/",
		"--endpoint=minio:9000",
		"--access-key=ak",
		"--secret-access-key=sk",
		"--region=ru-1",
		"--disable-tls",
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("args missing %q: %s", want, args)
		}
	}
	if strings.Contains(args, "http://") {
		t.Fatalf("endpoint must be bare host:port: %s", args)
	}
	// TLS-режим: без --disable-tls.
	cfg.Minio.Secure = true
	cfg.Minio.Endpoint = "https://minio.example.com"
	kargs, err = kopiaConnectArgs(t.Context(), cfg, nil, "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	args = strings.Join(kargs, " ")
	if strings.Contains(args, "--disable-tls") || strings.Contains(args, "https://") {
		t.Fatalf("secure args: %s", args)
	}
}

func TestParseKopiaSnapshotID(t *testing.T) {
	raw := []byte(`[
	  {"id":"abc123","rootID":"def456","path":"/data"},
	  {"id":"older","rootID":"x","path":"/data"}
	]`)
	if got := parseKopiaSnapshotID(raw); got != "abc123" {
		t.Fatalf("id %q", got)
	}
	if got := parseKopiaSnapshotID([]byte(`[]`)); got != "" {
		t.Fatalf("empty %q", got)
	}
	if got := parseKopiaSnapshotID([]byte(`not json`)); got != "" {
		t.Fatalf("garbage %q", got)
	}
}

func TestKopiaPrefixAndKeyHint(t *testing.T) {
	if got := kopiaPrefix("demo fs"); got != "kopia/demo-fs/" {
		t.Fatalf("prefix %q", got)
	}
	hint := kopiaKeyHint("demo", "abc")
	if !strings.HasPrefix(hint, "kopia/demo/manifests/abc") {
		t.Fatalf("hint %q", hint)
	}
}

func TestPolicyEngineParsing(t *testing.T) {
	rec := &composeRecord{ID: 1, Values: []composeRecordValue{
		{Name: "name", Value: "p"},
		{Name: "source", Value: "2"},
		{Name: "cron", Value: "0 2 * * *"},
		{Name: "engine", Value: "Kopia "},
	}}
	p := policyFrom(rec)
	if p.Engine != "kopia" {
		t.Fatalf("engine %q", p.Engine)
	}
	rec.Values = rec.Values[:3]
	if p := policyFrom(rec); p.Engine != "" {
		t.Fatalf("empty engine %q", p.Engine)
	}
}
