package cmp

import (
	"archive/zip"
	"io"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func finding(id, sev, status, desc, docRef, pr, docTgt, pt string) FindingIn {
	return FindingIn{ID: id, Kind: KindText, Severity: sev, Status: status, Description: "[детерм.] " + desc, DocRef: docRef, DocTarget: docTgt, PageRef: pr, PageTarget: pt,
		Evidence: `{"volumeRef":"Конструктивные решения [2024-15-КЖ]","volumeTarget":"Конструктивные решения [2024-15-КЖ1]"}`}
}

func issueInput() IssueInput {
	return IssueInput{
		ComparisonID: "c1", ComparisonTitle: "Сравнение 1",
		DocName: map[string]string{"d-pd": "pd-package.pdf", "d-rd": "rd-kzh.pdf"},
		Findings: []FindingIn{
			finding("f1", "high", "confirmed", "Расходятся значения: «глубина 2,4» → «глубина 2,6»", "d-pd", "1", "d-rd", "1"),
			finding("f2", "medium", "confirmed", "Изменена формулировка: «а» → «б»", "d-pd", "1", "d-rd", "1"),
			finding("f3", "medium", "confirmed", "Расходятся значения: «Ф-2 48» → «Ф-2 52»", "d-pd", "2", "d-rd", "2"),
			finding("f4", "high", "new", "Не подтверждено специалистом", "d-pd", "3", "d-rd", "3"),
			finding("f5", "high", "rejected", "Ложное срабатывание", "d-pd", "4", "d-rd", "4"),
			{ID: "f6", Kind: KindNotRelease, Severity: "low", Status: "confirmed", Description: "Том не выпущен", DocRef: "d-pd", PageRef: "6"},
			{ID: "f7", Kind: KindText, Severity: "high", Status: "confirmed", Description: "Старое расхождение без позиций", PageNumber: "5", DocRef: "d-pd", DocTarget: "d-rd"},
		},
	}
}

func TestBuildIssueGroupsConfirmedFindingsBySheet(t *testing.T) {
	res := BuildIssue(issueInput())
	if len(res.Active) != 3 || len(res.Fixed) != 0 {
		t.Fatalf("3 sheet pairs expected: %+v", res)
	}
	if res.Excluded != 3 { // not confirmed, rejected, a volume that is not released
		t.Errorf("excluded: %d", res.Excluded)
	}
	first := res.Active[0]
	if first.Count != 2 || first.Criticality != "critical" || strings.Join(first.FindingIDs, ",") != "f1,f2" {
		t.Errorf("sheet 1 holds two findings and is critical because of the worse one: %+v", first)
	}
	for _, want := range []string{"лист 1 РД / лист 1 ПД", "1) Расходятся значения", "2) Изменена формулировка", "(эталон — стадия П)"} {
		if !strings.Contains(first.Text, want) {
			t.Errorf("text lacks %q:\n%s", want, first.Text)
		}
	}
	if strings.Contains(first.Text, "[детерм.]") {
		t.Error("the origin tag is not part of a remark")
	}
	if first.Refs != "РД: rd-kzh.pdf, стр. 1; ПД: pd-package.pdf, стр. 1" || first.Pages != "РД стр. 1 · ПД стр. 1" || first.Section != "Конструктивные решения [2024-15-КЖ1]" {
		t.Errorf("refs/pages/section: %q | %q | %q", first.Refs, first.Pages, first.Section)
	}
	second := res.Active[1]
	if second.Count != 1 || second.Criticality != "noncritical" || strings.Contains(second.Text, "1)") {
		t.Errorf("a single finding is worded without a list: %+v", second)
	}
	// an old finding that only has a page number is still a remark on that sheet
	if last := res.Active[2]; !strings.Contains(last.Refs, "стр. 5") {
		t.Errorf("legacy page number: %+v", last)
	}
	if res.Active[0].GroupKey == res.Active[1].GroupKey {
		t.Error("group keys must differ")
	}
	// the key of a sheet pair is stable
	if again := BuildIssue(issueInput()); again.Active[0].GroupKey != first.GroupKey {
		t.Error("group key changed between runs")
	}
	Number(res.Active, "ПТО-", 2)
	if res.Active[0].Number != "ПТО-2.1" || res.Active[2].Number != "ПТО-2.3" {
		t.Errorf("numbers: %s %s", res.Active[0].Number, res.Active[2].Number)
	}
}

func TestBuildIssueSeparatesFixedSheets(t *testing.T) {
	in := issueInput()
	in.Findings[0].IterState = IterResolved
	in.Findings[1].IterState = IterResolved // sheet 1 is fully fixed
	in.Findings[2].IterState = IterPersisting
	res := BuildIssue(in)
	if len(res.Fixed) != 1 || res.Fixed[0].Refs == "" || len(res.Active) != 2 {
		t.Fatalf("fixed=%d active=%d", len(res.Fixed), len(res.Active))
	}
	// a sheet with one fixed and one still open finding stays active
	in.Findings[1].IterState = IterPersisting
	if res := BuildIssue(in); len(res.Fixed) != 0 || len(res.Active) != 3 {
		t.Errorf("fixed=%d active=%d", len(res.Fixed), len(res.Active))
	}
}

func TestIssuedExcelIsAnExpertiseExport(t *testing.T) {
	res := BuildIssue(issueInput())
	Number(res.Active, "ПТО-", 1)
	data, err := IssuedXLSX(IssueMeta{ObjectName: "Объект", ComparisonTitle: "Сравнение 1"}, res.Active)
	if err != nil {
		t.Fatal(err)
	}
	// it reads back with the same parser an expertise export goes through
	rows, err := ParseExport(data, "out.xlsx")
	if err != nil {
		t.Fatalf("the list must be in the export's format: %v", err)
	}
	if len(rows) != 3 || rows[0].Number != "ПТО-1.1" || rows[0].Criticality != "critical" || rows[0].Status != RemarkAwaiting ||
		!strings.Contains(rows[0].Text, "лист 1 РД") || rows[0].Iterations != 1 || rows[1].Criticality != "noncritical" || rows[0].Refs == "" {
		t.Errorf("round trip: %+v", rows)
	}
}

func TestIssuedWordLetter(t *testing.T) {
	res := BuildIssue(issueInput())
	Number(res.Active, "ПТО-", 1)
	data, err := IssuedDOCX(IssueMeta{ObjectName: "Мост №6 <&>", ComparisonTitle: "Сравнение 1"}, res.Active)
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(strings.NewReader(string(data)), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var doc string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			b, _ := io.ReadAll(rc)
			rc.Close()
			doc = string(b)
		}
	}
	for _, want := range []string{"Замечания по результатам сравнения", "Мост №6 &lt;&amp;&gt;", "ПТО-1.1", "<w:tbl>", "Критичное", "Некритичное", "<w:br/>"} {
		if !strings.Contains(doc, want) {
			t.Errorf("document.xml lacks %q", want)
		}
	}
	if strings.Count(doc, "<w:tr>") != 4 { // the heading row is <w:tr><w:trPr>
		t.Errorf("rows: %d", strings.Count(doc, "<w:tr>"))
	}
	// a malformed package would not open in Word: the three required parts are there
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["[Content_Types].xml"] || !names["_rels/.rels"] || !names["word/document.xml"] {
		t.Errorf("parts: %v", names)
	}
}

func TestIssueToRegistry(t *testing.T) {
	f := newFake()
	object := f.put(ModObjects, idcheck.V("name", "Объект"))
	in := issueInput()
	res := BuildIssue(in)

	reg, err := IssueToRegistry(f, "ns", object, in, res)
	if err != nil {
		t.Fatal(err)
	}
	if reg.Created != 3 || reg.Updated != 0 || strings.Join(reg.Numbers, ",") != "ПТО-1.1,ПТО-1.2,ПТО-1.3" {
		t.Fatalf("first registration: %+v", reg)
	}
	var first idcheck.Record
	for _, r := range f.recs[ModExpertise] {
		if r.Get("number") == "ПТО-1.1" {
			first = r
		}
	}
	if first.Get("status") != RemarkAwaiting || first.Get("origin") != OriginComparison || first.Get("source_comparison") != "c1" || first.Get("criticality") != "critical" ||
		len(first.Values["matched_discrepancies"]) != 2 || len(first.Values["linked_documents"]) != 2 || first.Get("iterations") != "1" {
		t.Errorf("registered remark: %+v", first.Values)
	}
	if len(f.recs[ModRemarkHistory]) != 3 {
		t.Errorf("one history entry per new remark: %d", len(f.recs[ModRemarkHistory]))
	}

	// the specialist works the remark; regenerating keeps that and does not duplicate
	_ = f.Update("ns", ModExpertise, first.ID, idcheck.V("status", RemarkBeingFix, "assignee", "Петров"))
	reg, err = IssueToRegistry(f, "ns", object, in, BuildIssue(in))
	if err != nil || reg.Created != 0 || reg.Updated != 3 || len(f.recs[ModExpertise]) != 3 {
		t.Fatalf("regeneration: %+v %v (%d remarks)", reg, err, len(f.recs[ModExpertise]))
	}
	got, _ := f.GetRecord("ns", ModExpertise, first.ID)
	if got.Get("status") != RemarkBeingFix || got.Get("assignee") != "Петров" {
		t.Errorf("the specialist's status and assignee survive: %+v", got.Values)
	}

	// a new confirmed finding on another sheet is a new remark of the same batch
	in2 := issueInput()
	in2.Findings = append(in2.Findings, finding("f8", "medium", "confirmed", "Ещё расхождение", "d-pd", "9", "d-rd", "9"))
	reg, _ = IssueToRegistry(f, "ns", object, in2, BuildIssue(in2))
	if reg.Created != 1 || len(reg.Numbers) != 1 || !strings.HasPrefix(reg.Numbers[0], "ПТО-1.") || reg.Numbers[0] == "ПТО-1.1" {
		t.Errorf("new sheet: %+v", reg)
	}

	// a later iteration fixed sheet 1: its remark is closed; when the difference returns it is reopened
	in3 := issueInput()
	in3.Findings[0].IterState, in3.Findings[1].IterState = IterResolved, IterResolved
	reg, _ = IssueToRegistry(f, "ns", object, in3, BuildIssue(in3))
	got, _ = f.GetRecord("ns", ModExpertise, first.ID)
	if reg.Closed != 1 || got.Get("status") != RemarkResolved {
		t.Fatalf("closed: %+v %v", reg, got.Values)
	}
	reg, _ = IssueToRegistry(f, "ns", object, issueInput(), BuildIssue(issueInput()))
	got, _ = f.GetRecord("ns", ModExpertise, first.ID)
	if reg.Reopened != 1 || got.Get("status") != RemarkAwaiting {
		t.Errorf("reopened: %+v %v", reg, got.Values)
	}
	// without an object there is nowhere to put the remarks
	if _, err := IssueToRegistry(f, "ns", "", in, res); err == nil {
		t.Error("a comparison without an object")
	}
	// a second comparison of the same object starts the next batch
	in4 := issueInput()
	in4.ComparisonID = "c2"
	if reg, _ := IssueToRegistry(f, "ns", object, in4, BuildIssue(in4)); len(reg.Numbers) == 0 || !strings.HasPrefix(reg.Numbers[0], "ПТО-2.") {
		t.Errorf("batch of the second comparison: %+v", reg)
	}
}
