package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func loadDoc(t *testing.T, id, side, rel string) *Document {
	t.Helper()
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skip("pdftotext not installed")
	}
	data, err := os.ReadFile(filepath.Join("..", "fixtures", rel))
	if err != nil {
		t.Skipf("fixture missing (run fixtures/gen_cmp_fixtures.mjs): %v", err)
	}
	pages, err := ExtractPages(id, data, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return &Document{ID: id, Name: filepath.Base(rel), Side: side, Pages: pages}
}

func find(res *Result, pred func(Finding) bool) []Finding {
	var out []Finding
	for _, f := range res.Findings {
		if pred(f) {
			out = append(out, f)
		}
	}
	return out
}

func contains(f Finding, s string) bool { return strings.Contains(f.Description, s) }

func TestPack1_PDPackageAgainstTwoRD(t *testing.T) {
	pd := loadDoc(t, "pd", SideRef, "pack-1/pd-package.pdf")
	rdk := loadDoc(t, "rdk", SideTarget, "pack-1/rd-kzh.pdf")
	rda := loadDoc(t, "rda", SideTarget, "pack-1/rd-ar.pdf")
	res := Compare([]*Document{pd}, []*Document{rdk, rda}, Options{})

	if got := len(res.Volumes["pd"]); got != 3 {
		t.Fatalf("one ПД file must split into 3 volumes by шифр, got %d: %+v", got, res.Volumes["pd"])
	}
	if res.Volumes["pd"][0].Code != "2024-15-КЖ" || res.Volumes["pd"][1].Code != "2024-15-АР" || res.Volumes["pd"][2].Code != "2024-15-ТХ" {
		t.Errorf("volume codes: %+v", res.Volumes["pd"])
	}
	if len(res.Pairs) != 2 {
		t.Fatalf("want 2 pairs (КЖ, АР), got %d", len(res.Pairs))
	}

	// depth 2,4 → 2,6 : high, with positions on both sides
	d := find(res, func(f Finding) bool { return contains(f, "Глубина заложения") })
	if len(d) != 1 || d[0].Severity != "high" || d[0].PageRef != 1 || d[0].PageTarget != 1 || d[0].BoxRef == nil || d[0].BoxTarget == nil {
		t.Fatalf("depth finding: %+v", d)
	}
	if d[0].RefDoc != "pd" || d[0].TargetDoc != "rdk" {
		t.Errorf("docs: %s %s", d[0].RefDoc, d[0].TargetDoc)
	}
	// table cell 48 → 52 is a table_row on page 2
	if r := find(res, func(f Finding) bool { return f.Kind == KindTableRow && contains(f, "48") }); len(r) != 1 || r[0].PageRef != 2 {
		t.Errorf("table row finding: %+v", r)
	}
	// concrete grade: reported once, as the edited line (not duplicated by the param rule)
	if c := find(res, func(f Finding) bool { return contains(f, "В25") || contains(f, "B25") }); len(c) != 1 {
		t.Errorf("concrete grade should be reported once, got %d: %+v", len(c), c)
	}
	// numbering only → low/numbering
	if n := find(res, func(f Finding) bool { return f.Type == "numbering" }); len(n) != 1 || n[0].Severity != "low" {
		t.Errorf("numbering: %+v", n)
	}
	// unreleased ТХ volume → informational, not an error
	nr := find(res, func(f Finding) bool { return f.Kind == KindNotRelease })
	if len(nr) != 1 || nr[0].Severity != "low" || !contains(nr[0], "Технологические") {
		t.Fatalf("not_released: %+v", nr)
	}
	// extra detail in РД is low only
	if e := find(res, func(f Finding) bool { return contains(f, "Закладные детали") }); len(e) != 1 || e[0].Severity != "low" {
		t.Errorf("detail line: %+v", e)
	}
	// a ПД line absent from РД is a (medium) finding
	if m := find(res, func(f Finding) bool { return contains(f, "Защитный слой") }); len(m) != 1 || m[0].Severity != "medium" {
		t.Errorf("missing line: %+v", m)
	}
	// the title block (шифр, лист, стадия) must never be compared
	if s := find(res, func(f Finding) bool { return contains(f, "Стадия") || contains(f, "Листов") }); len(s) != 0 {
		t.Errorf("title block leaked into findings: %+v", s)
	}
	if res.Stats.Similarity <= 50 || res.Stats.Similarity >= 100 {
		t.Errorf("similarity: %v", res.Stats.Similarity)
	}
	for _, f := range res.Findings {
		if f.Fingerprint == "" {
			t.Errorf("no fingerprint: %+v", f)
		}
	}
}

// One ПД volume with two structures vs a РД for the first structure only,
// with extra detail: must not raise any error.
func TestPack2_FragmentIsNotAnError(t *testing.T) {
	pd := loadDoc(t, "pd", SideRef, "pack-2/pd-is.pdf")
	rd := loadDoc(t, "rd", SideTarget, "pack-2/rd-is1.pdf")
	res := Compare([]*Document{pd}, []*Document{rd}, Options{})
	if len(res.Pairs) != 1 || !res.Pairs[0].Fragment {
		t.Fatalf("expected one fragment pair, got %+v", res.Pairs)
	}
	for _, f := range res.Findings {
		if f.Severity != "low" {
			t.Errorf("fragment comparison must not raise %s findings: %s", f.Severity, f.Description)
		}
	}
	if len(find(res, func(f Finding) bool { return contains(f, "Сооружение 2") || contains(f, "мост") })) != 0 {
		t.Error("structure 2 (outside the РД scope) must not be reported")
	}
}

func TestPackAndSingleFileSymmetry(t *testing.T) {
	// a pack of one-volume files on the ref side vs the same set in one file
	pd := loadDoc(t, "pd", SideRef, "pack-1/pd-package.pdf")
	rdk := loadDoc(t, "rdk", SideTarget, "pack-1/rd-kzh.pdf")
	res := Compare([]*Document{pd}, []*Document{rdk}, Options{})
	if len(res.Pairs) != 1 {
		t.Fatalf("pairs: %d", len(res.Pairs))
	}
	// АР and ТХ have no РД yet: two informational findings, no errors
	if nr := find(res, func(f Finding) bool { return f.Kind == KindNotRelease }); len(nr) != 2 {
		t.Errorf("want 2 not_released, got %d", len(nr))
	}
}

func TestManualVolumeOverride(t *testing.T) {
	pd := loadDoc(t, "pd", SideRef, "pack-1/pd-package.pdf")
	pd.ManualCode, pd.ManualTitle = "МАНУАЛ-1", "Весь пакет"
	vs := DetectVolumes(pd)
	if len(vs) != 1 || vs[0].RawCode != "МАНУАЛ-1" || vs[0].LastPage != 7 {
		t.Fatalf("manual override: %+v", vs)
	}
}

func TestNormCodeFoldsLookalikes(t *testing.T) {
	if NormCode("2024-15-KЖ1") != NormCode("2024/15/КЖ1") {
		t.Errorf("%q vs %q", NormCode("2024-15-KЖ1"), NormCode("2024/15/КЖ1"))
	}
}

func TestDiffKeys(t *testing.T) {
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"a", "x", "c", "e", "f"}
	eq, ch := groupChanges(diffKeys(a, b), a, b)
	if len(eq) != 3 { // a c e
		t.Errorf("equal pairs: %v", eq)
	}
	del, add := 0, 0
	for _, c := range ch {
		del += len(c.delOnly) + len(c.pairs)
		add += len(c.addOnly) + len(c.pairs)
	}
	if del != 2 || add != 2 { // b,d removed; x,f added
		t.Errorf("del=%d add=%d", del, add)
	}
}

func TestDiffKeysShiftedPages(t *testing.T) {
	// 100 lines, РД inserts 20 new lines at the top: the rest must still line up
	var a, b []string
	for i := 0; i < 100; i++ {
		a = append(a, "line "+strings.Repeat("x", i%7)+string(rune('a'+i%26))+string(rune('A'+i/26)))
	}
	for i := 0; i < 20; i++ {
		b = append(b, "new "+string(rune('a'+i)))
	}
	b = append(b, a...)
	eq, ch := groupChanges(diffKeys(a, b), a, b)
	if len(eq) != 100 {
		t.Errorf("equal=%d want 100 (changes %d)", len(eq), len(ch))
	}
}

func TestPairingNeedsSimilarity(t *testing.T) {
	a := []string{"глубина заложения фундамента 2.4 м"}
	b := []string{"глубина заложения фундамента 2.6 м"}
	_, ch := groupChanges(diffKeys(a, b), a, b)
	if len(ch) != 1 || len(ch[0].pairs) != 1 {
		t.Fatalf("similar lines must pair as an edit: %+v", ch)
	}
	a = []string{"защитный слой бетона 40 мм"}
	b = []string{"гидроизоляция по рабочим чертежам"}
	_, ch = groupChanges(diffKeys(a, b), a, b)
	if len(ch[0].pairs) != 0 || len(ch[0].delOnly) != 1 || len(ch[0].addOnly) != 1 {
		t.Fatalf("different lines must not pair: %+v", ch)
	}
}

func lines(texts ...string) []Line {
	var out []Line
	for _, s := range texts {
		out = append(out, Line{Text: s, Key: NormKey(s)})
	}
	return out
}

func TestExtractParams(t *testing.T) {
	p := ExtractParams(lines(
		"Бетон класса В25, W6, F150",
		"арматура A500C и А400 по ГОСТ",
		"бетон B7,5 для подготовки",
		"Ф-1 В2 шт", // not a grade
	))
	if _, ok := p[0]["B25"]; !ok {
		t.Error("B25 not found")
	}
	if _, ok := p[0]["B7.5"]; !ok {
		t.Error("B7.5 not found")
	}
	if _, ok := p[0]["B2"]; ok {
		t.Error("B2 is not a valid grade")
	}
	if _, ok := p[1]["W6"]; !ok {
		t.Error("W6 not found")
	}
	if _, ok := p[2]["F150"]; !ok {
		t.Error("F150 not found")
	}
	if _, ok := p[3]["A500C"]; !ok {
		t.Error("A500C not found")
	}
	if _, ok := p[3]["A400"]; !ok {
		t.Error("A400 (Cyrillic А) not found")
	}
}

func TestCompareParams(t *testing.T) {
	ref := lines("класс бетона В25", "арматура А500С")
	// conflict
	d := compareParams(ref, lines("класс бетона В30", "арматура А500С"), false)
	if len(d) != 1 || d[0].Mode != "conflict" || d[0].OnlyRef[0] != "B25" || d[0].OnlyTgt[0] != "B30" {
		t.Fatalf("conflict: %+v", d)
	}
	// target adds a grade but keeps the reference one: fine
	if d := compareParams(ref, lines("В25", "В30 для плиты", "А500С"), false); len(d) != 0 {
		t.Fatalf("superset must be accepted: %+v", d)
	}
	// target only lacks the value: low "missing", and never for fragments
	d = compareParams(ref, lines("бетон В25, арматура А400"), false)
	if len(d) == 0 {
		t.Fatal("expected a finding for the changed rebar class")
	}
	if d := compareParams(ref, lines("только текст без параметров"), false); len(d) != 0 {
		t.Fatalf("no values on target side: nothing to compare, got %+v", d)
	}
}

func TestParseBBoxAndRows(t *testing.T) {
	out := `<doc><page width="100.0" height="200.0">
<word xMin="10" yMin="10" xMax="30" yMax="20">Hello</word>
<word xMin="40" yMin="10.5" xMax="60" yMax="20.5">w&amp;rld</word>
<word xMin="10" yMin="50" xMax="30" yMax="60">second</word>
<word xMin="70" yMin="190" xMax="95" yMax="199">Лист</word>
</page></doc>`
	pages := ParseBBox("d", out)
	if len(pages) != 1 || len(pages[0].Lines) != 3 {
		t.Fatalf("%+v", pages)
	}
	if pages[0].Lines[0].Text != "Hello w&rld" {
		t.Errorf("row join/unescape: %q", pages[0].Lines[0].Text)
	}
	if b := pages[0].Lines[0].Box; b.X0 != 0.1 || b.X1 != 0.6 || b.Y0 != 0.05 {
		t.Errorf("box: %+v", b)
	}
	if !pages[0].Lines[2].Stamp {
		t.Error("bottom-right row must be the title block")
	}
	if pages[0].HasText {
		t.Error("4 words is not a text page")
	}
}

func solid(w, h int, c color.Gray) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = c.Y
	}
	return img
}

func TestDiffImages(t *testing.T) {
	a := solid(480, 480, color.Gray{255})
	b := solid(480, 480, color.Gray{255})
	if boxes := DiffImages(a, b); len(boxes) != 0 {
		t.Fatalf("identical images: %v", boxes)
	}
	for y := 100; y < 160; y++ { // a dark patch on b only
		for x := 200; x < 300; x++ {
			b.SetGray(x, y, color.Gray{0})
		}
	}
	// different pixel size, same content must still align
	boxes := DiffImages(a, b)
	if len(boxes) != 1 {
		t.Fatalf("want one region, got %v", boxes)
	}
	bx := boxes[0]
	if bx.X0 > 200.0/480 || bx.X1 < 300.0/480-0.03 || bx.Y0 > 100.0/480 || bx.Y1 < 160.0/480-0.03 {
		t.Errorf("box does not cover the patch: %+v", bx)
	}
	small := solid(240, 240, color.Gray{255})
	if boxes := DiffImages(a, small); len(boxes) != 0 {
		t.Errorf("resolution change alone must not be a difference: %v", boxes)
	}
}

func TestBoxRoundTrip(t *testing.T) {
	b := &Box{0.1, 0.2, 0.3, 0.4}
	got := ParseBox(b.String())
	if got == nil || *got != *b {
		t.Errorf("%v → %q → %v", b, b.String(), got)
	}
	if ParseBox("") != nil || ParseBox("junk") != nil {
		t.Error("malformed box must be nil")
	}
}

func TestBuildRemark(t *testing.T) {
	r := BuildRemark([]Fact{
		{Description: "[детерм.] Расходятся значения: A → B", DocRef: "pd.pdf", DocTarget: "rd.pdf", PageRef: "1", PageTarget: "2"},
		{Description: "Класс бетона: в ПД B25, в РД B30", DocRef: "pd.pdf", DocTarget: "rd.pdf", PageRef: "1", PageTarget: "2"},
	})
	if r.DocumentRef != "ПД: pd.pdf; РД: rd.pdf" || r.Sheet != "ПД стр. 1, РД стр. 2" {
		t.Errorf("refs: %q / %q", r.DocumentRef, r.Sheet)
	}
	if strings.Contains(r.WhatDiverges, "[детерм.]") || strings.Count(r.WhatDiverges, "\n") != 1 {
		t.Errorf("what: %q", r.WhatDiverges)
	}
	if r.Status != "draft" || r.MustAlign == "" || r.DraftText != "" {
		t.Errorf("%+v", r)
	}
}

func TestSuggest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			Model    string                           `json:"model"`
			Messages []struct{ Role, Content string } `json:"messages"`
		}
		_ = json.NewDecoder(req.Body).Decode(&body)
		if body.Model != "m1" || len(body.Messages) != 2 || !strings.Contains(body.Messages[1].Content, "Класс бетона") {
			http.Error(w, "bad request", 400)
			return
		}
		_, _ = w.Write([]byte(`{"message":{"content":" Замечание: привести в соответствие. "}}`))
	}))
	defer srv.Close()
	s := NewSuggester(srv.URL, "m1")
	got, err := s.Suggest(context.Background(), Remark{WhatDiverges: "— Класс бетона: в ПД B25, в РД B30"})
	if err != nil || got != "Замечание: привести в соответствие." {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := NewSuggester(srv.URL, "").Suggest(context.Background(), Remark{}); err == nil {
		t.Error("no model configured must fail")
	}
}

func mkLine(text string, x0, y0, x1, y1 float64) Line {
	l := Line{Text: text, Key: NormKey(text), Box: Box{x0, y0, x1, y1}}
	return l
}

func TestDrawingClassification(t *testing.T) {
	if !isDrawing(2384, 1684, nil) {
		t.Error("an A1 sheet is a drawing")
	}
	var text []Line
	for i := 0; i < 40; i++ {
		text = append(text, mkLine("Фундаменты выполнить из бетона класса В25, W6, F150.", 0.1, 0.1, 0.9, 0.12))
	}
	if isDrawing(595, 842, text) {
		t.Error("an A4 text page is not a drawing")
	}
	var scattered []Line
	for i := 0; i < 30; i++ {
		scattered = append(scattered, mkLine("+40", 0.1, 0.1, 0.12, 0.11))
	}
	if !isDrawing(842, 595, scattered) {
		t.Error("a sheet of scattered short labels is a drawing")
	}
}

func TestStampRowsAreRecognisedByContent(t *testing.T) {
	body := mkLine("постоянного отвода участка вдоль трассы", 0.39, 0.88, 0.9, 0.9)
	if isStampRow(false, body) {
		t.Error("body text at the bottom of a text page is content, not a title block")
	}
	for _, s := range []string{"Подпись Изм. Кол.уч. Лист Nдок. Подп. Дата", "2024-15-КЖ Стадия П Лист 1 Листов 3", "Взам. инв. №"} {
		if !isStampRow(false, mkLine(s, 0.1, 0.9, 0.9, 0.92)) {
			t.Errorf("%q must be a title-block row", s)
		}
	}
	if !isStampRow(false, mkLine("Проверил Маркелова 10.07.23", 0.1, 0.9, 0.5, 0.92)) {
		t.Error("signature row at the bottom is title block")
	}
	if isStampRow(false, mkLine("Проверил расчёт на прочность", 0.1, 0.3, 0.5, 0.32)) {
		t.Error("the word in the middle of a page is content")
	}
	// on a drawing sheet the corner is the title block by position
	if !isStampRow(true, mkLine("любая надпись", 0.8, 0.9, 0.95, 0.92)) {
		t.Error("drawing sheet: bottom-right corner is the title block")
	}
}

func TestStrictCodeShape(t *testing.T) {
	for _, bad := range []string{"02.24г", "Т380-71", "13-Расчеты", "ХС-710", "12.03.2024"} {
		if plausibleCode(NormCode(bad)) {
			t.Errorf("%q must not be taken for a шифр", bad)
		}
	}
	for _, good := range []string{"2024-15-КЖ1", "03102022/РД-3-ПИК1-НК2", "100720-1/П-3-ТКР5-1"} {
		if !plausibleCode(NormCode(good)) {
			t.Errorf("%q is a шифр", good)
		}
	}
}

func textDoc(id, side string, pages ...[]string) *Document {
	d := &Document{ID: id, Name: id + ".pdf", Side: side}
	for i, rows := range pages {
		p := Page{Number: i + 1, W: 595, H: 842, HasText: true}
		for j, r := range rows {
			l := mkLine(r, 0.1, 0.1+float64(j)*0.03, 0.8, 0.12+float64(j)*0.03)
			l.DocID, l.Page = id, i+1
			p.Lines = append(p.Lines, l)
		}
		d.Pages = append(d.Pages, p)
	}
	return d
}

func TestSingleVolumeAgainstSingleVolumeIsPairedWithoutOverlap(t *testing.T) {
	ref := textDoc("r", SideRef, []string{"Глубина заложения фундамента 2,4 м", "Толщина стен 300 мм"})
	tgt := textDoc("t", SideTarget, []string{"Совсем другой текст про кабели", "Прокладка траншей 1200 м"})
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	if len(res.Pairs) != 1 {
		t.Fatalf("a file against a file is the user's own pairing, got %d pairs", len(res.Pairs))
	}
	if !res.Stats.SimilarityKnown || res.Stats.Similarity > 1 {
		t.Errorf("no overlap must read as ~0%% similarity, got %+v", res.Stats)
	}
}

func TestCoverAndDatesAreNotCompared(t *testing.T) {
	cover := func(year string) []string {
		return []string{"РАБОЧАЯ ДОКУМЕНТАЦИЯ", "03102022/РД-3-ПИК1-НК2", "« » " + year + " г.", "Инв. " + year}
	}
	body := []string{"Трубы напорные ПЭ100 длина 2617,8 п.м", "Колодцы железобетонные 12 шт"}
	ref := textDoc("r", SideRef, cover("2023"), body)
	tgt := textDoc("t", SideTarget, cover("2026"), body)
	for _, d := range []*Document{ref, tgt} {
		d.Pages[0].Cover = true
	}
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	for _, f := range res.Findings {
		t.Errorf("cover page / dates must not produce findings: %s", f.Description)
	}
	if res.Stats.Similarity != 100 {
		t.Errorf("similarity %v", res.Stats.Similarity)
	}
}

func TestDrawingSheetsAreNotTextDiffed(t *testing.T) {
	sheet := func(id string, labels ...string) *Document {
		d := textDoc(id, SideRef, labels)
		d.Pages[0].Drawing = true
		d.Pages[0].W, d.Pages[0].H = 2384, 1684
		return d
	}
	ref := sheet("r", "+40 .75", "2x3")
	tgt := sheet("t", "+40", "3x3")
	tgt.Side = SideTarget
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	for _, f := range res.Findings {
		if f.Kind != KindDrawing {
			t.Errorf("drawing labels must not be text-diffed: %s", f.Description)
		}
	}
	if res.Stats.DrawingRef != 1 || res.Stats.DrawingTarget != 1 {
		t.Errorf("drawing counts: %+v", res.Stats)
	}
	if res.Stats.SimilarityKnown {
		t.Error("with nothing but drawings the similarity is unknown, not 0 or 100")
	}
	if len(res.Findings) != 1 || res.Findings[0].Severity != "low" {
		t.Errorf("without an image source one visual-check finding is expected: %+v", res.Findings)
	}
}

func TestDrawingsOnlyReferenceDoesNotFloodWithAddedLines(t *testing.T) {
	ref := textDoc("r", SideRef, []string{"метка"})
	ref.Pages[0].Drawing = true
	ref.Pages[0].Lines = nil
	var rows []string
	for i := 0; i < 30; i++ {
		rows = append(rows, fmt.Sprintf("Строка пояснительной записки номер %d", i))
	}
	tgt := textDoc("t", SideTarget, rows)
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	for _, f := range res.Findings {
		if f.Kind == KindText || f.Kind == KindTableRow {
			t.Fatalf("no reference text to compare: %s", f.Description)
		}
	}
}

func TestScannedTargetIsRecognizedAndCompared(t *testing.T) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not installed")
	}
	ocr := NewOCR("tesseract", "rus", t.TempDir())
	if !ocr.Enabled() {
		t.Skip("tesseract unavailable")
	}
	read := func(rel string) []byte {
		b, err := os.ReadFile(filepath.Join("..", "fixtures", rel))
		if err != nil {
			t.Skipf("fixture missing (run fixtures/gen_cmp_fixtures.mjs): %v", err)
		}
		return b
	}
	res, err := AnalyzeWith([]Input{
		{ID: "pd", Name: "pd.pdf", Side: SideRef, Data: read("pack-scan/pd.pdf")},
		{ID: "rd", Name: "rd.pdf", Side: SideTarget, Data: read("pack-scan/rd.pdf")},
	}, t.TempDir(), AnalyzeOptions{OCR: ocr})
	if err != nil {
		t.Fatal(err)
	}
	notes := strings.Join(res.Notes, " ")
	if !strings.Contains(notes, "Распознано (OCR) листов-сканов: 3") {
		t.Fatalf("scanned sheets must be recognized: %s", notes)
	}
	// the depth edit (2,4 → 2,6) is found through the recognized text
	var depth *Finding
	for i, f := range res.Findings {
		if strings.Contains(f.Description, "Глубина") {
			depth = &res.Findings[i]
		}
	}
	if depth == nil {
		t.Fatalf("the edited line was not found in the scan; findings: %d, notes: %s", len(res.Findings), notes)
	}
	if depth.Severity != "medium" || !depth.OCR || !strings.Contains(depth.Description, "распознанному") || depth.BoxTarget == nil || depth.PageTarget != 1 {
		t.Errorf("a finding from a scan is capped at medium, flagged and positioned: %+v", depth)
	}
	if !res.Stats.SimilarityKnown || res.Stats.Similarity < 60 {
		t.Errorf("recognition should keep most lines comparable: %+v", res.Stats)
	}
	// without OCR the same scan is just a drawing page
	res2, err := AnalyzeWith([]Input{
		{ID: "pd", Name: "pd.pdf", Side: SideRef, Data: read("pack-scan/pd.pdf")},
		{ID: "rd", Name: "rd.pdf", Side: SideTarget, Data: read("pack-scan/rd.pdf")},
	}, t.TempDir(), AnalyzeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Stats.SimilarityKnown {
		t.Errorf("without OCR there is no text to compare on the scan side: %+v", res2.Stats)
	}
}

func TestParseTSV(t *testing.T) {
	tsv := "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
		"1\t1\t0\t0\t0\t0\t0\t0\t1000\t500\t-1\t\n" +
		"5\t1\t1\t1\t1\t1\t100\t50\t200\t25\t96.5\tГлубина\n" +
		"5\t1\t1\t1\t1\t2\t320\t50\t60\t25\t12.0\tшум\n" +
		"5\t1\t1\t1\t1\t3\t400\t50\t60\t25\t90\t \n"
	ws, conf := parseTSV(tsv, 500, 250)
	if conf != 96.5 || len(ws) != 1 || ws[0].text != "Глубина" || ws[0].x0 != 50 || ws[0].y0 != 25 || ws[0].x1 != 150 || ws[0].y1 != 37.5 {
		t.Errorf("%+v", ws)
	}
}

func TestOCRResultIsCachedAndReadBack(t *testing.T) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract not installed")
	}
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "pack-scan", "rd.pdf"))
	if err != nil {
		t.Skip("fixture missing")
	}
	cache := t.TempDir()
	ocr := NewOCR("tesseract", "rus", cache)
	first, n1, _, err := ExtractPagesOCR(context.Background(), "d", data, cache, ocr)
	if err != nil || n1 == 0 {
		t.Fatalf("first run: %d %v", n1, err)
	}
	// the second run is served from the cache and must give the same text
	second, n2, _, err := ExtractPagesOCR(context.Background(), "d", data, cache, ocr)
	if err != nil || n2 != n1 {
		t.Fatalf("second run: %d %v", n2, err)
	}
	for i := range first {
		if len(first[i].Lines) != len(second[i].Lines) {
			t.Fatalf("page %d: %d lines vs %d from the cache", i+1, len(first[i].Lines), len(second[i].Lines))
		}
		for j := range first[i].Lines {
			if first[i].Lines[j].Text != second[i].Lines[j].Text {
				t.Fatalf("cached text differs: %q vs %q", first[i].Lines[j].Text, second[i].Lines[j].Text)
			}
		}
	}
}

func TestRelabelVersions(t *testing.T) {
	in := "Изменён текст: «a» (ПД, стр. 3) → «b» (РД, стр. 5); ПДФ и РДУ не трогаем, ПД2 тоже"
	want := "Изменён текст: «a» (До, стр. 3) → «b» (После, стр. 5); ПДФ и РДУ не трогаем, ПД2 тоже"
	if got := Relabel(in, ModeVersions); got != want {
		t.Fatalf("\n got %q\nwant %q", got, want)
	}
	if Relabel(in, ModeStages) != in || Relabel(in, "") != in {
		t.Fatal("stage names stay in the default mode")
	}
}
