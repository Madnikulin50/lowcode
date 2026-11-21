package idcheck

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fp(f float64) *float64 { return &f }

func TestParseActNumber(t *testing.T) {
	cases := []struct{ raw, want string }{
		{"1/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ", "1/2.14.1.9/ОП6/ИССО2.2/СКТРАСТ"},
		{"№2/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ", "2/2.14.1.9/ОП6/ИССО2.2/СКТРАСТ"},
		{"1/2.14.2.5/ОП№3ИССО 2.2/СКТРАСТ", "1/2.14.2.5/ОП3/ИССО2.2/СКТРАСТ"},
		{"1/2.14.2.5/OП№3/ИCCO 2.2/CKTPACT", "1/2.14.2.5/ОП3/ИССО2.2/СКТРАСТ"}, // Latin look-alikes
	}
	for _, c := range cases {
		if got := ParseActNumber(c.raw, '/').Key(); got != c.want {
			t.Errorf("%q: got %q want %q", c.raw, got, c.want)
		}
	}
}

func TestParseActFileName(t *testing.T) {
	cases := []struct {
		name, key, date string
	}{
		{"1_2.14.1.9_ОП№6_ИССО2.2_СКТРАСТ от 18.08.2024.pdf", "1/2.14.1.9/ОП6/ИССО2.2/СКТРАСТ", "18.08.2024"},
		{"2-2.14.1.9-ОП№6-ИССО2.2-СКТРАСТ от 18.08.2024.pdf", "2/2.14.1.9/ОП6/ИССО2.2/СКТРАСТ", "18.08.2024"},
		{"1_2.14.2.3_ОП№2_ИССО 2.2_СКТРАСТ от 02.10.24.pdf", "1/2.14.2.3/ОП2/ИССО2.2/СКТРАСТ", "02.10.2024"},
		{"10_2.52.24.1;2.52.24.2_ИССО17.2_Б4.3 –Б7.3_МТК от 28.10.2024.pdf", "10/2.52.24.1;2.52.24.2/ИССО17.2/Б4.3-Б7.3/МТК", "28.10.2024"},
	}
	for _, c := range cases {
		n, d := ParseActFileName(c.name)
		if n.Key() != c.key || FmtDate(d) != c.date {
			t.Errorf("%q: got %q %s, want %q %s", c.name, n.Key(), FmtDate(d), c.key, c.date)
		}
	}
}

func TestParseDate(t *testing.T) {
	for in, want := range map[string]string{
		"18.08.2024г.":         "18.08.2024",
		"«18» августа 2024 г.": "18.08.2024",
		"02.10.24":             "02.10.2024",
		"45639":                "13.12.2024",
		"1 мая 2024":           "01.05.2024",
	} {
		d, ok := ParseDate(in)
		if !ok || FmtDate(d) != want {
			t.Errorf("%q: got %s %v want %s", in, FmtDate(d), ok, want)
		}
	}
}

func TestChecksums(t *testing.T) {
	if !ValidOGRN("1187746140460") || !ValidINN("7731396110") {
		t.Error("valid OGRN/INN rejected")
	}
	if ValidOGRN("1187746140461") || ValidINN("77313966110") {
		t.Error("invalid OGRN/INN accepted")
	}
}

func testRegistry() *Registry {
	d := func(s string) time.Time { t, _ := ParseDate(s); return t }
	return &Registry{Rows: []RegistryRow{
		{Row: 9, Kind: RowPosition, Position: "2.14.1.9", Unit: "м3", QtyKS2: fp(26.3)},
		{Row: 10, Kind: RowAct, Position: "2.14.1.9", DocNumber: "1/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ", DocDate: d("18.08.2024")},
		{Row: 11, Kind: RowAct, Position: "2.14.1.9", DocNumber: "2/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ", DocDate: d("18.08.2024"), QtyProject: fp(26.3), QtyFact: fp(26.3)},
		{Row: 27, Kind: RowPosition, Position: "2.14.2.6", Unit: "м3", QtyKS2: fp(12)},
		{Row: 30, Kind: RowAct, Position: "2.14.2.6", DocNumber: "3/2.14.1.6/ОП№5/ИССО 2.2/СКТРАСТ", DocDate: d("07.08.2024")},
		{Row: 31, Kind: RowAct, Position: "2.14.2.6", DocNumber: "4/2.14.1.6/ОП№5/ИССО 2.2/СКТРАСТ", DocDate: d("09.08.2024"), QtyProject: fp(6), QtyFact: fp(5)},
	}}
}

func TestRegistryConsistencyAndMatch(t *testing.T) {
	reg := testRegistry()
	fs := CheckRegistryConsistency(reg)
	if len(fs) != 2 || fs[0].RegistryRow != 30 {
		t.Fatalf("expected rows 30/31 flagged, got %+v", fs)
	}
	num, date := ParseActFileName("3_2.14.2.6_ОП№5_ИССО 2.2_СКТРАСТ от 07.08.24.pdf")
	if r := MatchRegistryRow(reg, num, date); r == nil || r.Row != 30 {
		t.Fatalf("loose match failed: %+v", r)
	}
}

func TestReconcileKS2(t *testing.T) {
	reg := testRegistry()
	ks := &KS2{Rows: []KS2Row{{Position: "2.14.1.9", Qty: 26.3}, {Position: "2.14.2.6", Qty: 6}}}
	sum := ReconcileKS2(reg, ks)
	got := map[string]PositionSummary{}
	for _, s := range sum {
		got[s.Position] = s
	}
	if got["2.14.1.9"].Status != "ok" {
		t.Errorf("2.14.1.9: %+v", got["2.14.1.9"])
	}
	// min(6, 5) = 5 ≠ 6
	if s := got["2.14.2.6"]; s.Status != "mismatch" || s.RegistrySum != 5 {
		t.Errorf("2.14.2.6: %+v", s)
	}
	fs := Check8Findings(sum)
	if len(fs) != 2 { // mismatch + col 11 (12) ≠ КС-2 (6)
		t.Errorf("check 8 findings: %+v", fs)
	}
}

func baseAct() *ActData {
	return &ActData{
		Number: "№2/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ", Date: "«18» августа 2024 г.",
		P1Work: "Бетонирование насадки опоры №6 Vпр=26,3 м3 / Vф=26,3 м3", P1VolProject: fp(26.3), P1VolFact: fp(26.3),
		P2RDCode: "03102022/РД-ИССО2.2-ОП-КЖ4.1 (лист №1,5,6)", P2Text: "Путепровод ...",
		P5Start: "18.08.2024", P5End: "18.08.2024",
		P6Norms:  []string{`СП 48.13330.2019 "Организация строительства. СНиП 12-01-2004"`, "СП 46.13330.2012 Мосты и трубы. Актуализированная редакция СНиП 3.06.04-91", "СП 70.13330.2012 Несущие и ограждающие конструкции", "шифр ППР-ИССО2.2-ОП-01; РД 03102022/РД-ИССО2.2-ОП-КЖ4.1"},
		P6RDCode: "03102022/РД-ИССО2.2-ОП-КЖ4.1", SchemeRDCode: "03102022/РД-ИССО2.2-ОП-КЖ4.1",
		SchemeVolProj: fp(26.3), SchemeVolFact: fp(26.3),
		Appendix: []DocRef{{Name: "Акт проведения входного контроля", Number: "СТМ22-КБ180801", Date: "18.08.2024"}},
		P4Docs:   []DocRef{{Kind: "акт", Name: "Акт входного контроля", Number: "СТМ22-КБ180801", Date: "18.08.2024"}},
		People: []Person{
			{Role: "Представитель застройщика по вопросам строительного контроля", FIO: "Булатов В.Н.", OrderNo: "№1 от 10.01.2024", NRSID: "С-77-246659"},
			{Role: "Представитель лица, осуществляющего подготовку проектной документации", FIO: "Пуляевский Д.В.", OrderNo: "приказ №7 от 26.02.2032г."},
		},
		Orgs: []Org{{Name: "АО \"Технологии контроля безопасности\"", OGRN: "1187746140460", INN: "7731396110"}},
	}
}

func actInput(d *ActData) *ActInput {
	reg := testRegistry()
	num, date := ParseActFileName("2-2.14.1.9-ОП№6-ИССО2.2-СКТРАСТ от 18.08.2024.pdf")
	return &ActInput{Key: "a", FileNum: num, FileDate: date, Rows: MatchRegistryRows(reg, num, date), Data: d, Norms: NewNormCatalog(DefaultNorms)}
}

func findCheck(fs []Finding, check int, field string) *Finding {
	for i := range fs {
		if fs[i].Check == check && (field == "" || fs[i].Field == field) {
			return &fs[i]
		}
	}
	return nil
}

func TestCheckActClean(t *testing.T) {
	d := baseAct()
	d.People[1].OrderNo = "приказ №7 от 26.02.2023г."
	fs := CheckAct(actInput(d))
	for _, f := range fs {
		if f.Severity == High || f.Severity == Medium {
			t.Errorf("unexpected finding: %+v", f)
		}
	}
}

func TestCheckActFindings(t *testing.T) {
	d := baseAct()
	d.P1VolFact = fp(25)                             // check 1 + 4
	d.SchemeRDCode = "03102022/РД-ИССО2.2-ОП-КЖ4.2"  // check 3
	d.P6Norms = append(d.P6Norms, "ГОСТ 18105-2010") // check 5: superseded 2019
	d.Orgs[0].INN = "7731396111"                     // check 7 checksum
	d.Attached = []DocRef{{Kind: "паспорт", Number: "123", Date: "01.08.2024", ValidUntil: "10.08.2024", Page: 7}}
	fs := CheckAct(actInput(d))
	for _, want := range []struct {
		check int
		field string
	}{{1, "volume"}, {3, "scheme_rd_code"}, {4, "vol_fact"}, {5, "p6"}, {6, "order_date"}, {7, "inn"}, {2, "valid_until"}} {
		if findCheck(fs, want.check, want.field) == nil {
			t.Errorf("missing check %d/%s; got:\n%s", want.check, want.field, dump(fs))
		}
	}
	if f := findCheck(fs, 6, "order_date"); f != nil && !strings.Contains(f.Description, "из будущего") {
		t.Errorf("order date: %s", f.Description)
	}
}

func TestExtractNormCodes(t *testing.T) {
	got := ExtractNormCodes("СП 46.13330.2012 \"Мосты и трубы. Актуализированная редакция СНиП 3.06.04-91\", ГОСТ 10180-2012; ГОСТ Р 51872-2019")
	want := []string{"СП 46.13330.2012", "ГОСТ 10180-2012", "ГОСТ Р 51872-2019"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v", got)
	}
}

func TestParseNRS(t *testing.T) {
	html := `<tr>
<td style="width: 10%; max-width: 10%"><img src ="../imgdir/numbers/x.png"></td>
<td style="width: 15%"><img src="f.png"></td><td><img src="d.png"></td><td></td><td></td><td></td><td></td>
<td style="width: 20%">Организация выполнения работ</td><td style="width: 8%">Действует</td></tr>`
	r := parseNRS(html)
	if !r.Found || r.Excluded || r.Status != "Действует" {
		t.Errorf("%+v", r)
	}
}

func dump(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.Description + "\n")
	}
	return b.String()
}

func TestOCRTolerance(t *testing.T) {
	d := &ActData{P1Work: "Армирование насадки опоры №6 Мрд=3993,4 кг./ Мф=3993,4 кг."}
	postProcess(d)
	if d.P1VolProject == nil || *d.P1VolProject != 3993.4 || d.P1VolFact == nil || d.P1Unit != "кг" {
		t.Errorf("mass not parsed: %+v %+v %q", d.P1VolProject, d.P1VolFact, d.P1Unit)
	}
	if FirstDigitRun("77313966110, 119034") != "77313966110" {
		t.Error("FirstDigitRun")
	}
	a := baseAct()
	a.Number = "№2/2.14.19/ОП№6/ИССО 2.2/СКРАСТ" // OCR dropped a dot and a letter
	a.Orgs[0].INN = "77313966110, 119034"
	fs := CheckAct(actInput(a))
	if f := findCheck(fs, 1, "number"); f == nil || f.Severity != Low {
		t.Errorf("near-miss number should be low: %+v", f)
	}
	// "77313966110, 119034": a clean 11-digit run is a real typo in the act
	if f := findCheck(fs, 7, "inn"); f == nil || f.Severity != High {
		t.Errorf("11-digit INN should be high: %+v", f)
	}
	a.Orgs[0].INN = "7731396110119034" // glued to the postcode by OCR
	if f := findCheck(CheckAct(actInput(a)), 7, "inn"); f == nil || f.Severity != Low {
		t.Errorf("glued INN should be low: %+v", f)
	}
}

func TestSchemaDeterministic(t *testing.T) {
	a, _ := json.Marshal(obj(map[string]any{"a": str(), "b": str(), "c": str(), "d": str()}))
	for i := 0; i < 20; i++ {
		b, _ := json.Marshal(obj(map[string]any{"a": str(), "b": str(), "c": str(), "d": str()}))
		if string(a) != string(b) {
			t.Fatal("schema JSON is not deterministic — vision cache keys would change between runs")
		}
	}
}

func TestLocate(t *testing.T) {
	tsv := "level\tpage_num\tblock_num\tpar_num\tline_num\tword_num\tleft\ttop\twidth\theight\tconf\ttext\n" +
		"1\t1\t0\t0\t0\t0\t0\t0\t1000\t2000\t-1\t\n" +
		"5\t1\t1\t1\t1\t1\t100\t200\t50\t20\t90\tОГРН\n" +
		"5\t1\t1\t1\t1\t2\t160\t200\t150\t20\t90\t1187746140460,\n" +
		"5\t1\t1\t1\t1\t3\t320\t200\t40\t20\t90\tИНН\n" +
		"5\t1\t1\t1\t1\t4\t370\t200\t120\t20\t90\t77313966110,\n" +
		"5\t1\t1\t1\t2\t1\t100\t400\t300\t20\t90\tиЛОПА6ИССО\n" +
		"5\t1\t1\t1\t2\t2\t410\t400\t200\t20\t90\t2.2/СКТРАСТ\n"
	ws := parseTSV(tsv)
	if len(ws) != 6 {
		t.Fatalf("words: %d", len(ws))
	}
	f := Finding{Check: 7, Field: "inn", Actual: "77313966110"}
	p, b, ok := LocateFinding(&f, nil, map[int][]Word{1: ws}, []int{1})
	if !ok || p != 1 || b.Y != 0.1 || b.X != 0.1 { // widened to the whole line
		t.Errorf("inn: p=%d ok=%v box=%s", p, ok, b)
	}
	f = Finding{Check: 1, Field: "number", Actual: "№1/2.14.19/ОП№6/ИССО 2.2/СКТРАСТ"}
	if _, b, ok := LocateFinding(&f, nil, map[int][]Word{1: ws}, []int{1}); !ok || b.Y != 0.2 {
		t.Errorf("number tail: ok=%v box=%s", ok, b)
	}
}

func TestListJPEGSkipsDerived(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"p-01.jpg", "p-02.jpg", "p-03.jpg", "p-03-stamp.jpg", "p-03.tsv"} {
		_ = os.WriteFile(filepath.Join(dir, n), nil, 0o644)
	}
	if got := listJPEG(dir); len(got) != 3 {
		t.Errorf("pages: %v", got)
	}
}

func TestFillFromOCR(t *testing.T) {
	d := &ActData{People: []Person{{FIO: "Булатов В.Н.", OrderNo: "приказ организации ... не требуется"}}}
	fillFromOCR(d, "строительного контроля АО \"ТКБ\" Булатов В.Н., идентификационный № С-77-246659, приказ\nорганизации в области строительства\nкаптионы мелким шрифтом\nЛ от 10.01.2024, ОГРН 1187746140460")
	if d.People[0].OrderDate != "10.01.2024" || !strings.Contains(d.People[0].OrderNo, "от 10.01.2024") {
		t.Errorf("%+v", d.People[0])
	}
}
