package rulesgo

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilePersistenceRoundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rulechains.json")
	p := NewFilePersistence(path)
	chain := &Chain{ID: "cmdb-trigger-scan", Name: "CMDB scan", EntryNode: "http_scan"}
	if err := p.SaveChain(context.Background(), chain); err != nil {
		t.Fatal(err)
	}
	p2 := NewFilePersistence(path)
	list, err := p2.LoadChains(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != "cmdb-trigger-scan" {
		t.Fatalf("got %#v", list)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
