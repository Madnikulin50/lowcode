package agent

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func packTestArchive(t *testing.T, files map[string]string) *bytes.Buffer {
	t.Helper()
	src := t.TempDir()
	for name, body := range files {
		full := filepath.Join(src, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	walker, err := NewLocalFS(src)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, _, err := PackTarGz(context.Background(), &buf, walker, nil); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestBrowseTarGzPrefixAndLimit(t *testing.T) {
	archive := packTestArchive(t, map[string]string{
		"docs/a.txt": "AAA",
		"docs/b.txt": "BBB",
		"other/c.md": "CCC",
	})
	entries, err := browseTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("entries %d", len(entries))
	}
	entries, err = browseTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "docs", 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name, "docs/") {
			t.Fatalf("prefix leak %s", e.Name)
		}
	}
	// Лимит режет выдачу.
	entries, err = browseTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("limit %d", len(entries))
	}
}

func TestExtractFromTarGzFileAndDir(t *testing.T) {
	archive := packTestArchive(t, map[string]string{
		"docs/a.txt":     "AAA",
		"docs/sub/b.txt": "BBB",
		"top.txt":        "TOP",
	})
	// Один файл.
	dst := t.TempDir()
	files, n, err := extractFromTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "docs/a.txt", dst)
	if err != nil {
		t.Fatal(err)
	}
	if files != 1 || n != 3 {
		t.Fatalf("files=%d n=%d", files, n)
	}
	body, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil || string(body) != "AAA" {
		t.Fatalf("body %q %v", body, err)
	}
	// Поддерево.
	dst2 := t.TempDir()
	files, n, err = extractFromTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "docs", dst2)
	if err != nil {
		t.Fatal(err)
	}
	if files != 2 {
		t.Fatalf("dir files %d", files)
	}
	if _, err := os.Stat(filepath.Join(dst2, "sub", "b.txt")); err != nil {
		t.Fatal(err)
	}
	// Отсутствующий путь — ошибка, а не пустой успех.
	if _, _, err := extractFromTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "nope.txt", t.TempDir()); err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestExtractInlineTarGz(t *testing.T) {
	archive := packTestArchive(t, map[string]string{"hello.txt": "hello world"})
	inline, n, err := extractInlineTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if inline != "hello world" || n != 11 {
		t.Fatalf("inline %q n=%d", inline, n)
	}
	if _, _, err := extractInlineTarGz(context.Background(), bytes.NewReader(archive.Bytes()), "absent"); err == nil {
		t.Fatal("expected not-found")
	}
}

func TestParseLsOutput(t *testing.T) {
	out := "/data\n/data/docs\n/data/docs/a.txt\n/data/docs/b.txt\n"
	entries := parseLsOutput([]byte(out))
	// Корень (/data) пропускаем.
	if len(entries) != 3 {
		t.Fatalf("entries %d: %+v", len(entries), entries)
	}
	if entries[0].Name != "/data/docs" {
		t.Fatalf("first %q", entries[0].Name)
	}
}
