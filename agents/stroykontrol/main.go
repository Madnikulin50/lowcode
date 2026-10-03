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
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/madnikulin50/lowcode/agents/sdk"
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
	visionModel := flag.String("vision-model", envOr("IDCHECK_VISION_MODEL", "qwen3.5"), "Multimodal Ollama model used to read scanned acts (empty disables scan recognition)")
	tesseractBin := flag.String("tesseract", "tesseract", "tesseract binary for OCR of scanned acts (optional: without it pages are classified by the vision model)")
	ocrLang := flag.String("ocr-lang", "rus", "tesseract language")
	idcheckCache := flag.String("idcheck-cache", "var/idcheck-cache", "Directory for ИД check page images / OCR / model reply cache")
	maxAttached := flag.Int("max-attached", 6, "Attached quality-document pages per act sent to the vision model (-1 = all)")
	classifyPages := flag.Bool("classify-pages", false, "Without tesseract, classify every scanned page with the vision model (slow on CPU; default uses layout rules)")
	external := flag.Bool("external", true, "Check legal entities in ЕГРЮЛ (egrul.nalog.ru) and specialists in НРС (nrs.nostroy.ru)")
	staticDir := flag.String("static", "", "Serve the frontend from this directory instead of the embedded web/dist — point at web/dist while running `npx vite build --watch` to see frontend changes without rebuilding this binary")
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
		compose = NewComposeClient(*api, *token, mintTokenViaNode)
	}
	raster = NewRasterizer(*cacheDir)
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/comparison", handleComparison)
		r.Get("/pages", handlePages)
		r.Get("/page", handlePage)
		r.Get("/file", handleFile)
		r.Get("/text", handleText)
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
func mintTokenViaNode() (string, error) {
	cmd := exec.Command("node", "-e",
		"import('./compose/helpers.mjs').then(h=>h.mintToken()).then(t=>process.stdout.write(t))",
	)
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
		return "", fmt.Errorf("mint token via node: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
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

type discrepancyOut struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	PageNumber  string `json:"pageNumber"`
}

type comparisonOut struct {
	RecordID          string           `json:"recordID"`
	Title             string           `json:"title"`
	Status            string           `json:"status"`
	SimilarityPercent string           `json:"similarityPercent"`
	Comment           string           `json:"comment"`
	HasPdFile         bool             `json:"hasPdFile"`
	HasRdFile         bool             `json:"hasRdFile"`
	PdFileName        string           `json:"pdFileName"`
	RdFileName        string           `json:"rdFileName"`
	Discrepancies     []discrepancyOut `json:"discrepancies"`
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

	out := comparisonOut{
		RecordID:          recordID,
		Title:             rec.Title,
		Status:            rec.Status,
		SimilarityPercent: rec.SimilarityPercent,
		Comment:           rec.Comment,
		HasPdFile:         rec.PdFileID != "",
		HasRdFile:         rec.RdFileID != "",
		PdFileName:        compose.AttachmentName(ns, rec.PdFileID),
		RdFileName:        compose.AttachmentName(ns, rec.RdFileID),
		Discrepancies:     make([]discrepancyOut, 0, len(discs)),
	}
	for _, d := range discs {
		out.Discrepancies = append(out.Discrepancies, discrepancyOut{
			Type:        d.Type,
			Severity:    d.Severity,
			Description: d.Description,
			PageNumber:  d.PageNumber,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// fetchSideFile downloads the pd/rd file for a comparison record via
// whichever store backend is active (live Compose or local fixtures).
func fetchSideFile(ns, recordID, side string) (data []byte, mimetype string, err error) {
	return compose.FetchFile(ns, recordID, side)
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
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	side := r.URL.Query().Get("side")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}

	data, mimetype, err := fetchSideFile(ns, recordID, side)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
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

// handleFile streams the raw pd_file/rd_file attachment bytes so the browser
// can render it itself (currently: DOCX via docx-preview, client-side —
// there's no server-side page-image conversion without LibreOffice).
func handleFile(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	side := r.URL.Query().Get("side")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}
	data, mimetype, err := fetchSideFile(ns, recordID, side)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
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
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	side := r.URL.Query().Get("side")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}

	data, mimetype, err := fetchSideFile(ns, recordID, side)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
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
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	side := r.URL.Query().Get("side")
	pageStr := r.URL.Query().Get("n")
	page, _ := strconv.Atoi(pageStr)
	if ns == "" || recordID == "" || page < 1 {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID, namespaceID and n (>=1) are required"))
		return
	}

	data, mimetype, err := fetchSideFile(ns, recordID, side)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
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
