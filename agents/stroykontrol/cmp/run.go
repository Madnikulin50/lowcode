package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Backend is the slice of the Compose API the runner needs; the agent's
// ComposeClient adapter (idcheck.Backend) satisfies it.
type Backend = idcheck.Backend

// Values are record field values.
type Values = idcheck.Values

// Module handles of the «Стройконтроль ПТО» namespace used here.
const (
	ModComparisons   = "pd_rd_comparisons"
	ModDiscrepancies = "pd_rd_discrepancies"
	ModDocuments     = "cmp_documents"
	ModRemarks       = "remark_cards"
)

// Statuses of a discrepancy (reviewed by the specialist).
const (
	StatusNew       = "new"
	StatusConfirmed = "confirmed"
	StatusRejected  = "rejected"
	// StatusClarify: the specialist cannot decide and asks for a clarification (ТЗ 3.2.7); needs a comment.
	// The finding stays open, is not a remark and keeps the decision across re-runs and iterations.
	StatusClarify = "clarify"
)

// Input is one source file of a comparison.
type Input struct {
	ID          string
	Name        string
	Side        string
	Data        []byte
	ManualCode  string
	ManualTitle string
}

// AnalyzeOptions are the optional helpers of Analyze.
type AnalyzeOptions struct {
	// Rasterize renders one page of a PDF for the image comparison of drawing sheets.
	Rasterize func(data []byte, page int) (image.Image, error)
	// OCR recognizes scanned pages (no text layer).
	OCR *OCR
	// Mode is ModeStages (ПД → РД, default) or ModeVersions (two versions of one stage)
	Mode string
	// Pairing is the manual pairing of volumes and page ranges
	Pairing []ManualPair
	// Exclude are the areas of pages left out of the comparison
	Exclude []Region
}

// Analyze extracts every PDF input and compares the target set against the
// reference set. Non-PDF inputs are skipped and listed in Result.Notes.
func Analyze(inputs []Input, cacheDir string, rasterize func(data []byte, page int) (image.Image, error)) (*Result, error) {
	return AnalyzeWith(inputs, cacheDir, AnalyzeOptions{Rasterize: rasterize})
}

// AnalyzeWith is Analyze with options.
func AnalyzeWith(inputs []Input, cacheDir string, o AnalyzeOptions) (*Result, error) {
	var refs, tgts []*Document
	data := map[string][]byte{}
	var skipped []string
	var ocrDone, ocrLeft int
	var ocrByDoc []string
	for _, in := range inputs {
		if !IsPDF(in.Data) {
			skipped = append(skipped, in.Name)
			continue
		}
		pages, done, left, err := ExtractPagesOCR(context.Background(), in.ID, in.Data, cacheDir, o.OCR)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", in.Name, err)
		}
		ocrDone += done
		ocrLeft += left
		if done > 0 {
			ocrByDoc = append(ocrByDoc, fmt.Sprintf("%s — %d", in.Name, done))
		}
		d := &Document{ID: in.ID, Name: in.Name, Side: in.Side, ManualCode: in.ManualCode, ManualTitle: in.ManualTitle, Pages: pages}
		data[in.ID] = in.Data
		if in.Side == SideRef {
			refs = append(refs, d)
		} else {
			tgts = append(tgts, d)
		}
	}
	if len(refs) == 0 || len(tgts) == 0 {
		if o.Mode == ModeVersions {
			return nil, fmt.Errorf("нужен хотя бы один PDF-файл версии «до» и один — версии «после» (сейчас: %d и %d)", len(refs), len(tgts))
		}
		return nil, fmt.Errorf("нужен хотя бы один PDF-файл ПД и один PDF-файл РД (сейчас ПД: %d, РД: %d)", len(refs), len(tgts))
	}
	opt := Options{Mode: o.Mode, Pairing: o.Pairing, Exclude: o.Exclude}
	if o.Rasterize != nil {
		opt.Imager = func(docID string, page int) (image.Image, error) { return o.Rasterize(data[docID], page) }
	}
	res := Compare(refs, tgts, opt)
	if len(skipped) > 0 {
		res.Notes = append(res.Notes, "Пропущены не-PDF файлы: "+strings.Join(skipped, ", ")+".")
	}
	if ocrDone > 0 {
		note := fmt.Sprintf("Распознано (OCR) листов-сканов: %d (%s); найденные по ним расхождения менее надёжны, критичность ограничена средней.", ocrDone, strings.Join(ocrByDoc, "; "))
		if ocrLeft > 0 {
			note += fmt.Sprintf(" Ещё %d листов-сканов не распознавались (лимит).", ocrLeft)
		}
		res.Notes = append(res.Notes, note)
	}
	return res, nil
}

// Runner executes comparisons one at a time against a live Compose.
type Runner struct {
	Backend   Backend
	CacheDir  string
	Rasterize func(data []byte, page int) (image.Image, error)
	OCR       *OCR
	// LLM drafts wordings and reads ambiguous expertise remarks (optional)
	LLM *Suggester

	mu      sync.Mutex
	queue   chan job
	current map[string]string
}

type job struct{ ns, id string }

// NewRunner starts the worker goroutine.
func NewRunner(b Backend, cacheDir string, rasterize func(data []byte, page int) (image.Image, error)) *Runner {
	r := &Runner{Backend: b, CacheDir: cacheDir, Rasterize: rasterize, queue: make(chan job, 32), current: map[string]string{}}
	go r.loop()
	return r
}

// Enqueue schedules a comparison; false if it is already queued or running.
func (r *Runner) Enqueue(ns, comparisonID string) bool {
	r.mu.Lock()
	if _, busy := r.current[comparisonID]; busy {
		r.mu.Unlock()
		return false
	}
	r.current[comparisonID] = "в очереди"
	r.mu.Unlock()
	r.queue <- job{ns, comparisonID}
	return true
}

// Status is the progress text of a queued/running comparison ("" if idle).
func (r *Runner) Status(comparisonID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.current[comparisonID]
}

func (r *Runner) setStatus(id, s string) {
	r.mu.Lock()
	r.current[id] = s
	r.mu.Unlock()
}

func (r *Runner) loop() {
	for j := range r.queue {
		if err := r.Run(context.Background(), j.ns, j.id); err != nil {
			log.Printf("cmp: comparison %s: %v", j.id, err)
			_ = r.Backend.Update(j.ns, ModComparisons, j.id, idcheck.V("status", "failed", "comment", "Ошибка сравнения: "+err.Error()))
		}
		r.mu.Lock()
		delete(r.current, j.id)
		r.mu.Unlock()
	}
}

// docRef is a source file of a comparison as stored in Compose.
type docRef struct {
	ID         string // cmp_documents record ID, or legacy:pd / legacy:rd
	Side       string
	Attachment string
	Manual     bool
	Code       string
	Title      string
	Persist    bool // a real cmp_documents record (can be updated)
}

// Ancestors returns the comparison's own ID followed by its previous
// iterations, nearest first (cycles and absurd depth are cut off).
func Ancestors(b Backend, ns, id string) []string {
	out := []string{id}
	seen := map[string]bool{id: true}
	for cur := id; len(out) < 12; {
		rec, err := b.GetRecord(ns, ModComparisons, cur)
		if err != nil {
			break
		}
		p := rec.Get("parent_comparison")
		if p == "" || seen[p] {
			break
		}
		seen[p] = true
		out = append(out, p)
		cur = p
	}
	return out
}

// ListDocs returns the files of a comparison: its cmp_documents records, or —
// for comparisons made before packs existed — the single pd_file/rd_file.
func ListDocs(b Backend, ns, comparisonID string) ([]docRef, error) {
	recs, err := b.Search(ns, ModDocuments, fmt.Sprintf("comparison = '%s'", comparisonID))
	if err != nil {
		return nil, err
	}
	var out []docRef
	for _, rec := range recs {
		if rec.Get("file") == "" {
			continue
		}
		side := rec.Get("side")
		if side != SideRef && side != SideTarget {
			side = SideTarget
		}
		out = append(out, docRef{
			ID: rec.ID, Side: side, Attachment: rec.Get("file"),
			Manual: rec.Get("detected_by") == "manual" && rec.Get("volume_code") != "",
			Code:   rec.Get("volume_code"), Title: rec.Get("volume_title"), Persist: true,
		})
	}
	cmpRec, err := b.GetRecord(ns, ModComparisons, comparisonID)
	if err != nil {
		return nil, err
	}
	// files of the document sets the comparison points at (ПД set, РД set)
	have := map[string]bool{}
	for _, d := range out {
		have[d.ID] = true
	}
	for _, f := range []struct{ field, side string }{{"set_pd", SideRef}, {"set_rd", SideTarget}} {
		setID := cmpRec.Get(f.field)
		if setID == "" {
			continue
		}
		recs, err := b.Search(ns, ModDocuments, fmt.Sprintf("set = '%s'", setID))
		if err != nil {
			return nil, err
		}
		for _, rec := range recs {
			if rec.Get("file") == "" || have[rec.ID] {
				continue
			}
			have[rec.ID] = true
			out = append(out, docRef{
				ID: rec.ID, Side: f.side, Attachment: rec.Get("file"),
				Manual: rec.Get("detected_by") == "manual" && rec.Get("volume_code") != "",
				Code:   rec.Get("volume_code"), Title: rec.Get("volume_title"), Persist: true,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	// a later iteration normally brings only the corrected РД: the reference
	// set (ПД) is inherited from the nearest previous iteration that has one
	hasRef := false
	for _, d := range out {
		hasRef = hasRef || d.Side == SideRef
	}
	if !hasRef {
		for _, anc := range Ancestors(b, ns, comparisonID)[1:] {
			ar, err := b.Search(ns, ModDocuments, fmt.Sprintf("comparison = '%s'", anc))
			if err != nil {
				return nil, err
			}
			var inherited []docRef
			for _, rec := range ar {
				if rec.Get("side") == SideRef && rec.Get("file") != "" {
					inherited = append(inherited, docRef{ID: rec.ID, Side: SideRef, Attachment: rec.Get("file"), Persist: true})
				}
			}
			if len(inherited) > 0 {
				sort.SliceStable(inherited, func(i, j int) bool { return inherited[i].ID < inherited[j].ID })
				out = append(inherited, out...)
				break
			}
		}
	}
	if id := cmpRec.Get("pd_file"); id != "" {
		out = append([]docRef{{ID: "legacy:pd", Side: SideRef, Attachment: id}}, out...)
	}
	if id := cmpRec.Get("rd_file"); id != "" {
		out = append(out, docRef{ID: "legacy:rd", Side: SideTarget, Attachment: id})
	}
	return out, nil
}

// Run performs one comparison and writes the result back.
func (r *Runner) Run(ctx context.Context, ns, id string) error {
	b := r.Backend
	r.setStatus(id, "загрузка файлов")
	if err := b.Update(ns, ModComparisons, id, idcheck.V("status", "processing", "comment", "Сравнение выполняется…")); err != nil {
		return err
	}
	docs, err := ListDocs(b, ns, id)
	if err != nil {
		return err
	}
	cmpMode := ModeStages
	var pairing []ManualPair
	var exclude []Region
	if rec, err := b.GetRecord(ns, ModComparisons, id); err == nil {
		if rec.Get("kind") == ModeVersions {
			cmpMode = ModeVersions
		}
		var perr error
		if pairing, perr = ParsePairing(rec.Get("pairing")); perr != nil {
			return b.Update(ns, ModComparisons, id, idcheck.V("status", "new", "comment", perr.Error()))
		}
		if exclude, perr = ParseRegions(rec.Get("excluded_regions")); perr != nil {
			return b.Update(ns, ModComparisons, id, idcheck.V("status", "new", "comment", perr.Error()))
		}
	}
	var inputs []Input
	for _, d := range docs {
		data, name, err := b.Download(ns, d.Attachment)
		if err != nil {
			return fmt.Errorf("загрузка файла %s: %w", d.ID, err)
		}
		in := Input{ID: d.ID, Name: name, Side: d.Side, Data: data}
		if d.Manual {
			in.ManualCode, in.ManualTitle = d.Code, d.Title
		}
		inputs = append(inputs, in)
	}

	r.setStatus(id, "сравнение")
	res, err := AnalyzeWith(inputs, r.CacheDir, AnalyzeOptions{Rasterize: r.Rasterize, OCR: r.OCR, Mode: cmpMode, Pairing: pairing, Exclude: exclude})
	if err != nil {
		// not an infrastructure failure: tell the user what is missing
		return b.Update(ns, ModComparisons, id, idcheck.V("status", "new", "comment", err.Error()))
	}

	r.setStatus(id, "запись результата")
	if err := r.persistDocuments(ns, docs, inputs, res); err != nil {
		return err
	}
	if err := r.persistResult(ns, id, res); err != nil {
		return err
	}
	// keep the completeness of the object's sets and the "ПД v2 ↔ РД v1" label of the pair current
	if rec, err := b.GetRecord(ns, ModComparisons, id); err == nil && rec.Get("object") != "" {
		if _, err := r.CheckSets(ctx, ns, rec.Get("object")); err != nil {
			log.Printf("cmp: sets check of object %s: %v", rec.Get("object"), err)
		}
	}
	// an iteration is the corrected edition: it becomes a correction package and the remarks it concerns are re-checked
	if err := r.RecordCorrectionPackage(ctx, ns, id); err != nil {
		log.Printf("cmp: correction package of %s: %v", id, err)
	}
	return nil
}

func (r *Runner) persistDocuments(ns string, docs []docRef, inputs []Input, res *Result) error {
	pages := map[string]int{}
	for _, in := range inputs {
		for _, v := range res.Volumes[in.ID] {
			if v.To > pages[in.ID] {
				pages[in.ID] = v.To
			}
		}
	}
	for _, d := range docs {
		if !d.Persist {
			continue
		}
		vs := res.Volumes[d.ID]
		v := idcheck.V("pages", strconv.Itoa(pages[d.ID]))
		v["sheet_map"] = []string{FormatSheetMap(res.Sheets[d.ID])}
		v["headings_json"] = []string{HeadingsJSON(res.Headings[d.ID])}
		if !d.Manual {
			existing, _ := r.Backend.GetRecord(ns, ModDocuments, d.ID)
			for _, in := range inputs {
				if in.ID == d.ID {
					for k, x := range EditionValues(existing, hashKey(in.Data), vs) {
						v[k] = x
					}
				}
			}
		}
		if !d.Manual && len(vs) > 0 {
			var codes, titles []string
			for _, x := range vs {
				if x.Code != "" {
					codes = append(codes, x.Code)
				}
				titles = append(titles, x.Title)
			}
			if len(vs) == 1 {
				v["volume_code"] = []string{vs[0].Code}
				v["volume_title"] = []string{vs[0].Title}
			} else {
				v["volume_code"] = []string{strings.Join(codes, ", ")}
				v["volume_title"] = []string{fmt.Sprintf("%d томов: %s", len(vs), strings.Join(titles, "; "))}
			}
			v["detected_by"] = []string{"auto"}
		}
		if err := r.Backend.Update(ns, ModDocuments, d.ID, v); err != nil {
			return err
		}
	}
	return nil
}

func discrepancyValues(comparisonID string, f Finding) Values {
	page := f.PageRef
	if page == 0 {
		page = f.PageTarget
	}
	ev, _ := json.Marshal(f.Evidence)
	v := idcheck.V(
		"comparison", comparisonID,
		"discrepancy_type", f.Type,
		"severity", f.Severity,
		"description", "[детерм.] "+f.Description,
		"kind", f.Kind,
		"doc_ref", f.RefDoc,
		"doc_target", f.TargetDoc,
		"anchor_text", f.Anchor,
		"evidence", string(ev),
		"fingerprint", f.Fingerprint,
	)
	setNum := func(name string, n int) {
		if n > 0 {
			v[name] = []string{strconv.Itoa(n)}
		}
	}
	setNum("page_number", page)
	setNum("page_pd", f.PageRef)
	setNum("page_rd", f.PageTarget)
	if f.BoxRef != nil {
		v["bbox_pd"] = []string{f.BoxRef.String()}
	}
	if f.BoxTarget != nil {
		v["bbox_rd"] = []string{f.BoxTarget.String()}
	}
	return v
}

func (r *Runner) persistResult(ns, id string, res *Result) error {
	b := r.Backend
	old, err := b.Search(ns, ModDiscrepancies, fmt.Sprintf("comparison = '%s'", id))
	if err != nil {
		return err
	}
	// carry the reviewer's decisions over to findings that are still there
	status, comments := map[string]string{}, map[string]string{}
	for _, o := range old {
		if fp, st := o.Get("fingerprint"), o.Get("status"); fp != "" && st != "" && st != StatusNew {
			status[fp] = st
			if c := o.Get("reviewer_comment"); c != "" {
				comments[fp] = c
			}
		}
	}
	for _, o := range old {
		if err := b.Delete(ns, ModDiscrepancies, o.ID); err != nil {
			return err
		}
	}

	// a later iteration of the same pair: what the corrections fixed, what is still open
	state, resolved, iterNo, err := r.iteration(ns, id, res, status, comments)
	if err != nil {
		return err
	}
	kind := ""
	if c, err := b.GetRecord(ns, ModComparisons, id); err == nil {
		kind = c.Get("data_kind") // demo comparisons keep their findings out of the working lists
	}
	for _, f := range res.Findings {
		v := discrepancyValues(id, f)
		if kind != "" {
			v["data_kind"] = []string{kind}
		}
		st := status[f.Fingerprint]
		if st == "" {
			st = StatusNew
		}
		v["status"] = []string{st}
		if c := comments[f.Fingerprint]; c != "" && st != StatusNew {
			v["reviewer_comment"] = []string{c}
		}
		if s := state[f.Fingerprint]; s != "" {
			v["iter_state"] = []string{s}
		}
		if _, err := b.Create(ns, ModDiscrepancies, v); err != nil {
			return err
		}
	}
	for _, p := range resolved {
		v := Values{}
		for k, vals := range p.Values {
			v[k] = vals
		}
		v["comparison"] = []string{id}
		v["iter_state"] = []string{IterResolved}
		if kind != "" {
			v["data_kind"] = []string{kind}
		}
		if _, err := b.Create(ns, ModDiscrepancies, v); err != nil {
			return err
		}
	}

	matching := res.Stats.ComparedPages - res.Stats.DifferingPages
	if matching < 0 {
		matching = 0
	}
	comment := "Агент (детерминированное сравнение). " + strings.Join(res.Notes, " ")
	if iterNo > 0 {
		var nNew, nStill int
		for _, s := range state {
			if s == IterNew {
				nNew++
			} else {
				nStill++
			}
		}
		comment = fmt.Sprintf("Итерация %d: устранено %d, осталось %d, новых %d. ", iterNo, len(resolved), nStill, nNew) + comment
	}
	sim := ""
	if res.Stats.SimilarityKnown {
		sim = strconv.FormatFloat(res.Stats.Similarity, 'f', 1, 64)
	}
	upd := idcheck.V(
		"status", "done",
		"total_pages_pd", strconv.Itoa(res.Stats.RefPages),
		"total_pages_rd", strconv.Itoa(res.Stats.TargetPages),
		"matching_pages", strconv.Itoa(matching),
		"differing_pages", strconv.Itoa(res.Stats.DifferingPages),
		"comment", comment,
	)
	if iterNo > 0 {
		upd["iteration_no"] = []string{strconv.Itoa(iterNo)}
	}
	// Update overlays values: an explicit empty value (idcheck.V drops it)
	// removes a stale percentage left by an earlier run
	upd["similarity_percent"] = []string{sim}
	upd["pairs_json"] = []string{res.StateJSON()}
	return b.Update(ns, ModComparisons, id, upd)
}

// iteration compares the findings with those of the comparison's parent
// (the previous iteration). iterNo is 0 when there is no parent.
func (r *Runner) iteration(ns, id string, res *Result, status, comments map[string]string) (state map[string]string, resolved []Prior, iterNo int, err error) {
	b := r.Backend
	rec, err := b.GetRecord(ns, ModComparisons, id)
	if err != nil {
		return nil, nil, 0, err
	}
	parent := rec.Get("parent_comparison")
	if parent == "" || parent == id {
		return nil, nil, 0, nil
	}
	pRec, err := b.GetRecord(ns, ModComparisons, parent)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("предыдущая итерация %s: %w", parent, err)
	}
	pn, _ := strconv.Atoi(pRec.Get("iteration_no"))
	if pn < 1 {
		pn = 1
	}
	iterNo = pn + 1
	recs, err := b.Search(ns, ModDiscrepancies, fmt.Sprintf("comparison = '%s'", parent))
	if err != nil {
		return nil, nil, 0, err
	}
	var prior []Prior
	for _, p := range recs {
		if p.Get("iter_state") == IterResolved {
			continue // closed in the parent already
		}
		prior = append(prior, Prior{Fingerprint: p.Get("fingerprint"), Status: p.Get("status"), Values: p.Values})
		if fp, c := p.Get("fingerprint"), p.Get("reviewer_comment"); fp != "" && c != "" && comments[fp] == "" {
			comments[fp] = c
		}
	}
	st, inherited, resolved := Carry(prior, res.Findings)
	for fp, s := range inherited {
		if status[fp] == "" {
			status[fp] = s
		}
	}
	return st, resolved, iterNo, nil
}
