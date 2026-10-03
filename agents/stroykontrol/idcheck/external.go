package idcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

// EGRULRecord is what the public egrul.nalog.ru search returns for a
// legal entity (the free endpoint gives the name, ids, region and the
// termination date — not the full address).
type EGRULRecord struct {
	FullName  string `json:"n"`
	ShortName string `json:"c"`
	OGRN      string `json:"o"`
	INN       string `json:"i"`
	KPP       string `json:"p"`
	RegDate   string `json:"r"`
	EndDate   string `json:"e"` // set when the entity has ceased to exist
	Region    string `json:"rn"`
	Director  string `json:"g"`
}

// NRSRecord is the public part of a National Register of Specialists entry.
// nrs.nostroy.ru renders the name and dates as images on purpose, so only
// presence, status and "has an exclusion decision" are read.
type NRSRecord struct {
	Number   string
	Found    bool
	Status   string // Действует / Исключен
	Excluded bool
	WorkType string
}

// External looks entities up in public registries, with an in-memory cache
// and a polite delay between requests. Lookup failures are returned as
// errors; callers turn them into low-severity "could not verify" findings.
type External struct {
	EGRULURL string
	NRSURL   string
	Delay    time.Duration
	http     *http.Client

	mu    sync.Mutex
	last  time.Time
	egrul map[string]*EGRULRecord
	nrs   map[string]*NRSRecord
}

// NewExternal returns a client for the public registries.
func NewExternal() *External {
	return &External{
		EGRULURL: "https://egrul.nalog.ru",
		NRSURL:   "https://nrs.nostroy.ru",
		Delay:    700 * time.Millisecond,
		http:     &http.Client{Timeout: 30 * time.Second},
		egrul:    map[string]*EGRULRecord{},
		nrs:      map[string]*NRSRecord{},
	}
}

func (e *External) throttle() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if wait := e.Delay - time.Since(e.last); wait > 0 {
		time.Sleep(wait)
	}
	e.last = time.Now()
}

func (e *External) get(ctx context.Context, u string, form url.Values) ([]byte, error) {
	e.throttle()
	var req *http.Request
	var err error
	if form != nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, u, strings.NewReader(form.Encode()))
		if req != nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "stroykontrol-idcheck/1.0")
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s: HTTP %d", u, resp.StatusCode)
	}
	return b, nil
}

// EGRUL finds a legal entity by ОГРН or ИНН. Returns (nil, nil) if the
// registry has no such entity.
func (e *External) EGRUL(ctx context.Context, query string) (*EGRULRecord, error) {
	q := digits(query)
	if q == "" {
		return nil, fmt.Errorf("empty query")
	}
	e.mu.Lock()
	if r, ok := e.egrul[q]; ok {
		e.mu.Unlock()
		return r, nil
	}
	e.mu.Unlock()

	b, err := e.get(ctx, e.EGRULURL+"/", url.Values{"query": {q}, "region": {""}, "page": {""}})
	if err != nil {
		return nil, fmt.Errorf("ЕГРЮЛ: %w", err)
	}
	var tok struct {
		T               string `json:"t"`
		CaptchaRequired bool   `json:"captchaRequired"`
	}
	if err := json.Unmarshal(b, &tok); err != nil || tok.T == "" {
		return nil, fmt.Errorf("ЕГРЮЛ: unexpected reply %s", truncate(string(b), 120))
	}
	if tok.CaptchaRequired {
		return nil, fmt.Errorf("ЕГРЮЛ: сервис требует капчу — проверка отложена")
	}
	var res struct {
		Status string        `json:"status"`
		Rows   []EGRULRecord `json:"rows"`
	}
	for attempt := 0; attempt < 5; attempt++ {
		b, err = e.get(ctx, e.EGRULURL+"/search-result/"+tok.T, nil)
		if err != nil {
			return nil, fmt.Errorf("ЕГРЮЛ: %w", err)
		}
		res.Status, res.Rows = "", nil
		if err := json.Unmarshal(b, &res); err != nil {
			return nil, fmt.Errorf("ЕГРЮЛ: decode: %w", err)
		}
		if res.Status != "wait" {
			break
		}
		time.Sleep(time.Second)
	}
	var out *EGRULRecord
	for i := range res.Rows {
		r := res.Rows[i]
		if r.OGRN == q || r.INN == q {
			out = &r
			break
		}
	}
	e.mu.Lock()
	e.egrul[q] = out
	e.mu.Unlock()
	return out, nil
}

var (
	nrsRowRe    = regexp.MustCompile(`(?s)<tr>\s*(<td style="width: 10%.*?)</tr>`)
	nrsCellRe   = regexp.MustCompile(`(?s)<td[^>]*>(.*?)</td>`)
	tagRe       = regexp.MustCompile(`<[^>]+>`)
	spaceRe     = regexp.MustCompile(`\s+`)
	imgInCellRe = regexp.MustCompile(`<img\s`)
)

// NRS looks a specialist up by identification number ("С-77-246659").
func (e *External) NRS(ctx context.Context, number string) (*NRSRecord, error) {
	n := NormalizeNRS(number)
	if n == "" {
		return nil, fmt.Errorf("некорректный номер НРС %q", number)
	}
	// the register stores numbers with Cyrillic С/П; compact() already folded look-alikes
	e.mu.Lock()
	if r, ok := e.nrs[n]; ok {
		e.mu.Unlock()
		return r, nil
	}
	e.mu.Unlock()
	b, err := e.get(ctx, e.NRSURL+"/?"+url.Values{"s.registrationNumber": {n}}.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("НРС: %w", err)
	}
	rec := parseNRS(string(b))
	rec.Number = n
	e.mu.Lock()
	e.nrs[n] = rec
	e.mu.Unlock()
	return rec, nil
}

func parseNRS(html string) *NRSRecord {
	rec := &NRSRecord{}
	for _, m := range nrsRowRe.FindAllStringSubmatch(html, -1) {
		cells := nrsCellRe.FindAllStringSubmatch(m[1], -1)
		if len(cells) < 9 {
			continue
		}
		text := func(i int) string {
			return strings.TrimSpace(spaceRe.ReplaceAllString(tagRe.ReplaceAllString(cells[i][1], " "), " "))
		}
		rec.Found = true
		rec.WorkType = text(7)
		rec.Status = text(8)
		rec.Excluded = imgInCellRe.MatchString(cells[6][1]) || strings.Contains(strings.ToLower(rec.Status), "исключ")
		break
	}
	return rec
}

// CheckExternal runs check 7 (ЕГРЮЛ) and the online part of check 6 (НРС)
// for one act.
func (e *External) CheckExternal(ctx context.Context, in *ActInput) []Finding {
	if in.Data == nil {
		return nil
	}
	var out []Finding
	actDate := in.actDate()
	for _, o := range in.Data.Orgs {
		q := FirstDigitRun(o.OGRN)
		if !ValidOGRN(q) {
			q = FirstDigitRun(o.INN)
			if !ValidINN(q) {
				continue // checksum finding already raised offline
			}
		}
		r, err := e.EGRUL(ctx, q)
		if err != nil {
			f := in.find(7, Low, "egrul", "", q, fmt.Sprintf("%s: не удалось проверить по ЕГРЮЛ (%v)", o.Name, err))
			f.Source = "ext"
			out = append(out, f)
			continue
		}
		if r == nil {
			f := in.find(7, High, "egrul", "запись ЕГРЮЛ", q, fmt.Sprintf("%s: юрлицо с ОГРН/ИНН %s не найдено в ЕГРЮЛ", o.Name, q))
			f.Source = "ext"
			out = append(out, f)
			continue
		}
		add := func(sev Severity, field, exp, act, desc string) {
			f := in.find(7, sev, field, exp, act, desc)
			f.Source = "ext"
			out = append(out, f)
		}
		inn, ogrn := FirstDigitRun(o.INN), FirstDigitRun(o.OGRN)
		if ogrn != "" && inn != "" && inn != r.INN && ogrn == r.OGRN {
			add(High, "inn", r.INN, o.INN, fmt.Sprintf("%s: ИНН в акте %s, по ЕГРЮЛ для ОГРН %s — %s", o.Name, o.INN, r.OGRN, r.INN))
		}
		if ogrn != "" && ogrn != r.OGRN && inn == r.INN && ValidOGRN(ogrn) {
			add(High, "ogrn", r.OGRN, o.OGRN, fmt.Sprintf("%s: ОГРН в акте %s, по ЕГРЮЛ для ИНН %s — %s", o.Name, o.OGRN, r.INN, r.OGRN))
		}
		if o.Name != "" && !sameOrgName(o.Name, r) {
			add(Medium, "name", r.ShortName, o.Name, fmt.Sprintf("Наименование в акте «%s» не совпадает с ЕГРЮЛ: «%s» (%s)", o.Name, r.ShortName, r.FullName))
		}
		if r.EndDate != "" {
			end, _ := ParseDate(r.EndDate)
			sev := Medium
			if !end.IsZero() && !actDate.IsZero() && end.Before(actDate) {
				sev = High
			}
			add(sev, "status", "действующее", "прекращено "+r.EndDate, fmt.Sprintf("%s: деятельность прекращена %s", r.ShortName, r.EndDate))
		}
		if o.Address != "" && r.Region != "" && !addressRegionMatches(o.Address, r.Region) {
			add(Low, "address", r.Region, o.Address, fmt.Sprintf("%s: адрес в акте (%s) не относится к региону регистрации по ЕГРЮЛ (%s)", o.Name, o.Address, r.Region))
		}
	}
	for _, p := range in.Data.People {
		n := NormalizeNRS(p.NRSID)
		if n == "" {
			continue
		}
		r, err := e.NRS(ctx, n)
		f := func(sev Severity, exp, act, desc string) {
			x := in.find(6, sev, "nrs", exp, act, desc)
			x.Source = "ext"
			out = append(out, x)
		}
		switch {
		case err != nil:
			f(Low, "", n, fmt.Sprintf("%s: не удалось проверить номер %s в НРС (%v)", p.FIO, n, err))
		case !r.Found:
			f(High, "запись в НРС", n, fmt.Sprintf("%s: идентификационный номер %s не найден в НРС", p.FIO, n))
		case r.Excluded:
			f(High, "Действует", r.Status, fmt.Sprintf("%s: сведения %s исключены из НРС — проверьте дату исключения относительно даты акта %s", p.FIO, n, FmtDate(actDate)))
		}
	}
	return out
}

func sameOrgName(name string, r *EGRULRecord) bool {
	a := NormalizeOrgName(name)
	for _, b := range []string{r.ShortName, r.FullName} {
		nb := NormalizeOrgName(b)
		if a == nb || strings.Contains(nb, a) || strings.Contains(a, nb) {
			return true
		}
	}
	// "ТК Безопасность" vs "Технологии контроля безопасности": compare
	// significant words of the full name
	words := strings.Fields(strings.NewReplacer("-", " ", ".", " ").Replace(a))
	full := NormalizeOrgName(r.FullName)
	hit := 0
	for _, w := range words {
		if len([]rune(w)) > 3 && strings.Contains(full, w) {
			hit++
		}
	}
	return len(words) > 0 && hit*2 >= len(words)
}

func addressRegionMatches(addr, region string) bool {
	a := strings.ToLower(strings.ReplaceAll(addr, " ", ""))
	// an address without a locality ("Сеченовский переулок, д.6") can't be
	// compared — the extractor often splits the city off into another field
	hasLocality := false
	for _, m := range []string{"г.", "город", "обл", "край", "респ", "москва", "санкт-петербург", "пос.", "с.", "д.", "р-н", "район"} {
		if strings.Contains(a, m) && (m != "д." || strings.Contains(a, "р-н")) {
			hasLocality = true
		}
	}
	if !hasLocality {
		return true
	}
	r := strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(strings.TrimPrefix(region, "Г."), "г."), " ", ""))
	r = strings.TrimSuffix(strings.TrimSuffix(r, "область"), "обл.")
	if len([]rune(r)) > 5 {
		r = string([]rune(r)[:5]) // "рязанская" ~ "Рязань", "московская" ~ "Москва"
	}
	return strings.Contains(a, r)
}
