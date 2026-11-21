package agent

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestPackTarGzSymlink guards against a regression of "archive/tar: write
// too long": a symlink's Lstat().Size() is the length of the link text, but
// os.Open() on the path follows the link and reads the (differently sized)
// target. The walker must report symlinks via FileEntry.LinkTarget instead
// of an opener, so PackTarGz writes a bodyless tar.TypeSymlink record.
func TestPackTarGzSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(target, bytes.Repeat([]byte("x"), 4096), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	walker, err := NewLocalFS(dir)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	files, _, err := PackTarGz(context.Background(), &buf, walker, nil)
	if err != nil {
		t.Fatalf("PackTarGz: %v", err)
	}
	if files != 2 {
		t.Fatalf("files = %d, want 2", files)
	}

	dest := t.TempDir()
	if _, err := UnpackTarGz(context.Background(), &buf, dest); err != nil {
		t.Fatalf("UnpackTarGz: %v", err)
	}
	got, err := os.Readlink(filepath.Join(dest, "link"))
	if err != nil {
		t.Fatalf("restored entry is not a symlink: %v", err)
	}
	if got != target {
		t.Fatalf("link target = %q, want %q", got, target)
	}
}
