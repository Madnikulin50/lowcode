package agent

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// SaaS-экспорт: облачные приложения (Bitrix24, почта IMAP, CalDAV/CardDAV)
// выгружаются в поток tar.gz с JSON/.eml/.ics-файлами и дальше идут по
// обычному пайплайну packToMinio. Инкремент — курсором source.LastSync.

type saasExportFile struct {
	Name string
	Mod  time.Time
	Body func() (io.ReadCloser, error)
}

// saasExportStreams собирает файлы выгрузки по провайдеру.
func saasExportFiles(ctx context.Context, src *Source, cred *Credential, since time.Time) ([]saasExportFile, error) {
	switch strings.ToLower(strings.TrimSpace(src.Provider)) {
	case "bitrix24":
		return bitrix24Export(ctx, src, cred)
	case "imap":
		return imapExport(ctx, src, cred, since)
	case "caldav", "carddav":
		return caldavExport(ctx, src, cred, since)
	default:
		return nil, fmt.Errorf("unknown SaaS provider %q (bitrix24|imap|caldav)", src.Provider)
	}
}

// writeSaasTarGz пишет собранные файлы в w как tar.gz; body открываются
// лениво и закрываются после копирования. Тар требует размер до данных,
// поэтому каждый файл буферизуется в памяти (лимит 64 МБ на файл).
func writeSaasTarGz(ctx context.Context, w io.Writer, files []saasExportFile) (int, error) {
	gz := gzip.NewWriter(w)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	for _, f := range files {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		body, err := f.Body()
		if err != nil {
			return 0, fmt.Errorf("%s: %w", f.Name, err)
		}
		data, err := io.ReadAll(io.LimitReader(body, 64<<20))
		body.Close()
		if err != nil {
			return 0, fmt.Errorf("read %s: %w", f.Name, err)
		}
		hdr := &tar.Header{
			Name:     f.Name,
			Mode:     0o644,
			ModTime:  f.Mod,
			Size:     int64(len(data)),
			Format:   tar.FormatPAX,
			Typeflag: tar.TypeReg,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return 0, err
		}
		if _, err := tw.Write(data); err != nil {
			return 0, err
		}
	}
	return len(files), nil
}

// --- Bitrix24 --------------------------------------------------------------

var bitrixSafeName = regexp.MustCompile(`[^a-z0-9_.]+`)

// bitrix24Export выгружает основные CRM-сущности через входящий webhook.
// Токен — в credentials.extra: {"token":"..."} или просто строка; портал —
// source.BaseURL (https://portal.bitrix24.ru) или source.Host.
func bitrix24Export(ctx context.Context, src *Source, cred *Credential) ([]saasExportFile, error) {
	token := saasToken(cred)
	if token == "" {
		return nil, fmt.Errorf("bitrix24: empty webhook token in credentials.extra")
	}
	portal := strings.TrimRight(firstNonEmpty(src.BaseURL, src.Host), "/")
	if portal == "" {
		return nil, fmt.Errorf("bitrix24: empty portal URL (source base_url/host)")
	}
	methods := []string{
		"crm.lead.list", "crm.contact.list", "crm.company.list", "crm.deal.list",
		"tasks.task.list", "user.get",
	}
	var files []saasExportFile
	for _, m := range methods {
		rows, err := bitrixListAll(ctx, portal, token, m)
		if err != nil {
			// Одна сущность недоступна — не роняем всю выгрузку.
			rows = []map[string]any{{"error": err.Error()}}
		}
		name := "bitrix24/" + bitrixSafeName.ReplaceAllString(strings.ReplaceAll(m, ".", "_"), "_") + ".json"
		data, _ := json.MarshalIndent(rows, "", "  ")
		files = append(files, saasExportFile{
			Name: name,
			Mod:  time.Now(),
			Body: func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(string(data))), nil },
		})
	}
	return files, nil
}

func saasToken(cred *Credential) string {
	extra := strings.TrimSpace(cred.Extra)
	if extra == "" {
		return ResolveSecret(cred.Handle)
	}
	var obj map[string]string
	if json.Unmarshal([]byte(extra), &obj) == nil {
		return firstNonEmpty(obj["token"], obj["access_token"])
	}
	return extra
}

func bitrixListAll(ctx context.Context, portal, token, method string) ([]map[string]any, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	var out []map[string]any
	start := 0
	for {
		u := fmt.Sprintf("%s/rest/%s/%s.json?start=%d", portal, token, method, start)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return out, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return out, err
		}
		raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
		resp.Body.Close()
		if err != nil {
			return out, err
		}
		var payload struct {
			Result []map[string]any `json:"result"`
			Next   int              `json:"next"`
			Error  string           `json:"error"`
		}
		if err := json.Unmarshal(raw, &payload); err != nil {
			return out, fmt.Errorf("%s: %w", method, err)
		}
		if payload.Error != "" {
			return out, fmt.Errorf("%s: %s", method, payload.Error)
		}
		out = append(out, payload.Result...)
		if payload.Next <= start || len(payload.Result) == 0 {
			break
		}
		start = payload.Next
	}
	return out, nil
}

// --- IMAP ------------------------------------------------------------------

// imapExport выгружает письма из всех видимых папок ящика в .eml-файлы,
// новее since. Учётка — credentials.username + секрет по handle.
func imapExport(ctx context.Context, src *Source, cred *Credential, since time.Time) ([]saasExportFile, error) {
	host := firstNonEmpty(src.Host, src.BaseURL)
	if host == "" {
		return nil, fmt.Errorf("imap: empty host (source host/base_url)")
	}
	if !strings.Contains(host, ":") {
		host += ":993"
	}
	password := ResolveSecret(cred.Handle)
	if password == "" {
		password = cred.Extra
	}
	username := firstNonEmpty(cred.Username, src.Account)
	if username == "" {
		return nil, fmt.Errorf("imap: empty username")
	}

	// go-imap работает с net.Conn напрямую — пробрасываем deadline-контекст
	// через goroutine + Close при отмене.
	dialer := &tls.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", host)
	if err != nil {
		return nil, fmt.Errorf("imap dial %s: %w", host, err)
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	c, err := imapClientNew(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer c.Logout()
	if err := c.Login(username, password); err != nil {
		return nil, fmt.Errorf("imap login %s: %w", username, err)
	}

	folders, err := imapFolders(c)
	if err != nil {
		return nil, err
	}
	var files []saasExportFile
	for _, folder := range folders {
		messages, err := imapFetchSince(c, folder, since)
		if err != nil {
			files = append(files, saasErrorFile("imap/"+folder+".error.txt", err))
			continue
		}
		for i := range messages {
			msg := messages[i]
			name := fmt.Sprintf("imap/%s/%d.eml", sanitizeKey(folder), msg.UID)
			body := msg.Body
			files = append(files, saasExportFile{
				Name: name,
				Mod:  msg.InternalDate,
				Body: body,
			})
		}
	}
	return files, nil
}

func saasErrorFile(name string, err error) saasExportFile {
	return saasExportFile{
		Name: name,
		Mod:  time.Now(),
		Body: func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(err.Error())), nil
		},
	}
}

// --- CalDAV / CardDAV ------------------------------------------------------

// caldavExport выгружает события/контакты за период (since .. now+год) через
// REPORT calendar-query / addressbook-query. Простейший DAV-клиент на net/http.
func caldavExport(ctx context.Context, src *Source, cred *Credential, since time.Time) ([]saasExportFile, error) {
	base := strings.TrimRight(firstNonEmpty(src.BaseURL, src.Host), "/")
	if base == "" {
		return nil, fmt.Errorf("caldav: empty base_url")
	}
	password := ResolveSecret(cred.Handle)
	if password == "" {
		password = cred.Extra
	}
	username := firstNonEmpty(cred.Username, src.Account)
	hc := &http.Client{Timeout: 60 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}}
	report := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<C:calendar-query xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop><D:getetag/><C:calendar-data/></D:prop>
  <C:filter><C:comp-filter name="VCALENDAR"><C:comp-filter name="VEVENT">
    <C:time-range start="%s" end="%s"/>
  </C:comp-filter></C:comp-filter></C:filter>
</C:calendar-query>`, since.UTC().Format("20060102T150405Z"), time.Now().AddDate(1, 0, 0).UTC().Format("20060102T150405Z"))
	if strings.EqualFold(src.Provider, "carddav") {
		report = `<?xml version="1.0" encoding="utf-8"?>
<card:addressbook-query xmlns:D="DAV:" xmlns:card="urn:ietf:params:xml:ns:carddav">
  <D:prop><D:getetag/><card:address-data/></D:prop>
</card:addressbook-query>`
	}
	req, err := http.NewRequestWithContext(ctx, "REPORT", base+"/", strings.NewReader(report))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dav report %s: %w", base, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("dav report %s: HTTP %d: %s", base, resp.StatusCode, truncateStr(string(raw), 200))
	}
	kind := strings.ToLower(src.Provider)
	name := fmt.Sprintf("%s/%s-%s.xml", kind, kind, time.Now().UTC().Format("20060102T150405"))
	return []saasExportFile{{
		Name: name,
		Mod:  time.Now(),
		Body: func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader(string(raw))), nil },
	}}, nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// saasToMinio выгружает облачное приложение и упаковывает выгрузку в MinIO
// через общий tar.gz-пайплайн (hasher/метрики бесплатно).
func (a *Agent) saasToMinio(ctx context.Context, st *JobStatus, store Storage, src *Source, cred *Credential, key string) (*BackupResult, error) {
	files, err := saasExportFiles(ctx, src, cred, src.LastSync)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("saas %s: nothing to export", src.Provider)
	}
	st.Files = len(files)
	st.Message = fmt.Sprintf("exporting %d item(s) from %s", len(files), src.Provider)
	pr, pw := io.Pipe()
	var packErr error
	go func() {
		_, packErr = writeSaasTarGz(ctx, pw, files)
		if packErr != nil {
			pw.CloseWithError(packErr)
			return
		}
		pw.Close()
	}()
	written, err := store.Put(ctx, key, pr, -1, "application/gzip")
	if err != nil {
		return nil, err
	}
	if packErr != nil {
		return nil, packErr
	}
	return &BackupResult{
		Bucket:       store.Bucket(),
		Key:          key,
		BytesRead:    written,
		BytesWritten: written,
		Files:        len(files),
		Engine:       "saas-" + src.Provider,
		Kind:         KindFull,
	}, nil
}
