package idcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"
)

// Record is a Compose record as the runner needs it.
type Record struct {
	ID     string
	Values map[string][]string
}

// Get returns the first value of a field.
func (r Record) Get(name string) string {
	if v := r.Values[name]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// Values are record field values; multi-value fields carry several.
type Values map[string][]string

// V builds Values from name/value pairs, skipping empty values.
func V(kv ...string) Values {
	out := Values{}
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i+1] != "" {
			out[kv[i]] = []string{kv[i+1]}
		}
	}
	return out
}

// Backend is the slice of the Compose API the runner uses (implemented by
// the agent's ComposeClient; a fake in tests).
type Backend interface {
	GetRecord(ns, module, id string) (*Record, error)
	Search(ns, module, filter string) ([]Record, error)
	Create(ns, module string, values Values) (string, error)
	// Update overlays values onto the record (other fields are kept).
	Update(ns, module, id string, values Values) error
	Delete(ns, module, id string) error
	Download(ns, attachmentID string) (data []byte, name string, err error)
}

// Module handles of the «Проверка ИД (АОСР)» namespace.
const (
	ModPackages      = "packages"
	ModRegistryRows  = "registry_rows"
	ModActs          = "acts"
	ModParticipants  = "participants"
	ModOrganizations = "organizations"
	ModNRS           = "nrs_specialists"
	ModNorms         = "norm_docs"
	ModFindings      = "findings"
	ModPositions     = "position_summary"
)

// Runner executes package checks one at a time.
type Runner struct {
	Backend   Backend
	Extractor *Extractor
	External  *External // nil disables online checks

	mu      sync.Mutex
	queue   chan job
	current map[string]string // packageID → progress text
}

type job struct{ ns, packageID string }

// NewRunner starts the worker goroutine.
func NewRunner(b Backend, ex *Extractor, ext *External) *Runner {
	r := &Runner{Backend: b, Extractor: ex, External: ext, queue: make(chan job, 32), current: map[string]string{}}
	go r.loop()
	return r
}

// Enqueue schedules a package check; returns false if it is already queued
// or running.
func (r *Runner) Enqueue(ns, packageID string) bool {
	r.mu.Lock()
	if _, busy := r.current[packageID]; busy {
		r.mu.Unlock()
		return false
	}
	r.current[packageID] = "в очереди"
	r.mu.Unlock()
	r.queue <- job{ns, packageID}
	return true
}

// Status returns the progress text of a queued/running package ("" if idle).
func (r *Runner) Status(packageID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current[packageID]
}

func (r *Runner) setStatus(id, s string) {
	r.mu.Lock()
	r.current[id] = s
	r.mu.Unlock()
}

func (r *Runner) loop() {
	for j := range r.queue {
		ctx := context.Background()
		if err := r.Run(ctx, j.ns, j.packageID); err != nil {
			log.Printf("idcheck: package %s: %v", j.packageID, err)
			_ = r.Backend.Update(j.ns, ModPackages, j.packageID, V("status", "failed", "progress", "Ошибка: "+err.Error()))
		}
		r.mu.Lock()
		delete(r.current, j.packageID)
		r.mu.Unlock()
	}
}

func fmtNum(p *float64) string {
	if p == nil {
		return ""
	}
	return trimFloat(*p)
}

// Run checks one package end to end.
func (r *Runner) Run(ctx context.Context, ns, pkgID string) error {
	b := r.Backend
	pkg, err := b.GetRecord(ns, ModPackages, pkgID)
	if err != nil {
		return fmt.Errorf("package: %w", err)
	}
	progress := func(s string) {
		r.setStatus(pkgID, s)
		_ = b.Update(ns, ModPackages, pkgID, V("progress", s))
	}
	_ = b.Update(ns, ModPackages, pkgID, V("status", "processing", "progress", "Подготовка", "started_at", time.Now().UTC().Format(time.RFC3339)))

	progress("Очистка результатов прошлой проверки")
	for _, m := range []string{ModFindings, ModParticipants, ModPositions, ModActs, ModRegistryRows} {
		recs, err := b.Search(ns, m, fmt.Sprintf("package = '%s'", pkgID))
		if err != nil {
			return fmt.Errorf("search %s: %w", m, err)
		}
		for _, rec := range recs {
			_ = b.Delete(ns, m, rec.ID)
		}
	}

	progress("Разбор реестра и КС-2")
	reg, err := r.loadRegistry(ns, pkg)
	if err != nil {
		return err
	}
	var ks *KS2
	if id := pkg.Get("ks2_xlsx"); id != "" {
		data, _, err := b.Download(ns, id)
		if err != nil {
			return fmt.Errorf("КС-2: %w", err)
		}
		if ks, err = ParseKS2(data); err != nil {
			return err
		}
	}
	norms := r.loadNorms(ns)

	// registry rows
	rowIDs := map[int]string{}
	for i, row := range reg.Rows {
		if i%25 == 0 {
			progress(fmt.Sprintf("Запись строк реестра %d/%d", i, len(reg.Rows)))
		}
		v := V("package", pkgID, "row_no", fmt.Sprint(row.Row), "seq", row.Seq, "kind", string(row.Kind),
			"position", row.Position, "work_name", row.WorkName, "doc_name", row.DocName,
			"doc_number", row.DocNumber, "doc_date", isoDate(row.DocDate), "doc_date_raw", row.DocDateRaw, "unit", row.Unit,
			"qty_vrc", fmtNum(row.QtyVRC), "qty_total", fmtNum(row.QtyTotal), "qty_ks2", fmtNum(row.QtyKS2),
			"qty_project", fmtNum(row.QtyProject), "qty_fact", fmtNum(row.QtyFact))
		if row.Kind == RowAct {
			v["doc_number_norm"] = []string{ParseActNumber(row.DocNumber, '/').Key()}
		}
		id, err := b.Create(ns, ModRegistryRows, v)
		if err != nil {
			return fmt.Errorf("registry row %d: %w", row.Row, err)
		}
		rowIDs[row.Row] = id
	}

	var all []Finding
	all = append(all, CheckRegistryConsistency(reg)...)
	if ks != nil {
		progress("Сверка реестра с КС-2 (проверка 8)")
		sum := ReconcileKS2(reg, ks)
		for _, s := range sum {
			_, err := b.Create(ns, ModPositions, V("package", pkgID, "position", s.Position, "name", s.Name, "unit", s.Unit,
				"registry_sum", trimFloat(s.RegistrySum), "registry_ks2", fmtNum(s.RegistryKS2), "ks2_qty", fmtNum(s.KS2Qty),
				"delta", trimFloat(s.Delta), "status", s.Status))
			if err != nil {
				return fmt.Errorf("position summary: %w", err)
			}
		}
		all = append(all, Check8Findings(sum)...)
	} else {
		all = append(all, Finding{Check: 8, Severity: High, Source: "det", Description: "Форма КС-2 (xlsx) не загружена — проверка 8 не выполнена"})
	}
	if err := r.writeFindings(ns, pkgID, "", all, rowIDs); err != nil {
		return err
	}

	// acts
	files := pkg.Values["act_files"]
	type actStat struct{ high, medium, low int }
	var total, ok, withIssues int
	counts := map[Severity]int{}
	for _, f := range all {
		counts[f.Severity]++
	}
	actsByKey := map[string]bool{}
	for i, attID := range files {
		data, name, err := b.Download(ns, attID)
		if err != nil {
			all = append(all, Finding{Check: 1, Severity: Low, Source: "det", Description: fmt.Sprintf("Не удалось скачать файл акта %s: %v", attID, err)})
			continue
		}
		if !strings.HasSuffix(strings.ToLower(name), ".pdf") {
			continue
		}
		total++
		prefix := fmt.Sprintf("Акт %d/%d (%s): ", i+1, len(files), name)
		progress(prefix + "подготовка")
		num, fdate := ParseActFileName(name)
		in := &ActInput{Key: name, FileNum: num, FileDate: fdate, Rows: MatchRegistryRows(reg, num, fdate), Norms: norms}
		actsByKey[num.Key()+"|"+FmtDate(fdate)] = true

		actID, err := b.Create(ns, ModActs, V("package", pkgID, "file_name", name, "attachment_id", attID,
			"number", num.Raw, "number_norm", num.Key(), "date", isoDate(fdate), "positions", strings.Join(num.Positions, "; "),
			"ocr_status", "processing"))
		if err != nil {
			return fmt.Errorf("act record: %w", err)
		}
		var res *ExtractResult
		if r.Extractor != nil {
			res, err = r.Extractor.Extract(ctx, ActFile{Name: name, Data: data}, func(s string) { progress(prefix + s) })
			if err != nil {
				log.Printf("idcheck: %s: %v", name, err)
			}
		}
		ocrStatus, ocrNote := "done", ""
		if res == nil {
			ocrStatus, ocrNote = "failed", fmt.Sprint(err)
		} else {
			in.Data = res.Data
			ocrNote = strings.Join(res.Notes, "\n")
		}
		fs := CheckAct(in)
		if r.External != nil && in.Data != nil {
			progress(prefix + "проверка ЕГРЮЛ / НРС")
			fs = append(fs, r.External.CheckExternal(ctx, in)...)
			r.cacheExternal(ns, in)
		}
		if res != nil && len(res.Words) > 0 {
			for i := range fs {
				if p, box, ok := LocateFinding(&fs[i], res.Pages, res.Words, res.ActPages); p > 0 {
					fs[i].Page = p
					if ok {
						b := box
						fs[i].Box = &b
					}
				}
			}
		}
		st := actStat{}
		for _, f := range fs {
			counts[f.Severity]++
			switch f.Severity {
			case High:
				st.high++
			case Medium:
				st.medium++
			case Low:
				st.low++
			}
		}
		checkStatus := "ok"
		if st.high+st.medium > 0 {
			checkStatus = "issues"
			withIssues++
		} else {
			ok++
		}
		upd := V("ocr_status", ocrStatus, "ocr_notes", ocrNote, "check_status", checkStatus,
			"issues_high", fmt.Sprint(st.high), "issues_medium", fmt.Sprint(st.medium), "issues_low", fmt.Sprint(st.low))
		if len(in.Rows) > 0 {
			upd["registry_row"] = []string{rowIDs[in.Rows[0].Row]}
		}
		if res != nil {
			pj, _ := json.Marshal(res.Pages)
			upd["pages_json"] = []string{string(pj)}
		}
		if d := in.Data; d != nil {
			ej, _ := json.MarshalIndent(d, "", "  ")
			for k, v := range V("act_type", d.Type, "doc_number", d.Number, "doc_date", isoDateStr(d.Date),
				"p1_work", d.P1Work, "p1_vol_project", fmtNum(d.P1VolProject), "p1_vol_fact", fmtNum(d.P1VolFact), "p1_unit", d.P1Unit,
				"p2_rd_code", d.P2RDCode, "p2_text", d.P2Text, "p3_materials", docList(d.P3Materials), "p4_docs", docList(d.P4Docs),
				"p5_start", isoDateStr(d.P5Start), "p5_end", isoDateStr(d.P5End), "p6_norms", strings.Join(d.P6Norms, "\n"),
				"p6_rd_code", d.P6RDCode, "appendix", docList(d.Appendix), "attached", docList(d.Attached),
				"scheme_rd_code", d.SchemeRDCode, "scheme_vol_project", fmtNum(d.SchemeVolProj), "scheme_vol_fact", fmtNum(d.SchemeVolFact),
				"extracted_json", string(ej)) {
				upd[k] = v
			}
		}
		if err := b.Update(ns, ModActs, actID, upd); err != nil {
			log.Printf("idcheck: update act %s: %v", name, err)
		}
		if in.Data != nil {
			for _, p := range in.Data.People {
				_, _ = b.Create(ns, ModParticipants, V("package", pkgID, "act", actID, "role", p.Role, "fio", p.FIO,
					"position_title", p.Position, "org", p.Org, "order_no", p.OrderNo, "order_date", isoDateStr(p.OrderDate),
					"nrs_id", NormalizeNRS(p.NRSID)))
			}
		}
		if err := r.writeFindings(ns, pkgID, actID, fs, rowIDs); err != nil {
			return err
		}
		_ = b.Update(ns, ModPackages, pkgID, V("acts_total", fmt.Sprint(total), "acts_ok", fmt.Sprint(ok),
			"acts_with_issues", fmt.Sprint(withIssues), "findings_high", fmt.Sprint(counts[High]),
			"findings_medium", fmt.Sprint(counts[Medium]), "findings_low", fmt.Sprint(counts[Low])))
	}

	// registry acts without a delivered file (only within the package's
	// positions: the registry lists every act ever, files are a subset)
	missing := r.missingFiles(reg, actsByKey, files)
	if len(missing) > 0 {
		if err := r.writeFindings(ns, pkgID, "", missing, rowIDs); err != nil {
			return err
		}
		counts[Medium] += len(missing)
	}

	summary := fmt.Sprintf("Проверено актов: %d (без замечаний: %d, с замечаниями: %d). Замечания: высокие %d, средние %d, низкие %d.",
		total, ok, withIssues, counts[High], counts[Medium], counts[Low])
	if r.Extractor == nil || !r.Extractor.Vision.Enabled() {
		summary += " Распознавание сканов отключено — проверки 2–7 не выполнялись."
	} else if !r.Extractor.OCR.Available() {
		summary += " Tesseract не установлен — страницы классифицированы моделью."
	}
	return b.Update(ns, ModPackages, pkgID, V("status", "done", "progress", summary,
		"acts_total", fmt.Sprint(total), "acts_ok", fmt.Sprint(ok), "acts_with_issues", fmt.Sprint(withIssues),
		"findings_high", fmt.Sprint(counts[High]), "findings_medium", fmt.Sprint(counts[Medium]), "findings_low", fmt.Sprint(counts[Low]),
		"finished_at", time.Now().UTC().Format(time.RFC3339)))
}

// missingFiles reports registry act rows (for positions that have at least
// one delivered file) whose act file is absent from the package.
func (r *Runner) missingFiles(reg *Registry, have map[string]bool, files []string) []Finding {
	if len(files) == 0 {
		return nil
	}
	positions := map[string]bool{}
	for k := range have {
		n := ParseActNumber(strings.SplitN(k, "|", 2)[0], '/')
		for _, p := range n.Positions {
			positions[p] = true
		}
	}
	var out []Finding
	seen := map[string]bool{}
	for _, row := range reg.Acts() {
		n := ParseActNumber(row.DocNumber, '/')
		if !positions[compact(row.Position)] || seen[n.Key()] {
			continue
		}
		seen[n.Key()] = true
		found := false
		for k := range have {
			kn := ParseActNumber(strings.SplitN(k, "|", 2)[0], '/')
			if sameActKey(kn, n, n.Key()) || (kn.Seq == n.Seq && strings.HasSuffix(k, "|"+FmtDate(row.DocDate)) && kn.HasPosition(row.Position)) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, Finding{Check: 1, Severity: Medium, Source: "det", Field: "file", RegistryRow: row.Row, Position: row.Position,
				Expected: row.DocNumber, Description: fmt.Sprintf("Акт %s от %s указан в реестре (строка %d), но файл акта в пакете отсутствует", row.DocNumber, FmtDate(row.DocDate), row.Row)})
		}
	}
	return out
}

func (r *Runner) writeFindings(ns, pkgID, actID string, fs []Finding, rowIDs map[int]string) error {
	sort.SliceStable(fs, func(i, j int) bool {
		if fs[i].Check != fs[j].Check {
			return fs[i].Check < fs[j].Check
		}
		return sevRank(fs[i].Severity) < sevRank(fs[j].Severity)
	})
	for _, f := range fs {
		v := V("package", pkgID, "act", actID, "check_no", fmt.Sprint(f.Check), "check_name", CheckNames[f.Check],
			"severity", string(f.Severity), "field", f.Field, "expected", f.Expected, "actual", f.Actual,
			"description", f.Description, "source", f.Source, "position", f.Position)
		if f.Page > 0 {
			v["page"] = []string{fmt.Sprint(f.Page)}
		}
		if f.Box != nil {
			v["bbox"] = []string{f.Box.String()}
		}
		if f.RegistryRow > 0 {
			v["registry_row"] = []string{rowIDs[f.RegistryRow]}
			v["registry_row_no"] = []string{fmt.Sprint(f.RegistryRow)}
		}
		if _, err := r.Backend.Create(ns, ModFindings, v); err != nil {
			return fmt.Errorf("finding: %w", err)
		}
	}
	return nil
}

func sevRank(s Severity) int {
	switch s {
	case High:
		return 0
	case Medium:
		return 1
	case Low:
		return 2
	}
	return 3
}

func (r *Runner) loadRegistry(ns string, pkg *Record) (*Registry, error) {
	id := pkg.Get("registry_xlsx")
	if id == "" {
		return nil, fmt.Errorf("в пакете не загружен реестр ИД (xlsx)")
	}
	data, _, err := r.Backend.Download(ns, id)
	if err != nil {
		return nil, fmt.Errorf("реестр: %w", err)
	}
	return ParseRegistry(data)
}

// loadNorms overlays the namespace's norm_docs records onto DefaultNorms.
func (r *Runner) loadNorms(ns string) *NormCatalog {
	recs, err := r.Backend.Search(ns, ModNorms, "")
	if err != nil {
		log.Printf("idcheck: norm_docs: %v (using built-in catalog)", err)
		return NewNormCatalog(DefaultNorms)
	}
	var over []Norm
	for _, rec := range recs {
		n := Norm{Code: rec.Get("code"), Title: rec.Get("title"), ReplacedBy: rec.Get("replaced_by")}
		if n.Code == "" {
			continue
		}
		n.ValidFrom, _ = ParseDate(rec.Get("valid_from"))
		n.ValidTo, _ = ParseDate(rec.Get("valid_to"))
		for _, t := range strings.Split(rec.Get("work_tags"), ",") {
			if t = strings.TrimSpace(t); t != "" {
				n.WorkTags = append(n.WorkTags, t)
			}
		}
		over = append(over, n)
	}
	return NewNormCatalog(DefaultNorms, over...)
}

// cacheExternal stores looked-up organizations/specialists in the
// reference modules so users see what the registries returned.
func (r *Runner) cacheExternal(ns string, in *ActInput) {
	b := r.Backend
	now := time.Now().UTC().Format(time.RFC3339)
	for _, o := range in.Data.Orgs {
		key := FirstDigitRun(o.OGRN)
		if !ValidOGRN(key) {
			continue
		}
		rec, _ := r.External.EGRUL(context.Background(), key)
		v := V("name", o.Name, "ogrn", key, "inn", FirstDigitRun(o.INN), "address", o.Address, "sro_name", o.SRO, "checked_at", now)
		if rec != nil {
			for k, x := range V("egrul_name", rec.FullName, "egrul_short_name", rec.ShortName, "egrul_inn", rec.INN,
				"egrul_kpp", rec.KPP, "egrul_region", rec.Region, "egrul_reg_date", isoDateStr(rec.RegDate),
				"egrul_end_date", isoDateStr(rec.EndDate), "egrul_director", rec.Director) {
				v[k] = x
			}
			v["egrul_status"] = []string{"active"}
			if rec.EndDate != "" {
				v["egrul_status"] = []string{"ceased"}
			}
		} else {
			v["egrul_status"] = []string{"not_found"}
		}
		upsert(b, ns, ModOrganizations, fmt.Sprintf("ogrn = '%s'", key), v)
	}
	for _, p := range in.Data.People {
		n := NormalizeNRS(p.NRSID)
		if n == "" {
			continue
		}
		rec, err := r.External.NRS(context.Background(), n)
		if err != nil {
			continue
		}
		status := "not_found"
		if rec.Found {
			status = "active"
			if rec.Excluded {
				status = "excluded"
			}
		}
		upsert(b, ns, ModNRS, fmt.Sprintf("nrs_id = '%s'", n), V("nrs_id", n, "fio", p.FIO, "status", status,
			"status_text", rec.Status, "work_type", rec.WorkType, "checked_at", now))
	}
}

func upsert(b Backend, ns, module, filter string, v Values) {
	recs, err := b.Search(ns, module, filter)
	if err == nil && len(recs) > 0 {
		_ = b.Update(ns, module, recs[0].ID, v)
		return
	}
	_, _ = b.Create(ns, module, v)
}

func isoDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func isoDateStr(s string) string {
	if t, ok := ParseDate(s); ok {
		return isoDate(t)
	}
	return ""
}

func docList(ds []DocRef) string {
	var lines []string
	for _, d := range ds {
		s := describeDoc(d)
		if d.Material != "" {
			s += " — " + d.Material
		}
		if d.ValidUntil != "" {
			s += " (годен до " + d.ValidUntil + ")"
		}
		if d.Page > 0 {
			s += fmt.Sprintf(" [стр. %d]", d.Page)
		}
		lines = append(lines, s)
	}
	return strings.Join(lines, "\n")
}
