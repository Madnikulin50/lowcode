package cmp

import (
	"regexp"
	"sort"
	"strings"
)

// Key parameters that must agree between stages: concrete grade, water
// resistance, frost resistance, rebar class. Extracted with regexes from
// the text of a volume; values are canonicalized ("B25", "W6", "F150",
// "A500C").
type paramKind struct {
	Name  string // human label
	Re    *regexp.Regexp
	Canon func(m []string) string
	Valid map[string]bool
}

func set(vals ...string) map[string]bool {
	m := map[string]bool{}
	for _, v := range vals {
		m[v] = true
	}
	return m
}

var paramKinds = []paramKind{
	{
		Name: "Класс бетона",
		// "В25", "B25", "В 7,5" (Cyrillic or Latin В), not part of a longer word
		Re:    regexp.MustCompile(`(?:^|[^0-9A-Za-zА-Яа-я])[ВB]\s?(\d{1,2}(?:[.,]\d)?)(?:[^0-9A-Za-zА-Яа-я]|$)`),
		Canon: func(m []string) string { return "B" + strings.Replace(m[1], ",", ".", 1) },
		Valid: set("B7.5", "B10", "B12.5", "B15", "B20", "B22.5", "B25", "B27.5", "B30", "B35", "B40", "B45", "B50", "B55", "B60", "B70", "B80"),
	},
	{
		Name:  "Водонепроницаемость",
		Re:    regexp.MustCompile(`(?:^|[^0-9A-Za-zА-Яа-я])W\s?(\d{1,2})(?:[^0-9A-Za-zА-Яа-я]|$)`),
		Canon: func(m []string) string { return "W" + m[1] },
		Valid: set("W2", "W4", "W6", "W8", "W10", "W12", "W14", "W16", "W18", "W20"),
	},
	{
		Name:  "Морозостойкость",
		Re:    regexp.MustCompile(`(?:^|[^0-9A-Za-zА-Яа-я])F(?:[12]\s)?\s?(\d{2,4})(?:[^0-9A-Za-zА-Яа-я]|$)`),
		Canon: func(m []string) string { return "F" + m[1] },
		Valid: set("F50", "F75", "F100", "F150", "F200", "F300", "F400", "F500", "F600", "F800", "F1000"),
	},
	{
		Name: "Класс арматуры",
		// "А500С", "A400", "А-III" excluded: only numeric classes
		Re: regexp.MustCompile(`(?:^|[^0-9A-Za-zА-Яа-я])[АA]\s?-?(240|300|400|500|600|800|1000)\s?([СC]?)(?:[^0-9A-Za-zА-Яа-я]|$)`),
		Canon: func(m []string) string {
			s := "A" + m[1]
			if m[2] != "" {
				s += "C"
			}
			return s
		},
		Valid: set("A240", "A300", "A400", "A500", "A500C", "A600", "A800", "A1000"),
	},
}

// ParamHit is where a parameter value occurs.
type ParamHit struct {
	Value string
	Line  Line
}

// ExtractParams returns, per parameter kind, the canonical values found in
// lines together with the first line each was seen on.
func ExtractParams(lines []Line) []map[string]ParamHit {
	out := make([]map[string]ParamHit, len(paramKinds))
	for i := range out {
		out[i] = map[string]ParamHit{}
	}
	for _, l := range lines {
		for i, pk := range paramKinds {
			// overlapping matches: allow a delimiter to be shared between
			// neighbours by matching on a sliding basis
			text := l.Text
			for off := 0; off < len(text); {
				loc := pk.Re.FindStringSubmatchIndex(text[off:])
				if loc == nil {
					break
				}
				m := make([]string, len(loc)/2)
				for g := range m {
					if loc[2*g] >= 0 {
						m[g] = text[off+loc[2*g] : off+loc[2*g+1]]
					}
				}
				if v := pk.Canon(m); pk.Valid[v] {
					if _, seen := out[i][v]; !seen {
						out[i][v] = ParamHit{Value: v, Line: l}
					}
				}
				// resume right after the parameter token itself (not the
				// trailing delimiter) so adjacent values are not skipped
				end := loc[3]
				if end <= 0 {
					end = loc[1]
				}
				off += end
			}
		}
	}
	return out
}

// compareParams reports parameters whose values conflict between the
// reference and the target. A target that merely lacks a value the reference
// has, while adding none of its own, is not a conflict (the working stage
// may carry less on a given sheet) — it is reported at low severity only in
// full (non-fragment) pairs.
func compareParams(refLines, tgtLines []Line, fragment bool) []paramDiff {
	r := ExtractParams(refLines)
	t := ExtractParams(tgtLines)
	var out []paramDiff
	for i, pk := range paramKinds {
		var onlyRef, onlyTgt []string
		for v := range r[i] {
			if _, ok := t[i][v]; !ok {
				onlyRef = append(onlyRef, v)
			}
		}
		for v := range t[i] {
			if _, ok := r[i][v]; !ok {
				onlyTgt = append(onlyTgt, v)
			}
		}
		sort.Strings(onlyRef)
		sort.Strings(onlyTgt)
		switch {
		case len(onlyRef) > 0 && len(onlyTgt) > 0:
			out = append(out, paramDiff{pk.Name, "conflict", onlyRef, onlyTgt, r[i], t[i]})
		case len(onlyRef) > 0 && !fragment && len(t[i]) > 0:
			out = append(out, paramDiff{pk.Name, "missing", onlyRef, nil, r[i], t[i]})
		}
	}
	return out
}

type paramDiff struct {
	Name    string
	Mode    string // conflict | missing
	OnlyRef []string
	OnlyTgt []string
	RefHits map[string]ParamHit
	TgtHits map[string]ParamHit
}
