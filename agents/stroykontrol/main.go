// Command stroykontrol-web is a small standalone agent service: it serves
// its own UI (embedded static files) for a document-comparison viewer that
// Compose embeds via an IFrame page block on the pd_rd_comparison record
// card. All comparison-specific visualization logic lives here, outside the
// lowcode platform itself — the platform only needed a working iframe block
// (see client3/web/compose/src/components/PageBlocks/IFrameBase.vue).
//
//	go run . --listen=:8092 --api=http://localhost:3333/compose --token=$TOKEN --namespace=<nsID>
//
// Or skip --token/$TOKEN entirely with AGENT_SHARED_SECRET set the same in
// this process's env and the server's (self-enrolls via POST /agents/enroll,
// see agents/sdk/authtoken.go).
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/madnikulin50/lowcode/agents/sdk"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

//go:embed web/dist
var webFS embed.FS

var (
	compose            store
	raster             *Rasterizer
	defaultNamespaceID string
	idRunner           *idcheck.Runner
	idBackend          idcheck.Backend
	idCacheDir         string
	assistVision       *cmp.Suggester // multimodal model for pictures of sheets (--vision-model)
	assistSkillsDir    string
	comparisonModule   = "pd_rd_comparisons"
	discrepancyModule  = "pd_rd_discrepancies"
)

func main() {
	listen := flag.String("listen", ":8092", "HTTP listen address")
	api := flag.String("api", "http://localhost:3333/compose", "Compose API base URL")
	token := flag.String("token", "", "Compose API token")
	tokenFile := flag.String("token-file", "", "Read the Compose API token from this file (trimmed) if --token/TOKEN are both empty — refreshed by mint-token.sh before each run, meant for a GoLand \"Before launch\" step (see README's GoLand section) rather than manual copy-pasting")
	namespace := flag.String("namespace", "", "Default namespace ID (used if a request doesn't pass ?namespaceID=)")
	cacheDir := flag.String("cache", "var/stroykontrol-raster-cache", "Directory for rasterized page cache")
	fixtures := flag.String("fixtures", "", "Run standalone against a local fixtures directory instead of a live Compose instance (see fixtures/README.md) — skips --api/--token/--namespace entirely")
	ollamaURL := flag.String("ollama", envOr("OLLAMA_URL", "http://localhost:11434"), "Ollama base URL for the ИД check vision model")
	visionModel := flag.String("vision-model", envOr("IDCHECK_VISION_MODEL", "deepseek-v2"), "Multimodal Ollama model used to read scanned acts (empty disables scan recognition)")
	tesseractBin := flag.String("tesseract", "tesseract", "tesseract binary for OCR of scanned acts (optional: without it pages are classified by the vision model)")
	ocrLang := flag.String("ocr-lang", "rus", "tesseract language")
	idcheckCache := flag.String("idcheck-cache", "var/idcheck-cache", "Directory for ИД check page images / OCR / model reply cache")
	maxAttached := flag.Int("max-attached", 6, "Attached quality-document pages per act sent to the vision model (-1 = all)")
	classifyPages := flag.Bool("classify-pages", false, "Without tesseract, classify every scanned page with the vision model (slow on CPU; default uses layout rules)")
	external := flag.Bool("external", true, "Check legal entities in ЕГРЮЛ (egrul.nalog.ru) and specialists in НРС (nrs.nostroy.ru)")
	remarkModel := flag.String("remark-model", envOr("REMARK_MODEL", "deepseek-v2:latest"), "Ollama text model that drafts remark-card wording on request (empty disables the button)")
	staticDir := flag.String("static", "", "Serve the frontend from this directory instead of the embedded web/dist — point at web/dist while running `npx vite build --watch` to see frontend changes without rebuilding this binary")
	smtpAddr := flag.String("smtp", envOr("SMTP_ADDR", ""), "SMTP server host:port for the notifications by e-mail (empty — the feed only)")
	smtpUser := flag.String("smtp-user", envOr("SMTP_USER", ""), "SMTP login")
	smtpFrom := flag.String("smtp-from", envOr("SMTP_FROM", ""), "Sender of the notification letters, e.g. «Стройконтроль <noreply@example.ru>»")
	smtpTLS := flag.String("smtp-tls", envOr("SMTP_TLS", "starttls"), "SMTP encryption: starttls | tls | none")
	publicURL := flag.String("public-url", envOr("PUBLIC_URL", ""), "Base URL of this agent's UI as users open it — for the link in a notification letter")
	notifyEvery := flag.Duration("notify-every", 15*time.Minute, "How often the agent looks for new notifications (0 — only on request and after a hand-made change)")
	notifySoon := flag.Int("notify-soon-days", 3, "Notify that a due date is near this many days before it")
	dumpNorms := flag.Bool("dump-norms", false, "Print the built-in regulatory documents catalog (idcheck.DefaultNorms) as JSON and exit — used by compose_id/apply_id.mjs to seed «Нормативные документы»")
	flag.Parse()
	if *dumpNorms {
		_ = json.NewEncoder(os.Stdout).Encode(idcheck.DefaultNorms)
		return
	}

	if *fixtures != "" {
		fs, err := NewFixtureStore(*fixtures)
		if err != nil {
			log.Fatal(err)
		}
		compose = fs
		log.Printf("stroykontrol-web running standalone against fixtures at %s (no Compose backend)", *fixtures)
	} else {
		if *token == "" {
			*token = os.Getenv("TOKEN")
		}
		if *token == "" && *tokenFile != "" {
			b, err := os.ReadFile(*tokenFile)
			if err != nil {
				log.Fatalf("--token-file: %v (run mint-token.sh first, or pass --token/TOKEN directly)", err)
			}
			*token = strings.TrimSpace(string(b))
		}
		if *token == "" {
			// Self-enroll via AGENT_SHARED_SECRET, same as backup/cmdb/invest
			// (see agents/sdk/authtoken.go) — no manual mint-token.sh step.
			// The enroll endpoint is mounted on the server's root router,
			// not under /api or /compose (see mountAgentEnroll in
			// server/app/servers.go, called before the /api route is
			// mounted) — so this needs the bare origin, whatever path
			// suffix --api happens to carry.
			if root, err := rootOrigin(*api); err != nil {
				log.Printf("self-enroll: could not derive server origin from --api=%q: %v", *api, err)
			} else {
				*token = sdk.SelfToken(root)
			}
		}
		if *token == "" {
			log.Fatal("--token (or TOKEN env, AGENT_SHARED_SECRET, or --token-file) is required (or pass --fixtures=<dir> to run standalone, see fixtures/README.md)")
		}
		// Prefer re-enrolling via AGENT_SHARED_SECRET (works anywhere the
		// agent can reach the server); fall back to the node dev helper.
		remint := mintTokenViaNode
		if strings.TrimSpace(os.Getenv("AGENT_SHARED_SECRET")) != "" {
			if root, err := rootOrigin(*api); err == nil {
				remint = func() (string, error) {
					if t, err := sdk.EnrollToken(root); err == nil {
						return t, nil
					} else if t2, err2 := mintTokenViaNode(); err2 == nil {
						return t2, nil
					} else {
						return "", fmt.Errorf("enroll: %v; node: %v", err, err2)
					}
				}
			}
		}
		cc := NewComposeClient(*api, *token, nil)
		cc.remint = withTokenFileFallback(remint, *tokenFile, cc.getToken)
		compose = cc
	}
	raster = NewRasterizer(*cacheDir)
	cmpCacheDir = *cacheDir
	cmpOCR = cmp.NewOCR(*tesseractBin, *ocrLang, *cacheDir)
	if !cmpOCR.Enabled() {
		log.Printf("cmp: tesseract (%s) not found — scanned pages of ПД/РД are compared as images only", *tesseractBin)
	}
	suggester = cmp.NewSuggester(*ollamaURL, *remarkModel)
	// the assistant looks at page pictures with the same multimodal model that reads scanned acts
	assistVision = cmp.NewSuggester(*ollamaURL, *visionModel)
	assistSkillsDir = firstNonEmpty(os.Getenv("AI_SKILLS_DIR"), "skills")
	if fx, ok := compose.(*FixtureStore); ok {
		fx.analyze = func(in []cmp.Input, mode string, pairing []cmp.ManualPair, excluded []cmp.Region) (*cmp.Result, error) {
			return cmp.AnalyzeWith(in, cmpCacheDir, cmp.AnalyzeOptions{Rasterize: rasterPage, OCR: cmpOCR, Mode: mode, Pairing: pairing, Exclude: excluded})
		}
	}
	if cc, ok := compose.(*ComposeClient); ok {
		ocr := idcheck.NewTesseract(*tesseractBin, *ocrLang)
		if !ocr.Available() {
			log.Printf("idcheck: tesseract (%s, lang %s) not found — scanned pages will be classified by the vision model only", *tesseractBin, *ocrLang)
		}
		var ext *idcheck.External
		if *external {
			ext = idcheck.NewExternal()
		}
		idBackend, idCacheDir = composeBackend{cc}, *idcheckCache
		cmpRunner = cmp.NewRunner(composeBackend{cc}, cmpCacheDir, rasterPage)
		cmpRunner.OCR = cmpOCR
		cmpRunner.LLM = suggester
		idRunner = idcheck.NewRunner(idBackend, &idcheck.Extractor{
			CacheDir:      *idcheckCache,
			DPI:           150,
			OCR:           ocr,
			Vision:        idcheck.NewVision(*ollamaURL, *visionModel, *idcheckCache, 0),
			MaxAttached:   *maxAttached,
			ClassifyPages: *classifyPages,
		}, ext)
	}
	defaultNamespaceID = *namespace
	if cmpRunner != nil {
		notif = &notifier{b: cmpRunner.Backend, opt: cmp.NotifyOptions{SoonDays: *notifySoon},
			cfg: smtpConfig{Addr: *smtpAddr, User: *smtpUser, Password: os.Getenv("SMTP_PASSWORD"), From: *smtpFrom, TLS: *smtpTLS, Link: *publicURL}}
		notif.opt.Channel = notif.cfg.enabled()
		if notif.cfg.enabled() {
			log.Printf("notify: e-mail via %s from %s", *smtpAddr, *smtpFrom)
		} else {
			log.Printf("notify: no SMTP (--smtp/--smtp-from) — notifications go to the feed only")
		}
		go notif.run(defaultNamespaceID, *notifyEvery)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/comparison", handleComparison)
		r.Get("/pages", handlePages)
		r.Get("/sheets", handleSheets)
		r.Get("/page", handlePage)
		r.Get("/file", handleFile)
		r.Get("/text", handleText)
		r.Post("/cmp/run", handleCmpRun)
		r.Get("/cmp/run", handleCmpRun)
		r.Get("/cmp/status", handleCmpStatus)
		r.Post("/discrepancy/status", handleDiscrepancyStatus)
		r.Post("/expertise/import", handleExpertiseImport)
		r.Get("/expertise/import", handleExpertiseImport)
		r.Post("/expertise/check", handleExpertiseCheck)
		r.Get("/sets/overview", handleSetsOverview)
		r.Post("/sets/check", handleSetsCheck)
		r.Get("/sets/check", handleSetsCheck)
		r.Post("/volumes/suggest", handleVolumesSuggest)
		r.Post("/volumes/resolve", handleVolumesResolve)
		r.Get("/remarks/board", handleRemarksBoard)
		r.Get("/remarks/history", handleRemarkHistory)
		r.Get("/notifications", handleNotifications)
		r.Post("/notifications/read", handleNotificationsRead)
		r.Post("/notifications/scan", handleNotificationsScan)
		r.Get("/remarks/export", handleRemarksExport)
		r.Get("/reports/index", handleReportIndex)
		r.Post("/remarks/update", handleRemarkUpdate)
		r.Post("/remarks/requirements", handleRemarkRequirements)
		r.Post("/remarks/ai", handleRemarkAI)
		r.Get("/issue/preview", handleIssuePreview)
		r.Post("/issue/register", handleIssueRegister)
		r.Get("/issue/register", handleIssueRegister)
		r.Get("/issue/export", handleIssueExport)
		r.Get("/report/export", handleReportExport)
		r.Get("/pairing", handlePairingGet)
		r.Post("/pairing", handlePairingSet)
		r.Post("/exclusions", handleExclusionsSet)
		r.Post("/remark", handleRemarkCreate)
		r.Post("/remark/save", handleRemarkSave)
		r.Post("/remark/suggest", handleRemarkSuggest)
		r.Post("/dialog", handleDialog)
		r.Post("/assist", handleAssist)
		r.Get("/assist/result", handleAssistResult)
		r.Get("/assist/doc", handleAssistDoc)
		r.Get("/assist/page", handleAssistPage)
		r.Post("/idcheck/run", handleIDCheckRun)
		r.Get("/idcheck/run", handleIDCheckRun)
		r.Get("/idcheck/status", handleIDCheckStatus)
		r.Get("/idcheck/view", handleIDCheckView)
		r.Get("/idcheck/page", handleIDCheckPage)
	})

	var static http.Handler
	if *staticDir != "" {
		static = http.FileServer(http.Dir(*staticDir))
	} else {
		sub, err := fs.Sub(webFS, "web/dist")
		if err != nil {
			log.Fatalf("embedded frontend: %v", err)
		}
		static = http.FileServer(http.FS(sub))
	}
	// No Cache-Control was set here at all before, so http.FileServer sent
	// none either — browsers then fall back to heuristic caching for a
	// same-URL GET with no cache directives, which can keep serving a
	// pre-rebuild copy of index.html (loaded earlier in the same tab/iframe,
	// e.g. via the Compose IFrame block) after this binary is rebuilt and
	// restarted, with nothing about the request/response visibly wrong.
	// This is a small, frequently-iterated dev UI embedded in an iframe, not
	// a CDN-fronted asset bundle, so unconditionally disabling caching costs
	// nothing and removes that whole class of "I restarted but nothing
	// changed" confusion.
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		static.ServeHTTP(w, req)
	}))

	log.Printf("stroykontrol-web listening on %s (api=%s ns=%s)", *listen, *api, *namespace)
	log.Fatal(http.ListenAndServe(*listen, r))
}

// rootOrigin strips whatever path --api carries (/compose, /api/compose,
// ...) down to just scheme://host[:port], since the agent self-enrollment
// endpoint lives on the server's root router, not under the Compose API path.
func rootOrigin(apiURL string) (string, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return "", err
	}
	u.Path, u.RawQuery, u.Fragment = "", "", ""
	return u.String(), nil
}

// mintTokenViaNode re-mints a Compose dev token the same way run.sh and
// mint-token.sh do at startup, so a long-lived agent process can recover
// once its initial token (short-lived, ~2h) expires instead of failing
// every request from then on (see ComposeClient.refreshToken). Best-effort:
// requires node + DB/auth access on this host, same as the scripts it
// mirrors, so it's a no-op-on-failure dev convenience, not a real refresh
// mechanism for a deployment with a dedicated service token.
// findNode locates node even when the agent runs from an IDE: a non-login shell has no nvm on PATH.
func findNode() (string, error) {
	if p, err := exec.LookPath("node"); err == nil {
		return p, nil
	}
	var cands []string
	if home, err := os.UserHomeDir(); err == nil {
		matches, _ := filepath.Glob(filepath.Join(home, ".nvm", "versions", "node", "*", "bin", "node"))
		sort.Strings(matches)
		for i := len(matches) - 1; i >= 0; i-- { // newest version first
			cands = append(cands, matches[i])
		}
	}
	cands = append(cands, "/usr/local/bin/node", "/usr/bin/node", "/snap/bin/node")
	for _, c := range cands {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("node не найден ни в PATH, ни в ~/.nvm (нужен для обновления токена: compose/helpers.mjs)")
}

// agentDir is the directory that holds compose/helpers.mjs (the agent's own folder).
func agentDir() (string, error) {
	var cands []string
	if wd, err := os.Getwd(); err == nil {
		cands = append(cands, wd)
	}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Dir(exe), filepath.Dir(filepath.Dir(exe)))
	}
	for _, d := range cands {
		if _, err := os.Stat(filepath.Join(d, "compose", "helpers.mjs")); err == nil {
			return d, nil
		}
	}
	return "", fmt.Errorf("не найден compose/helpers.mjs: запустите агент из каталога agents/stroykontrol")
}

func mintTokenViaNode() (string, error) {
	node, err := findNode()
	if err != nil {
		return "", err
	}
	dir, err := agentDir()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(node, "-e",
		"import('./compose/helpers.mjs').then(h=>h.mintToken()).then(t=>process.stdout.write(t))",
	)
	cmd.Dir = dir
	// mintToken() returns $TOKEN as-is when set — and the agent itself is
	// usually started with TOKEN=…, so the child would hand back the very
	// token that just expired. Drop it to force a real re-mint.
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "TOKEN=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	out, err := cmd.Output()
	if err != nil {
		msg := ""
		if ee, ok := err.(*exec.ExitError); ok {
			msg = strings.TrimSpace(string(ee.Stderr))
		}
		return "", fmt.Errorf("mint token via node: %w %s", err, trimTo(msg, 200))
	}
	return strings.TrimSpace(string(out)), nil
}

// withTokenFileFallback tries remint and, when it fails, takes a fresher token from --token-file
// (an IDE "before launch" step or a person may have refreshed it) — as long as it is not the expired one.
func withTokenFileFallback(remint func() (string, error), file string, current func() string) func() (string, error) {
	return func() (string, error) {
		t, err := remint()
		if err == nil && t != "" {
			return t, nil
		}
		if file != "" {
			if b, rerr := os.ReadFile(file); rerr == nil {
				if ft := strings.TrimSpace(string(b)); ft != "" && ft != current() {
					return ft, nil
				}
			}
		}
		if err == nil {
			err = fmt.Errorf("пустой токен")
		}
		return "", err
	}
}

func nsFromQuery(r *http.Request) string {
	if ns := r.URL.Query().Get("namespaceID"); ns != "" {
		return ns
	}
	return defaultNamespaceID
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

type comparisonOut struct {
	RecordID          string              `json:"recordID"`
	Title             string              `json:"title"`
	Status            string              `json:"status"`
	SimilarityPercent string              `json:"similarityPercent"`
	Comment           string              `json:"comment"`
	Running           string              `json:"running"`
	SetsLabel         string              `json:"setsLabel,omitempty"` // "ПД v2 ↔ РД v1"
	Kind              string              `json:"kind"`                // stages | versions
	RefName           string              `json:"refName"`             // "ПД" | "До"
	TargetName        string              `json:"targetName"`          // "РД" | "После"
	Iteration         *iterationOut       `json:"iteration,omitempty"`
	ObjectID          string              `json:"objectID"`
	Documents         []documentRecord    `json:"documents"`
	Discrepancies     []discrepancyRecord `json:"discrepancies"`
	Remarks           []cmp.Remark        `json:"remarks"`
	Expertise         []remarkRecord      `json:"expertise"`
}

// iterationOut summarizes a re-check of the same pair after corrections.
type iterationOut struct {
	No          string `json:"no"`
	ParentID    string `json:"parentID"`
	ParentTitle string `json:"parentTitle"`
	Resolved    int    `json:"resolved"`
	Persisting  int    `json:"persisting"`
	New         int    `json:"new"`
}

func iterationSummary(rec *comparisonRecord, discs []discrepancyRecord) *iterationOut {
	if rec.ParentID == "" {
		return nil
	}
	it := &iterationOut{No: rec.IterationNo, ParentID: rec.ParentID, ParentTitle: rec.ParentTitle}
	for _, d := range discs {
		switch d.IterState {
		case "resolved":
			it.Resolved++
		case "persisting":
			it.Persisting++
		case "new":
			it.New++
		}
	}
	return it
}

func handleComparison(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}

	rec, err := compose.Comparison(ns, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	discs, err := compose.Discrepancies(ns, recordID)
	if err != nil {
		log.Printf("discrepancies lookup failed: %v", err)
		discs = nil
	}
	docs, err := compose.Documents(ns, recordID)
	if err != nil {
		log.Printf("documents lookup failed: %v", err)
		docs = nil
	}
	remarks, err := compose.Remarks(ns, recordID)
	if err != nil {
		log.Printf("remarks lookup failed: %v", err)
		remarks = nil
	}

	expertise, err := compose.Expertise(ns, recordID)
	if err != nil {
		log.Printf("expertise remarks lookup failed: %v", err)
		expertise = nil
	}

	out := comparisonOut{
		RecordID:          recordID,
		Title:             rec.Title,
		Status:            rec.Status,
		SimilarityPercent: rec.SimilarityPercent,
		Comment:           rec.Comment,
		Documents:         docs,
		Discrepancies:     discs,
		Remarks:           remarks,
		Expertise:         expertise,
		ObjectID:          rec.ObjectID,
		SetsLabel:         rec.SetsLabel,
		Kind:              modeOf(rec.Kind),
	}
	out.RefName, out.TargetName = cmp.SideNames(out.Kind)
	if out.Expertise == nil {
		out.Expertise = []remarkRecord{}
	}
	if out.Documents == nil {
		out.Documents = []documentRecord{}
	}
	if out.Discrepancies == nil {
		out.Discrepancies = []discrepancyRecord{}
	}
	if out.Remarks == nil {
		out.Remarks = []cmp.Remark{}
	}
	out.Iteration = iterationSummary(rec, discs)
	if cmpRunner != nil {
		out.Running = cmpRunner.Status(recordID)
	}
	writeJSON(w, http.StatusOK, out)
}

// fetchDoc downloads the document a request names (?doc=, or legacy ?side=).
func fetchDoc(r *http.Request) (data []byte, mimetype string, status int, err error) {
	ns, recordID, doc, err := resolveDoc(r)
	if err != nil {
		return nil, "", http.StatusBadRequest, err
	}
	data, mimetype, err = compose.FetchDocument(ns, recordID, doc)
	if err != nil {
		return nil, "", http.StatusBadGateway, err
	}
	return data, mimetype, http.StatusOK, nil
}

// pagesOut.Kind tells the viewer how to get page images for a side:
//   - "pdf":  server rasterizes: PageCount is already known, fetch each page
//     from /api/page.
//   - "docx": no page-image concept server-side (no LibreOffice on this
//     host); PageCount is 0 — the browser fetches the raw file from
//     /api/file and paginates it client-side (docx-preview + html2canvas).
//   - "dxf":  same client-side pattern as docx, no CAD rasterizer on this
//     host either — the browser fetches /api/file and draws the entities
//     (LINE/LWPOLYLINE/CIRCLE/ARC/TEXT/MTEXT) onto a canvas itself. Always
//     one page (a DXF's ENTITIES section is one model-space sheet, no
//     multi-sheet/layout concept here).
//   - "xlsx": same client-side pattern — the browser fetches /api/file,
//     parses the workbook and draws each sheet as a grid, split into pages
//     of a fixed number of rows so long sheets stay pageable.
//   - "other": genuinely unsupported, falls back to the text-only diff list.
type pagesOut struct {
	Supported bool   `json:"supported"` // kept for older clients: true unless kind=="other"
	Kind      string `json:"kind"`
	PageCount int    `json:"pageCount"`
}

func handlePages(w http.ResponseWriter, r *http.Request) {
	data, mimetype, status, err := fetchDoc(r)
	if err != nil {
		writeErr(w, status, err)
		return
	}
	switch {
	case IsPDF(mimetype, data):
		_, count, err := raster.Pages(data)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, pagesOut{Supported: true, Kind: "pdf", PageCount: count})
	case IsDOCX(mimetype, data):
		writeJSON(w, http.StatusOK, pagesOut{Supported: true, Kind: "docx", PageCount: 0})
	case IsDXF(mimetype, data):
		writeJSON(w, http.StatusOK, pagesOut{Supported: true, Kind: "dxf", PageCount: 0})
	case IsXLSX(mimetype, data):
		writeJSON(w, http.StatusOK, pagesOut{Supported: true, Kind: "xlsx", PageCount: 0})
	default:
		writeJSON(w, http.StatusOK, pagesOut{Supported: false, Kind: "other"})
	}
}

// handleSheets serves the sheet numbers read from the title blocks of a PDF: {"sheets": {"<page>": <sheet>}}.
// Asked for apart from /api/pages: reading the title blocks needs the text layer with positions, which
// is cached by the file's content but takes seconds the first time.
func handleSheets(w http.ResponseWriter, r *http.Request) {
	data, mimetype, status, err := fetchDoc(r)
	if err != nil {
		writeErr(w, status, err)
		return
	}
	out := map[string]int{}
	if IsPDF(mimetype, data) {
		pages, err := cmp.ExtractPages("sheets", data, cmpCacheDir)
		if err != nil {
			writeErr(w, http.StatusUnprocessableEntity, err)
			return
		}
		for pg, sh := range cmp.SheetMap(pages) {
			out[strconv.Itoa(pg)] = sh
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sheets": out})
}

// handleFile streams the raw pd_file/rd_file attachment bytes so the browser
// can render it itself (currently: DOCX via docx-preview, client-side —
// there's no server-side page-image conversion without LibreOffice).
func handleFile(w http.ResponseWriter, r *http.Request) {
	data, mimetype, status, err := fetchDoc(r)
	if err != nil {
		writeErr(w, status, err)
		return
	}
	if mimetype == "" {
		mimetype = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimetype)
	w.Header().Set("Cache-Control", "private, max-age=300")
	_, _ = w.Write(data)
}

type textOut struct {
	Pages []string `json:"pages"`
}

// handleText serves per-page extracted text for the "page-by-page text
// comparison" view mode. PDF only: DOCX text comes from the client's own
// docx-preview render (it already has the page-split DOM, no round trip
// needed). Callers should treat a non-PDF side as "no server text" rather
// than an error — the client only calls this for kind=="pdf".
func handleText(w http.ResponseWriter, r *http.Request) {
	data, mimetype, status, err := fetchDoc(r)
	if err != nil {
		writeErr(w, status, err)
		return
	}
	if !IsPDF(mimetype, data) {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("text endpoint only supports PDF; render DOCX client-side"))
		return
	}
	pages, err := raster.Text(data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, textOut{Pages: pages})
}

func handlePage(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("n"))
	if page < 1 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("n (>=1) is required"))
		return
	}
	data, mimetype, status, err := fetchDoc(r)
	if err != nil {
		writeErr(w, status, err)
		return
	}
	if !IsPDF(mimetype, data) {
		writeErr(w, http.StatusUnprocessableEntity, fmt.Errorf("not a PDF, no page images available"))
		return
	}
	dir, count, err := raster.Pages(data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err)
		return
	}
	path, err := raster.PagePath(dir, count, page)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=300")
	http.ServeFile(w, r, path)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// handleIDCheckRun queues an ИД package check. Called by the Compose rule
// chain "stroykontrol-id-run" (http node) with ?packageID=&namespaceID= or
// a JSON body {"packageID": "...", "namespaceID": "..."}; returns at once —
// the run writes its progress into the package record.
func handleIDCheckRun(w http.ResponseWriter, r *http.Request) {
	if idRunner == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("ИД check needs a live Compose backend (not available with --fixtures)"))
		return
	}
	var body struct {
		PackageID   string `json:"packageID"`
		NamespaceID string `json:"namespaceID"`
	}
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	pkg := strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("packageID"), body.PackageID))
	ns := strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("namespaceID"), body.NamespaceID, defaultNamespaceID))
	if pkg == "" || ns == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("packageID and namespaceID are required"))
		return
	}
	queued := idRunner.Enqueue(ns, pkg)
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"packageID": pkg, "queued": queued, "status": idRunner.Status(pkg)})
}

func handleIDCheckStatus(w http.ResponseWriter, r *http.Request) {
	if idRunner == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("ИД check not available"))
		return
	}
	pkg := r.URL.Query().Get("packageID")
	writeJSON(w, http.StatusOK, map[string]interface{}{"packageID": pkg, "status": idRunner.Status(pkg)})
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

// handleIDCheckView returns everything the in-place findings viewer needs
// (?namespaceID=&packageID= or &actID=).
func handleIDCheckView(w http.ResponseWriter, r *http.Request) {
	if idBackend == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("ИД check not available"))
		return
	}
	q := r.URL.Query()
	v, err := idcheck.BuildView(idBackend, firstNonEmpty(q.Get("namespaceID"), defaultNamespaceID), q.Get("packageID"), q.Get("actID"), q.Get("findingID"))
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

// handleIDCheckPage serves one rendered page of an act scan as JPEG
// (?namespaceID=&actID=&n=).
func handleIDCheckPage(w http.ResponseWriter, r *http.Request) {
	if idBackend == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("ИД check not available"))
		return
	}
	q := r.URL.Query()
	n, _ := strconv.Atoi(q.Get("n"))
	path, _, err := idcheck.ActPageImage(r.Context(), idBackend, idCacheDir, firstNonEmpty(q.Get("namespaceID"), defaultNamespaceID), q.Get("actID"), n, 150)
	if err != nil {
		writeErr(w, http.StatusNotFound, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeFile(w, r, path)
}
