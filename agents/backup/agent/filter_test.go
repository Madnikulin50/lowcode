package agent

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseExclude(t *testing.T) {
	got := parseExclude("*.tmp, node_modules\n# comment\ncache/**\n  \n")
	want := []string{"*.tmp", "node_modules", "cache/**"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if parseExclude("") != nil || parseExclude("# only comment\n") != nil {
		t.Fatal("empty spec")
	}
}

func TestMatchExclude(t *testing.T) {
	patterns := parseExclude("*.tmp, node_modules, cache/**, docs/*.md")
	cases := []struct {
		rel  string
		want bool
	}{
		{"a.tmp", true},
		{"sub/dir/b.tmp", true}, // glob без '/' — по base name
		{"keep.txt", false},
		{"node_modules/pkg/index.js", true}, // имя каталога — всё поддерево
		{"src/node_modules/x.js", true},     // на любом уровне
		{"cache/a/b/c.bin", true},           // cache/**
		{"cached/x", false},                 // префикс «cache» без слэша не матчит
		{"docs/readme.md", true},
		{"docs/sub/readme.md", false}, // glob docs/*.md не глубже
	}
	for _, tc := range cases {
		if got := matchExclude(tc.rel, patterns); got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.rel, got, tc.want)
		}
	}
}

func TestExcludeWalkerFiltersPack(t *testing.T) {
	src := t.TempDir()
	write := func(name, body string) {
		full := filepath.Join(src, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("keep.txt", "K")
	write("junk.tmp", "J")
	write("node_modules/m/index.js", "N")

	inner, err := NewLocalFS(src)
	if err != nil {
		t.Fatal(err)
	}
	walker := newExcludeWalker(inner, "*.tmp, node_modules")
	var seen []string
	if err := walker.Walk(context.Background(), func(e FileEntry, _ openFn) error {
		seen = append(seen, e.RelPath)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1 || seen[0] != "keep.txt" {
		t.Fatalf("seen %v", seen)
	}
	// Пустой spec — тот же walker (обёртка не мешает).
	if newExcludeWalker(inner, "  ") != FileWalker(inner) {
		t.Fatal("empty spec should return inner walker")
	}
}

func TestKeyForDestPrefix(t *testing.T) {
	pol := &Policy{DestPrefix: "tenant-a"}
	got := keyFor(pol, nil, "demo", "job1", "archive.tar.gz")
	want := "tenant-a/demo/" // ObjectKey добавит дату/джоб/имя
	if got[:len(want)] != want {
		t.Fatalf("key %s", got)
	}
	noPol := keyFor(nil, nil, "demo", "job1", "archive.tar.gz")
	if noPol != ObjectKey("demo", "job1", "archive.tar.gz") {
		t.Fatalf("noPol %s", noPol)
	}
	empty := keyFor(&Policy{}, nil, "demo", "job1", "archive.tar.gz")
	if empty != noPol {
		t.Fatalf("empty prefix %s", empty)
	}
}

func TestKeyForDestinationPrefix(t *testing.T) {
	dest := &Destination{Prefix: "team-b"}
	// Префикс назначения применяется, когда у политики своего нет.
	got := keyFor(nil, dest, "demo", "job1", "archive.tar.gz")
	want := "team-b/demo/"
	if got[:len(want)] != want {
		t.Fatalf("dest prefix key %s", got)
	}
	// dest_prefix политики побеждает префикс назначения.
	got = keyFor(&Policy{DestPrefix: "tenant-a"}, dest, "demo", "job1", "archive.tar.gz")
	want = "tenant-a/demo/"
	if got[:len(want)] != want {
		t.Fatalf("policy prefix must win, key %s", got)
	}
	// Пустой префикс назначения — без эффекта.
	empty := keyFor(nil, &Destination{}, "demo", "job1", "archive.tar.gz")
	if empty != ObjectKey("demo", "job1", "archive.tar.gz") {
		t.Fatalf("empty dest prefix %s", empty)
	}
}
