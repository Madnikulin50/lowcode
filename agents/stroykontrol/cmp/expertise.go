package cmp

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

// Remark of the expertise (замечание экспертизы) as it comes in an Excel
// table, optionally with the designer's answer.
type Expert struct {
	Number  string `json:"number"`
	Text    string `json:"text"`
	Answer  string `json:"answer"`
	Where   string `json:"where"` // раздел / том / лист / пункт, as written
	Source  string `json:"source"`
	Section string `json:"section"`
}

var (
	hdrNumber = regexp.MustCompile(`^(№|n|номер|№ ?п/п|п/п|№ замечания|номер замечания)\b|^№`)
	hdrText   = regexp.MustCompile(`замечани|содержани|недостат|несоответств`)
	hdrAnswer = regexp.MustCompile(`ответ|пояснени|решени|исполнени|устранен|принят`)
	hdrWhere  = regexp.MustCompile(`раздел|том\b|шифр|лист|пункт|ссылк|документ|расположени|стр\.|подраздел|книга`)
)

func norm(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(s, " ", " ")), " "))
}

// ParseExpertise reads the first sheet of a workbook that has a header row
// with a remark column. Columns are recognized by their headings (the
// formats of different experts vary): a number, the remark text, the
// designer's answer and a location reference. A row with an answer only
// continues the previous remark; a one-cell row is a section heading.
func ParseExpertise(data []byte, source string) ([]Expert, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть Excel: %w", err)
	}
	defer f.Close()
	var lastErr error = fmt.Errorf("не найдена строка заголовка со столбцом «Замечание»")
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			lastErr = err
			continue
		}
		if out, ok := parseExpertRows(rows, source); ok {
			return out, nil
		}
	}
	return nil, lastErr
}

func parseExpertRows(rows [][]string, source string) ([]Expert, bool) {
	hdr, num, txt, ans := -1, -1, -1, -1
	var where []int
	for r := 0; r < len(rows) && r < 40; r++ {
		n, t, a := -1, -1, -1
		var w []int
		for c, cell := range rows[r] {
			h := norm(cell)
			if h == "" || len([]rune(h)) > 60 {
				continue
			}
			switch {
			case hdrNumber.MatchString(h) && n < 0:
				n = c
			case hdrAnswer.MatchString(h) && a < 0:
				a = c
			case hdrText.MatchString(h) && t < 0:
				t = c
			case hdrWhere.MatchString(h):
				w = append(w, c)
			}
		}
		if t >= 0 && (n >= 0 || a >= 0 || len(w) > 0) {
			hdr, num, txt, ans, where = r, n, t, a, w
			break
		}
	}
	if hdr < 0 {
		return nil, false
	}
	cell := func(r []string, c int) string {
		if c < 0 || c >= len(r) {
			return ""
		}
		return strings.TrimSpace(strings.ReplaceAll(r[c], " ", " "))
	}
	var out []Expert
	section := ""
	for _, row := range rows[hdr+1:] {
		filled, lone := 0, -1
		for c, v := range row {
			if strings.TrimSpace(v) != "" {
				filled++
				lone = c
			}
		}
		if filled == 0 {
			continue
		}
		// a lone cell outside the remark column is a section heading; in the
		// remark column it is the wrapped tail of the previous remark
		if filled == 1 && lone != txt && lone != ans {
			section = cell(row, lone)
			continue
		}
		e := Expert{Number: cell(row, num), Text: cell(row, txt), Answer: cell(row, ans), Source: source, Section: section}
		var wh []string
		for _, c := range where {
			if v := cell(row, c); v != "" {
				wh = append(wh, v)
			}
		}
		e.Where = strings.Join(wh, "; ")
		if e.Number == "" && e.Text == "" && e.Answer == "" {
			continue
		}
		if e.Number == "" && e.Text == "" && e.Answer != "" && len(out) > 0 {
			out[len(out)-1].Answer = strings.TrimSpace(out[len(out)-1].Answer + " " + e.Answer)
			continue
		}
		if e.Number == "" && e.Text != "" && len(out) > 0 && out[len(out)-1].Number != "" && e.Answer == "" && e.Where == "" {
			out[len(out)-1].Text += " " + e.Text // wrapped text of the previous remark
			continue
		}
		if e.Number == "" {
			e.Number = fmt.Sprint(len(out) + 1)
		}
		out = append(out, e)
	}
	return out, len(out) > 0
}

// MergeExperts joins remark lists by number: the expert's table and the
// designer's answer table often arrive as separate files.
func MergeExperts(lists ...[]Expert) []Expert {
	var out []Expert
	idx := map[string]int{}
	for _, l := range lists {
		for _, e := range l {
			key := norm(e.Number)
			i, ok := idx[key]
			if !ok || key == "" {
				idx[key] = len(out)
				out = append(out, e)
				continue
			}
			m := &out[i]
			if m.Text == "" {
				m.Text = e.Text
			}
			if m.Answer == "" {
				m.Answer = e.Answer
			}
			if m.Where == "" {
				m.Where = e.Where
			}
		}
	}
	return out
}

// Verdicts of matching a remark with the comparison's findings.
const (
	VerdictStillDiffers = "still_differs" // the difference the remark is about is still there
	VerdictResolved     = "resolved"      // it was there and is gone in this iteration
	VerdictNoMatch      = "no_match"      // no related finding: not a ПД/РД difference, or not visible to the engine
)

// Candidate is a finding offered for matching.
type Candidate struct {
	ID        string
	Text      string // description + fragments
	IterState string
	Severity  string
}

// RemarkMatch is the outcome for one remark.
type RemarkMatch struct {
	Number       string
	Verdict      string
	Score        float64
	FindingIDs   []string
	Note         string
	ClaimMissing bool // the answer says it was corrected, the difference remains
}

var (
	remWordRe   = regexp.MustCompile(`[\p{L}\p{N}][\p{L}\p{N}.,\-]*[\p{L}\p{N}]|[\p{L}\p{N}]`)
	claimRe     = regexp.MustCompile(`уточнен|принят|исправл|заменен|дополнен|добавлен|учтен|учтён|откорректирован|скорректирован|внесен|внесён|исправлен|устранен|устранён|приведен|приведён|доработан|изменен|изменён|выполнен`)
	stopWords   = map[string]bool{"что": true, "для": true, "при": true, "или": true, "как": true, "это": true, "его": true, "они": true, "все": true, "так": true, "быть": true, "были": true, "было": true, "также": true, "после": true, "более": true, "менее": true, "стадии": true, "стадия": true, "замечание": true, "рассмотрено": true, "документация": true, "документации": true, "проектной": true, "рабочей": true, "рабочая": true, "обосновать": true, "соответствует": true, "соответствие": true}
	lookalikeLC = strings.NewReplacer("a", "а", "b", "в", "c", "с", "e", "е", "h", "н", "k", "к", "m", "м", "o", "о", "p", "р", "t", "т", "x", "х", "y", "у")
)

type weighted map[string]float64

// remarkTokens: stems of meaningful words (weight 1) and codes/numbers (weight 3):
// "17п", "С15-35Т6", "2,4", "В25" identify a place or a value, words only a topic.
func remarkTokens(s string) weighted {
	out := weighted{}
	for _, w := range remWordRe.FindAllString(strings.ToLower(s), -1) {
		w = strings.Trim(w, ".,-")
		w = strings.ReplaceAll(w, ",", ".")
		hasDigit, hasLetter := false, false
		for _, r := range w {
			hasDigit = hasDigit || unicode.IsDigit(r)
			hasLetter = hasLetter || unicode.IsLetter(r)
		}
		rs := []rune(w)
		switch {
		case hasDigit:
			if !hasLetter && len(rs) < 3 && !strings.Contains(w, ".") {
				continue // a bare small number carries no identity
			}
			out[lookalikeLC.Replace(w)] = 3
		case len(rs) >= 4 && !stopWords[w]:
			if len(rs) > 5 {
				rs = rs[:5]
			}
			if _, ok := out[string(rs)]; !ok {
				out[string(rs)] = 1
			}
		}
	}
	return out
}

func shared(a, b weighted) (score float64, common []string) {
	for k, w := range a {
		if _, ok := b[k]; ok {
			score += w
			common = append(common, k)
		}
	}
	sort.Strings(common)
	return
}

func total(a weighted) (s float64) {
	for _, w := range a {
		s += w
	}
	return
}

const (
	matchMinScore = 0.22
	maxMatches    = 3
)

// MatchRemarks links every remark to the findings of the comparison it is
// about, by the words and — more strongly — the codes and values they share
// (a pile mark, a pier number, a size). It is deterministic and explains
// itself: which tokens matched. The verdict describes the *current* state of
// the documents: the difference is still there, was fixed in this iteration,
// or no related difference is visible.
func MatchRemarks(experts []Expert, cands []Candidate) []RemarkMatch {
	ct := make([]weighted, len(cands))
	for i, c := range cands {
		ct[i] = remarkTokens(c.Text)
	}
	out := make([]RemarkMatch, 0, len(experts))
	for _, e := range experts {
		// the remark's own words and the designer's answer both describe the place
		rt := remarkTokens(e.Text + " " + e.Where + " " + e.Answer)
		wr := total(rt)
		type hit struct {
			i     int
			score float64
			words []string
		}
		var hits []hit
		for i := range cands {
			s, common := shared(rt, ct[i])
			if len(common) == 0 || wr == 0 || total(ct[i]) == 0 {
				continue
			}
			hasCode := false
			for _, k := range common {
				if rt[k] == 3 {
					hasCode = true
				}
			}
			if len(common) < 2 && !hasCode {
				continue
			}
			sc := s / math.Sqrt(wr*total(ct[i]))
			if sc >= matchMinScore {
				hits = append(hits, hit{i, sc, common})
			}
		}
		sort.Slice(hits, func(a, b int) bool { return hits[a].score > hits[b].score })
		if len(hits) > maxMatches {
			hits = hits[:maxMatches]
		}
		m := RemarkMatch{Number: e.Number, Verdict: VerdictNoMatch}
		if len(hits) == 0 {
			m.Note = "Связанных расхождений между ПД и РД не найдено: проверьте вручную (замечание может касаться не различий стадий)."
			out = append(out, m)
			continue
		}
		open, closed := 0, 0
		var why []string
		for _, h := range hits {
			c := cands[h.i]
			m.FindingIDs = append(m.FindingIDs, c.ID)
			if c.IterState == IterResolved {
				closed++
			} else {
				open++
			}
			why = append(why, fmt.Sprintf("%s (%.0f%%)", strings.Join(h.words, ", "), 100*h.score))
		}
		m.Score = hits[0].score
		switch {
		case open > 0:
			m.Verdict = VerdictStillDiffers
			m.Note = fmt.Sprintf("Расхождение, о котором замечание, сохраняется (связанных находок: %d). Совпали: %s.", open, strings.Join(why, "; "))
			if claimRe.MatchString(strings.ToLower(e.Answer)) {
				m.ClaimMissing = true
				m.Note += " Ответ проектировщика заявляет правку, но различие в документах осталось."
			}
		default:
			m.Verdict = VerdictResolved
			m.Note = fmt.Sprintf("Связанное расхождение устранено в этой итерации (находок: %d). Совпали: %s.", closed, strings.Join(why, "; "))
		}
		out = append(out, m)
	}
	return out
}
