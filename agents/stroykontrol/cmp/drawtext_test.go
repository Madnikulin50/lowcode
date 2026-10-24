package cmp

import "testing"

func labelsPage(ws ...Word) Page {
	return Page{Number: 1, W: 1000, H: 700, HasText: true, Drawing: true, Labels: ws}
}

func TestSheetTextFindings(t *testing.T) {
	w := func(s string, x, y float64) Word { return Word{s, Box{x, y, x + 0.05, y + 0.01}} }
	ref := &Document{ID: "a", Pages: []Page{labelsPage(w("Бетон", 0.1, 0.1), w("B30", 0.16, 0.1), w("Ø16", 0.5, 0.5), w("Защитный", 0.2, 0.3), w("слой", 0.26, 0.3), w("40", 0.32, 0.3), w("УДАЛЕНО", 0.7, 0.2))}}
	tgt := &Document{ID: "b", Pages: []Page{labelsPage(w("Бетон", 0.1, 0.1), w("B35", 0.16, 0.1), w("Ø16", 0.5, 0.5), w("Защитный", 0.2, 0.3), w("слой", 0.26, 0.3), w("40", 0.32, 0.3), w("НОВОЕ", 0.8, 0.8))}}
	var st Stats
	fs := sheetTextFindings(sheet{ref, 1}, sheet{tgt, 1}, "vol", "", nil, &st)
	var changed, missing, added int
	for _, f := range fs {
		switch {
		case f.BoxRef != nil && f.BoxTarget != nil:
			changed++
		case f.BoxRef != nil:
			missing++
		case f.BoxTarget != nil:
			added++
		}
	}
	if changed != 1 || missing != 1 || added != 1 {
		t.Fatalf("changed=%d missing=%d added=%d: %+v", changed, missing, added, fs)
	}
	if st.simDen == 0 {
		t.Fatal("similarity not counted")
	}
}
