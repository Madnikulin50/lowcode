package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// --- list of remarks made from a comparison ------------------------------------

// issueInputFor reads a comparison, its findings and documents.
func issueInputFor(ns, recordID string) (cmp.IssueInput, *comparisonRecord, error) {
	rec, err := compose.Comparison(ns, recordID)
	if err != nil {
		return cmp.IssueInput{}, nil, err
	}
	discs, err := compose.Discrepancies(ns, recordID)
	if err != nil {
		return cmp.IssueInput{}, nil, err
	}
	docs, err := compose.Documents(ns, recordID)
	if err != nil {
		return cmp.IssueInput{}, nil, err
	}
	in := cmp.IssueInput{ComparisonID: recordID, ComparisonTitle: rec.Title, DocName: map[string]string{}, Mode: modeOf(rec.Kind)}
	for _, d := range docs {
		in.DocName[d.ID] = d.Name
	}
	for _, d := range discs {
		in.Findings = append(in.Findings, cmp.FindingIn{
			ID: d.ID, Kind: d.Kind, Severity: d.Severity, Description: d.Description, Anchor: d.Anchor, Status: firstNonEmpty(d.Status, cmp.StatusNew),
			IterState: d.IterState, DocRef: d.DocRef, DocTarget: d.DocTarget, PageRef: d.PageRef, PageTarget: d.PageTarget, PageNumber: d.PageNumber, BoxRef: d.BboxRef, BoxTarget: d.BboxTarget, Evidence: d.Evidence,
		})
	}
	return in, rec, nil
}

// numbered gives the remarks their registry numbers where they have them,
// plain 1, 2, 3… otherwise.
func numbered(ns, recordID string, list []cmp.Issued) {
	reg, _ := compose.IssuedNumbers(ns, recordID)
	for i := range list {
		if n := reg[list[i].GroupKey]; n != "" {
			list[i].Number = n
		} else {
			list[i].Number = fmt.Sprint(i + 1)
		}
	}
}

type issueOut struct {
	ComparisonID string       `json:"comparisonID"`
	Title        string       `json:"title"`
	ObjectID     string       `json:"objectID"`
	ObjectName   string       `json:"objectName"`
	Active       []cmp.Issued `json:"active"`
	Fixed        []cmp.Issued `json:"fixed"`
	Excluded     int          `json:"excluded"`
	Confirmed    int          `json:"confirmed"`
}

func handleIssuePreview(w http.ResponseWriter, r *http.Request) {
	ns, recordID := nsFromQuery(r), r.URL.Query().Get("recordID")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	in, rec, err := issueInputFor(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	res := cmp.BuildIssue(in)
	numbered(ns, recordID, res.Active)
	numbered(ns, recordID, res.Fixed)
	out := issueOut{ComparisonID: recordID, Title: rec.Title, ObjectID: rec.ObjectID, ObjectName: compose.ObjectName(ns, rec.ObjectID),
		Active: nonNilIssued(res.Active), Fixed: nonNilIssued(res.Fixed), Excluded: res.Excluded}
	for _, f := range in.Findings {
		if f.Status == cmp.StatusConfirmed {
			out.Confirmed++
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func nonNilIssued(l []cmp.Issued) []cmp.Issued {
	if l == nil {
		return []cmp.Issued{}
	}
	return l
}

func handleIssueRegister(w http.ResponseWriter, r *http.Request) {
	ns, recordID := nsFromQuery(r), r.URL.Query().Get("recordID")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	in, _, err := issueInputFor(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	res := cmp.BuildIssue(in)
	if len(res.Active) == 0 && len(res.Fixed) == 0 {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("нет подтверждённых находок: подтвердите расхождения в списке, чтобы сформировать замечания"))
		return
	}
	out, err := compose.RegisterIssued(ns, recordID, in, res)
	if err != nil {
		code := http.StatusBadGateway
		if strings.Contains(err.Error(), "не указан объект") {
			code = http.StatusUnprocessableEntity
		}
		writeErr(w, code, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func handleIssueExport(w http.ResponseWriter, r *http.Request) {
	ns, recordID := nsFromQuery(r), r.URL.Query().Get("recordID")
	format := r.URL.Query().Get("format")
	if ns == "" || recordID == "" || (format != "xlsx" && format != "docx") {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID, namespaceID and format (xlsx|docx) are required"))
		return
	}
	in, rec, err := issueInputFor(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	res := cmp.BuildIssue(in)
	if len(res.Active) == 0 {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("нет подтверждённых находок для списка замечаний"))
		return
	}
	numbered(ns, recordID, res.Active)
	meta := cmp.IssueMeta{ObjectName: compose.ObjectName(ns, rec.ObjectID), ComparisonTitle: rec.Title, Date: time.Now(), Mode: modeOf(rec.Kind)}
	var data []byte
	var ctype string
	if format == "xlsx" {
		data, err = cmp.IssuedXLSX(meta, res.Active)
		ctype = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else {
		data, err = cmp.IssuedDOCX(meta, res.Active)
		ctype = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	name := "Замечания " + rec.Title + " " + time.Now().Format("02.01.2006") + "." + format
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}
