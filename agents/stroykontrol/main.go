// Command stroykontrol-web is a small standalone agent service: it serves
// its own UI (embedded static files) for a document-comparison viewer that
// Compose embeds via an IFrame page block on the pd_rd_comparison record
// card. All comparison-specific visualization logic lives here, outside the
// lowcode platform itself — the platform only needed a working iframe block
// (see client3/web/compose/src/components/PageBlocks/IFrameBase.vue).
//
//	go run . --listen=:8092 --api=http://localhost:3333/compose --token=$TOKEN --namespace=<nsID>
package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed web/static
var webFS embed.FS

var (
	compose            store
	raster             *Rasterizer
	defaultNamespaceID string
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
	flag.Parse()

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
			log.Fatal("--token (or TOKEN env, or --token-file) is required (or pass --fixtures=<dir> to run standalone, see fixtures/README.md)")
		}
		compose = NewComposeClient(*api, *token, mintTokenViaNode)
	}
	raster = NewRasterizer(*cacheDir)
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
	})

	sub, err := fs.Sub(webFS, "web/static")
	if err != nil {
		log.Fatalf("embedded frontend: %v", err)
	}
	r.Handle("/*", http.FileServer(http.FS(sub)))

	log.Printf("stroykontrol-web listening on %s (api=%s ns=%s)", *listen, *api, *namespace)
	log.Fatal(http.ListenAndServe(*listen, r))
}

// mintTokenViaNode re-mints a Compose dev token the same way run.sh and
// mint-token.sh do at startup, so a long-lived agent process can recover
// once its initial token (short-lived, ~2h) expires instead of failing
// every request from then on (see ComposeClient.refreshToken). Best-effort:
// requires node + DB/auth access on this host, same as the scripts it
// mirrors, so it's a no-op-on-failure dev convenience, not a real refresh
// mechanism for a deployment with a dedicated service token.
func mintTokenViaNode() (string, error) {
	out, err := exec.Command("node", "-e",
		"import('./compose/helpers.mjs').then(h=>h.mintToken()).then(t=>process.stdout.write(t))",
	).Output()
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
