package cmp

import (
	"fmt"
	"strconv"
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
	Kind         string // one of the Hist* kinds; "" = import
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

// CheckStatusChange validates a hand-made change of the working status. Confirming a
// remark as resolved or sending it back to the designer must be explained (ТЗ 3.7.4): the
// comment is required and goes to the history. Other changes need none.
func CheckStatusChange(from, to, comment string) error {
	if _, ok := RemarkStatusLabels[to]; !ok {
		return fmt.Errorf("unknown status %q", to)
	}
	if to == from {
		return nil
	}
	if (to == RemarkResolved || to == RemarkRework) && strings.TrimSpace(comment) == "" {
		return fmt.Errorf("statusComment is required to set the status «%s»", RemarkStatusLabels[to])
	}
	return nil
}

// ParseEstimate reads the labour estimate of a remark in hours ("" clears it).
func ParseEstimate(s string) (string, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	if s == "" {
		return "", nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 || f > 10000 {
		return "", fmt.Errorf("estimateH must be a number of hours from 0 to 10000")
	}
	return strconv.FormatFloat(f, 'f', -1, 64), nil
}

// AutoHandover decides whether a remark goes to the quality control by itself (ТЗ 3.7.3): the
// designer has answered, a new edition was loaded and compared, and the difference the remark is
// about is gone (the check against the documents turned «resolved» just now). The remark then
// waits for the specialist's review; the final «Устранено» is never set automatically.
func AutoHandover(status, answer, oldVerdict, newVerdict string) (string, bool) {
	if newVerdict != VerdictResolved || oldVerdict == VerdictResolved || strings.TrimSpace(answer) == "" {
		return status, false
	}
	switch status {
	case RemarkAwaiting, RemarkBeingFix, RemarkRework:
		return RemarkAnswerRev, true
	}
	return status, false
}

// StatusSnapshot is the picture of an object's remarks at a day: how many are in each
// working status, how many are overdue. A series of them is the dynamics of the corrections (ТЗ 3.9.4).
type StatusSnapshot struct {
	Date     string         `json:"date"` // YYYY-MM-DD
	Total    int            `json:"total"`
	ByStatus map[string]int `json:"byStatus"`
	Overdue  int            `json:"overdue"`
}

// StatusRow is what a snapshot needs of one remark.
type StatusRow struct {
	Status  string
	DueDate string // YYYY-MM-DD…, may be empty
}

// CountStatuses builds the snapshot of the given (live, not archived) remarks at the day.
func CountStatuses(rows []StatusRow, day string) StatusSnapshot {
	s := StatusSnapshot{Date: day, ByStatus: map[string]int{}}
	for _, code := range RemarkStatusOrder {
		s.ByStatus[code] = 0
	}
	for _, r := range rows {
		code := r.Status
		if _, ok := RemarkStatusLabels[code]; !ok {
			code = RemarkAwaiting
		}
		s.ByStatus[code]++
		s.Total++
		if code != RemarkResolved && len(r.DueDate) >= 10 && r.DueDate[:10] < day {
			s.Overdue++
		}
	}
	return s
}

// Kinds of an entry of the history of a remark.
const (
	HistImport      = "import"       // a state read from an export of the expertise
	HistStatus      = "status"       // the working status changed by hand
	HistAssignee    = "assignee"     // the remark was assigned (or the assignment was taken back)
	HistDue         = "due"          // the due date was set or changed
	HistNote        = "note"         // the specialist's note
	HistDecision    = "decision"     // the decision on the result of the check
	HistAuto        = "auto"         // a change the agent made (handover to the control, closing by a comparison)
	HistComparison  = "comparison"   // the remark was made from a comparison
	HistRequirement = "requirements" // the list of the requirements was edited, or a decision was made on one
)

// RequirementChanges lists the history entries an edit of the list of requirements makes: one for the list
// itself when its texts changed, one for every decision on a requirement that is new or different.
func RequirementChanges(old, now RequirementSet) []WorkChange {
	var out []WorkChange
	same := len(old.Items) == len(now.Items)
	oldBy := map[string]Requirement{}
	for i, q := range old.Items {
		oldBy[strings.ToLower(q.Text)] = q
		if same && i < len(now.Items) && !strings.EqualFold(q.Text, now.Items[i].Text) {
			same = false
		}
	}
	if !same {
		out = append(out, WorkChange{HistRequirement, fmt.Sprintf("список требований изменён: %d → %d", len(old.Items), len(now.Items)), ""})
	}
	for _, q := range now.Items {
		o := oldBy[strings.ToLower(q.Text)]
		if q.OutcomeFinal != "" && (o.OutcomeFinal != q.OutcomeFinal || o.OutcomeComment != q.OutcomeComment) {
			out = append(out, WorkChange{HistRequirement, fmt.Sprintf("решение по требованию %d: %s", q.Idx, OutcomeLabels[q.OutcomeFinal]), q.OutcomeComment})
		}
	}
	return out
}

// WorkChange is one entry the specialist's own fields (assignee, due date, note) add to the history.
type WorkChange struct {
	Kind, Source, Note string
}

// WorkChanges lists what a hand-made update changed in the assignee, the due date and the note: each
// change is a history entry, so that the history tells who was given the remark, when it was due and
// what was said about it (ТЗ 3.6.3). Unchanged values and nil (not sent) make no entry.
func WorkChanges(oldAssignee, oldDue, oldNote string, assignee, due, note *string) []WorkChange {
	var out []WorkChange
	if assignee != nil && strings.TrimSpace(*assignee) != strings.TrimSpace(oldAssignee) {
		a := strings.TrimSpace(*assignee)
		if a == "" {
			out = append(out, WorkChange{HistAssignee, "назначение снято", "был: " + strings.TrimSpace(oldAssignee)})
		} else {
			out = append(out, WorkChange{HistAssignee, "назначен: " + a, ""})
		}
	}
	short := func(s string) string {
		if len(s) >= 10 {
			return s[:10]
		}
		return s
	}
	if due != nil && short(*due) != short(oldDue) {
		if *due == "" {
			out = append(out, WorkChange{HistDue, "срок снят", ""})
		} else {
			out = append(out, WorkChange{HistDue, "срок: " + short(*due), ""})
		}
	}
	if note != nil && strings.TrimSpace(*note) != strings.TrimSpace(oldNote) {
		out = append(out, WorkChange{HistNote, "заметка ПТО", strings.TrimSpace(*note)})
	}
	return out
}
