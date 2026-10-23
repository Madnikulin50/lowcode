package cmp

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Требования замечания (ТЗ 3.4.3, 3.6.6, 3.9.5): замечание экспертизы часто содержит несколько
// требований («1) … 2) …», «Необходимо … Следует …»). Каждое проверяется отдельно: по документам, по
// разбору ИИ и решением специалиста; результат замечания складывается из результатов требований.
// Деление детерминированное, по нумерации, маркерам и оборотам «необходимо / следует / представить»;
// специалист может его поправить, и поправленный список агент не перезаписывает.

// Requirement is one requirement of a remark and what became of it.
type Requirement struct {
	Idx    int    `json:"idx"`
	Text   string `json:"text"`
	Manual bool   `json:"manual,omitempty"` // typed or edited by the specialist

	// the check against the documents
	Verdict    string   `json:"verdict,omitempty"`
	Score      float64  `json:"score,omitempty"`
	Note       string   `json:"note,omitempty"`
	FindingIDs []string `json:"findingIDs,omitempty"`
	// the AI reading (never changes anything else)
	AIVerdict string `json:"aiVerdict,omitempty"`
	AINote    string `json:"aiNote,omitempty"`
	// the result in the four values of ТЗ 3.4.6, derived from the two above
	Outcome string `json:"outcome,omitempty"`
	// the specialist's decision on the requirement
	OutcomeFinal   string `json:"outcomeFinal,omitempty"`
	OutcomeComment string `json:"outcomeComment,omitempty"`
	OutcomeBy      string `json:"outcomeBy,omitempty"`
	OutcomeAt      string `json:"outcomeAt,omitempty"`
}

// Effective is the result of the requirement: the specialist's decision, else the system's outcome.
func (q Requirement) Effective() string {
	if q.OutcomeFinal != "" {
		return q.OutcomeFinal
	}
	if q.Outcome != "" {
		return q.Outcome
	}
	return OutcomeCheck
}

// RequirementSet is what a remark stores: the list and the text it was made for.
type RequirementSet struct {
	For    string        `json:"for"`              // hash of the text of the remark the list was made from
	Manual bool          `json:"manual,omitempty"` // the specialist edited the list
	Stale  bool          `json:"stale,omitempty"`  // the text of the remark changed after the list was edited by hand
	Items  []Requirement `json:"items"`
}

// TextHash identifies a text of a remark.
func TextHash(s string) string {
	sum := sha1.Sum([]byte(strings.Join(strings.Fields(s), " ")))
	return hex.EncodeToString(sum[:6])
}

// ParseRequirementSet reads the stored JSON ("" is no list).
func ParseRequirementSet(s string) (RequirementSet, error) {
	var set RequirementSet
	if strings.TrimSpace(s) == "" {
		return set, nil
	}
	if err := json.Unmarshal([]byte(s), &set); err != nil {
		return RequirementSet{}, fmt.Errorf("требования замечания: %w", err)
	}
	for i := range set.Items {
		set.Items[i].Idx = i + 1
	}
	return set, nil
}

// JSON is the stored form.
func (s RequirementSet) JSON() string {
	if len(s.Items) == 0 {
		return ""
	}
	b, _ := json.Marshal(s)
	return string(b)
}

const (
	maxRequirements   = 30
	maxRequirementLen = 600
	minRequirementLen = 8
)

var (
	// "1)", "а)" in the middle of a text; "1." and the dashes only at the start of a line (a decimal or a date is not a marker)
	inlineMarkRe = regexp.MustCompile(`(?:^|[\s;:.,])(\d{1,2}\)|[а-яё]\))\s`)
	lineMarkRe   = regexp.MustCompile(`(?m)^\s*(?:\d{1,2}[.)]|[а-яё]\)|[–—•*]|-)\s+`)
	modalRe      = regexp.MustCompile(`(?i)необходим|следует|должн|нужно|требуется|представить|предоставить|предусмотреть|указать|уточнить|привести|обосновать|откорректировать|скорректировать|дополнить|исключить|устранить|внести|согласовать|отразить|подтвердить|разработать|добавить|исправить|учесть|проверить|исключ`)
	sentenceEnd  = regexp.MustCompile(`[.!?]\s+`)
)

func cleanRequirement(s string) string {
	s = strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	s = strings.TrimLeft(s, "–—•*-;, ")
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > maxRequirementLen {
		s = string(r[:maxRequirementLen]) + "…"
	}
	return s
}

// splitSentences cuts a text at the end of a sentence (a full stop, a capital letter after it); a stop inside
// an abbreviation or a number ("п. 4.2", "стр. 5") does not end one.
func splitSentences(text string) []string {
	var out []string
	start := 0
	for _, loc := range sentenceEnd.FindAllStringIndex(text, -1) {
		end := loc[1]
		if end >= len(text) {
			continue
		}
		next := []rune(text[end:])[0]
		prev := strings.TrimSpace(text[start:loc[0]])
		last := prev
		if i := strings.LastIndexAny(prev, " \n"); i >= 0 {
			last = prev[i+1:]
		}
		short := len([]rune(last)) <= 3 && strings.IndexFunc(last, unicode.IsLetter) >= 0 && strings.ToLower(last) == last // "п", "стр", "рис"
		if !unicode.IsUpper(next) || short {
			continue
		}
		out = append(out, text[start:loc[0]+1])
		start = end
	}
	if rest := strings.TrimSpace(text[start:]); rest != "" {
		out = append(out, text[start:])
	}
	return out
}

// SplitRequirements cuts the text of a remark into its requirements. A remark in which nothing can
// be told apart is one requirement — the whole text.
func SplitRequirements(text string) []Requirement {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	whole := []Requirement{{Idx: 1, Text: cleanRequirement(text)}}

	var parts []string
	// 1. a numbered or bulleted list
	if loc := lineMarkRe.FindAllStringIndex(text, -1); len(loc) >= 2 {
		parts = cutAt(text, loc, 0)
	} else if loc := inlineMarkRe.FindAllStringSubmatchIndex(text, -1); len(loc) >= 2 {
		var at [][]int
		for _, m := range loc {
			at = append(at, []int{m[2], m[3]})
		}
		parts = cutAt(text, at, 0)
	}
	if len(parts) >= 2 {
		// the part before the first marker is the preamble: «Необходимо:» is a heading, a sentence with a modal word is a requirement
		var out []Requirement
		for i, p := range parts {
			c := cleanRequirement(p)
			if len([]rune(c)) < minRequirementLen {
				continue
			}
			if i == 0 && (!modalRe.MatchString(c) || strings.HasSuffix(strings.TrimSpace(p), ":")) {
				continue
			}
			out = append(out, Requirement{Text: c})
		}
		if len(out) >= 2 {
			return number(out)
		}
		return whole
	}
	// 2. sentences with a modal word
	var out []Requirement
	for _, sn := range splitSentences(text) {
		if c := cleanRequirement(sn); len([]rune(c)) >= minRequirementLen && modalRe.MatchString(c) {
			out = append(out, Requirement{Text: c})
		}
	}
	if len(out) >= 2 {
		return number(out)
	}
	return whole
}

// cutAt cuts text at the given marker spans; the first element is what precedes the first marker.
func cutAt(text string, marks [][]int, _ int) []string {
	var parts []string
	prev := 0
	for _, m := range marks {
		parts = append(parts, text[prev:m[0]])
		prev = m[1]
	}
	return append(parts, text[prev:])
}

func number(rs []Requirement) []Requirement {
	seen := map[string]bool{}
	var out []Requirement
	for _, r := range rs {
		k := strings.ToLower(r.Text)
		if seen[k] || len(out) >= maxRequirements {
			continue
		}
		seen[k] = true
		r.Idx = len(out) + 1
		out = append(out, r)
	}
	return out
}

// AggregateOutcome folds the results of the requirements into the result of the remark. All fixed → fixed;
// a requirement that cannot be told stops the remark from being called fixed («Требует проверки») unless one
// of them is certainly not fixed; not fixed alone → not fixed; anything else mixed → partly fixed.
func AggregateOutcome(outcomes []string) string {
	var fixed, partial, notFixed, check int
	for _, o := range outcomes {
		switch o {
		case OutcomeFixed:
			fixed++
		case OutcomePartial:
			partial++
		case OutcomeNotFixed:
			notFixed++
		default:
			check++
		}
	}
	switch {
	case len(outcomes) == 0:
		return OutcomeCheck
	case notFixed == 0 && partial == 0 && check == 0:
		return OutcomeFixed
	case notFixed == 0 && check > 0:
		return OutcomeCheck
	case fixed == 0 && partial == 0:
		return OutcomeNotFixed
	}
	return OutcomePartial
}

// AggregateRequirements is AggregateOutcome over the effective results of a list.
func AggregateRequirements(items []Requirement) string {
	var o []string
	for _, q := range items {
		o = append(o, q.Effective())
	}
	return AggregateOutcome(o)
}

// RequirementCounts counts the requirements by their effective result.
func RequirementCounts(items []Requirement) map[string]int {
	out := map[string]int{}
	for _, q := range items {
		out[q.Effective()]++
	}
	return out
}

// EnsureRequirements returns the list a remark should have for its text: the stored one when the text is
// the same (or the specialist edited the list), a new one when the text is new or changed. changed says
// whether the list is not the stored one.
func EnsureRequirements(stored RequirementSet, text string) (set RequirementSet, changed bool) {
	h := TextHash(text)
	if len(stored.Items) > 0 && stored.For == h {
		return stored, false
	}
	if len(stored.Items) > 0 && stored.Manual {
		// the specialist's list outlives a change of the text; it says so and waits to be looked at
		if !stored.Stale {
			stored.Stale = true
			return stored, true
		}
		return stored, false
	}
	items := SplitRequirements(text)
	if len(items) == 0 {
		return RequirementSet{}, len(stored.Items) > 0
	}
	return RequirementSet{For: h, Items: items}, true
}

// RequirementInput is a requirement as the specialist sends it when editing the list.
type RequirementInput struct {
	Text           string `json:"text"`
	OutcomeFinal   string `json:"outcomeFinal"`
	OutcomeComment string `json:"outcomeComment"`
}

// ApplyRequirementEdit builds the list the specialist asked for: a requirement whose text did not change
// keeps its results, a new or changed one starts with none (the next check fills them). Decisions are
// validated like the decision on a remark: another result than the system's needs a comment.
func ApplyRequirementEdit(stored RequirementSet, text string, in []RequirementInput, by, at string) (RequirementSet, error) {
	if len(in) == 0 {
		return RequirementSet{}, fmt.Errorf("requirements: the list must not be empty")
	}
	if len(in) > maxRequirements {
		return RequirementSet{}, fmt.Errorf("requirements: at most %d", maxRequirements)
	}
	byText := map[string]Requirement{}
	for _, q := range stored.Items {
		byText[strings.ToLower(q.Text)] = q
	}
	out := RequirementSet{For: TextHash(text), Manual: true}
	for i, x := range in {
		t := cleanRequirement(x.Text)
		if len([]rune(t)) < 2 {
			return RequirementSet{}, fmt.Errorf("requirements: requirement %d is empty", i+1)
		}
		q := Requirement{Idx: i + 1, Text: t, Manual: true}
		if old, ok := byText[strings.ToLower(t)]; ok {
			q = old
			q.Idx, q.Manual = i+1, true
		}
		if kind, err := ResolveDecision(q.Outcome, x.OutcomeFinal, strings.TrimSpace(x.OutcomeComment)); err != nil {
			return RequirementSet{}, fmt.Errorf("requirements: requirement %d: %w", i+1, err)
		} else if kind == "" {
			q.OutcomeFinal, q.OutcomeComment, q.OutcomeBy, q.OutcomeAt = "", "", "", ""
		} else if q.OutcomeFinal != x.OutcomeFinal || q.OutcomeComment != strings.TrimSpace(x.OutcomeComment) {
			q.OutcomeFinal, q.OutcomeComment, q.OutcomeBy, q.OutcomeAt = x.OutcomeFinal, strings.TrimSpace(x.OutcomeComment), by, at
		}
		out.Items = append(out.Items, q)
	}
	return out, nil
}
