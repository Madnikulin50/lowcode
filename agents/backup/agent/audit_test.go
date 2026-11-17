package agent

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestGuessClass(t *testing.T) {
	cases := map[string]string{
		"список_персональных_данных.xlsx": "personal_special",
		"medical_card_2024.pdf":           "personal_special",
		"Паспорт скан.jpg":                "personal_special",
		"договор_аренды.docx":             "confidential",
		"report.pdf":                      "confidential",
		"base.1cd":                        "internal",
		"dump.sql":                        "internal",
		"notes.txt":                       "internal",
	}
	for name, want := range cases {
		if got := guessClass(name); got != want {
			t.Fatalf("%s: got %s want %s", name, got, want)
		}
	}
}

func TestMatchMasks(t *testing.T) {
	masks := []string{"*.docx", "*.xls*"}
	if !matchMasks("договор.docx", masks) {
		t.Fatal("docx should match")
	}
	if !matchMasks("sub/отчет.xlsx", masks) {
		t.Fatal("base name without slash pattern")
	}
	if matchMasks("photo.jpg", masks) {
		t.Fatal("jpg should not match")
	}
	if !matchMasks("anything.bin", nil) {
		t.Fatal("empty masks match all")
	}
}

func TestCoversPath(t *testing.T) {
	sources := []Source{
		{Type: SourceFS, Path: "/data/shared", Enabled: true},
		{Type: SourceFS, Path: "/data/other", Enabled: false},
		{Type: SourceDatabase, DBName: "erp", Enabled: true},
	}
	cases := []struct {
		path string
		want bool
	}{
		{"/data/shared/reports/a.xlsx", true},
		{"/data/shared2/x", false}, // префикс-ловушка: shared2 ≠ shared
		{"/data/other/secret.docx", false},
		{"/elsewhere/file.txt", false},
	}
	for _, tc := range cases {
		if got := coversPath(tc.path, sources); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.path, got, tc.want)
		}
	}
	// SMB UNC-пути
	smbSources := []Source{
		{Type: SourceSMB, Host: "nas", Share: "docs", SMBPath: "arch", Enabled: true},
	}
	if !coversPath("nas/docs/arch/2020/file.pdf", smbSources) {
		t.Fatal("unc covered")
	}
	if coversPath("nas/docs/other/file.pdf", smbSources) {
		t.Fatal("unc uncovered")
	}
}

func TestEnforceDataClass(t *testing.T) {
	// personal_special fs без kopia — отказ
	err := enforceDataClass(&Policy{Name: "p", Engine: "archive"}, &Source{DataClass: "personal_special", Type: SourceFS})
	if err == nil || !strings.Contains(err.Error(), "kopia") {
		t.Fatalf("pdn fs: %v", err)
	}
	if err := enforceDataClass(&Policy{Name: "p", Engine: "kopia"}, &Source{DataClass: "personal_special", Type: SourceFS}); err != nil {
		t.Fatalf("pdn fs kopia: %v", err)
	}
	// personal_special БД: kopia недоступен — достаточно vault
	if err := enforceDataClass(&Policy{Name: "p"}, &Source{DataClass: "personal_special", Type: SourceDatabase}); err == nil {
		t.Fatal("pdn db without vault should fail")
	}
	if err := enforceDataClass(&Policy{Name: "p", Vault: true}, &Source{DataClass: "personal_special", Type: SourceDatabase}); err != nil {
		t.Fatalf("pdn db vault: %v", err)
	}
	// confidential без vault — отказ
	err = enforceDataClass(&Policy{Name: "p"}, &Source{DataClass: "confidential"})
	if err == nil || !strings.Contains(err.Error(), "vault") {
		t.Fatalf("conf: %v", err)
	}
	if err := enforceDataClass(&Policy{Name: "p", Vault: true}, &Source{DataClass: "confidential"}); err != nil {
		t.Fatalf("conf vault: %v", err)
	}
	// internal/public — без ограничений
	if err := enforceDataClass(&Policy{Name: "p"}, &Source{DataClass: "internal"}); err != nil {
		t.Fatal(err)
	}
}

func TestHashTarGz(t *testing.T) {
	archive := packTestArchive(t, map[string]string{
		"a.txt":    "AAA",
		"dir/b.md": "BBB",
	})
	entries, err := hashTarGz(context.Background(), bytes.NewReader(archive.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries %d", len(entries))
	}
	for _, e := range entries {
		if len(e.SHA256) != 64 {
			t.Fatalf("sha %q", e.SHA256)
		}
	}
}

func TestAuditScopeFrom(t *testing.T) {
	rec := composeRecord{ID: 7, Values: []composeRecordValue{
		{Name: "name", Value: "файловая помойка"},
		{Name: "type", Value: "fs"},
		{Name: "path", Value: "/srv/share"},
		{Name: "masks", Value: "*.docx, *.xls*\n*.pdf"},
		{Name: "min_size_mb", Value: "10"},
		{Name: "enabled", Value: "1"},
	}}
	scope := auditScopeFrom(rec)
	if scope.ID != 7 || scope.Path != "/srv/share" || !scope.Enabled {
		t.Fatalf("scope %+v", scope)
	}
	if len(scope.Masks) != 3 {
		t.Fatalf("masks %v", scope.Masks)
	}
	if scope.MinSizeMB != 10 {
		t.Fatalf("minSize %d", scope.MinSizeMB)
	}
}

func TestParseKopiaLsLong(t *testing.T) {
	out := "-rw-r--r-- 1234 2024-01-02 03:04:05 /data/docs/a.txt\n-rw-r--r-- 0 2024-01-02 03:04:05 /data/empty.bin\n"
	entries := parseKopiaLsLong(out)
	if len(entries) != 2 || entries[0].Path != "/data/docs/a.txt" || entries[0].Size != 1234 {
		t.Fatalf("entries %+v", entries)
	}
}

func TestParseResticLs(t *testing.T) {
	out := "/data\n/data/docs\n/data/docs/a.txt\n"
	entries := parseResticLs(out)
	if len(entries) != 2 || entries[0].Path != "/data/docs" {
		t.Fatalf("entries %+v", entries)
	}
}

func TestComplianceMarkdown(t *testing.T) {
	d := map[string]any{
		"generatedAt": "2026-11-14T00:00:00Z",
		"sources": []map[string]any{
			{"name": "crm-db", "type": SourceDatabase, "dataClass": "confidential", "enabled": true},
		},
		"policies": []map[string]any{
			{"name": "nightly", "cron": "0 2 * * *", "retentionDays": 14, "engine": "archive", "vault": true, "archive": false},
		},
		"snapshots":         map[string]any{"total": 10, "restorable": 8, "verified": 5, "withManifest": 3},
		"auditOpenFindings": 2,
		"wormVault":         map[string]any{"enabled": true, "bucket": "vault", "mode": "compliance", "retentionDays": 30},
	}
	md := complianceMarkdown(d)
	for _, want := range []string{"План резервирования", "crm-db", "0 2 * * *", "WORM-сейф", "Открытых находок аудита"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown missing %q:\n%s", want, md)
		}
	}
}
