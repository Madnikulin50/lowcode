package cmp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func setsFixture() (*fakeBackend, map[string]string) {
	f := newFake()
	ids := map[string]string{}
	ids["obj"] = f.put(ModObjects, idcheck.V("name", "ЖК Пример"))
	set := func(key, stage, version, issued, status string) {
		ids[key] = f.put(ModSets, idcheck.V("object", ids["obj"], "stage", stage, "version", version, "issued_at", issued, "status", status))
	}
	set("pd1", SideRef, "v1", "2024-01-10", SetSuperseded)
	set("pd2", SideRef, "v2", "2024-03-01", SetAccepted)
	set("rd1", SideTarget, "v1", "2024-04-01", SetAccepted)
	doc := func(set, code string) string {
		return f.put(ModDocuments, idcheck.V("set", ids[set], "file", set+"-"+code+".pdf", "volume_code", code, "pages", "10", "side", SideRef))
	}
	doc("pd1", "КЖ1")
	doc("pd2", "КЖ1")
	doc("pd2", "АР")
	doc("pd2", "ПЗ")
	doc("pd2", "ПЗ") // duplicate
	doc("rd1", "100-24/РД-КЖ1")
	doc("rd1", "100-24/РД-ОВ") // not in the statement
	for _, c := range []string{"КЖ1", "АР", "ПЗ", "ЭОМ"} {
		f.put(ModVolumes, idcheck.V("object", ids["obj"], "volume_code", c, "title", "Том "+c))
	}
	ids["cmp"] = f.put(ModComparisons, idcheck.V("title", "ПД↔РД", "object", ids["obj"], "set_pd", ids["pd1"], "set_rd", ids["rd1"], "status", "done"))
	return f, ids
}

func TestObjectDocsCompletenessAndPairs(t *testing.T) {
	f, ids := setsFixture()
	od, err := LoadObjectDocs(f, "ns", ids["obj"])
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]SetInfo{}
	for _, s := range od.Sets {
		by[s.ID] = s
	}
	pd2 := by[ids["pd2"]]
	if !pd2.Latest || by[ids["pd1"]].Latest || by[ids["pd1"]].Newer != "v2" || pd2.Newer != "" {
		t.Fatalf("latest ПД must be v2, got pd1=%v pd2=%v", by[ids["pd1"]].Latest, pd2.Latest)
	}
	if pd2.Total != 4 || pd2.Present != 3 || pd2.Pct != 75 {
		t.Fatalf("ПД v2: %d of %d (%d%%), want 3 of 4 (75%%): %s", pd2.Present, pd2.Total, pd2.Pct, pd2.Note)
	}
	if len(pd2.Missing) != 1 || pd2.Missing[0] != "ЭОМ" || len(pd2.Duplicates) != 1 {
		t.Fatalf("missing=%v dups=%v", pd2.Missing, pd2.Duplicates)
	}
	rd1 := by[ids["rd1"]]
	if rd1.Present != 1 || len(rd1.Extra) != 1 || !strings.Contains(rd1.Note, "лишние") {
		t.Fatalf("РД v1: present=%d extra=%v note=%s", rd1.Present, rd1.Extra, rd1.Note)
	}

	if len(od.Pairs) != 1 {
		t.Fatalf("pairs: %+v", od.Pairs)
	}
	p := od.Pairs[0]
	if p.Label != "ПД v1 ↔ РД v1" {
		t.Fatalf("label %q", p.Label)
	}
	joined := strings.Join(p.Warnings, " | ")
	if !strings.Contains(joined, "есть более новая — v2") || !strings.Contains(joined, "РД v1 неполон") {
		t.Fatalf("warnings: %s", joined)
	}

	// matrix: the four expected volumes plus the one only РД has
	rows := map[string]MatrixRow{}
	for _, r := range od.Matrix {
		rows[r.Code] = r
	}
	if rows["ЭОМ"].Ref.State != CellMissing || rows["ЭОМ"].Target.State != CellMissing {
		t.Fatalf("ЭОМ: %+v", rows["ЭОМ"])
	}
	if rows["КЖ1"].Ref.State != CellPresent || rows["КЖ1"].Target.State != CellPresent || rows["КЖ1"].Ref.Version != "v2" {
		t.Fatalf("КЖ1: %+v", rows["КЖ1"])
	}
	if extra, ok := rows["100-24/РД-ОВ"]; !ok || extra.Target.State != CellExtra {
		t.Fatalf("extra row missing: %+v", od.Matrix)
	}
}

func TestObjectDocsWithoutStatement(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	s := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v1"))
	f.put(ModDocuments, idcheck.V("set", s, "file", "a.pdf", "volume_code", "КЖ1", "pages", "3"))
	f.put(ModDocuments, idcheck.V("set", s, "file", "b.pdf")) // not read yet
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	got := od.Sets[0]
	if got.Total != 0 || got.Unparsed != 1 || !strings.Contains(got.Note, "ведомость состава не задана") {
		t.Fatalf("%+v", got)
	}
	if od.Matrix[0].Ref.State != CellPresent || od.Matrix[0].Target.State != CellNoSet {
		t.Fatalf("matrix: %+v", od.Matrix)
	}
}

func TestCheckSetsWritesBack(t *testing.T) {
	f, ids := setsFixture()
	r := &Runner{Backend: f}
	if _, err := r.CheckSets(context.Background(), "ns", ids["obj"]); err != nil {
		t.Fatal(err)
	}
	rec, _ := f.GetRecord("ns", ModSets, ids["pd2"])
	if rec.Get("volumes_total") != "4" || rec.Get("volumes_present") != "3" || rec.Get("completeness_pct") != "75" {
		t.Fatalf("set values: %+v", rec.Values)
	}
	c, _ := f.GetRecord("ns", ModComparisons, ids["cmp"])
	if c.Get("sets_label") != "ПД v1 ↔ РД v1" {
		t.Fatalf("label %q", c.Get("sets_label"))
	}
}

func TestListDocsTakesFilesFromSets(t *testing.T) {
	f, ids := setsFixture()
	docs, err := ListDocs(f, "ns", ids["cmp"])
	if err != nil {
		t.Fatal(err)
	}
	var ref, tgt int
	for _, d := range docs {
		if d.Side == SideRef {
			ref++
		} else {
			tgt++
		}
	}
	if ref != 1 || tgt != 2 { // ПД v1 (КЖ1) and РД v1 (two files); not the ПД v2 files
		t.Fatalf("ref=%d target=%d docs=%+v", ref, tgt, docs)
	}
}

func TestUndatedSetsAreNotVersions(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	a := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "Исходная (2023)", "status", SetAccepted))
	f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "Исходная (2024)", "status", SetAccepted))
	f.put(ModComparisons, idcheck.V("title", "С", "object", obj, "set_pd", a))
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range od.Pairs {
		for _, w := range p.Warnings {
			if strings.Contains(w, "более новая") {
				t.Fatalf("packs without dates must not be called outdated: %v", p.Warnings)
			}
		}
	}
}

func TestVersionsPairLabelUsesArrowAndSetNames(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	a := f.put(ModSets, idcheck.V("object", obj, "stage", SideTarget, "version", "Изм. 1", "issued_at", "2024-01-01"))
	b := f.put(ModSets, idcheck.V("object", obj, "stage", SideTarget, "version", "Изм. 2", "issued_at", "2024-03-01"))
	f.put(ModComparisons, idcheck.V("title", "до/после", "object", obj, "kind", ModeVersions, "set_pd", a, "set_rd", b))
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	p := od.Pairs[0]
	if p.Label != "РД Изм. 1 → РД Изм. 2" || p.Kind != ModeVersions {
		t.Fatalf("label %q kind %q", p.Label, p.Kind)
	}
	if len(p.Warnings) == 0 || !strings.Contains(p.Warnings[0], "РД: сравнивается Изм. 1, есть более новая — Изм. 2") {
		t.Fatalf("warnings %v", p.Warnings)
	}
}

func TestSetDocWithoutCodesSerializesAnEmptyList(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	s := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v1"))
	f.put(ModDocuments, idcheck.V("set", s, "file", "a.pdf")) // not read yet: no codes
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(od)
	if strings.Contains(string(b), `"codes":null`) || strings.Contains(string(b), ":null") {
		t.Fatalf("the viewer reads .length of every list, null breaks it: %s", b)
	}
}

func TestSetFlagsATitleThatDiffersFromTheStatement(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "Объект"))
	set := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v1", "issued_at", "2024-01-10", "status", SetAccepted))
	doc := func(code, title string) {
		f.put(ModDocuments, idcheck.V("set", set, "file", code+".pdf", "volume_code", code, "volume_title", title, "pages", "5", "side", SideRef))
	}
	doc("КЖ1", "Фундаментная плита. Конструктивные решения")
	doc("АР", "Схема планировочной организации земельного участка")
	doc("ПЗ", "")
	for _, v := range [][2]string{{"КЖ1", "Конструктивные решения фундаментов"}, {"АР", "Архитектурные решения"}, {"ПЗ", "Пояснительная записка"}} {
		f.put(ModVolumes, idcheck.V("object", obj, "volume_code", v[0], "title", v[1]))
	}
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	s := od.Sets[0]
	if len(s.Renamed) != 1 || !strings.HasPrefix(s.Renamed[0], "АР:") || !strings.Contains(s.Note, "наименования не согласованы") {
		t.Fatalf("only АР has a title with no word in common (КЖ1 shares «конструктивные», ПЗ has no title): %+v\n%s", s.Renamed, s.Note)
	}
}
