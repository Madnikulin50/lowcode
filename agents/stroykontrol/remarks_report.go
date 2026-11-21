package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// remarkFilter selects the remarks of the registry the way the workbench filters do, so that an
// export holds what the specialist sees. The period selects the remarks that had any activity in it:
// changed in the expertise, or an entry of their history.
type remarkFilter struct {
	Status                            map[string]bool
	Crit, Expert, Discipline, Section string
	Assignee, Verdict, Outcome, Q     string
	MinIter                           int
	Overdue, Claim, Archived          bool
	Materials                         bool   // only the remarks whose declared materials were not found
	From, To                          string // YYYY-MM-DD, inclusive; "" = open
}

func filterFromQuery(q url.Values) (remarkFilter, error) {
	f := remarkFilter{
		Crit: q.Get("crit"), Expert: q.Get("expert"), Discipline: q.Get("discipline"), Section: q.Get("section"),
		Assignee: q.Get("assignee"), Verdict: q.Get("verdict"), Outcome: q.Get("outcome"), Q: strings.ToLower(strings.TrimSpace(q.Get("q"))),
		Overdue: q.Get("overdue") == "1", Claim: q.Get("claim") == "1", Materials: q.Get("materials") == "1", Archived: q.Get("archived") == "1",
		From: q.Get("from"), To: q.Get("to"),
	}
	if s := q.Get("status"); s != "" {
		f.Status = map[string]bool{}
		for _, c := range strings.Split(s, ",") {
			c = strings.TrimSpace(c)
			if _, ok := cmp.RemarkStatusLabels[c]; !ok {
				return f, fmt.Errorf("unknown status %q", c)
			}
			f.Status[c] = true
		}
	}
	if s := q.Get("iter"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 {
			return f, fmt.Errorf("iter must be a number")
		}
		f.MinIter = n
	}
	for _, d := range []string{f.From, f.To} {
		if d != "" && !validDate(d) {
			return f, fmt.Errorf("from and to must be YYYY-MM-DD")
		}
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		return f, fmt.Errorf("from is after to")
	}
	return f, nil
}

func (f remarkFilter) describe() []string {
	var out []string
	if len(f.Status) > 0 {
		var names []string
		for _, c := range cmp.RemarkStatusOrder {
			if f.Status[c] {
				names = append(names, cmp.RemarkStatusLabels[c])
			}
		}
		out = append(out, "статус: "+strings.Join(names, ", "))
	}
	add := func(label, v string) {
		if v != "" {
			out = append(out, label+": "+v)
		}
	}
	add("критичность", map[string]string{"critical": "критичное", "noncritical": "некритичное"}[f.Crit])
	add("эксперт", f.Expert)
	add("НД", f.Discipline)
	add("раздел", f.Section)
	add("ответственный", f.Assignee)
	add("проверка по документам", cmp.VerdictLabels[f.Verdict])
	add("результат проверки", cmp.OutcomeLabels[f.Outcome])
	add("поиск", f.Q)
	if f.MinIter > 0 {
		out = append(out, fmt.Sprintf("итерация не меньше %d", f.MinIter))
	}
	if f.Overdue {
		out = append(out, "только просроченные")
	}
	if f.Claim {
		out = append(out, "ответ ≠ документы")
	}
	if f.Materials {
		out = append(out, "заявленные материалы не найдены")
	}
	if f.Archived {
		out = append(out, "включая пропавшие из выгрузки")
	}
	if f.From != "" || f.To != "" {
		out = append(out, "период: "+firstNonEmpty(f.From, "…")+" — "+firstNonEmpty(f.To, "…"))
	}
	return out
}

func (r remarkRecord) overdue(today string) bool {
	return r.DueDate != "" && r.Status != cmp.RemarkResolved && len(r.DueDate) >= 10 && r.DueDate[:10] < today
}

// match applies the filter to a remark; history is needed only for the period.
func (f remarkFilter) match(r remarkRecord, today string, history []remarkHistoryRecord) bool {
	switch {
	case r.Archived && !f.Archived,
		len(f.Status) > 0 && !f.Status[r.Status],
		f.Crit != "" && r.Criticality != f.Crit,
		f.Expert != "" && r.Expert != f.Expert,
		f.Discipline != "" && r.Discipline != f.Discipline,
		f.Section != "" && r.Section != f.Section,
		f.Assignee == "—" && r.Assignee != "",
		f.Assignee != "" && f.Assignee != "—" && r.Assignee != f.Assignee,
		f.MinIter > 0 && r.Iterations < f.MinIter,
		f.Verdict != "" && r.Verdict != f.Verdict,
		f.Outcome != "" && firstNonEmpty(r.OutcomeFinal, r.Outcome) != f.Outcome,
		f.Overdue && !r.overdue(today),
		f.Claim && !r.ClaimMismatch:
		return false
	case f.Materials && r.MaterialsState != cmp.MaterialsMissing:
		return false
	}
	if f.Q != "" && !strings.Contains(strings.ToLower(r.Text+" "+r.Answer+" "+r.Number+" "+r.Refs), f.Q) {
		return false
	}
	if f.From != "" || f.To != "" {
		in := func(at string) bool {
			if len(at) < 10 {
				return false
			}
			d := at[:10]
			return (f.From == "" || d >= f.From) && (f.To == "" || d <= f.To)
		}
		active := in(r.ChangedAt)
		for _, h := range history {
			active = active || in(h.At)
		}
		if !active {
			return false
		}
	}
	return true
}

func placesText(links []cmp.RemarkLocation, docName func(string) string) string {
	var parts []string
	for _, l := range links {
		switch l.Kind {
		case cmp.LocPlace:
			s := firstNonEmpty(l.Doc, docName(l.DocID))
			if l.Page > 0 {
				s += fmt.Sprintf(", стр. %d", l.Page)
			}
			if l.Sheet > 0 {
				s += fmt.Sprintf(" (лист %d)", l.Sheet)
			}
			parts = append(parts, s)
		case cmp.LocFinding:
			parts = append(parts, "находка: "+l.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// registryReportFor selects the remarks of the object and gathers their history.
func registryReportFor(ns, objectID string, f remarkFilter) (cmp.RegistryReport, error) {
	board, err := compose.Board(ns, objectID)
	if err != nil {
		return cmp.RegistryReport{}, err
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	var allTasks []cmp.Task
	tasksOf := map[string]cmp.TaskCount{}
	if taskB != nil {
		if ts, err := cmp.ListTasks(taskB, ns, cmp.TaskFilter{Object: objectID}, now); err == nil {
			allTasks, tasksOf = ts, cmp.CountTasks(ts)
		}
	}
	rep := cmp.RegistryReport{ObjectName: board.ObjectName, Date: now, Filters: f.describe(), Dynamics: board.Dynamics}
	for _, r := range board.Remarks {
		needHistory := f.From != "" || f.To != ""
		var hist []remarkHistoryRecord
		if needHistory {
			if hist, err = compose.RemarkHistory(ns, objectID, r.ID); err != nil {
				return rep, err
			}
		}
		if !f.match(r, today, hist) {
			continue
		}
		if hist == nil {
			if hist, err = compose.RemarkHistory(ns, objectID, r.ID); err != nil {
				return rep, err
			}
		}
		rep.Remarks = append(rep.Remarks, cmp.RegistryRemark{
			ID: r.ID, TaskCount: tasksOf[r.ID],
			Number: firstNonEmpty(r.Number, r.Key), Origin: r.Origin, Criticality: r.Criticality, Section: r.Section, Discipline: r.Discipline, Expert: r.Expert,
			Text: r.Text, Refs: r.Refs, Answer: r.Answer, RejectReason: r.RejectReason, SourceStatus: r.SourceStatus, Status: r.Status,
			Verdict: r.Verdict, VerdictNote: r.MatchNote, AIVerdict: r.AIVerdict, AINote: r.AINote,
			Outcome: r.Outcome, OutcomeFinal: r.OutcomeFinal, OutcomeComment: r.OutcomeComment, OutcomeBy: r.OutcomeBy,
			Assignee: r.Assignee, DueDate: dateOnly(r.DueDate), Note: r.Note, ChangedAt: r.ChangedAt, EstimateH: r.EstimateH,
			Iterations: r.Iterations, Overdue: r.overdue(today),
			Places:       placesText(r.Links, func(id string) string { return board.Documents[id] }),
			Requirements: r.Requirements,
			Materials:    r.Materials,
		})
		for _, h := range hist {
			rep.History = append(rep.History, cmp.RegistryHistory{
				Number: firstNonEmpty(r.Number, r.Key), At: h.At, Kind: h.Kind, Source: h.Source, Author: h.Author, Status: h.Status,
				Iteration: h.Iteration, Answer: h.Answer, RejectReason: h.RejectReason, Note: h.Note,
			})
		}
	}
	// the tasks of the exported remarks and the tasks of the object that belong to no remark
	chosen := map[string]bool{}
	for _, x := range rep.Remarks {
		chosen[x.ID] = true
	}
	for _, t := range allTasks {
		if t.Remark == "" || chosen[t.Remark] {
			rep.Tasks = append(rep.Tasks, t)
		}
	}
	return rep, nil
}

func dateOnly(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// handleRemarksExport — GET /api/remarks/export?objectID=&namespaceID=&format=xlsx plus the filters of the workbench.
func handleRemarksExport(w http.ResponseWriter, r *http.Request) {
	ns, obj, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if fm := r.URL.Query().Get("format"); fm != "" && fm != "xlsx" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("format must be xlsx"))
		return
	}
	f, err := filterFromQuery(r.URL.Query())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	rep, err := registryReportFor(ns, obj, f)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	data, err := cmp.RegistryXLSX(rep)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	name := "Реестр замечаний " + rep.ObjectName + " " + rep.Date.Format("02.01.2006") + ".xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}

// handleReportIndex — GET /api/reports/index?namespaceID=: the objects and comparisons the «Отчёты» page offers.
func handleReportIndex(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	if ns == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("namespaceID is required"))
		return
	}
	out, err := compose.ReportIndex(ns)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	if out == nil {
		out = []reportObject{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"objects": out})
}
