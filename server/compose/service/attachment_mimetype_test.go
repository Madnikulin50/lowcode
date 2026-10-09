package service

import (
	"reflect"
	"testing"

	"github.com/madnikulin50/lowcode/server/compose/types"
)

func TestFieldAllowedAttachmentTypes(t *testing.T) {
	fallback := []string{"image/png"}

	both := &types.ModuleField{Kind: "File", Options: types.ModuleFieldOptions{
		"allowImages":    true,
		"allowDocuments": true,
	}}
	if got := fieldAllowedAttachmentTypes(both, fallback); got != nil {
		t.Fatalf("both flags should clear the constraint, got %v", got)
	}

	docs := &types.ModuleField{Kind: "File", Options: types.ModuleFieldOptions{
		"allowImages":    false,
		"allowDocuments": true,
	}}
	got := fieldAllowedAttachmentTypes(docs, fallback)
	if len(got) == 0 {
		t.Fatal("documents-only field should return a document mime list")
	}
	foundPDF := false
	for _, m := range got {
		if m == "application/pdf" {
			foundPDF = true
			break
		}
	}
	if !foundPDF {
		t.Fatalf("documents-only list missing application/pdf: %v", got)
	}
	wantExt := map[string]bool{".dxf": true, ".dwg": true, ".ifc": true, ".pln": true, ".bimx": true}
	for _, m := range got {
		delete(wantExt, m)
	}
	if len(wantExt) > 0 {
		t.Fatalf("documents-only list missing CAD/BIM extensions %v in %v", wantExt, got)
	}

	explicit := &types.ModuleField{Kind: "File", Options: types.ModuleFieldOptions{
		"mimetypes":      "application/pdf, .docx",
		"allowImages":    true,
		"allowDocuments": true,
	}}
	if got := fieldAllowedAttachmentTypes(explicit, fallback); !reflect.DeepEqual(got, []string{"application/pdf", ".docx"}) {
		t.Fatalf("explicit mimetypes: got %v", got)
	}

	none := &types.ModuleField{Kind: "File", Options: types.ModuleFieldOptions{}}
	if got := fieldAllowedAttachmentTypes(none, fallback); !reflect.DeepEqual(got, fallback) {
		t.Fatalf("no flags should keep fallback, got %v", got)
	}
}
