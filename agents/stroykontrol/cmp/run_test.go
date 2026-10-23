package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// fakeBackend is an in-memory Compose: records per module, attachments are
// files of the fixtures directory (attachment ID = path relative to it).
type fakeBackend struct {
	recs  map[string]map[string]idcheck.Record
	files map[string][]byte // in-memory attachments (checked before the fixtures dir)
	seq   int
}

func newFake() *fakeBackend {
	return &fakeBackend{recs: map[string]map[string]idcheck.Record{}, files: map[string][]byte{}}
}

func (f *fakeBackend) put(module string, v idcheck.Values) string {
	f.seq++
	id := fmt.Sprintf("%s-%d", module, f.seq)
	if f.recs[module] == nil {
		f.recs[module] = map[string]idcheck.Record{}
	}
	cp := idcheck.Values{}
	for k, x := range v {
		cp[k] = append([]string(nil), x...)
	}
	f.recs[module][id] = idcheck.Record{ID: id, Values: cp}
	return id
}

func (f *fakeBackend) GetRecord(_, module, id string) (*idcheck.Record, error) {
	r, ok := f.recs[module][id]
	if !ok {
		return nil, fmt.Errorf("no %s/%s", module, id)
	}
	return &r, nil
}

var eqRe = regexp.MustCompile(`^(\w+) = '([^']*)'$`)

func (f *fakeBackend) Search(_, module, filter string) ([]idcheck.Record, error) {
	var out []idcheck.Record
	m := eqRe.FindStringSubmatch(filter)
	for _, r := range f.recs[module] {
		if m == nil || r.Get(m[1]) == m[2] {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeBackend) Create(_, module string, v idcheck.Values) (string, error) {
	return f.put(module, v), nil
}

func (f *fakeBackend) Update(_, module, id string, v idcheck.Values) error {
	r, ok := f.recs[module][id]
	if !ok {
		return fmt.Errorf("no %s/%s", module, id)
	}
	for k, x := range v {
		if len(x) == 0 || x[0] == "" {
			delete(r.Values, k)
			continue
		}
		r.Values[k] = x
	}
	return nil
}

func (f *fakeBackend) Delete(_, module, id string) error {
	delete(f.recs[module], id)
	return nil
}

func (f *fakeBackend) Download(_, att string) ([]byte, string, error) {
	if b, ok := f.files[att]; ok {
		return b, att, nil
	}
	b, err := os.ReadFile(filepath.Join("..", "fixtures", att))
	return b, filepath.Base(att), err
}

func (f *fakeBackend) discrepancies(comparison string) map[string]idcheck.Record {
	out := map[string]idcheck.Record{}
	for _, r := range f.recs[ModDiscrepancies] {
		if r.Get("comparison") == comparison {
			out[r.Get("description")] = r
		}
	}
	return out
}

func findDesc(m map[string]idcheck.Record, sub string) (idcheck.Record, bool) {
	for d, r := range m {
		if strings.Contains(d, sub) {
			return r, true
		}
	}
	return idcheck.Record{}, false
}

func skipWithoutPoppler(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("poppler not installed")
	}
	if _, err := os.Stat("../fixtures/pack-1-v2/rd-kzh.pdf"); err != nil {
		t.Skip("fixtures missing (run fixtures/gen_cmp_fixtures.mjs)")
	}
}

func addPackage(f *fakeBackend, comparison, dir string) {
	f.put(ModDocuments, idcheck.V("comparison", comparison, "side", SideRef, "file", dir+"/pd-package.pdf"))
	f.put(ModDocuments, idcheck.V("comparison", comparison, "side", SideTarget, "file", dir+"/rd-kzh.pdf"))
	f.put(ModDocuments, idcheck.V("comparison", comparison, "side", SideTarget, "file", dir+"/rd-ar.pdf"))
}

func TestRunWritesResultAndKeepsDecisions(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	c := f.put(ModComparisons, idcheck.V("title", "v1", "status", "new"))
	addPackage(f, c, "pack-1")
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	if err := r.Run(context.Background(), "ns", c); err != nil {
		t.Fatal(err)
	}
	rec, _ := f.GetRecord("ns", ModComparisons, c)
	if rec.Get("status") != "done" || rec.Get("similarity_percent") == "" || rec.Get("total_pages_pd") != "7" {
		t.Fatalf("comparison not updated: %+v", rec.Values)
	}
	d, ok := findDesc(f.discrepancies(c), "Глубина заложения")
	if !ok || d.Get("status") != "new" || d.Get("page_pd") != "1" || d.Get("bbox_pd") == "" || d.Get("doc_ref") == "" {
		t.Fatalf("depth finding not stored with positions: %+v", d.Values)
	}
	// a decision survives a re-run of the same comparison
	_ = f.Update("ns", ModDiscrepancies, d.ID, idcheck.V("status", StatusConfirmed))
	if err := r.Run(context.Background(), "ns", c); err != nil {
		t.Fatal(err)
	}
	d2, _ := findDesc(f.discrepancies(c), "Глубина заложения")
	if d2.Get("status") != StatusConfirmed {
		t.Errorf("decision lost on re-run: %v", d2.Values)
	}
	// a request for a clarification keeps its comment across a re-run (ТЗ 3.2.7)
	_ = f.Update("ns", ModDiscrepancies, d2.ID, idcheck.V("status", StatusClarify, "reviewer_comment", "уточнить у ГИПа, какая глубина верна"))
	if err := r.Run(context.Background(), "ns", c); err != nil {
		t.Fatal(err)
	}
	d3, _ := findDesc(f.discrepancies(c), "Глубина заложения")
	if d3.Get("status") != StatusClarify || d3.Get("reviewer_comment") != "уточнить у ГИПа, какая глубина верна" {
		t.Errorf("clarify decision or its comment lost on re-run: %v", d3.Values)
	}
	_ = f.Update("ns", ModDiscrepancies, d3.ID, idcheck.V("status", StatusConfirmed))
	// the version of every uploaded file is its content hash; the sheet numbers are kept for the sheet → page links
	for _, dr := range f.recs[ModDocuments] {
		if dr.Get("file") != "" && dr.Get("file_hash") == "" {
			t.Errorf("a document row must carry the hash of its file: %+v", dr.Values)
		}
	}
	// volume info lands on the document rows
	var coded bool
	for _, dr := range f.recs[ModDocuments] {
		if strings.Contains(dr.Get("volume_code"), "КЖ") {
			coded = true
		}
	}
	if !coded {
		t.Error("detected volume codes were not written to cmp_documents")
	}
}

func TestIterationCarriesDecisionsAndClosesFixed(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}

	v1 := f.put(ModComparisons, idcheck.V("title", "v1", "status", "new"))
	addPackage(f, v1, "pack-1")
	if err := r.Run(context.Background(), "ns", v1); err != nil {
		t.Fatal(err)
	}
	// the reviewer confirms the depth finding and the table quantity, rejects the concrete grade
	for sub, st := range map[string]string{"Глубина заложения": StatusConfirmed, "Ф-2": StatusConfirmed, "В25": StatusRejected} {
		d, ok := findDesc(f.discrepancies(v1), sub)
		if !ok {
			t.Fatalf("v1 has no finding about %q", sub)
		}
		_ = f.Update("ns", ModDiscrepancies, d.ID, idcheck.V("status", st))
	}

	v2 := f.put(ModComparisons, idcheck.V("title", "v2", "status", "new", "parent_comparison", v1))
	addPackage(f, v2, "pack-1-v2")
	if err := r.Run(context.Background(), "ns", v2); err != nil {
		t.Fatal(err)
	}
	rec, _ := f.GetRecord("ns", ModComparisons, v2)
	if rec.Get("iteration_no") != "2" || !strings.Contains(rec.Get("comment"), "Итерация 2: устранено") {
		t.Fatalf("iteration not recorded: %v", rec.Values)
	}
	got := f.discrepancies(v2)

	// fixed: the depth was corrected → closed, with the old decision kept for the record
	d, ok := findDesc(got, "Глубина заложения")
	if !ok || d.Get("iter_state") != IterResolved || d.Get("status") != StatusConfirmed {
		t.Errorf("fixed finding must be resolved: %+v", d.Values)
	}
	// not fixed: the quantity is still 52 → persisting, and the earlier decision is inherited
	d, ok = findDesc(got, "Ф-2")
	if !ok || d.Get("iter_state") != IterPersisting || d.Get("status") != StatusConfirmed {
		t.Errorf("unfixed finding must persist with its decision: %+v", d.Values)
	}
	// new slip: wall 300 → 350
	d, ok = findDesc(got, "350")
	if !ok || d.Get("iter_state") != IterNew || d.Get("status") != StatusNew {
		t.Errorf("new finding: %+v", d.Values)
	}
	// a false alarm that was rejected is not reported as "fixed"
	if d, ok := findDesc(got, "В25"); ok && d.Get("iter_state") == IterResolved {
		t.Errorf("a rejected finding must not become 'resolved': %+v", d.Values)
	}
	// resolved findings are not carried into a third iteration as open ones
	v3 := f.put(ModComparisons, idcheck.V("title", "v3", "status", "new", "parent_comparison", v2))
	addPackage(f, v3, "pack-1-v2")
	if err := r.Run(context.Background(), "ns", v3); err != nil {
		t.Fatal(err)
	}
	if d, ok := findDesc(f.discrepancies(v3), "Глубина заложения"); ok {
		t.Errorf("a finding closed in the previous iteration must not come back: %+v", d.Values)
	}
	rec, _ = f.GetRecord("ns", ModComparisons, v3)
	if rec.Get("iteration_no") != "3" {
		t.Errorf("iteration number: %v", rec.Values)
	}
}

func TestIterationInheritsReferenceDocuments(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	v1 := f.put(ModComparisons, idcheck.V("title", "v1", "status", "new"))
	addPackage(f, v1, "pack-1")
	v2 := f.put(ModComparisons, idcheck.V("title", "v2", "status", "new", "parent_comparison", v1))
	// only the corrected РД is uploaded to the new iteration
	f.put(ModDocuments, idcheck.V("comparison", v2, "side", SideTarget, "file", "pack-1-v2/rd-kzh.pdf"))
	f.put(ModDocuments, idcheck.V("comparison", v2, "side", SideTarget, "file", "pack-1-v2/rd-ar.pdf"))
	if err := r.Run(context.Background(), "ns", v2); err != nil {
		t.Fatal(err)
	}
	rec, _ := f.GetRecord("ns", ModComparisons, v2)
	if rec.Get("status") != "done" || rec.Get("total_pages_pd") != "7" {
		t.Fatalf("the ПД of the previous iteration was not used: %v", rec.Values)
	}
	if got := Ancestors(f, "ns", v2); len(got) != 2 || got[1] != v1 {
		t.Errorf("ancestors: %v", got)
	}
}

func objectWithComparison(t *testing.T, f *fakeBackend, r *Runner) (object, comparison string) {
	t.Helper()
	object = f.put(ModObjects, idcheck.V("name", "Объект"))
	comparison = f.put(ModComparisons, idcheck.V("title", "v1", "status", "new", "object", object))
	addPackage(f, comparison, "pack-1")
	if err := r.Run(context.Background(), "ns", comparison); err != nil {
		t.Fatal(err)
	}
	return
}

func remarksByKey(f *fakeBackend) map[string]idcheck.Record {
	out := map[string]idcheck.Record{}
	for _, x := range f.recs[ModExpertise] {
		out[x.Get("remark_key")] = x
	}
	return out
}

func TestImportObjectRemarksWorkflow(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	object, comparison := objectWithComparison(t, f, r)
	_ = comparison

	f.files["101"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м)", "Том 3, 2024-15-ГП – КЖ", "Откорректировано в соответствии со стадией П", "", "", "", "Требуется рассмотрение ответа", "1"),
		exportRow("30.2", "Не представлено обоснование по технологии", "Том 5, 2024-15-ГП – ЭС9", "", "", "", "", "Ответ не предоставлен", "1"),
		exportRow("30.3", "Класс бетона не указан", "", "Принято", "Нет изменений в документации", "Принято ранее", "Нет изменений", "Устраняется заявителем", "2"),
		exportRow("30.4", "Старое замечание", "", "Ответ", "", "", "", "Устранено", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101"}

	sum, err := r.ImportObjectRemarks(context.Background(), "ns", object, false)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Files != 1 || sum.Remarks != 4 || sum.New != 4 || sum.Skipped != 0 {
		t.Fatalf("summary: %+v", sum)
	}
	got := remarksByKey(f)
	r1, r2, r3 := got["30.1"], got["30.2"], got["30.3"]
	if r1.Get("status") != RemarkAnswerRev || r1.Get("source_status") != "Требуется рассмотрение ответа" || r1.Get("batch") != "30" || r1.Get("iterations") != "1" {
		t.Errorf("remark 1: %+v", r1.Values)
	}
	// history: remark 3 is on its second iteration → previous + current; the others just current
	hist := map[string][]idcheck.Record{}
	for _, h := range f.recs[ModRemarkHistory] {
		hist[h.Get("remark")] = append(hist[h.Get("remark")], h)
	}
	if len(hist[r3.ID]) != 2 || len(hist[r1.ID]) != 1 {
		t.Errorf("history sizes: r3=%d r1=%d", len(hist[r3.ID]), len(hist[r1.ID]))
	}
	// the check against documents: КЖ is a loaded volume (the ПД), the depth difference is still there
	if r1.Get("verdict") != VerdictStillDiffers || r1.Get("claim_mismatch") == "" || len(r1.Values["linked_documents"]) == 0 || len(r1.Values["matched_discrepancies"]) == 0 {
		t.Errorf("remark 1 must be linked to the ПД volume and its depth finding: %+v", r1.Values)
	}
	if r2.Get("verdict") != VerdictNoDocument {
		t.Errorf("a volume that is not loaded: %+v", r2.Values)
	}
	// where the remark leads: the volume it names and the region of the finding it was matched to (ТЗ 3.5)
	var links []RemarkLocation
	if err := json.Unmarshal([]byte(r1.Get("links_json")), &links); err != nil {
		t.Fatalf("links_json: %v (%q)", err, r1.Get("links_json"))
	}
	var place, found bool
	for _, l := range links {
		place = place || l.Kind == LocPlace && l.DocID != ""
		found = found || l.Kind == LocFinding && l.FindingID != "" && (l.RefBox != "" || l.TargetBox != "")
	}
	if !place || !found {
		t.Errorf("remark 1 must lead to its volume and to the region of its finding: %+v", links)
	}
	if r2.Get("links_json") != "" {
		t.Errorf("a volume that is not loaded leads nowhere: %q", r2.Get("links_json"))
	}
	// the journal
	var journal []idcheck.Record
	for _, j := range f.recs[ModExpertImports] {
		journal = append(journal, j)
	}
	if len(journal) != 1 || journal[0].Get("total") != "4" || journal[0].Get("new_count") != "4" || journal[0].Get("attachment") != "101" {
		t.Fatalf("journal: %+v", journal)
	}

	// the same export again: nothing to do
	sum, err = r.ImportObjectRemarks(context.Background(), "ns", object, false)
	if err != nil || sum.Files != 0 || sum.Skipped != 1 || len(f.recs[ModExpertImports]) != 1 || len(f.recs[ModRemarkHistory]) != 4+1 {
		t.Fatalf("an export is imported once: %+v %v (history %d)", sum, err, len(f.recs[ModRemarkHistory]))
	}

	// the specialist sets a status by hand, assigns the remark
	_ = f.Update("ns", ModExpertise, r1.ID, idcheck.V("status", RemarkBeingFix, "assignee", "Петров"))
	_ = f.Update("ns", ModExpertise, r3.ID, idcheck.V("status", RemarkResolved))

	// the next export: remark 1 unchanged, remark 2 answered, remark 3 changed on the expertise's side, 30.4 dropped, a new one added
	f.files["102"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м)", "Том 3, 2024-15-ГП – КЖ", "Откорректировано в соответствии со стадией П", "", "", "", "Требуется рассмотрение ответа", "1"),
		exportRow("30.2", "Не представлено обоснование по технологии", "Том 5, 2024-15-ГП – ЭС9", "Обоснование добавлено", "", "", "", "Требуется рассмотрение ответа", "1"),
		exportRow("30.3", "Класс бетона не указан", "", "Принято", "Снова нет изменений", "Принято ранее", "Нет изменений", "Ответ не предоставлен", "3"),
		exportRow("31.1", "Новое замечание", "", "", "", "", "", "Ответ не предоставлен", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101", "102"}
	sum, err = r.ImportObjectRemarks(context.Background(), "ns", object, false)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Files != 1 || sum.New != 1 || sum.Archived != 1 || sum.StatusChanges < 1 {
		t.Fatalf("second export: %+v", sum)
	}
	got = remarksByKey(f)
	if got["30.1"].Get("status") != RemarkBeingFix || got["30.1"].Get("assignee") != "Петров" {
		t.Errorf("the hand-made status and assignee survive an unchanged remark: %+v", got["30.1"].Values)
	}
	if got["30.3"].Get("status") != RemarkAwaiting || got["30.3"].Get("iterations") != "3" {
		t.Errorf("the expertise sent remark 3 back: %+v", got["30.3"].Values)
	}
	if got["30.2"].Get("answer") != "Обоснование добавлено" || got["30.2"].Get("status") != RemarkAnswerRev {
		t.Errorf("remark 2 was answered: %+v", got["30.2"].Values)
	}
	if got["30.4"].Get("archived") == "" || got["31.1"].Get("status") != RemarkAwaiting {
		t.Errorf("archived / new: %+v %+v", got["30.4"].Values, got["31.1"].Values)
	}
	n := 0
	for _, h := range f.recs[ModRemarkHistory] {
		if h.Get("remark") == got["30.3"].ID {
			n++
		}
	}
	if n != 3 {
		t.Errorf("remark 3 has previous + current + the new iteration, got %d entries", n)
	}
	// forced re-import of the last export changes nothing
	before := len(f.recs[ModRemarkHistory])
	if _, err := r.ImportObjectRemarks(context.Background(), "ns", object, true); err != nil {
		t.Fatal(err)
	}
	if len(f.recs[ModRemarkHistory]) != before {
		t.Errorf("a re-import of the same export must not add history: %d → %d", before, len(f.recs[ModRemarkHistory]))
	}
}

func TestCorrectionPackageOfAnIteration(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	object, v1 := objectWithComparison(t, f, r)
	if n := len(f.recs[ModCorrection]); n != 0 {
		t.Fatalf("the first comparison corrects nothing: %d packages", n)
	}
	f.files["101"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м)", "Том 3, 2024-15-ГП – КЖ", "Откорректировано в соответствии со стадией П", "", "", "", "Устраняется заявителем", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101"}
	if _, err := r.ImportObjectRemarks(context.Background(), "ns", object, false); err != nil {
		t.Fatal(err)
	}
	remark := remarksByKey(f)["30.1"]

	v2 := f.put(ModComparisons, idcheck.V("title", "v2", "status", "new", "object", object, "parent_comparison", v1))
	addPackage(f, v2, "pack-1-v2")
	if err := r.Run(context.Background(), "ns", v2); err != nil {
		t.Fatal(err)
	}
	pk := f.recs[ModCorrection]
	if len(pk) != 1 {
		t.Fatalf("one package for the iteration: %d", len(pk))
	}
	var p idcheck.Record
	for _, x := range pk {
		p = x
	}
	if p.Get("comparison") != v2 || p.Get("object") != object || p.Get("iteration_no") != "2" || p.Get("runs") != "1" || p.Get("parent_comparison") != v1 {
		t.Errorf("package: %+v", p.Values)
	}
	if p.Get("n_documents") == "0" || p.Get("n_documents") == "" || len(p.Values["documents"]) == 0 {
		t.Errorf("the corrected files are the package: %+v", p.Values)
	}
	if p.Get("resolved") == "" || p.Get("resolved") == "0" {
		t.Errorf("the depth was corrected in v2: %+v", p.Values)
	}
	linked := false
	for _, id := range p.Values["remarks"] {
		linked = linked || id == remark.ID
	}
	if !linked {
		t.Errorf("the remark about the depth belongs to the package: %+v (remark %s)", p.Values["remarks"], remark.ID)
	}
	// the corrected edition fixed the difference and the designer answered: the remark waits for the specialist (ТЗ 3.7.3)
	if got := f.recs[ModExpertise][remark.ID].Get("status"); got != RemarkAnswerRev {
		t.Errorf("the remark must be handed over to the control by the new edition: %s", got)
	}
	// the same iteration again: the package is refreshed, not doubled, and keeps its history
	if err := r.Run(context.Background(), "ns", v2); err != nil {
		t.Fatal(err)
	}
	if len(f.recs[ModCorrection]) != 1 {
		t.Fatalf("a package per iteration: %d", len(f.recs[ModCorrection]))
	}
	for _, x := range f.recs[ModCorrection] {
		var hist []PackageRun
		if err := json.Unmarshal([]byte(x.Get("history_json")), &hist); err != nil || len(hist) != 2 || x.Get("runs") != "2" {
			t.Errorf("history of the checks: %v %s", err, x.Get("history_json"))
		}
		if PackageSummary(x) == "" {
			t.Error("summary")
		}
	}
}

func TestEachRequirementOfARemarkIsCheckedOnItsOwn(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	object, _ := objectWithComparison(t, f, r)
	f.files["101"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "1) Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). 2) Представить обоснование по технологии производства работ.", "Том 3, 2024-15-ГП – КЖ", "Откорректировано", "", "", "", "Устраняется заявителем", "1"),
		exportRow("30.2", "Класс бетона не указан", "Том 3, 2024-15-ГП – КЖ", "", "", "", "", "Ответ не предоставлен", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101"}
	if _, err := r.ImportObjectRemarks(context.Background(), "ns", object, false); err != nil {
		t.Fatal(err)
	}
	rem := remarksByKey(f)["30.1"]
	set, err := ParseRequirementSet(rem.Get("requirements_json"))
	if err != nil || len(set.Items) != 2 || set.For != TextHash(rem.Get("text")) || set.Manual {
		t.Fatalf("two requirements: %v %+v", err, set)
	}
	if set.Items[0].Verdict != VerdictStillDiffers || len(set.Items[0].FindingIDs) == 0 || set.Items[0].Outcome != OutcomeNotFixed {
		t.Errorf("the first requirement is about the depth that is still different: %+v", set.Items[0])
	}
	if set.Items[1].Verdict == VerdictStillDiffers || set.Items[1].Outcome == OutcomeNotFixed {
		t.Errorf("the second one (technology) has nothing to do with the depth finding: %+v", set.Items[1])
	}
	one, _ := ParseRequirementSet(remarksByKey(f)["30.2"].Get("requirements_json"))
	if len(one.Items) != 1 {
		t.Errorf("a remark with nothing to split is one requirement: %+v", one)
	}
	// a hand-made list is kept when the check runs again
	set.Manual = true
	set.Items[1].Text = "Представить обоснование технологии работ и ППР"
	if err := f.Update("ns", ModExpertise, rem.ID, idcheck.V("requirements_json", set.JSON())); err != nil {
		t.Fatal(err)
	}
	if _, err := r.CheckObjectRemarks(context.Background(), "ns", object); err != nil {
		t.Fatal(err)
	}
	kept, _ := ParseRequirementSet(f.recs[ModExpertise][rem.ID].Get("requirements_json"))
	if len(kept.Items) != 2 || kept.Items[1].Text != "Представить обоснование технологии работ и ППР" || !kept.Manual || kept.Stale {
		t.Errorf("the specialist's list stays while the remark's text is the same: %+v", kept)
	}
}

func TestTheModelReadsEveryRequirementOfARemark(t *testing.T) {
	skipWithoutPoppler(t)
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Messages []Msg `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		calls = append(calls, body.Messages[len(body.Messages)-1].Content)
		out, _ := json.Marshal(map[string]any{"message": map[string]string{"content": `{"verdict":"unclear","explanation":"по фрагментам судить нельзя","cited":[]}`}})
		_, _ = w.Write(out)
	}))
	defer srv.Close()
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}, LLM: NewSuggester(srv.URL, "m")}
	object, _ := objectWithComparison(t, f, r)
	f.files["101"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "1) Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м). 2) Представить обоснование по технологии производства работ.", "Том 3, 2024-15-ГП – КЖ", "Откорректировано", "", "", "", "Устраняется заявителем", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101"}
	if _, err := r.ImportObjectRemarks(context.Background(), "ns", object, false); err != nil {
		t.Fatal(err)
	}
	rem := remarksByKey(f)["30.1"]
	if _, err := r.AnalyzeRemarkAI(context.Background(), "ns", rem.ID); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 3 {
		t.Fatalf("one reading of the remark and one of each requirement: %d calls", len(calls))
	}
	if !strings.Contains(calls[1], "Представить") && !strings.Contains(calls[2], "Представить") || strings.Contains(calls[1], "технологии") && strings.Contains(calls[2], "технологии") {
		t.Errorf("each reading is about its own requirement:\n%s\n%s", calls[1], calls[2])
	}
	set, _ := ParseRequirementSet(f.recs[ModExpertise][rem.ID].Get("requirements_json"))
	for _, q := range set.Items {
		if q.AIVerdict != AIUnclear || q.AINote == "" {
			t.Errorf("the reading is kept on the requirement: %+v", q)
		}
	}
	if got := f.recs[ModExpertise][rem.ID].Get("status"); got != RemarkBeingFix {
		t.Errorf("the model never changes the working status: %s", got)
	}
}

func TestMaterialsDeclaredInAnAnswerAreLookedFor(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	object, _ := objectWithComparison(t, f, r)
	f.files["101"] = xlsxBytes(t, "Выгрузка данных", [][]string{
		exportHeader,
		exportRow("30.1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м)", "Том 3, 2024-15-ГП – КЖ",
			"Исправлено на листе 1 тома КЖ1. Расчёт осадок приложен. Письмо № 7 прилагается.", "", "", "", "Требуется рассмотрение ответа", "1"),
		exportRow("30.2", "Класс бетона не указан", "Том 3, 2024-15-ГП – КЖ", "Принято", "", "", "", "Ответ не предоставлен", "1"),
	})
	f.recs[ModObjects][object].Values["remarks_file"] = []string{"101"}
	if _, err := r.ImportObjectRemarks(context.Background(), "ns", object, false); err != nil {
		t.Fatal(err)
	}
	rem := remarksByKey(f)["30.1"]
	ms := ParseMaterialsJSON(rem.Get("materials_json"))
	if len(ms) != 3 || rem.Get("materials_state") != MaterialsMissing {
		t.Fatalf("volume, calculation, letter; nothing attached yet: %s %s", rem.Get("materials_state"), rem.Get("materials_json"))
	}
	if ms[0].Kind != MatVolume || ms[0].State != MatFound || ms[0].DocID == "" {
		t.Errorf("the volume КЖ1 is loaded: %+v", ms[0])
	}
	if ms[2].Doc != "письмо" || ms[2].State != MatNotFound {
		t.Errorf("the letter is not there yet: %+v", ms[2])
	}
	if got := remarksByKey(f)["30.2"]; got.Get("materials_json") != "" || got.Get("materials_state") != "" {
		t.Errorf("an answer that names nothing declares nothing: %+v", got.Values)
	}
	// the letter comes as a supporting file of the remark
	f.files["mat-test-letter"] = []byte("not a pdf")
	f.recs[ModExpertise][rem.ID].Values["materials_files"] = []string{"mat-test-letter"}
	if _, err := r.CheckObjectRemarks(context.Background(), "ns", object); err != nil {
		t.Fatal(err)
	}
	ms = ParseMaterialsJSON(f.recs[ModExpertise][rem.ID].Get("materials_json"))
	if len(ms) != 3 || ms[2].State != MatNotFound {
		// the file is named by its attachment ID in the fake: it says nothing about being a letter
		t.Errorf("a file whose name says nothing is not the letter: %+v", ms)
	}
	f.files["Письмо ГИП №7.pdf"] = []byte("not a pdf")
	f.recs[ModExpertise][rem.ID].Values["materials_files"] = []string{"Письмо ГИП №7.pdf"}
	if _, err := r.CheckObjectRemarks(context.Background(), "ns", object); err != nil {
		t.Fatal(err)
	}
	ms = ParseMaterialsJSON(f.recs[ModExpertise][rem.ID].Get("materials_json"))
	if len(ms) != 3 || ms[2].State != MatFound || ms[2].In != "Письмо ГИП №7.pdf" {
		t.Errorf("the attached letter is found: %+v", ms)
	}
	if f.recs[ModExpertise][rem.ID].Get("materials_state") != MaterialsMissing {
		t.Errorf("the calculation is still not there: %s", f.recs[ModExpertise][rem.ID].Get("materials_state"))
	}
}
