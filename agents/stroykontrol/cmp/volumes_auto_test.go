package cmp

import (
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func contentsPages() []Page {
	d := textDoc("d", SideRef,
		[]string{"Состав проектной документации", "Обозначение Наименование Примечание",
			"100720-1/П-1-ПЗ Пояснительная записка стр. 3",
			"100720-1/П-2-ПЗУ Схема планировочной организации",
			"земельного участка",
			"100720-1/П-3-ТКР5-1.2 Книга 1.2. Мостовые сооружения",
			"100720-1/П-3-ТКР5-1.2-2.1 Часть тома",
			"ПК33+94-ПК47+20 трасса", "Солнцево-Бутово-Видное-39.2-24 название объекта"},
		[]string{"100720-1/П-4-ОВ Отопление и вентиляция", "03.2024 дата"},
	)
	return d.Pages
}

func TestParseContentsFindsVolumeCodesAndTitles(t *testing.T) {
	es := ParseContents(contentsPages(), "100720-1/П-3-ТКР5-1.2")
	got := map[string]string{}
	for _, e := range es {
		got[NormCode(e.Code)] = e.Title
	}
	if got[NormCode("100720-1/П-1-ПЗ")] != "Пояснительная записка" {
		t.Fatalf("ПЗ: %+v", es)
	}
	if got[NormCode("100720-1/П-2-ПЗУ")] != "Схема планировочной организации" && !strings.Contains(got[NormCode("100720-1/П-2-ПЗУ")], "Схема") {
		t.Fatalf("a title that goes on the next line is not required, but the code must be found: %+v", es)
	}
	if _, ok := got[NormCode("100720-1/П-4-ОВ")]; !ok {
		t.Fatalf("the second page of the table is read too: %+v", es)
	}
	// the volume itself and its parts are not other volumes; stationing is not a code
	for code := range got {
		if strings.HasPrefix(code, NormCode("100720-1/П-3-ТКР5-1.2")) || strings.Contains(code, "ПК33") || strings.Contains(code, "СОЛНЦЕВО") {
			t.Fatalf("own volume / part / stationing proposed: %s", code)
		}
	}
	if ParseContents(textDoc("x", SideRef, []string{"Просто текст", "100720-1/П-1-ПЗ Пояснительная записка"}).Pages, "") != nil {
		t.Fatal("no contents heading → nothing")
	}
}

func volumeSetup() (*fakeBackend, string, string) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "О"))
	s1 := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v1", "issued_at", "2024-01-01"))
	s2 := f.put(ModSets, idcheck.V("object", obj, "stage", SideRef, "version", "v2", "issued_at", "2024-03-01"))
	for _, c := range []string{"КЖ1", "АР", "ПЗ"} {
		f.put(ModDocuments, idcheck.V("set", s1, "file", "v1-"+c+".pdf", "volume_code", c, "volume_title", "Том "+c, "pages", "3"))
	}
	f.put(ModDocuments, idcheck.V("set", s2, "file", "v2-КЖ1.pdf", "volume_code", "КЖ1", "pages", "3"))
	return f, obj, s2
}

func TestProposeFromTheEarliestSetAndAcceptReject(t *testing.T) {
	f, obj, _ := volumeSetup()
	od, _ := LoadObjectDocs(f, "ns", obj)
	props := proposeFromSets(od, map[string][]string{})
	if len(props) != 3 || props[0].Source != SrcSet || !strings.Contains(props[0].Evidence, "первом комплекте ПД v1") {
		t.Fatalf("%+v", props)
	}
	// a stage that already has a statement is left alone
	if got := proposeFromSets(od, map[string][]string{SideRef: {"КЖ1"}}); len(got) != 0 {
		t.Fatalf("%+v", got)
	}

	r := &Runner{Backend: f}
	// write the proposals the way SuggestVolumes does and resolve them
	for _, p := range props {
		f.put(ModVolumes, idcheck.V("object", obj, "stage", p.Stage, "volume_code", p.Code, "title", p.Title, "source", p.Source, "proposed", "1"))
	}
	od, _ = LoadObjectDocs(f, "ns", obj)
	if len(od.Proposals) != 3 {
		t.Fatalf("proposals: %+v", od.Proposals)
	}
	for _, s := range od.Sets {
		if s.Total != 0 {
			t.Fatalf("a proposal must not count as the statement yet: %+v", s)
		}
	}
	var accept, reject string
	for _, p := range od.Proposals {
		// the proposals come in no fixed order: accept КЖ1 (the one volume set v2 has), reject АР
		switch p.Code {
		case "КЖ1":
			accept = p.ID
		case "АР":
			reject = p.ID
		}
	}
	acc, rej, err := r.ResolveVolumes("ns", obj, []string{accept}, []string{reject}, false)
	if err != nil || acc != 1 || rej != 1 {
		t.Fatalf("%d %d %v", acc, rej, err)
	}
	od, _ = LoadObjectDocs(f, "ns", obj)
	if len(od.Proposals) != 1 {
		t.Fatalf("one accepted, one rejected, one still proposed: %+v", od.Proposals)
	}
	byVersion := map[string]SetInfo{}
	for _, s := range od.Sets {
		byVersion[s.Version] = s
	}
	if byVersion["v2"].Total != 1 || byVersion["v2"].Present != 1 {
		t.Fatalf("the accepted row is the statement now: %+v", byVersion["v2"])
	}
	if acc, _, _ := r.ResolveVolumes("ns", obj, nil, nil, true); acc != 1 {
		t.Fatalf("accept all takes the rest: %d", acc)
	}
	// a confirmed row is never deleted by a reject
	confirmed := f.put(ModVolumes, idcheck.V("object", obj, "stage", SideRef, "volume_code", "ОВ", "proposed", "0"))
	if _, rj, _ := r.ResolveVolumes("ns", obj, nil, []string{confirmed}, false); rj != 0 {
		t.Fatal("a confirmed row must stay")
	}
}
