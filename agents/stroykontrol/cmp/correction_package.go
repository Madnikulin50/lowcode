package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Пакет исправления (ТЗ 3.4.11, 3.6.7): исправленная редакция, которую загрузили в новую итерацию
// сравнения, вместе с замечаниями, к которым она относится, и результатом проверки. Пакет хранит
// состав документов, связь с замечаниями и историю проверок (по записи на каждый прогон), так что
// видно, что и когда проверялось. Отдельные ответы, расчёты и приложения в пакет не собираются: он
// строится из документов итерации.

// ModCorrection is the module of the correction packages.
const ModCorrection = "correction_packages"

// maxPackageHistory bounds the history of checks kept in a package.
const maxPackageHistory = 30

// PackageRun is one check of a package: what the corrections did to the findings.
type PackageRun struct {
	At         string `json:"at"`
	Resolved   int    `json:"resolved"`
	Persisting int    `json:"persisting"`
	New        int    `json:"new"`
	Remarks    int    `json:"remarks"`
}

// RecordCorrectionPackage creates or refreshes the package of an iteration of a comparison: the
// corrected files of the iteration, the remarks of the object they concern and the result of the check.
// Nothing is done for a comparison that is not an iteration (no earlier one to correct).
func (r *Runner) RecordCorrectionPackage(ctx context.Context, ns, comparisonID string) error {
	b := r.Backend
	rec, err := b.GetRecord(ns, ModComparisons, comparisonID)
	if err != nil {
		return err
	}
	iter, _ := strconv.Atoi(rec.Get("iteration_no"))
	object := rec.Get("object")
	if iter < 2 || object == "" {
		return nil
	}
	// the remarks are re-linked to the documents and re-checked first: a corrected edition that fixed
	// the difference hands the remark over to the control (ТЗ 3.7.3)
	if _, err := r.CheckObjectRemarks(ctx, ns, object); err != nil {
		log.Printf("cmp: check of the remarks of object %s: %v", object, err)
	}
	docs, err := b.Search(ns, ModDocuments, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return err
	}
	var docIDs []string
	for _, d := range docs {
		if d.Get("side") == SideTarget {
			docIDs = append(docIDs, d.ID)
		}
	}
	finds, err := b.Search(ns, ModDiscrepancies, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return err
	}
	run := PackageRun{At: dateValue(time.Now().UTC())}
	for _, f := range finds {
		switch f.Get("iter_state") {
		case IterResolved:
			run.Resolved++
		case IterPersisting:
			run.Persisting++
		case IterNew:
			run.New++
		}
	}
	inPackage := map[string]bool{}
	for _, id := range docIDs {
		inPackage[id] = true
	}
	inFinding := map[string]bool{}
	for _, f := range finds {
		inFinding[f.ID] = true
	}
	remarks, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", object))
	if err != nil {
		return err
	}
	var remarkIDs []string
	for _, rm := range remarks {
		if rm.Get("archived") != "" {
			continue
		}
		hit := false
		for _, id := range append(append([]string{}, rm.Values["linked_documents"]...), rm.Values["matched_discrepancies"]...) {
			hit = hit || inPackage[id] || inFinding[id]
		}
		if hit {
			remarkIDs = append(remarkIDs, rm.ID)
		}
	}
	run.Remarks = len(remarkIDs)

	v := idcheck.V("object", object, "comparison", comparisonID, "parent_comparison", rec.Get("parent_comparison"),
		"title", fmt.Sprintf("Исправление: %s (итерация %d)", rec.Get("title"), iter))
	v["iteration_no"] = []string{strconv.Itoa(iter)}
	v["documents"] = nonBlank(docIDs)
	v["remarks"] = nonBlank(remarkIDs)
	v["n_documents"] = []string{strconv.Itoa(len(docIDs))}
	v["n_remarks"] = []string{strconv.Itoa(len(remarkIDs))}
	v["resolved"] = []string{strconv.Itoa(run.Resolved)}
	v["persisting"] = []string{strconv.Itoa(run.Persisting)}
	v["new_findings"] = []string{strconv.Itoa(run.New)}
	v["last_run"] = []string{run.At}
	if k := rec.Get("data_kind"); k != "" {
		v["data_kind"] = []string{k}
	}

	existing, err := b.Search(ns, ModCorrection, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return err
	}
	var history []PackageRun
	runs := 0
	if len(existing) > 0 {
		_ = json.Unmarshal([]byte(existing[0].Get("history_json")), &history)
		runs, _ = strconv.Atoi(existing[0].Get("runs"))
	}
	history = append(history, run)
	if len(history) > maxPackageHistory {
		history = history[len(history)-maxPackageHistory:]
	}
	hj, _ := json.Marshal(history)
	v["history_json"] = []string{string(hj)}
	v["runs"] = []string{strconv.Itoa(runs + 1)}
	if len(existing) > 0 {
		return b.Update(ns, ModCorrection, existing[0].ID, v)
	}
	_, err = b.Create(ns, ModCorrection, v)
	return err
}

// PackageSummary is the one line a package shows.
func PackageSummary(rec idcheck.Record) string {
	return strings.TrimSpace(fmt.Sprintf("документов: %s, замечаний: %s; устранено %s, осталось %s, новых %s",
		rec.Get("n_documents"), rec.Get("n_remarks"), rec.Get("resolved"), rec.Get("persisting"), rec.Get("new_findings")))
}
