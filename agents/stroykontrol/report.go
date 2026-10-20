package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// reportFor reads a comparison and arranges its report.
func reportFor(ns, recordID string) (cmp.Report, *comparisonRecord, error) {
	in, rec, err := issueInputFor(ns, recordID)
	if err != nil {
		return cmp.Report{}, nil, err
	}
	docs, err := compose.Documents(ns, recordID)
	if err != nil {
		return cmp.Report{}, nil, err
	}
	refName, tgtName := cmp.SideNames(modeOf(rec.Kind))
	ri := cmp.ReportInput{
		RefName: refName, TargetName: tgtName,
		Title: rec.Title, ObjectName: compose.ObjectName(ns, rec.ObjectID), SetsLabel: rec.SetsLabel, Status: rec.Status,
		Similarity: rec.SimilarityPercent, Comment: rec.Comment, Date: time.Now(), Findings: in.Findings, RemarkNo: map[string]string{},
	}
	for _, d := range docs {
		ri.Documents = append(ri.Documents, cmp.ReportDoc{ID: d.ID, Side: d.Side, Name: d.Name, VolumeCode: d.VolumeCode, VolumeTitle: d.VolumeTitle, Pages: d.Pages})
	}
	// numbers of the remarks the confirmed findings went into
	issue := cmp.BuildIssue(in)
	numbered(ns, recordID, issue.Active)
	numbered(ns, recordID, issue.Fixed)
	for _, list := range [][]cmp.Issued{issue.Active, issue.Fixed} {
		for _, is := range list {
			for _, id := range is.FindingIDs {
				ri.RemarkNo[id] = is.Number
			}
		}
	}
	if discs, err := compose.Discrepancies(ns, recordID); err == nil {
		if it := iterationSummary(rec, discs); it != nil {
			ri.Iteration = &cmp.ReportIteration{No: it.No, ParentTitle: it.ParentTitle, Resolved: it.Resolved, Persisting: it.Persisting, New: it.New}
		}
	}
	if p, err := compose.Pairing(ns, recordID); err == nil && p.State != nil {
		ri.Params = p.State.Params
	}
	return cmp.BuildReport(ri), rec, nil
}

func handleReportExport(w http.ResponseWriter, r *http.Request) {
	ns, recordID := nsFromQuery(r), r.URL.Query().Get("recordID")
	format := r.URL.Query().Get("format")
	if ns == "" || recordID == "" || (format != "xlsx" && format != "docx") {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID, namespaceID and format (xlsx|docx) are required"))
		return
	}
	rep, rec, err := reportFor(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	var data []byte
	var ctype string
	if format == "xlsx" {
		data, err = cmp.ReportXLSX(rep)
		ctype = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	} else {
		data, err = cmp.ReportDOCX(rep)
		ctype = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	name := "Отчёт " + rec.Title + " " + time.Now().Format("02.01.2006") + "." + format
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+url.PathEscape(name))
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}
