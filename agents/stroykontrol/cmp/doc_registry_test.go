package cmp

import (
	"context"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func TestSectionOf(t *testing.T) {
	for in, want := range map[string]string{
		"2024-15-КЖ2": "КЖ",
		"2024-15-КЖ2, 2024-15-КЖ3, 2024-15-АР": "КЖ, АР",
		"03102022/РД-ИССО3.10-ОП-КЖ1.13":       "КЖ",
		"": "",
	} {
		if got := SectionOf(in); got != want {
			t.Errorf("SectionOf(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatchStructure(t *testing.T) {
	ss := []Structure{{ID: "s1", Name: "Путепровод на ПК 12+40", Code: "ИССО3"}, {ID: "s2", Name: "Мост через р. Ока", Code: "ИССО4"}}
	if got := MatchStructure([]string{"", "Опоры путепровода. Путепровод на ПК 12+40"}, ss); got != "s1" {
		t.Errorf("by name: %q", got)
	}
	if got := MatchStructure([]string{"ИССО4"}, ss); got != "s2" {
		t.Errorf("by code: %q", got)
	}
	if got := MatchStructure([]string{"ИССО3 и ИССО4"}, ss); got != "" {
		t.Errorf("two structures named must not link: %q", got)
	}
	if got := MatchStructure([]string{"ИССО31"}, ss); got != "" {
		t.Errorf("a code inside another code: %q", got)
	}
}

func TestReadSetDocsRegistryAndJournal(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	f.put(ModStructures, idcheck.V("object", obj, "name", "Фундаменты здания", "code", "КЖ-Ф"))
	pd := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v1"))
	rd := f.put(ModSets, idcheck.V("object", obj, "stage", SideTarget, "version", "v1"))
	d1 := f.put(ModDocuments, idcheck.V("set", pd, "file", "pack-1/rd-kzh.pdf", "structure_object", "Фундаменты здания", SysCreatedBy, "u-7", SysCreatedAt, "2026-10-20T10:00:00Z"))
	d2 := f.put(ModDocuments, idcheck.V("set", rd, "file", "pack-1/rd-kzh.pdf", "section", "КР")) // the same file again; the section set by hand
	d3 := f.put(ModDocuments, idcheck.V("set", rd, "file", "pack-1/rd-ar.pdf"))
	r := &Runner{Backend: f}
	od, err := r.CheckSets(context.Background(), "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	got := func(id, k string) string { return f.recs[ModDocuments][id].Get(k) }
	if got(d1, "section") != "КЖ" || got(d1, "structure") == "" {
		t.Errorf("d1: section %q structure %q", got(d1, "section"), got(d1, "structure"))
	}
	if got(d2, "section") != "КР" {
		t.Errorf("a section set by hand was overwritten: %q", got(d2, "section"))
	}
	if got(d3, "section") != "АР" || got(d3, "structure") != "" {
		t.Errorf("d3: %q %q", got(d3, "section"), got(d3, "structure"))
	}
	kinds := map[string]string{}
	for _, u := range f.recs[ModUploads] {
		kinds[u.Get("document")] = u.Get("kind")
		if u.Get("document") == d1 && (u.Get("uploaded_by") != "u-7" || u.Get("uploaded_at") != "2026-10-20T10:00:00Z") {
			t.Errorf("author of the upload: %v", u.Values)
		}
	}
	// the first of the two equal files is new, the second is its repeat
	if len(kinds) != 3 || kinds[d3] != UploadNew || (kinds[d1] == UploadRepeat) == (kinds[d2] == UploadRepeat) {
		t.Errorf("journal kinds %v", kinds)
	}
	if len(od.Uploads) != 3 || len(od.Tree) != 2 || od.Tree[0].Name != "Фундаменты здания (КЖ-Ф)" || od.Tree[1].Name != "сооружение не указано" {
		t.Fatalf("tree %+v uploads %d", od.Tree, len(od.Uploads))
	}
	// a second check does not journal the same files again
	if _, err := r.CheckSets(context.Background(), "ns", obj); err != nil {
		t.Fatal(err)
	}
	if n := len(f.recs[ModUploads]); n != 3 {
		t.Errorf("journal grew to %d", n)
	}
}

func TestUploadKindEdition(t *testing.T) {
	others := []idcheck.Record{{ID: "a", Values: map[string][]string{"volume_code": {"2024-15-КЖ2"}, "file_hash": {"h1"}, "change_no": {"1"}}}}
	if k, n := uploadKind("h2", "2024-15-КЖ2", "b", others); k != UploadEdition || n != "уже есть 2024-15-КЖ2 изм. 1" {
		t.Errorf("%s %s", k, n)
	}
	if k, _ := uploadKind("h2", "", "b", others); k != UploadNoCode {
		t.Errorf("no code: %s", k)
	}
}
