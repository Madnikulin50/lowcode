package idcheck

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// latinLookalikes folds Latin letters that look like Cyrillic ones onto the
// Cyrillic letter — scanned/typed act numbers mix both freely ("OП№6",
// "CKTPACT"), and OCR does too.
var latinLookalikes = strings.NewReplacer(
	"A", "А", "B", "В", "C", "С", "E", "Е", "H", "Н", "K", "К", "M", "М",
	"O", "О", "P", "Р", "T", "Т", "X", "Х", "Y", "У",
)

// compact upper-cases s, folds look-alike letters, unifies dashes and
// strips whitespace and "№".
func compact(s string) string {
	s = strings.ToUpper(s)
	s = latinLookalikes.Replace(s)
	s = strings.NewReplacer("–", "-", "—", "-", "‐", "-", "№", "", "N°", "", "«", "", "»", "", "\"", "").Replace(s)
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

var missingSlashBeforeObject = regexp.MustCompile(`([^/])(ИССО)`)

// ActNumber is a canonical act number "seq/positions/part/object/contractor".
type ActNumber struct {
	Raw       string
	Seq       string
	Positions []string
	Rest      []string
}

// Key is the comparable form of the whole number.
func (a ActNumber) Key() string {
	parts := append([]string{a.Seq, strings.Join(a.Positions, ";")}, a.Rest...)
	return strings.Join(parts, "/")
}

// HasPosition reports whether the number names position p.
func (a ActNumber) HasPosition(p string) bool {
	p = compact(p)
	for _, x := range a.Positions {
		if x == p {
			return true
		}
	}
	return false
}

// ParseActNumber canonicalizes an act number as written in the registry
// or in the act ("№2/2.14.1.9/ОП№6/ИССО 2.2/СКТРАСТ"). sep is the token
// separator ('/' for registry/acts, '_' or '-' for file names).
func ParseActNumber(raw string, sep rune) ActNumber {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "№")
	if sep != '/' {
		s = strings.ReplaceAll(s, "/", string(sep))
	}
	tokens := strings.Split(s, string(sep))
	var out []string
	for _, t := range tokens {
		t = compact(t)
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	joined := strings.Join(out, "/")
	joined = missingSlashBeforeObject.ReplaceAllString(joined, "$1/$2")
	out = strings.Split(joined, "/")
	an := ActNumber{Raw: raw}
	if len(out) > 0 {
		an.Seq = out[0]
	}
	if len(out) > 1 {
		for _, p := range strings.Split(out[1], ";") {
			if p = strings.Trim(p, ".,"); p != "" {
				an.Positions = append(an.Positions, p)
			}
		}
	}
	if len(out) > 2 {
		an.Rest = out[2:]
	}
	return an
}

var fileDateRe = regexp.MustCompile(`(?i)\s*от\s*([0-9]{1,2}\.[0-9]{1,2}\.[0-9]{2,4})\s*(?:г\.?)?\s*$`)
var dashSeparatedRe = regexp.MustCompile(`^\d+-\d+\.`)

// ParseActFileName extracts number and date from file names like
// "1_2.14.1.9_ОП№6_ИССО2.2_СКТРАСТ от 18.08.2024.pdf" or
// "2-2.14.1.9-ОП№6-ИССО2.2-СКТРАСТ от 18.08.2024.pdf".
func ParseActFileName(name string) (ActNumber, time.Time) {
	base := name
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	base = strings.TrimSuffix(base, ".pdf")
	base = strings.TrimSuffix(base, ".PDF")
	var date time.Time
	if m := fileDateRe.FindStringSubmatch(base); m != nil {
		date, _ = ParseDate(m[1])
		base = base[:len(base)-len(m[0])]
	}
	sep := '_'
	if dashSeparatedRe.MatchString(base) {
		sep = '-'
	}
	return ParseActNumber(base, sep), date
}

var monthsRu = map[string]time.Month{
	"январ": 1, "феврал": 2, "март": 3, "апрел": 4, "ма": 5, "июн": 6,
	"июл": 7, "август": 8, "сентябр": 9, "октябр": 10, "ноябр": 11, "декабр": 12,
}

var (
	numDateRe  = regexp.MustCompile(`(\d{1,2})[.\-/](\d{1,2})[.\-/](\d{2,4})`)
	isoDateRe  = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})`)
	wordDateRe = regexp.MustCompile(`(?i)(\d{1,2})\s*[»"]?\s*([а-яё]+)\s+(\d{4})`)
)

// ParseDate understands "18.08.2024г.", "18.08.24", "2024-08-18",
// "«18» августа 2024 г." and Excel serial numbers.
func ParseDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if m := isoDateRe.FindStringSubmatch(s); m != nil {
		return mkDate(m[1], m[2], m[3])
	}
	if m := numDateRe.FindStringSubmatch(s); m != nil {
		y := m[3]
		if len(y) == 2 {
			y = "20" + y
		}
		return mkDate(y, m[2], m[1])
	}
	if m := wordDateRe.FindStringSubmatch(s); m != nil {
		w := strings.ToLower(m[2])
		for prefix, mon := range monthsRu {
			if strings.HasPrefix(w, prefix) && (prefix != "ма" || strings.HasPrefix(w, "мая") || strings.HasPrefix(w, "май")) {
				d, _ := strconv.Atoi(m[1])
				y, _ := strconv.Atoi(m[3])
				return time.Date(y, mon, d, 0, 0, 0, 0, time.UTC), true
			}
		}
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && f > 20000 && f < 80000 {
		return excelSerial(f), true
	}
	return time.Time{}, false
}

func mkDate(y, m, d string) (time.Time, bool) {
	yi, _ := strconv.Atoi(y)
	mi, _ := strconv.Atoi(m)
	di, _ := strconv.Atoi(d)
	if yi < 1950 || yi > 2100 || mi < 1 || mi > 12 || di < 1 || di > 31 {
		return time.Time{}, false
	}
	return time.Date(yi, time.Month(mi), di, 0, 0, 0, 0, time.UTC), true
}

func excelSerial(f float64) time.Time {
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return base.AddDate(0, 0, int(math.Floor(f)))
}

// FmtDate renders a date the way Russian documents do.
func FmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02.01.2006")
}

var numRe = regexp.MustCompile(`-?\d+(?:[.,]\d+)?`)

// ParseNum parses "26,3", "26.3", " 26.3 м3" → 26.3.
func ParseNum(s string) (float64, bool) {
	s = strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, false
	}
	if f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64); err == nil {
		return f, true
	}
	m := numRe.FindString(s)
	if m == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.ReplaceAll(m, ",", "."), 64)
	return f, err == nil
}

func numPtr(s string) *float64 {
	if f, ok := ParseNum(s); ok {
		return &f
	}
	return nil
}

// NearlyEqual compares quantities with a tolerance that scales with the
// value (registry quantities are often 4-decimal tonnes).
func NearlyEqual(a, b float64) bool {
	tol := math.Max(0.01, math.Abs(b)*0.0005)
	return math.Abs(a-b) <= tol
}

var rdSheetsRe = regexp.MustCompile(`(?i)\(?\s*(лист|листы|л\.)[^)]*\)?`)

// NormalizeRDCode reduces a working-documentation code ("шифр") to a
// comparable form: drops sheet lists, punctuation spacing and look-alikes.
func NormalizeRDCode(s string) string {
	s = rdSheetsRe.ReplaceAllString(s, "")
	s = strings.TrimPrefix(strings.TrimSpace(s), "шифр")
	s = compact(s)
	return strings.Trim(s, ".,;:")
}

var rdTokenSplit = regexp.MustCompile(`[\s;,()«»"]+`)

// FindRDCodes pulls code-looking tokens ("03102022/РД-ИССО2.2-ОП-КЖ4.1")
// out of free text: whitespace-separated tokens containing a digit and a
// "-" or "/", at least 8 characters long.
func FindRDCodes(s string) []string {
	var out []string
	seen := map[string]bool{}
	for _, tok := range rdTokenSplit.Split(s, -1) {
		t := strings.Trim(compact(tok), ".,;:-")
		if len([]rune(t)) < 8 || !strings.ContainsAny(t, "0123456789") || !strings.ContainsAny(t, "-/") || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// ownershipForms are stripped before comparing company names.
var ownershipForms = regexp.MustCompile(`(?i)^(общество с ограниченной ответственностью|акционерное общество|публичное акционерное общество|закрытое акционерное общество|открытое акционерное общество|ооо|ао|пао|зао|оао)\s+`)

// NormalizeOrgName reduces "ООО «РАД»" / "Общество с ограниченной
// ответственностью "РАД"" to "РАД".
func NormalizeOrgName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = ownershipForms.ReplaceAllString(s, "")
	s = strings.NewReplacer("«", "", "»", "", "\"", "", "'", "", "“", "", "”", "").Replace(s)
	return strings.ToUpper(strings.Join(strings.Fields(s), " "))
}

// digits keeps only ASCII digits.
func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
