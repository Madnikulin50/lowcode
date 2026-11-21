package cmp

import (
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func TestProjectPrefixAndVolumeRefs(t *testing.T) {
	for in, want := range map[string]string{"2024-15-КЖ2": "2024-15-", "03102022/РД-ИССО3.10-ОП-КЖ1.13": "03102022/РД-ИССО3.10-ОП-", "КЖ1": ""} {
		if got := projectPrefix(in); got != want {
			t.Errorf("prefix(%q) = %q, want %q", in, got, want)
		}
	}
	pages := []Page{fragPage(1, "Армирование см. 2024-15-КЖ3 и том 2024-15-АР.", "Узлы по 2024-15-КЖ2 (этот том)"), {Number: 2, HasText: true, Lines: []Line{{Text: "2024-15-ОВ", Stamp: true}}}}
	got := ExtractVolumeRefs(pages, []string{"2024-15-КЖ2"})
	if len(got) != 2 || got[0].Code != "2024-15-АР" || got[1].Code != "2024-15-КЖ3" || got[1].Page != 1 {
		t.Errorf("refs %+v", got)
	}
}

func TestStatementUnits(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	s := f.put(ModSets, idcheck.V("object", obj, "stage", SideTarget, "version", "v1"))
	f.put(ModDocuments, idcheck.V("set", s, "file", "a.pdf", "volume_code", "2024-15-КЖ2", "volume_title", "Фундаменты", "pages", "9",
		"headings_json", `[{"p":7,"t":"Приложение А. Расчёт","k":"appendix","id":"А"}]`, "volume_refs_json", `[{"code":"2024-15-ОВ","page":3},{"code":"2024-15-АР","page":4}]`,
		"contents_json", `[{"Code":"2024-15-АР","Title":"Архитектурные решения","Page":2}]`))
	f.put(ModDocuments, idcheck.V("set", s, "file", "b.pdf", "volume_code", "2024-15-АР", "volume_title", "Отопление и вентиляция", "pages", "4"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "2024-15-КЖ2", "required", "1"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "КЖ", "kind", UnitSection, "required", "1"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "ЭС", "kind", UnitSection, "required", "1"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "Приложение А", "kind", UnitAppendix, "parent_code", "2024-15-КЖ2", "required", "1"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "Приложение Б", "kind", UnitAppendix, "required", "1"))
	f.put(ModVolumes, idcheck.V("object", obj, "volume_code", "ТУ-1", "title", "Технические условия на подключение", "kind", UnitReference, "required", "0"))
	od, err := LoadObjectDocs(f, "ns", obj)
	if err != nil {
		t.Fatal(err)
	}
	set := od.Sets[0]
	// 1 volume + 4 required units; present: the volume, КЖ, Приложение А
	if set.Total != 5 || set.Present != 3 || set.Pct != 60 {
		t.Fatalf("total %d present %d pct %d (%s)", set.Total, set.Present, set.Pct, set.Note)
	}
	if !strings.Contains(strings.Join(set.Missing, "|"), "раздел ЭС") || !strings.Contains(strings.Join(set.Missing, "|"), "приложение Приложение Б") {
		t.Errorf("missing %v", set.Missing)
	}
	if len(set.Units) != 5 || set.Units[4].Kind != UnitReference || set.Units[4].Required || set.Units[4].Present || set.Units[0].Where != "2024-15-КЖ2" {
		t.Errorf("units %+v", set.Units)
	}
	if len(od.Matrix) != 2 { // the matrix is of volumes only
		t.Errorf("matrix %+v", od.Matrix)
	}
	// 2024-15-ОВ is not loaded (for information), 2024-15-АР is
	if len(set.RefsMissing) != 1 || !strings.HasPrefix(set.RefsMissing[0], "2024-15-ОВ (ссылка в 2024-15-КЖ2, стр. 3)") {
		t.Errorf("refs missing %v", set.RefsMissing)
	}
	// the contents call 2024-15-АР «Архитектурные решения», its file «Отопление и вентиляция»
	if !strings.Contains(strings.Join(set.Renamed, "|"), "в составе проекта") || !strings.Contains(set.Note, "наименования не согласованы") {
		t.Errorf("names %v / %s", set.Renamed, set.Note)
	}
}
