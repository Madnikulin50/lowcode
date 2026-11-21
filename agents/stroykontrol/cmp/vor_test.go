package cmp

import (
	"strings"
	"testing"
)

func w(t string, x0, y float64) Word {
	return Word{T: t, B: Box{x0, y - 0.006, x0 + 0.012*float64(len([]rune(t))), y + 0.006}}
}

// tablePage lays a ВОР out like a drawing sheet table: position, name,
// unit, quantity columns under a header.
func tablePage(num int, title string, rows [][3]string, withHeader bool) Page {
	p := Page{Number: num, W: 595, H: 842, HasText: true}
	addRow := func(y float64, ws ...Word) {
		l := Line{Page: num, Box: ws[0].B}
		var parts []string
		for _, x := range ws {
			parts = append(parts, x.T)
			l.Box = l.Box.union(x.B)
		}
		l.Text = strings.Join(parts, " ")
		l.Key = NormKey(l.Text)
		p.Lines = append(p.Lines, l)
		p.Words = append(p.Words, ws...)
	}
	if title != "" {
		addRow(0.08, w(title, 0.2, 0.08))
	}
	if withHeader {
		addRow(0.15, w("№", 0.10, 0.15), w("Наименование", 0.2, 0.15), w("Ед.", 0.5, 0.15), w("Кол-во", 0.62, 0.15))
		addRow(0.19, w("1", 0.10, 0.19), w("2", 0.3, 0.19), w("3", 0.5, 0.19), w("4", 0.62, 0.19))
	}
	y := 0.25
	for i, r := range rows {
		parts := strings.Split(r[0], "\n") // wrapped name
		for k, part := range parts {
			var ws []Word
			x := 0.16
			for _, tok := range strings.Fields(part) { // names are separate words in a real PDF
				ws = append(ws, w(tok, x, y+float64(k)*0.017))
				x += 0.012*float64(len([]rune(tok))) + 0.008
			}
			if k == len(parts)/2 {
				ws = append(ws, w("1", 0.11, y+float64(k)*0.017))
				ws = append(ws, w(r[1], 0.5, y+float64(k)*0.017))
				for qi, q := range strings.Fields(r[2]) { // "6 322,83" → two words
					ws = append(ws, w(q, 0.62+0.03*float64(qi), y+float64(k)*0.017))
				}
			}
			addRow(y+float64(k)*0.017, ws...)
		}
		y += 0.017*float64(len(parts)) + 0.012 + float64(i)*0
	}
	p.HasText = true
	return p
}

func docOf(id, side string, pages ...Page) *Document {
	d := &Document{ID: id, Name: id + ".pdf", Side: side, Pages: pages}
	for i := range d.Pages {
		for j := range d.Pages[i].Lines {
			d.Pages[i].Lines[j].DocID = id
		}
	}
	return d
}

func TestExtractItemsReadsWrappedNamesAndThousands(t *testing.T) {
	p := tablePage(1, "Ведомость объемов работ", [][3]string{
		{"Разработка грунта\nв траншеях вручную", "м3", "6 322,83"},
		{"Устройство песчаной подушки", "м3", "19,84"},
		{"Перевозка грунта", "т", "77,395"},
	}, true)
	items := ExtractItems("d", p, pageTableKind(p))
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %+v", items)
	}
	if items[0].Qty != 6322.83 || items[0].Unit != "м3" || !strings.Contains(items[0].Name, "траншеях") || !strings.Contains(items[0].Name, "Разработка") {
		t.Errorf("a number with a thousands space and a wrapped name: %+v", items[0])
	}
	if items[2].Unit != "т" || items[2].Qty != 77.395 {
		t.Errorf("%+v", items[2])
	}
	if got := ExtractItems("d", p, ""); got != nil {
		t.Error("a page that is not a table yields nothing")
	}
}

func TestContinuationSheetsInheritTheTableKind(t *testing.T) {
	first := tablePage(1, "Ведомость объемов работ", [][3]string{{"Бетон B25", "м3", "10"}, {"Арматура A500C", "т", "2"}, {"Опалубка", "м2", "40"}}, true)
	next := tablePage(2, "", [][3]string{{"Гидроизоляция стен", "м2", "30"}, {"Засыпка пазух", "м3", "15"}, {"Уплотнение грунта", "м3", "15"}}, true)
	d := docOf("d", SideRef, first, next)
	items, pages := volumeItems(d, 1, 2)
	if len(items) != 6 || !pages[1] || !pages[2] {
		t.Fatalf("both sheets are the bill: %d items, pages %v", len(items), pages)
	}
	for _, it := range items {
		if it.Table != TableVOR {
			t.Errorf("kind not inherited: %+v", it)
		}
	}
}

func volume1(items ...[3]string) []Page {
	return []Page{tablePage(1, "Ведомость объемов работ", items, true)}
}

func TestBillQuantityChangesBetweenStages(t *testing.T) {
	ref := docOf("r", SideRef, volume1(
		[3]string{"Бетонирование плиты из бетона B25", "м3", "12,5"},
		[3]string{"Армирование плиты арматурой A500C", "т", "3,2"},
		[3]string{"Гидроизоляция стен", "м2", "40"},
		[3]string{"Выравнивающая стяжка", "м2", "100"},
	)...)
	tgt := docOf("t", SideTarget, volume1(
		[3]string{"Бетонирование плиты из бетона B25", "м3", "14,0"}, // +12% → high
		[3]string{"Армирование плиты арматурой A500C", "т", "3,25"},  // +1.6% → low
		[3]string{"Гидроизоляция стен", "м2", "40"},                  // same
		[3]string{"Устройство дренажной трубы ПНД", "м", "25"},       // new
	)...)
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{})
	var high, low, info int
	for _, f := range res.Findings {
		if f.Kind != KindVOR {
			t.Errorf("only table findings expected: %s", f.Description)
		}
		switch {
		case strings.Contains(f.Description, "изменение +12.0%"):
			high++
			if f.Severity != "high" || f.PageRef != 1 || f.BoxRef == nil || f.BoxTarget == nil {
				t.Errorf("12%% change: %+v", f)
			}
		case strings.Contains(f.Description, "изменение +1.6%"):
			low++
			if f.Severity != "low" {
				t.Errorf("1.6%% change: %+v", f)
			}
		case strings.Contains(f.Description, "не найдена в РД") || strings.Contains(f.Description, "отсутствует в ПД"):
			info++
			if f.Severity != "low" {
				t.Errorf("an unmatched position is information only: %+v", f)
			}
		default:
			t.Errorf("unexpected: %s", f.Description)
		}
	}
	if high != 1 || low != 1 || info != 2 {
		t.Errorf("high=%d low=%d unmatched=%d: %+v", high, low, info, res.Findings)
	}
	if !res.Stats.SimilarityKnown || res.Stats.TableItems != 8 {
		t.Errorf("a volume of tables still has a similarity: %+v", res.Stats)
	}
}

func TestImplausibleJumpIsNeverHigh(t *testing.T) {
	ref := docOf("r", SideRef, volume1([3]string{"Бетонирование плиты B25", "м3", "30883"}, [3]string{"Опалубка", "м2", "10"}, [3]string{"Арматура", "т", "5"})...)
	tgt := docOf("t", SideTarget, volume1([3]string{"Бетонирование плиты B25", "м3", "1"}, [3]string{"Опалубка", "м2", "10"}, [3]string{"Арматура", "т", "5"})...)
	for _, f := range Compare([]*Document{ref}, []*Document{tgt}, Options{}).Findings {
		if f.Severity == "high" {
			t.Errorf("a 30000-fold jump is a misread, not a finding to escalate: %s", f.Description)
		}
	}
}

func TestBillAgainstSpecificationInsideAVolume(t *testing.T) {
	bill := tablePage(1, "Ведомость объемов работ", [][3]string{
		{"Армирование плиты арматурой A500C", "т", "3,2"},
		{"Гидроизоляция обмазочная стен", "м2", "40"},
		{"Бетон B25 фундаментной плиты", "м3", "12,5"},
	}, true)
	spec := tablePage(2, "Спецификация оборудования, изделий и материалов", [][3]string{
		{"Арматура A500C d16 для плиты", "т", "2,0"},   // the bill's 3.2 is split in two lines…
		{"Арматура A500C d12 для плиты", "т", "1,2"},   // …whose sum agrees
		{"Гидроизоляция обмазочная стен", "м2", "45"},  // disagrees
		{"Бетон B25 фундаментной плиты", "м3", "12,5"}, // agrees
	}, true)
	v := buildVolume(docOf("d", SideTarget, bill, spec), 0, "X-1-КЖ", "T", 1, 2, "d", true)
	got := internalConsistency(v, true)
	if len(got) != 1 || !strings.Contains(got[0].Description, "Гидроизоляция") || got[0].Severity != "medium" || got[0].TargetDoc != "d" || got[0].PageTarget != 1 {
		t.Fatalf("only the waterproofing disagrees (aggregated reinforcement is fine): %+v", got)
	}
}

func TestPackVORFixture(t *testing.T) {
	pd := loadDoc(t, "pd", SideRef, "pack-vor/pd.pdf")
	rd := loadDoc(t, "rd", SideTarget, "pack-vor/rd.pdf")
	res := Compare([]*Document{pd}, []*Document{rd}, Options{})
	by := map[string]Finding{}
	for _, f := range res.Findings {
		if f.Kind != KindVOR {
			t.Errorf("only table findings expected: %s", f.Description)
		}
		for _, k := range []string{"плиты из бетона B25", "Бетон B25", "Гидроизоляция", "дренаж"} {
			if strings.Contains(f.Description, k) {
				by[k+"|"+f.Severity] = f
			}
		}
	}
	for _, want := range []string{"плиты из бетона B25|high", "Бетон B25|high", "Гидроизоляция|high", "Гидроизоляция|medium"} {
		if _, ok := by[want]; !ok {
			t.Errorf("missing finding %q; got %d findings: %+v", want, len(res.Findings), res.Findings)
		}
	}
	if res.Stats.TableItems != 16 {
		t.Errorf("table items: %d", res.Stats.TableItems)
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Description, "Армирование") || strings.Contains(f.Description, "Арматура") {
			t.Errorf("the reinforcement is consistent (3,2 = 2,0 + 1,2): %s", f.Description)
		}
	}
}
