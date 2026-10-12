package cmp

import (
	"context"
	"fmt"
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

func TestImportRemarksMatchesAndKeepsManualStatus(t *testing.T) {
	skipWithoutPoppler(t)
	f := newFake()
	r := &Runner{Backend: f, CacheDir: t.TempDir(), current: map[string]string{}}
	c := f.put(ModComparisons, idcheck.V("title", "v1", "status", "new"))
	addPackage(f, c, "pack-1")
	if err := r.Run(context.Background(), "ns", c); err != nil {
		t.Fatal(err)
	}
	f.files["zam.xlsx"] = xlsxBytes(t, "Замечания", [][]string{
		{"№", "Замечание", "Ответ проектировщика"},
		{"1", "Глубина заложения фундамента 2,6 м не соответствует стадии П (2,4 м)", "Откорректировано в соответствии со стадией П."},
		{"2", "Количество свай марки С15-35Т6 опоры №17п больше, чем в проекте", "Учтено."},
	})
	if _, err := r.ImportRemarks(context.Background(), "ns", c); err == nil {
		t.Fatal("an import without an attached file must say so")
	}
	rec := f.recs[ModComparisons][c]
	rec.Values["remarks_file"] = []string{"zam.xlsx"}

	sum, err := r.ImportRemarks(context.Background(), "ns", c)
	if err != nil {
		t.Fatal(err)
	}
	if sum.Remarks != 2 || sum.StillDiffers != 1 || sum.NoMatch != 1 || sum.ClaimMismatch != 1 {
		t.Fatalf("summary: %+v", sum)
	}
	byNum := map[string]idcheck.Record{}
	for _, x := range f.recs[ModExpertise] {
		byNum[x.Get("number")] = x
	}
	d1 := byNum["1"]
	if d1.Get("verdict") != VerdictStillDiffers || d1.Get("status") != "open" || d1.Get("claim_mismatch") == "" || len(d1.Values["matched_discrepancies"]) == 0 {
		t.Fatalf("remark 1: %+v", d1.Values)
	}
	linked := f.recs[ModDiscrepancies][d1.Values["matched_discrepancies"][0]]
	if !strings.Contains(linked.Get("description"), "Глубина заложения") {
		t.Errorf("remark 1 is linked to the wrong finding: %v", linked.Values)
	}
	if byNum["2"].Get("verdict") != VerdictNoMatch {
		t.Errorf("remark 2: %+v", byNum["2"].Values)
	}

	// the specialist closes remark 2; a re-import (file now has only remark 1 + a new one) keeps that and prunes the dropped
	_ = f.Update("ns", ModExpertise, byNum["2"].ID, idcheck.V("status", "closed"))
	f.files["zam.xlsx"] = xlsxBytes(t, "Замечания", [][]string{
		{"№", "Замечание", "Ответ проектировщика"},
		{"2", "Количество свай марки С15-35Т6 опоры №17п больше, чем в проекте", "Учтено."},
		{"3", "Толщина стен подвала не соответствует стадии П", ""},
	})
	if _, err := r.ImportRemarks(context.Background(), "ns", c); err != nil {
		t.Fatal(err)
	}
	byNum = map[string]idcheck.Record{}
	for _, x := range f.recs[ModExpertise] {
		byNum[x.Get("number")] = x
	}
	if len(byNum) != 2 || byNum["2"].Get("status") != "closed" || byNum["1"].ID != "" {
		t.Errorf("re-import: %+v", byNum)
	}
	if byNum["3"].Get("status") != "open" {
		t.Errorf("new remark: %+v", byNum["3"].Values)
	}
}
