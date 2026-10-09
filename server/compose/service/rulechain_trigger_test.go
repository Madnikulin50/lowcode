package service

import (
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

func TestFileFieldKey(t *testing.T) {
	rec := &types.Record{Values: types.RecordValueSet{
		{Name: "file", Value: "22", Place: 1},
		{Name: "file", Value: "11", Place: 0},
		{Name: "title", Value: "x"},
	}}
	got := fileFieldKey(rec, "file")
	if got != "11,22" {
		t.Fatalf("got %q", got)
	}
	if fileFieldKey(rec, "missing") != "" {
		t.Fatal("expected empty")
	}
}

func TestRecordTriggerBagJoinsMultiFile(t *testing.T) {
	rec := &types.Record{ID: 9, NamespaceID: 1, Values: types.RecordValueSet{
		{Name: "file", Value: "11"},
		{Name: "file", Value: "22", Place: 1},
	}}
	mod := &types.Module{ID: 2, Handle: "documents"}
	bag := recordTriggerBag(rec, mod, 1)
	if bag["recordID"] != "9" {
		t.Fatalf("%v", bag["recordID"])
	}
	if bag["file"] != "11\n22" {
		t.Fatalf("file %v", bag["file"])
	}
}
