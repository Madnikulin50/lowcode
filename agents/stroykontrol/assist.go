package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Ассистент при любом просмотре документа: вопросы по листу, который открыт на экране. Источник
// документа — сравнение ПД/РД, вложение Compose, документ Compose, документ комплекта или акт ИД;
// модель видит текст листа (и его картинку — для чертежей и сканов) и отвечает только по ним.

// assistSource is a document the assistant is asked about.
type assistSource struct {
	Source   string `json:"source"` // comparison | comparison_doc | attachment | cmp_document | document | act
	ID       string `json:"id"`
	RecordID string `json:"recordID"` // comparison and comparison_doc: the comparison the file belongs to
}

type assistContext struct {
	assistSource
	Page          int    `json:"page"`
	WithImage     bool   `json:"withImage"`
	DiscrepancyID string `json:"discrepancyID"` // comparison: the finding the question is about
	Scope         string `json:"scope"`         // comparison: finding | ref | target | both
	Ref           struct {
		Doc  string `json:"doc"`
		Page int    `json:"page"`
	} `json:"ref"`
	Target struct {
		Doc  string `json:"doc"`
		Page int    `json:"page"`
	} `json:"target"`
}

// loadedDoc is a file ready to be read page by page.
type loadedDoc struct {
	Data     []byte
	Mimetype string
	Name     string
	Title    string
	Extra    string // facts about the document that are not on its pages (an act's recognised fields)
}

func isImageMime(m string) bool { return strings.HasPrefix(m, "image/") }

// assistVisionTimeout bounds one call of the multimodal model (env ASSIST_VISION_TIMEOUT_SEC).
var assistVisionTimeout = func() time.Duration {
	if n, err := strconv.Atoi(os.Getenv("ASSIST_VISION_TIMEOUT_SEC")); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return 150 * time.Second
}()

var (
	sourceCacheMu sync.Mutex
	sourceCache   = map[string]cachedDoc{}
)

type cachedDoc struct {
	doc *loadedDoc
	at  time.Time
}

// loadSource downloads a document once per couple of minutes: a conversation asks about the same
// file again and again, and the viewer fetches its pages one by one.
func loadSource(ns string, s assistSource) (*loadedDoc, error) {
	key := ns + "|" + s.Source + "|" + s.RecordID + "|" + s.ID
	sourceCacheMu.Lock()
	if c, ok := sourceCache[key]; ok && time.Since(c.at) < 2*time.Minute {
		sourceCacheMu.Unlock()
		return c.doc, nil
	}
	sourceCacheMu.Unlock()
	doc, err := loadSourceUncached(ns, s)
	if err != nil {
		return nil, err
	}
	sourceCacheMu.Lock()
	if len(sourceCache) >= 8 {
		for k, c := range sourceCache {
			if time.Since(c.at) >= 2*time.Minute {
				delete(sourceCache, k)
			}
		}
		if len(sourceCache) >= 8 {
			sourceCache = map[string]cachedDoc{}
		}
	}
	sourceCache[key] = cachedDoc{doc, time.Now()}
	sourceCacheMu.Unlock()
	return doc, nil
}

// loadSourceUncached reads the document from Compose; the namespace comes from the request.
func loadSourceUncached(ns string, s assistSource) (*loadedDoc, error) {
	if s.Source == "comparison_doc" { // a file of a comparison (also legacy pd/rd files, also --fixtures mode)
		if s.RecordID == "" {
			return nil, fmt.Errorf("recordID is required")
		}
		data, mt, err := compose.FetchDocument(ns, s.RecordID, s.ID)
		if err != nil {
			return nil, err
		}
		doc := &loadedDoc{Data: data, Mimetype: mt, Name: s.ID}
		if docs, err := compose.Documents(ns, s.RecordID); err == nil {
			for _, d := range docs {
				if d.ID == s.ID {
					doc.Name, doc.Title = firstNonEmpty(d.Name, s.ID), strings.TrimSpace(d.VolumeCode+" "+d.VolumeTitle)
				}
			}
		}
		return doc, nil
	}
	cc, _ := compose.(*ComposeClient)
	if cc == nil {
		return nil, fmt.Errorf("документы Compose доступны только при работе с Compose (не в режиме --fixtures)")
	}
	switch s.Source {
	case "attachment":
		data, mt, err := cc.downloadOriginal(ns, s.ID)
		if err != nil {
			return nil, err
		}
		name := s.ID
		if att, err := cc.getAttachment(ns, s.ID); err == nil && att.Name != "" {
			name = att.Name
		}
		return &loadedDoc{Data: data, Mimetype: mt, Name: name}, nil
	case "document":
		data, mt, err := cc.downloadDocument(ns, s.ID)
		if err != nil {
			return nil, err
		}
		return &loadedDoc{Data: data, Mimetype: mt, Name: "документ " + s.ID}, nil
	case "cmp_document":
		rec, err := cc.getRecord(ns, cmp.ModDocuments, s.ID)
		if err != nil {
			return nil, err
		}
		att := rec.val("file")
		if att == "" {
			return nil, fmt.Errorf("у документа комплекта нет файла")
		}
		data, mt, err := cc.downloadOriginal(ns, att)
		if err != nil {
			return nil, err
		}
		name := firstNonEmpty(rec.val("volume_code"), s.ID)
		if a, err := cc.getAttachment(ns, att); err == nil && a.Name != "" {
			name = a.Name
		}
		return &loadedDoc{Data: data, Mimetype: mt, Name: name, Title: strings.TrimSpace(rec.val("volume_code") + " " + rec.val("volume_title"))}, nil
	case "act":
		rec, err := cc.getRecord(ns, idcheck.ModActs, s.ID)
		if err != nil {
			return nil, err
		}
		att := rec.val("attachment_id")
		if att == "" {
			return nil, fmt.Errorf("у акта нет вложения")
		}
		data, mt, err := cc.downloadOriginal(ns, att)
		if err != nil {
			return nil, err
		}
		extra := ""
		for _, f := range []struct{ k, label string }{{"act_type", "Тип акта"}, {"doc_number", "№ в акте"}, {"doc_date", "Дата в акте"}, {"p1_work", "п.1 Работы"}, {"p2_rd_code", "п.2 Шифр РД"}, {"p3_materials", "п.3 Материалы"}, {"p6_norms", "п.6 Нормы"}} {
			if v := rec.val(f.k); v != "" {
				extra += fmt.Sprintf("%s: %s\n", f.label, trimTo(v, 400))
			}
		}
		return &loadedDoc{Data: data, Mimetype: mt, Name: firstNonEmpty(rec.val("file_name"), "акт "+s.ID), Extra: strings.TrimSpace(extra)}, nil
	}
	return nil, fmt.Errorf("неизвестный источник документа %q", s.Source)
}

// sheetOf reads a page of a loaded file: its text, picture (for PDF and images) and the page count.
type sheetOf struct {
	Count     int
	Text      string
	Thin      bool
	ImagePath string
	ImageData []byte
}

func readSheet(docKey string, data []byte, mimetype string, page int) (*sheetOf, error) {
	switch {
	case cmp.IsPDF(data):
		pages, err := cmp.ExtractPages(docKey, data, cmpCacheDir)
		if err != nil {
			return nil, err
		}
		if len(pages) == 0 {
			return nil, fmt.Errorf("в файле нет страниц")
		}
		if page < 1 {
			page = 1
		}
		if page > len(pages) {
			page = len(pages)
		}
		var rows []string
		words := 0
		for _, l := range pages[page-1].Lines {
			rows = append(rows, l.Text)
			words += len(strings.Fields(l.Text))
		}
		sh := &sheetOf{Count: len(pages), Text: trimTo(strings.Join(rows, "\n"), 7000), Thin: pages[page-1].Drawing || words < 25}
		if dir, count, err := raster.Pages(data); err == nil {
			if p, err := raster.PagePath(dir, count, page); err == nil {
				sh.ImagePath = p
			}
		}
		return sh, nil
	case isImageMime(mimetype) || (len(data) > 8 && (string(data[:4]) == "\x89PNG" || string(data[:3]) == "\xff\xd8\xff")):
		return &sheetOf{Count: 1, Thin: true, ImageData: data}, nil
	}
	return nil, fmt.Errorf("ассистент читает PDF и изображения; этот тип файла (%s) не поддерживается", mimetype)
}

func (s *sheetOf) imageB64() string {
	data := s.ImageData
	if data == nil && s.ImagePath != "" {
		data, _ = os.ReadFile(s.ImagePath)
	}
	if len(data) == 0 || len(data) > 6<<20 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

// assistFactsFor builds the facts and the pictures for a request.
func assistFactsFor(ns string, c assistContext) (cmp.DialogFacts, []string, error) {
	var facts cmp.DialogFacts
	var images []string
	var sheets []*sheetOf

	if c.Source == "comparison" {
		if c.RecordID == "" {
			return facts, nil, fmt.Errorf("recordID is required")
		}
		var err error
		if c.Scope == "" || c.Scope == "finding" || c.DiscrepancyID != "" {
			facts, err = dialogFactsFor(ns, c.RecordID, c.DiscrepancyID)
		} else {
			facts, err = dialogFactsFor(ns, c.RecordID, "")
		}
		if err != nil {
			return facts, nil, err
		}
		docs, _ := compose.Documents(ns, c.RecordID)
		name := map[string]documentRecord{}
		for _, d := range docs {
			name[d.ID] = d
		}
		add := func(side, doc string, page int) error {
			if doc == "" || page < 1 {
				return nil
			}
			data, mt, err := compose.FetchDocument(ns, c.RecordID, doc)
			if err != nil {
				return err
			}
			sh, err := readSheet(doc, data, mt, page)
			if err != nil {
				return nil // a file the assistant cannot read is simply not in the facts
			}
			d := name[doc]
			facts.Pages = append(facts.Pages, cmp.DialogPage{Side: side, Doc: doc, Name: firstNonEmpty(d.Name, doc), Title: strings.TrimSpace(d.VolumeCode + " " + d.VolumeTitle),
				Page: min(max(page, 1), sh.Count), Count: sh.Count, Text: sh.Text, Thin: sh.Thin})
			sheets = append(sheets, sh)
			return nil
		}
		if c.Scope == "ref" || c.Scope == "both" {
			if err := add(cmp.SideRef, c.Ref.Doc, c.Ref.Page); err != nil {
				return facts, nil, err
			}
		}
		if c.Scope == "target" || c.Scope == "both" {
			if err := add(cmp.SideTarget, c.Target.Doc, c.Target.Page); err != nil {
				return facts, nil, err
			}
		}
	} else {
		ld, err := loadSource(ns, c.assistSource)
		if err != nil {
			return facts, nil, err
		}
		key := c.Source + "-" + c.RecordID + "-" + c.ID
		sh, err := readSheet(key, ld.Data, ld.Mimetype, c.Page)
		if err != nil {
			return facts, nil, err
		}
		page := min(max(c.Page, 1), sh.Count)
		facts.Title = ld.Name
		text := sh.Text
		if ld.Extra != "" {
			text += "\n\nУже извлечено из акта:\n" + ld.Extra
		}
		facts.Pages = []cmp.DialogPage{{Doc: key, Name: ld.Name, Title: ld.Title, Page: page, Count: sh.Count, Text: text, Thin: sh.Thin}}
		sheets = append(sheets, sh)
	}

	// the picture: for a drawing or a scan, or when the person asks for it
	drawing := false
	for i, p := range facts.Pages {
		if p.Thin {
			drawing = true
		}
		if p.Thin || c.WithImage {
			if b := sheets[i].imageB64(); b != "" {
				images = append(images, b)
			}
		}
	}
	if len(images) > 2 {
		images = images[:2]
	}
	switch {
	case len(images) > 0 && assistVision.Enabled():
		facts.ImageNote = "К сообщению приложено изображение открытого листа (листов): читай его вместе с текстом."
	case len(images) > 0:
		facts.ImageNote = "Изображение листа не передано: модель для изображений не настроена. Отвечай только по тексту и скажи, если для ответа нужен чертёж."
		images = nil
	}
	if drawing || c.WithImage {
		facts.Skill = cmp.LoadSkillText("read-drawings", 4500, assistSkillsDir, "agents/stroykontrol/skills")
	}
	return facts, images, nil
}

type assistBody struct {
	NamespaceID string        `json:"namespaceID"`
	Messages    []cmp.Msg     `json:"messages"`
	Context     assistContext `json:"context"`
	// Async: answer at once with a job ID and let the client poll /api/assist/result. A local model can
	// think for minutes, longer than the API gateway lets a request wait (30 s by default).
	Async bool `json:"async"`
}

type assistError struct {
	code int
	err  error
}

func (e *assistError) Error() string { return e.err.Error() }

// runAssist builds the facts, asks the model and returns the answer; it can take minutes.
func runAssist(ctx context.Context, ns string, body assistBody) (map[string]any, *assistError) {
	facts, images, err := assistFactsFor(ns, body.Context)
	if err != nil {
		code := http.StatusBadGateway
		if strings.Contains(err.Error(), "не поддерживается") || strings.Contains(err.Error(), "режиме --fixtures") {
			code = http.StatusUnprocessableEntity
		}
		return nil, &assistError{code, err}
	}
	msgs := append([]cmp.Msg{}, body.Messages...)
	llm := suggester
	note := ""
	if len(images) > 0 {
		for i := len(msgs) - 1; i >= 0; i-- {
			if msgs[i].Role == "user" {
				msgs[i].Images = images
				break
			}
		}
		llm = assistVision
	}
	callCtx := ctx
	if len(images) > 0 {
		// a picture model on a CPU can take many minutes: wait a limited time, then answer from the text
		var cancelVision context.CancelFunc
		callCtx, cancelVision = context.WithTimeout(ctx, assistVisionTimeout)
		defer cancelVision()
	}
	text, jumps, err := llm.Converse(callCtx, facts, msgs)
	if err != nil && len(images) > 0 {
		// the picture model is missing or failed: answer from the text of the sheet and say so
		for i := range msgs {
			msgs[i].Images = nil
		}
		facts.ImageNote = "Изображение листа не передано (модель для изображений недоступна). Отвечай только по тексту."
		note = "Изображение листа не удалось отдать модели (" + trimTo(err.Error(), 120) + ") — ответ только по тексту листа.\n\n"
		text, jumps, err = suggester.Converse(ctx, facts, msgs)
	}
	if err != nil {
		return nil, &assistError{http.StatusBadGateway, err}
	}
	if jumps == nil {
		jumps = []cmp.Jump{}
	}
	pages := make([]map[string]any, 0, len(facts.Pages))
	for _, p := range facts.Pages {
		pages = append(pages, map[string]any{"side": p.Side, "name": p.Name, "page": p.Page, "count": p.Count, "thin": p.Thin})
	}
	return map[string]any{"reply": note + text, "jumps": jumps, "pages": pages, "image": len(images) > 0, "skill": facts.Skill != ""}, nil
}

// --- jobs: answers that are collected by polling ---

type assistJob struct {
	started time.Time
	done    bool
	result  map[string]any
	err     string
}

var (
	assistJobsMu sync.Mutex
	assistJobs   = map[string]*assistJob{}
)

func newJobID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func startAssistJob(ns string, body assistBody) string {
	id := newJobID()
	job := &assistJob{started: time.Now()}
	assistJobsMu.Lock()
	for k, j := range assistJobs { // forget finished jobs after half an hour
		if j.done && time.Since(j.started) > 30*time.Minute {
			delete(assistJobs, k)
		}
	}
	assistJobs[id] = job
	assistJobsMu.Unlock()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
		defer cancel()
		res, aerr := runAssist(ctx, ns, body)
		assistJobsMu.Lock()
		defer assistJobsMu.Unlock()
		job.done = true
		if aerr != nil {
			job.err = aerr.Error()
		} else {
			job.result = res
		}
	}()
	return id
}

// handleAssist answers a question about the document on screen.
func handleAssist(w http.ResponseWriter, r *http.Request) {
	var body assistBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || len(body.Messages) == 0 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("namespaceID and messages are required"))
		return
	}
	if !suggester.Enabled() {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("модель не настроена (--remark-model / REMARK_MODEL): ассистент недоступен, остальное работает"))
		return
	}
	if body.Async {
		writeJSON(w, http.StatusAccepted, map[string]any{"jobID": startAssistJob(ns, body), "status": "running"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Minute)
	defer cancel()
	res, aerr := runAssist(ctx, ns, body)
	if aerr != nil {
		writeErr(w, aerr.code, aerr.err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// handleAssistResult returns the state of an asynchronous question: running, done (with the answer) or error.
func handleAssistResult(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("job")
	assistJobsMu.Lock()
	job := assistJobs[id]
	var out map[string]any
	switch {
	case job == nil:
		out = map[string]any{"status": "error", "error": "ответ не найден: агент мог перезапуститься — задайте вопрос ещё раз"}
	case !job.done:
		out = map[string]any{"status": "running", "elapsedSec": int(time.Since(job.started).Seconds())}
	case job.err != "":
		out = map[string]any{"status": "error", "error": job.err}
	default:
		out = map[string]any{"status": "done"}
		for k, v := range job.result {
			out[k] = v
		}
	}
	assistJobsMu.Unlock()
	writeJSON(w, http.StatusOK, out)
}

// handleAssistDoc tells the document viewer how many pages a source has.
func handleAssistDoc(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	s := assistSource{Source: r.URL.Query().Get("source"), ID: r.URL.Query().Get("id"), RecordID: r.URL.Query().Get("recordID")}
	if ns == "" || s.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("source, id and namespaceID are required"))
		return
	}
	ld, err := loadSource(ns, s)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	sh, err := readSheet(s.Source+"-"+s.ID, ld.Data, ld.Mimetype, 1)
	if err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": ld.Name, "title": ld.Title, "pageCount": sh.Count, "mimetype": ld.Mimetype})
}

// handleAssistPage serves one page of a source as an image.
func handleAssistPage(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	s := assistSource{Source: r.URL.Query().Get("source"), ID: r.URL.Query().Get("id"), RecordID: r.URL.Query().Get("recordID")}
	page, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if ns == "" || s.ID == "" || page < 1 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("source, id, namespaceID and n are required"))
		return
	}
	ld, err := loadSource(ns, s)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=300")
	if isImageMime(ld.Mimetype) {
		w.Header().Set("Content-Type", ld.Mimetype)
		_, _ = w.Write(ld.Data)
		return
	}
	if !cmp.IsPDF(ld.Data) {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("нет изображений страниц для этого типа файла"))
		return
	}
	dir, count, err := raster.Pages(ld.Data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	path, err := raster.PagePath(dir, count, page)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	http.ServeFile(w, r, path)
}
