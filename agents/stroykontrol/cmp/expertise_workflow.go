package cmp

import (
	"fmt"
	"strings"
	"time"
)

// Change kinds recorded in the import journal.
const (
	ChangeNew       = "new"       // a remark first seen in this export
	ChangeAnswer    = "answer"    // the designer's answer changed
	ChangeIteration = "iteration" // the remark came back for another iteration
	ChangeStatus    = "status"    // the expertise changed the status
	ChangeReopened  = "reopened"  // resolved earlier, open again
	ChangeArchived  = "archived"  // no longer in the export
	ChangeRestored  = "restored"  // in the export again
)

// StoredRemark is what the importer needs to know of a remark already in the
// registry.
type StoredRemark struct {
	Status       string // the working status (may have been set by hand)
	SourceStatus string // the status text of the last export
	Answer       string
	RejectReason string
	Iterations   int
	Archived     bool
}

// Snapshot is one state of a remark at some moment: the history of its
// iterations.
type Snapshot struct {
	Iteration    int
	Answer       string
	RejectReason string
	Status       string // working status code, "" when unknown
	SourceStatus string
	At           time.Time
	Source       string
}

// Change is one line of the journal of an import.
type Change struct {
	Key    string
	Kind   string
	Detail string
}

// Reconciled is what an import does with one remark.
type Reconciled struct {
	Status      string // the working status after the import
	History     []Snapshot
	Changes     []Change
	Unchanged   bool // nothing but the registry's own data changed
	StatusReset bool // the import overwrote a manual status
}

func statusLabel(code string) string {
	if l, ok := RemarkStatusLabels[code]; ok {
		return l
	}
	return "—"
}

func orAwaiting(code string) string {
	if code == "" {
		return RemarkAwaiting
	}
	return code
}

// Reconcile decides what an export row does to the registry:
//   - a new remark takes the export's status; its history starts with the
//     previous iteration (the export carries the previous answer and the
//     reason it was rejected) and the current state;
//   - a known remark keeps its working status unless the expertise changed
//     its own status since the last export: the decision of the expertise is
//     newer than a hand-made one, and is applied and logged;
//   - a new iteration, a changed answer or status add a snapshot to the history.
func Reconcile(stored *StoredRemark, row ExportRow, now time.Time) Reconciled {
	at := row.ChangedAt
	if at.IsZero() {
		at = now
	}
	src := "выгрузка " + row.Source
	snap := func() Snapshot {
		return Snapshot{Iteration: max(row.Iterations, 1), Answer: row.Answer, RejectReason: row.RejectReason, Status: row.Status, SourceStatus: row.SourceStatus, At: at, Source: src}
	}
	if stored == nil {
		r := Reconciled{Status: orAwaiting(row.Status)}
		if row.PrevAnswer != "" || row.PrevRejectReason != "" || row.Iterations >= 2 {
			r.History = append(r.History, Snapshot{
				Iteration: max(row.Iterations-1, 1), Answer: row.PrevAnswer, RejectReason: row.PrevRejectReason,
				SourceStatus: "предыдущая итерация", At: at, Source: src,
			})
		}
		r.History = append(r.History, snap())
		r.Changes = []Change{{row.Key, ChangeNew, fmt.Sprintf("итерация %d, %s", max(row.Iterations, 1), statusLabel(row.Status))}}
		return r
	}

	r := Reconciled{Status: stored.Status}
	if r.Status == "" {
		r.Status = orAwaiting(row.Status)
	}
	if stored.Archived {
		r.Changes = append(r.Changes, Change{row.Key, ChangeRestored, "замечание снова в выгрузке"})
	}
	newIter := row.Iterations > stored.Iterations && stored.Iterations > 0
	answer := norm(row.Answer) != norm(stored.Answer)
	reject := norm(row.RejectReason) != norm(stored.RejectReason)
	srcChanged := norm(row.SourceStatus) != norm(stored.SourceStatus)

	if newIter {
		d := fmt.Sprintf("итерация %d → %d", stored.Iterations, row.Iterations)
		if stored.Status == RemarkResolved || stored.SourceStatus == RemarkStatusLabels[RemarkResolved] {
			r.Changes = append(r.Changes, Change{row.Key, ChangeReopened, d})
		} else {
			r.Changes = append(r.Changes, Change{row.Key, ChangeIteration, d})
		}
	}
	if answer && !newIter {
		r.Changes = append(r.Changes, Change{row.Key, ChangeAnswer, "ответ проектировщика изменён"})
	}
	if srcChanged {
		// the expertise's own decision overrides a hand-made status
		if code := row.Status; code != "" && code != r.Status {
			r.Changes = append(r.Changes, Change{row.Key, ChangeStatus, fmt.Sprintf("%s → %s (по выгрузке)", statusLabel(r.Status), statusLabel(code))})
			r.StatusReset = stored.Status != "" && StatusCode(stored.SourceStatus) != "" && stored.Status != StatusCode(stored.SourceStatus)
			r.Status = code
		}
	}
	if newIter || answer || reject || srcChanged {
		r.History = append(r.History, snap())
	}
	r.Unchanged = len(r.Changes) == 0 && len(r.History) == 0
	return r
}

// JournalText renders the changes of an import for the journal.
func JournalText(ch []Change, limit int) string {
	label := map[string]string{
		ChangeNew: "новое", ChangeAnswer: "ответ", ChangeIteration: "итерация", ChangeStatus: "статус",
		ChangeReopened: "возвращено", ChangeArchived: "нет в выгрузке", ChangeRestored: "снова в выгрузке",
	}
	var b strings.Builder
	for i, c := range ch {
		if limit > 0 && i >= limit {
			fmt.Fprintf(&b, "… и ещё %d\n", len(ch)-limit)
			break
		}
		fmt.Fprintf(&b, "%s — %s: %s\n", c.Key, label[c.Kind], c.Detail)
	}
	return strings.TrimSpace(b.String())
}
