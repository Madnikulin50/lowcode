package cmp

import (
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

func TestCheckStatusChangeNeedsAReasonForResolvedAndRework(t *testing.T) {
	for _, to := range []string{RemarkResolved, RemarkRework} {
		if err := CheckStatusChange(RemarkAnswerRev, to, "  "); err == nil {
			t.Errorf("→ %s without a comment must fail", to)
		}
		if err := CheckStatusChange(RemarkAnswerRev, to, "сверено с листом 12"); err != nil {
			t.Errorf("→ %s with a comment: %v", to, err)
		}
	}
	if err := CheckStatusChange(RemarkAwaiting, RemarkBeingFix, ""); err != nil {
		t.Errorf("other changes need no reason: %v", err)
	}
	if err := CheckStatusChange(RemarkResolved, RemarkResolved, ""); err != nil {
		t.Errorf("no change, no reason: %v", err)
	}
	if err := CheckStatusChange(RemarkAwaiting, "bogus", "x"); err == nil {
		t.Error("unknown status")
	}
}

func TestParseEstimate(t *testing.T) {
	for in, want := range map[string]string{"": "", " 2,5 ": "2.5", "8": "8", "0": "0"} {
		if got, err := ParseEstimate(in); err != nil || got != want {
			t.Errorf("ParseEstimate(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"много", "-1", "100000"} {
		if _, err := ParseEstimate(bad); err == nil {
			t.Errorf("ParseEstimate(%q) must fail", bad)
		}
	}
}

func TestAutoHandover(t *testing.T) {
	cases := []struct {
		status, answer, oldV, newV string
		want                       string
		ok                         bool
	}{
		{RemarkBeingFix, "Исправлено", VerdictStillDiffers, VerdictResolved, RemarkAnswerRev, true},
		{RemarkAwaiting, "Исправлено", VerdictNoMatch, VerdictResolved, RemarkAnswerRev, true},
		{RemarkRework, "Исправлено повторно", VerdictStillDiffers, VerdictResolved, RemarkAnswerRev, true},
		{RemarkBeingFix, "", VerdictStillDiffers, VerdictResolved, RemarkBeingFix, false},           // the designer has not answered
		{RemarkBeingFix, "Исправлено", VerdictResolved, VerdictResolved, RemarkBeingFix, false},     // not new: the specialist may have moved it back
		{RemarkBeingFix, "Исправлено", VerdictNoMatch, VerdictStillDiffers, RemarkBeingFix, false},  // the difference is still there
		{RemarkResolved, "Исправлено", VerdictStillDiffers, VerdictResolved, RemarkResolved, false}, // already closed
		{RemarkAnswerRev, "Исправлено", VerdictStillDiffers, VerdictResolved, RemarkAnswerRev, false},
	}
	for i, c := range cases {
		got, ok := AutoHandover(c.status, c.answer, c.oldV, c.newV)
		if got != c.want || ok != c.ok {
			t.Errorf("case %d: got %q/%v, want %q/%v", i, got, ok, c.want, c.ok)
		}
	}
}

func TestReworkSurvivesAnUnchangedExport(t *testing.T) {
	if StatusCode("Возвращено на доработку") != RemarkRework {
		t.Skip("rework is a registry status; an export is not expected to carry it")
	}
	r := row()
	stored := &StoredRemark{Status: RemarkRework, SourceStatus: r.SourceStatus, Answer: r.Answer, Iterations: 1}
	if rec := Reconcile(stored, r, t0); rec.Status != RemarkRework {
		t.Errorf("an unchanged export keeps the status set by hand: %+v", rec)
	}
}

func TestSnapshotsCountAndKeepOnePerDay(t *testing.T) {
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "Объект"))
	mk := func(status, due, archived string) string {
		v := idcheck.V("object", obj, "status", status)
		if due != "" {
			v["due_date"] = []string{due}
		}
		if archived != "" {
			v["archived"] = []string{archived}
		}
		return f.put(ModExpertise, v)
	}
	mk(RemarkAwaiting, "2024-01-01", "") // overdue
	mk(RemarkBeingFix, "2999-01-01", "")
	mk(RemarkRework, "", "")
	mk(RemarkResolved, "2024-01-01", "") // resolved: never overdue
	mk(RemarkAwaiting, "", "да")         // archived: not counted
	day1 := time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)
	s, err := RecordSnapshot(f, "ns", obj, day1)
	if err != nil {
		t.Fatal(err)
	}
	if s.Total != 4 || s.Overdue != 1 || s.ByStatus[RemarkAwaiting] != 1 || s.ByStatus[RemarkRework] != 1 || s.ByStatus[RemarkResolved] != 1 {
		t.Errorf("snapshot: %+v", s)
	}
	// the same day again replaces, the next day adds
	mk(RemarkAnswerRev, "", "")
	if _, err := RecordSnapshot(f, "ns", obj, day1.Add(3*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := RecordSnapshot(f, "ns", obj, day1.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	dyn, err := Dynamics(f, "ns", obj)
	if err != nil || len(dyn) != 2 {
		t.Fatalf("one snapshot per day: %+v %v", dyn, err)
	}
	if dyn[0].Date != "2026-10-21" || dyn[0].Total != 5 || dyn[1].Date != "2026-10-22" || dyn[0].ByStatus[RemarkAnswerRev] != 1 {
		t.Errorf("dynamics: %+v", dyn)
	}
}

func TestWorkChangesAreHistoryEntries(t *testing.T) {
	s := func(v string) *string { return &v }
	got := WorkChanges("", "", "", s("Петров"), s("2026-11-01T00:00:00Z"), s(" проверить лист 5 "))
	if len(got) != 3 || got[0].Kind != HistAssignee || got[0].Source != "назначен: Петров" || got[1].Source != "срок: 2026-11-01" || got[2].Note != "проверить лист 5" {
		t.Fatalf("%+v", got)
	}
	if got := WorkChanges("Петров", "2026-11-01", "заметка", s("Петров"), s("2026-11-01T10:00:00Z"), s("заметка ")); len(got) != 0 {
		t.Errorf("nothing changed: %+v", got)
	}
	if got := WorkChanges("Петров", "2026-11-01", "", s(""), s(""), nil); len(got) != 2 || got[0].Source != "назначение снято" || got[0].Note != "был: Петров" || got[1].Source != "срок снят" {
		t.Errorf("taken back: %+v", got)
	}
	if got := WorkChanges("Петров", "", "", nil, nil, nil); len(got) != 0 {
		t.Errorf("fields not sent make no entries: %+v", got)
	}
}
