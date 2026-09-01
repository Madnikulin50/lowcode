package service

import (
	"context"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/server/pkg/rulesgo"
)

type fakeAttLoader struct {
	files map[uint64]*rulesgo.AttachmentBytes
}

func (f fakeAttLoader) Load(_ context.Context, _ uint64, id uint64) (*rulesgo.AttachmentBytes, error) {
	return f.files[id], nil
}

func TestDocumentExtractPicksIFC(t *testing.T) {
	ifc := []byte("ISO-10303-21;\nDATA;\n#1=IFCPROJECT('x',$,'Башня',$,$,$,$,$,$);\nENDSEC;\n")
	pln := append([]byte("GRAPHISOFT"), []byte("junk\x00BuildingX")...)
	ex := &documentExtractExecutor{load: fakeAttLoader{files: map[uint64]*rulesgo.AttachmentBytes{
		1: {ID: 1, Name: "house.pln", Data: pln},
		2: {ID: 2, Name: "house.ifc", Data: ifc},
	}}}
	ec := &rulesgo.ExecutionContext{
		Variables: map[string]interface{}{},
		Results:   map[string]interface{}{},
		Input:     map[string]interface{}{"file": "1\n2", "namespaceID": "9"},
	}
	out, err := ex.Execute(context.Background(), rulesgo.ChainNode{ID: "ex", Type: "document.extract"}, ec)
	if err != nil {
		t.Fatal(err)
	}
	if out["extract_kind"] != "ifc" {
		t.Fatalf("kind %v", out["extract_kind"])
	}
	text, _ := out["extracted_text"].(string)
	if !strings.Contains(text, "Башня") {
		t.Fatalf("text %q", text)
	}
	if out["extract_ok"] != true {
		t.Fatalf("ok %+v", out)
	}
}

func TestDocumentExtractNoFile(t *testing.T) {
	ex := &documentExtractExecutor{load: fakeAttLoader{files: map[uint64]*rulesgo.AttachmentBytes{}}}
	ec := &rulesgo.ExecutionContext{Variables: map[string]interface{}{}, Results: map[string]interface{}{}, Input: map[string]interface{}{}}
	out, err := ex.Execute(context.Background(), rulesgo.ChainNode{Type: "document.extract"}, ec)
	if err != nil {
		t.Fatal(err)
	}
	if out["extract_ok"] != false {
		t.Fatalf("%+v", out)
	}
}
