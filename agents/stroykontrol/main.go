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
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed web/static
var webFS embed.FS

var (
	compose            *ComposeClient
	raster             *Rasterizer
	defaultNamespaceID string
	comparisonModule   = "pd_rd_comparisons"
	discrepancyModule  = "pd_rd_discrepancies"
)

func main() {
	listen := flag.String("listen", ":8092", "HTTP listen address")
	api := flag.String("api", "http://localhost:3333/compose", "Compose API base URL")
	token := flag.String("token", "", "Compose API token")
	namespace := flag.String("namespace", "", "Default namespace ID (used if a request doesn't pass ?namespaceID=)")
	cacheDir := flag.String("cache", "var/stroykontrol-raster-cache", "Directory for rasterized page cache")
	flag.Parse()

	if *token == "" {
		*token = os.Getenv("TOKEN")
	}
	if *token == "" {
		log.Fatal("--token (or TOKEN env) is required")
	}

	compose = NewComposeClient(*api, *token)
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

// attachmentName is a best-effort lookup: the header shows it if available,
// but falls back to the plain "ПД"/"РД" label client-side if this fails —
// it's a label, not something worth failing the whole request over.
func attachmentName(ns, attID string) string {
	if attID == "" {
		return ""
	}
	att, err := compose.getAttachment(ns, attID)
	if err != nil {
		log.Printf("attachment name lookup failed for %s: %v", attID, err)
		return ""
	}
	return att.Name
}

func handleComparison(w http.ResponseWriter, r *http.Request) {
	ns := nsFromQuery(r)
	recordID := r.URL.Query().Get("recordID")
	if ns == "" || recordID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("recordID and namespaceID are required"))
		return
	}

	rec, err := compose.getRecord(ns, comparisonModule, recordID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}

	discs, err := compose.searchRecords(ns, discrepancyModule, fmt.Sprintf("comparison = '%s'", recordID))
	if err != nil {
		log.Printf("discrepancies lookup failed: %v", err)
		discs = nil
	}

	pdFileID := rec.val("pd_file")
	rdFileID := rec.val("rd_file")
	out := comparisonOut{
		RecordID:          recordID,
		Title:             rec.val("title"),
		Status:            rec.val("status"),
		SimilarityPercent: rec.val("similarity_percent"),
		Comment:           rec.val("comment"),
		HasPdFile:         pdFileID != "",
		HasRdFile:         rdFileID != "",
		PdFileName:        attachmentName(ns, pdFileID),
		RdFileName:        attachmentName(ns, rdFileID),
		Discrepancies:     make([]discrepancyOut, 0, len(discs)),
	}
	for _, d := range discs {
		out.Discrepancies = append(out.Discrepancies, discrepancyOut{
			Type:        d.val("discrepancy_type"),
			Severity:    d.val("severity"),
			Description: d.val("description"),
			PageNumber:  d.val("page_number"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// fetchSideFile downloads the pd_file/rd_file attachment for a comparison
// record and reports whether it's PDF (rasterizable).
func fetchSideFile(ns, recordID, side string) (data []byte, mimetype string, err error) {
	field := "pd_file"
	if side == "rd" {
		field = "rd_file"
	} else if side != "pd" {
		return nil, "", fmt.Errorf("side must be 'pd' or 'rd'")
	}
	rec, err := compose.getRecord(ns, comparisonModule, recordID)
	if err != nil {
		return nil, "", err
	}
	attID := rec.val(field)
	if attID == "" {
		return nil, "", fmt.Errorf("record has no %s attachment", field)
	}
	return compose.downloadOriginal(ns, attID)
}

// pagesOut.Kind tells the viewer how to get page images for a side:
//   - "pdf":  server rasterizes: PageCount is already known, fetch each page
//     from /api/page.
//   - "docx": no page-image concept server-side (no LibreOffice on this
//     host); PageCount is 0 — the browser fetches the raw file from
//     /api/file and paginates it client-side (docx-preview + html2canvas).
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
