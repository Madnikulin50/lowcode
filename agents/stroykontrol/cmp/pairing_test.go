package cmp

import (
	"strings"
	"testing"
)

func pairingDocs() (*Document, *Document) {
	ref := textDoc("r", SideRef,
		[]string{"Опоры сооружения А: бетон класса В25", "Глубина заложения опор А 3,0 м"},
		[]string{"Пролётное строение А: сталь С345", "Длина пролёта А 24 м"},
		[]string{"Опоры сооружения Б: бетон класса В30", "Глубина заложения опор Б 2,6 м"},
		[]string{"Пролётное строение Б: сталь С390", "Длина пролёта Б 18 м"},
	)
	tgt := textDoc("t", SideTarget,
		[]string{"Опоры сооружения Б: бетон класса В30", "Глубина заложения опор Б 2,4 м"},
		[]string{"Пролётное строение Б: сталь С390", "Длина пролёта Б 18 м"},
	)
	return ref, tgt
}

func TestManualPairComparesOnlyTheChosenFragment(t *testing.T) {
	ref, tgt := pairingDocs()
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{Pairing: []ManualPair{
		{Ref: Slice{Doc: "r", From: 3, To: 4}, Targets: []Slice{{Doc: "t"}}},
	}})
	if len(res.Pairs) != 1 || !res.Pairs[0].Manual || res.Pairs[0].Ref.From != 3 || res.Pairs[0].Ref.To != 4 {
		t.Fatalf("pairs: %+v", res.Pairs)
	}
	// the difference is in Б, found; А is not compared and is not reported as missing from the target
	var depth bool
	for _, f := range res.Findings {
		if strings.Contains(f.Description, "2,6") && strings.Contains(f.Description, "2,4") {
			depth = true
		}
		if strings.Contains(f.Description, "сооружения А") && f.Kind != KindNotRelease {
			t.Fatalf("the part of the reference that was not paired must not read as removed: %q", f.Description)
		}
	}
	if !depth {
		t.Fatalf("expected the Б depth finding, got %+v", res.Findings)
	}
	// pages 1–2 of the reference stay without a pair and are listed as such
	if len(res.RefOnly) != 1 || res.RefOnly[0].From != 1 || res.RefOnly[0].To != 2 {
		t.Fatalf("refOnly: %+v", res.RefOnly)
	}
}

func TestManualPairKeepsAutoMatchingOfTheRest(t *testing.T) {
	ref, tgt := pairingDocs()
	other := textDoc("o", SideTarget, []string{"Опоры сооружения А: бетон класса В25", "Глубина заложения опор А 3,0 м"})
	res := Compare([]*Document{ref}, []*Document{tgt, other}, Options{Pairing: []ManualPair{
		{Ref: Slice{Doc: "r", From: 3, To: 4}, Targets: []Slice{{Doc: "t"}}},
	}})
	if len(res.Pairs) != 2 {
		t.Fatalf("the manual pair and the auto pair of А with «o»: %+v", res.Pairs)
	}
	if !res.Pairs[0].Manual || res.Pairs[1].Manual || res.Pairs[1].Ref.From != 1 || res.Pairs[1].Ref.To != 2 {
		t.Fatalf("%+v", res.Pairs)
	}
}

func TestManualPairIsReportedWhenItDoesNotFit(t *testing.T) {
	ref, tgt := pairingDocs()
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{Pairing: []ManualPair{
		{Ref: Slice{Doc: "r", From: 9, To: 12}, Targets: []Slice{{Doc: "t"}}},
		{Ref: Slice{Doc: "nope", From: 1, To: 2}, Targets: []Slice{{Doc: "t"}}},
		{Ref: Slice{Doc: "t", From: 1, To: 1}, Targets: []Slice{{Doc: "t"}}}, // a target file on the reference side
	}})
	notes := strings.Join(res.Notes, "\n")
	for _, want := range []string{"вне документа", "не найден среди документов", "не на той стороне"} {
		if !strings.Contains(notes, want) {
			t.Fatalf("notes lack %q:\n%s", want, notes)
		}
	}
	if len(res.Pairs) != 1 || res.Pairs[0].Manual { // falls back to the automatic pair of the two files
		t.Fatalf("pairs: %+v", res.Pairs)
	}
}

func TestPairingJSONAndState(t *testing.T) {
	p, err := ParsePairing(`[{"ref":{"doc":"r","from":3,"to":4},"targets":[{"doc":"t","from":1,"to":0}]}]`)
	if err != nil || len(p) != 1 || p[0].Targets[0].To != 0 {
		t.Fatalf("%+v %v", p, err)
	}
	if got, _ := ParsePairing("  "); got != nil {
		t.Fatal("empty is no pairing")
	}
	if _, err := ParsePairing("{oops"); err == nil {
		t.Fatal("bad JSON must fail")
	}
	ref, tgt := pairingDocs()
	res := Compare([]*Document{ref}, []*Document{tgt}, Options{Pairing: p})
	st := res.State()
	if len(st.Pairs) != 1 || !st.Pairs[0].Manual || len(st.Volumes["r"]) == 0 || len(st.RefOnly) != 1 {
		t.Fatalf("state: %+v", st)
	}
	if !strings.Contains(res.StateJSON(), `"manual":true`) {
		t.Fatalf("json: %s", res.StateJSON())
	}
}

func TestSubtractRanges(t *testing.T) {
	got := subtract(1, 10, []pageRange{{3, 4}, {8, 12}})
	if len(got) != 2 || got[0] != (pageRange{1, 2}) || got[1] != (pageRange{5, 7}) {
		t.Fatalf("%v", got)
	}
	if len(subtract(2, 3, []pageRange{{1, 5}})) != 0 {
		t.Fatal("fully covered")
	}
}
