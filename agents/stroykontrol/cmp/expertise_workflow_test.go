package cmp

import (
	"strings"
	"testing"
	"time"
)

var t0 = time.Date(2024, 11, 7, 12, 0, 0, 0, time.UTC)

func row() ExportRow {
	return ExportRow{
		Key: "30.1", Number: "30.1", Text: "Расхождение по протяжённости экранов", Answer: "Принято", SourceStatus: "Требуется рассмотрение ответа",
		Status: RemarkAnswerRev, Iterations: 1, Source: "v1.xlsx", ChangedAt: time.Date(2024, 9, 30, 15, 17, 0, 0, time.UTC),
	}
}

func TestNewRemarkStartsHistoryWithThePreviousIteration(t *testing.T) {
	r := row()
	r.Iterations = 3
	r.PrevAnswer, r.PrevRejectReason = "Предыдущий ответ", "Не приведено в соответствие"
	rec := Reconcile(nil, r, t0)
	if rec.Status != RemarkAnswerRev || len(rec.Changes) != 1 || rec.Changes[0].Kind != ChangeNew {
		t.Fatalf("%+v", rec)
	}
	if len(rec.History) != 2 || rec.History[0].Iteration != 2 || rec.History[0].Answer != "Предыдущий ответ" || rec.History[0].RejectReason == "" || rec.History[1].Iteration != 3 {
		t.Errorf("history: %+v", rec.History)
	}
	// a first-iteration remark has just its current state
	one := Reconcile(nil, row(), t0)
	if len(one.History) != 1 || one.History[0].Iteration != 1 {
		t.Errorf("%+v", one.History)
	}
	// an unknown status in the export is awaiting an answer, never empty
	r.Status, r.SourceStatus = "", "что-то новое"
	if Reconcile(nil, r, t0).Status != RemarkAwaiting {
		t.Error("unknown status")
	}
}

func TestHandMadeStatusIsKeptUnlessTheExpertiseChangedItsOwn(t *testing.T) {
	r := row()
	stored := &StoredRemark{Status: RemarkResolved, SourceStatus: r.SourceStatus, Answer: r.Answer, Iterations: 1}
	// nothing changed in the export: the specialist's status stays
	rec := Reconcile(stored, r, t0)
	if rec.Status != RemarkResolved || !rec.Unchanged {
		t.Fatalf("a hand-made status survives an unchanged export: %+v", rec)
	}
	// the expertise moved the remark on: its decision wins, logged
	r.SourceStatus, r.Status = "Устранено", RemarkResolved
	stored2 := &StoredRemark{Status: RemarkBeingFix, SourceStatus: "Требуется рассмотрение ответа", Answer: r.Answer, Iterations: 1}
	rec = Reconcile(stored2, r, t0)
	if rec.Status != RemarkResolved || len(rec.Changes) != 1 || rec.Changes[0].Kind != ChangeStatus || !strings.Contains(rec.Changes[0].Detail, "Устраняется заявителем → Устранено") {
		t.Fatalf("%+v", rec)
	}
	if !rec.StatusReset {
		t.Error("the import replaced a status that had been set by hand: say so")
	}
	if len(rec.History) != 1 {
		t.Errorf("a status change is a history entry: %+v", rec.History)
	}
}

func TestNewIterationReopensAResolvedRemark(t *testing.T) {
	r := row()
	r.Iterations, r.SourceStatus, r.Status = 2, "Ответ не предоставлен", RemarkAwaiting
	r.Answer, r.PrevAnswer = "", "Принято"
	stored := &StoredRemark{Status: RemarkResolved, SourceStatus: "Устранено", Answer: "Принято", Iterations: 1}
	rec := Reconcile(stored, r, t0)
	if rec.Status != RemarkAwaiting {
		t.Errorf("the expertise sent it back: %+v", rec)
	}
	kinds := map[string]bool{}
	for _, c := range rec.Changes {
		kinds[c.Kind] = true
	}
	if !kinds[ChangeReopened] || !kinds[ChangeStatus] {
		t.Errorf("changes: %+v", rec.Changes)
	}
	if len(rec.History) != 1 || rec.History[0].Iteration != 2 {
		t.Errorf("history: %+v", rec.History)
	}
}

func TestAnswerChangeAndRestoredRemark(t *testing.T) {
	r := row()
	r.Answer = "Новый ответ"
	stored := &StoredRemark{Status: RemarkAnswerRev, SourceStatus: r.SourceStatus, Answer: "Принято", Iterations: 1, Archived: true}
	rec := Reconcile(stored, r, t0)
	kinds := map[string]bool{}
	for _, c := range rec.Changes {
		kinds[c.Kind] = true
	}
	if !kinds[ChangeAnswer] || !kinds[ChangeRestored] || len(rec.History) != 1 || rec.History[0].Answer != "Новый ответ" {
		t.Errorf("%+v", rec)
	}
	if got := JournalText(rec.Changes, 1); !strings.Contains(got, "30.1") || !strings.Contains(got, "и ещё 1") {
		t.Errorf("journal: %q", got)
	}
}

func TestParseVolumeRefs(t *testing.T) {
	got := ParseVolumeRefs("Том 4.1.5, 28/12-22-ГП-599 – ИЛО1.5;\r\nТом 6.1, 28/12-22-ГП-599 – ООС1.")
	if len(got) != 2 || got[0].Tom != "4.1.5" || got[0].Code != "ИЛО1.5" || got[1].Tom != "6.1" || got[1].Code != "ООС1" {
		t.Errorf("%+v", got)
	}
	got = ParseVolumeRefs("том 3.5.4.1, 28/12-22-ГП-599-ТКР5.4.1, раздел 2.3, лист 19 (стр. 25);\r\n\r\nтом 3.5.7.1, 28/12-22-ГП-599 – ТКР5.7.1, раздел 2.3, лист 14 (стр. 20)")
	if len(got) != 2 || got[0].Code != "ТКР5.4.1" || got[0].Sheet != 19 || got[0].Page != 25 || got[1].Code != "ТКР5.7.1" || got[1].Sheet != 14 {
		t.Errorf("%+v", got)
	}
	if got := ParseVolumeRefs("Том 3.1.2.7. 28/12-22-ГП-599-ТКР1.2.7; \r\n\r\nТом 4.1.4. 28/12-22-ГП-599-ИЛО1.4"); len(got) != 2 || got[0].Code != "ТКР1.2.7" || got[1].Code != "ИЛО1.4" {
		t.Errorf("%+v", got)
	}
	if got := ParseVolumeRefs("ЛСР № 02-01-05 водопропускные трубы"); len(got) != 0 {
		t.Errorf("an estimate reference is not a volume: %+v", got)
	}
}

func TestMatchVolumeByCode(t *testing.T) {
	ref := VolumeRef{Code: "ТКР5.1"}
	if !MatchVolume(ref, "100720-1/П-3-ТКР5-1", "") {
		t.Error("ТКР5.1 is the volume 100720-1/П-3-ТКР5-1")
	}
	if !MatchVolume(VolumeRef{Code: "КЖ1.13"}, "КЖ1.13-Расчеты", "") || !MatchVolume(VolumeRef{Code: "КЖ1.13"}, "03102022/РД-ИССО3.10-ОП-КЖ1.13-С", "") {
		t.Error("a code inside the шифр")
	}
	if MatchVolume(ref, "100720-1/П-3-ТКР5-10", "") || MatchVolume(ref, "2024-15-КЖ", "") || MatchVolume(VolumeRef{Code: "КЖ1"}, "КЖ1.13-Расчеты", "") {
		t.Error("a different volume matched")
	}
	if !MatchVolume(VolumeRef{Code: "ИЛО1.5"}, "", "Том 4.1.5 ИЛО1.5 Охрана") {
		t.Error("a volume named in the title")
	}
}
