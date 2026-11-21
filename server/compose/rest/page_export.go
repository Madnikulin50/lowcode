package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-chi/chi/v5"
	"github.com/madnikulin50/lowcode/server/pkg/auth"
)

const (
	pageExportTokenTTL = 3 * time.Minute
	pageExportTimeout  = 90 * time.Second
	pageExportSettle   = 5 * time.Second
)

type (
	PageExport struct{}

	// pageExportData is a page snapshot taken from the rendered DOM in print mode
	pageExportData struct {
		Title    string              `json:"title"`
		Sections []pageExportSection `json:"sections"`
	}

	pageExportSection struct {
		Tab     string     `json:"tab"`
		Title   string     `json:"title"`
		Headers []string   `json:"headers"`
		Rows    [][]string `json:"rows"`
		// Total is the number of records in the list, including records that are not loaded
		Total int `json:"total"`
	}
)

func MountPageExportRoutes(r chi.Router) {
	r.Get("/namespace/{namespaceID}/page/{pageID}/export", PageExport{}.Export)
	r.Get("/namespace/{namespaceID}/page/{pageID}/export-schedule", PageExport{}.ScheduleGet)
	r.Put("/namespace/{namespaceID}/page/{pageID}/export-schedule", PageExport{}.ScheduleSet)
	r.Post("/namespace/{namespaceID}/page/{pageID}/export-schedule/send", PageExport{}.ScheduleSendNow)
}

// Export renders the page in headless Chromium using the same print rules as the browser print
// and returns it as PDF, XLSX or PPTX.
//
// Query:
//
//	format - pdf | xlsx | pptx
//	url    - address of the page in the webapp (same host as the request)
func (PageExport) Export(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		format = strings.ToLower(r.URL.Query().Get("format"))
	)

	if format != "pdf" && format != "xlsx" && format != "pptx" {
		http.Error(w, "unsupported format", http.StatusBadRequest)
		return
	}

	identity := auth.GetIdentityFromContext(ctx)
	if identity == nil || identity.Identity() == 0 {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	target, err := pageExportTarget(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	body, contentType, status, err := pageExportBuild(ctx, identity, target, format)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="page-%s.%s"`, chi.URLParam(r, "pageID"), format))
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	_, _ = w.Write(body)
}

// pageExportBuild issues a short-lived token for the identity, renders the page and
// returns the file body, its content type and (on error) the suitable http status.
func pageExportBuild(ctx context.Context, identity auth.Identifiable, target, format string) (body []byte, contentType string, status int, err error) {
	token, err := auth.TokenIssuer.Issue(ctx,
		auth.WithIdentity(identity),
		auth.WithExpiration(pageExportTokenTTL),
		auth.WithScope("profile", "api"),
	)
	if err != nil {
		return nil, "", http.StatusInternalServerError, fmt.Errorf("could not issue export token")
	}

	var data pageExportData

	if format == "pdf" {
		body, err = pageExportRender(ctx, target, string(token), nil)
		contentType = "application/pdf"
	} else {
		_, err = pageExportRender(ctx, target, string(token), &data)
	}

	if err != nil {
		return nil, "", http.StatusBadGateway, fmt.Errorf("export failed: %w", err)
	}

	switch format {
	case "xlsx":
		body, err = pageExportXLSX(data)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "pptx":
		body, err = pageExportPPTX(data)
		contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	}

	if err != nil {
		return nil, "", http.StatusInternalServerError, fmt.Errorf("export failed: %w", err)
	}

	return body, contentType, 0, nil
}

func requestHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	if i := strings.Index(host, ","); i >= 0 {
		host = strings.TrimSpace(host[:i])
	}
	return host
}

// pageExportTarget validates the page url (protection against SSRF) and
// optionally rewrites the origin so the server can reach itself.
func pageExportTarget(r *http.Request) (string, error) {
	if err := pageExportCheckURL(r.URL.Query().Get("url"), requestHost(r)); err != nil {
		return "", err
	}
	return pageExportRewriteURL(r.URL.Query().Get("url"))
}

// pageExportCheckURL accepts only urls that belong to the given host or to EXPORT_ALLOWED_HOSTS
func pageExportCheckURL(raw, host string) error {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid page url")
	}

	// the webapp may be served from another origin than the API (dev server, separate domain)
	allowed := strings.EqualFold(u.Host, host)
	for _, h := range strings.Split(os.Getenv("EXPORT_ALLOWED_HOSTS"), ",") {
		if h = strings.TrimSpace(h); h != "" && strings.EqualFold(u.Host, h) {
			allowed = true
		}
	}

	if !allowed {
		return fmt.Errorf("page url must belong to the same host")
	}

	return nil
}

func pageExportRewriteURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid page url")
	}

	if base := os.Getenv("EXPORT_BASE_URL"); base != "" {
		b, err := url.Parse(base)
		if err != nil || b.Host == "" {
			return "", fmt.Errorf("invalid EXPORT_BASE_URL")
		}
		u.Scheme, u.Host = b.Scheme, b.Host
	}

	return u.String(), nil
}

// pageExportRender opens the page, waits until the content is loaded and either prints it
// to PDF or (when data is not nil) collects tables from the DOM.
func pageExportRender(ctx context.Context, target, token string, data *pageExportData) (pdf []byte, err error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1100, 1500),
	)

	if p := os.Getenv("EXPORT_CHROME_PATH"); p != "" {
		opts = append(opts, chromedp.ExecPath(p))
	}

	// the render must not depend on the client connection, but must be bound by timeout
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), pageExportTimeout)
	defer cancel()

	ctx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	ctx, cancelCtx := chromedp.NewContext(ctx)
	defer cancelCtx()

	settle := pageExportSettle
	if s, perr := strconv.Atoi(os.Getenv("EXPORT_SETTLE_SECONDS")); perr == nil && s > 0 {
		settle = time.Duration(s) * time.Second
	}

	// the webapp picks up the token instead of starting the OAuth2 flow
	inject := fmt.Sprintf("window.__EXPORT_TOKEN__ = %q;", token)

	err = chromedp.Do(ctx,
		chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
			_, err := cdp.Call(ctx, t, page.AddScriptToEvaluateOnNewDocument, page.AddScriptToEvaluateOnNewDocumentParams{Source: inject})
			return err
		}),
		chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
			_, err := cdp.Call(ctx, t, emulation.SetEmulatedMedia, emulation.SetEmulatedMediaParams{Media: "print"})
			return err
		}),
		chromedp.Navigate(target),
		// tabs are rendered in the background after the page is opened, see TabsBase
		chromedp.Sleep(settle),
	)
	if err != nil {
		return nil, err
	}

	if data != nil {
		raw, err := chromedp.Run(ctx, chromedp.Evaluate[string](pageExportExtractJS))
		if err != nil {
			return nil, err
		}
		return nil, json.Unmarshal([]byte(raw), data)
	}

	var (
		yes    = true
		margin = 0.4
	)

	res, err := chromedp.Run(ctx, func(ctx context.Context, t *chromedp.Target) (page.PrintToPDFResult, error) {
		return cdp.Call(ctx, t, page.PrintToPDF, page.PrintToPDFParams{
			PaperWidth:      8.27,
			PaperHeight:     11.69,
			MarginTop:       &margin,
			MarginBottom:    &margin,
			MarginLeft:      &margin,
			MarginRight:     &margin,
			PrintBackground: &yes,
		})
	})
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}

const pageExportExtractJS = `(() => {
  const text = el => (el.innerText || el.textContent || '').replace(/\s+/g, ' ').trim();
  const shown = el => el.getClientRects().length > 0;
  const h = document.querySelector('h1, .page-title');
  const out = { title: (h && text(h)) || document.title || '', sections: [] };

  document.querySelectorAll('.rl-root').forEach(root => {
    if (!shown(root)) return;
    const table = root.querySelector('table');
    if (!table) return;

    const ths = Array.from(table.querySelectorAll('thead tr:last-child th'));
    const cols = [];
    // columns without a title are service ones (row checkbox, row actions)
    ths.forEach((th, i) => { if (shown(th) && text(th)) cols.push(i); });

    const rows = [];
    table.querySelectorAll('tbody tr').forEach(tr => {
      const tds = Array.from(tr.children).filter(c => c.tagName === 'TD');
      // rows may have extra trailing cells (actions); rows with fewer cells are placeholders
      if (tds.length < ths.length) return;
      rows.push(cols.map(i => text(tds[i])));
    });

    const more = root.querySelector('.rl-print-more');
    const m = more ? (more.textContent || '').match(/\d+/) : null;
    const total = m ? parseInt(m[0], 10) + 7 : rows.length;

    const card = root.closest('.card');
    const header = card && card.querySelector('.card-header');
    const pane = root.closest('.tab-pane');
    const tabEl = pane && pane.getAttribute('aria-labelledby') && document.getElementById(pane.getAttribute('aria-labelledby'));

    out.sections.push({
      tab: tabEl ? text(tabEl) : '',
      title: header ? text(header) : '',
      headers: cols.map(i => text(ths[i])),
      rows,
      total: Math.max(total, rows.length),
    });
  });

  return JSON.stringify(out);
})()`
