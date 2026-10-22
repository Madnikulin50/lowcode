package cmp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// A list of remarks (список замечаний) is made from the findings of a ПД/РД
// comparison that the specialist confirmed: findings on the same sheet pair
// become one remark. The list goes to the registry of the object's remarks,
// to Excel in the format of an expertise export, or to a Word letter.

// FindingIn is a finding as the list builder needs it.
type FindingIn struct {
	ID          string
	Kind        string
	Severity    string
	Description string
	Anchor      string
	Status      string // new | confirmed | rejected
	IterState   string // "" | new | persisting | resolved
	DocRef      string
	DocTarget   string
	PageRef     string
	PageTarget  string
	PageNumber  string
	BoxRef      string // region on the reference page, "x0,y0,x1,y1" in page fractions
	BoxTarget   string
	Evidence    string // JSON of Evidence
}

// IssueInput is a comparison with its findings.
type IssueInput struct {
	ComparisonID    string
	ComparisonTitle string
	Findings        []FindingIn
	DocName         map[string]string // document ID → name shown in the text
	Mode            string            // ModeStages (default) | ModeVersions
}

// Issued is one remark of the list.
type Issued struct {
	GroupKey    string   `json:"groupKey"` // stable across re-generations of the same sheet pair
	Number      string   `json:"number"`   // assigned when the list is numbered
	Text        string   `json:"text"`     // «Вывод о несоответствии»
	Refs        string   `json:"refs"`     // «Ссылка на материалы»
	Section     string   `json:"section"`
	Basis       string   `json:"basis"`
	Criticality string   `json:"criticality"` // critical | noncritical
	Count       int      `json:"count"`       // findings in the remark
	FindingIDs  []string `json:"findingIDs"`
	DocIDs      []string `json:"docIDs"`
	Pages       string   `json:"pages"` // "ПД стр. 3 · РД стр. 5"
}

// IssueResult is the outcome of BuildIssue.
type IssueResult struct {
	Active   []Issued `json:"active"`   // confirmed findings that are still there
	Fixed    []Issued `json:"fixed"`    // sheet pairs whose confirmed findings were all fixed (iteration)
	Excluded int      `json:"excluded"` // findings left out (not confirmed / informational)
}

const (
	versionsBasis = "Предыдущая версия документации (эталон): новая версия должна ей соответствовать либо изменения должны быть обоснованы"
	versionsAlign = "Требуется пояснить и обосновать изменение относительно предыдущей версии документации либо устранить расхождение."
	issueBasis    = "Проектная документация стадии П (эталон): рабочая документация должна ей соответствовать"
	issueAlign    = "Требуется привести рабочую документацию в соответствие с проектной документацией (эталон — стадия П) либо представить обоснование изменения."
)

// informational kinds are never remarks: a volume that is not released yet or
// out of the compared scope is not a defect of the documents.
func informational(kind string) bool { return kind == KindNotRelease || kind == KindScopeDiff }

func cleanDescription(s string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(s, "[детерм.] "), "[ИИ] "))
}

func pageOf(f FindingIn) (ref, tgt string) {
	ref, tgt = f.PageRef, f.PageTarget
	if ref == "" && tgt == "" && f.PageNumber != "" { // old findings carry one page number only
		ref, tgt = f.PageNumber, f.PageNumber
	}
	return
}

// BuildIssue groups the confirmed findings by the sheet pair they are on and
// words each group as one remark.
func BuildIssue(in IssueInput) IssueResult {
	type group struct {
		key              string
		docRef, docTgt   string
		pageRef, pageTgt string
		items            []FindingIn
		resolved, open   int
		volRef, volTgt   string
		severity         string
		formatOnly       bool // only formatting findings: worded as a remark on the documents' form
	}
	groups := map[string]*group{}
	var order []string
	var res IssueResult
	rank := map[string]int{"high": 0, "medium": 1, "low": 2}
	for _, f := range in.Findings {
		if informational(f.Kind) || f.Status != StatusConfirmed {
			res.Excluded++
			continue
		}
		pr, pt := pageOf(f)
		k := fmt.Sprintf("%s|%s|%s|%s", f.DocRef, pr, f.DocTarget, pt)
		g := groups[k]
		if g == nil {
			g = &group{key: "cmp-" + shortHash(in.ComparisonID+"|"+k), docRef: f.DocRef, docTgt: f.DocTarget, pageRef: pr, pageTgt: pt, severity: "low", formatOnly: true}
			groups[k] = g
			order = append(order, k)
		}
		g.items = append(g.items, f)
		if f.Kind != KindFormat {
			g.formatOnly = false
		}
		if f.IterState == IterResolved {
			g.resolved++
		} else {
			g.open++
		}
		if rank[f.Severity] < rank[g.severity] {
			g.severity = f.Severity
		}
		var ev Evidence
		if f.Evidence != "" {
			_ = json.Unmarshal([]byte(f.Evidence), &ev)
		}
		if g.volRef == "" {
			g.volRef = ev.VolumeRef
		}
		if g.volTgt == "" {
			g.volTgt = ev.VolumeTgt
		}
	}
	// reading order: by the target sheet, then the reference sheet
	sort.SliceStable(order, func(i, j int) bool {
		a, b := groups[order[i]], groups[order[j]]
		if a.docTgt != b.docTgt {
			return a.docTgt < b.docTgt
		}
		if x, y := atoiOr(a.pageTgt, 1<<30), atoiOr(b.pageTgt, 1<<30); x != y {
			return x < y
		}
		return atoiOr(a.pageRef, 1<<30) < atoiOr(b.pageRef, 1<<30)
	})
	name := func(id, fallback string) string {
		if n := in.DocName[id]; n != "" {
			return n
		}
		return fallback
	}
	for _, k := range order {
		g := groups[k]
		basis, align, intro := issueBasis, issueAlign, "выявлены расхождения между проектной и рабочей документацией.\n"
		if in.Mode == ModeVersions {
			basis, align, intro = versionsBasis, versionsAlign, "выявлены расхождения между версиями документации.\n"
		}
		if g.formatOnly {
			basis = "Требования к оформлению проектной и рабочей документации"
			align = "Требуется устранить замечания по оформлению документации."
			intro = "выявлены замечания по оформлению документации.\n"
		}
		is := Issued{GroupKey: g.key, Basis: basis, Criticality: "noncritical", Count: len(g.items)}
		if g.severity == "high" {
			is.Criticality = "critical"
		}
		var where, refs, pages []string
		if g.pageTgt != "" {
			where = append(where, fmt.Sprintf("лист %s РД", g.pageTgt))
			refs = append(refs, fmt.Sprintf("РД: %s, стр. %s", name(g.docTgt, "РД"), g.pageTgt))
			pages = append(pages, "РД стр. "+g.pageTgt)
		}
		if g.pageRef != "" {
			where = append(where, fmt.Sprintf("лист %s ПД", g.pageRef))
			refs = append(refs, fmt.Sprintf("ПД: %s, стр. %s", name(g.docRef, "ПД"), g.pageRef))
			pages = append(pages, "ПД стр. "+g.pageRef)
		}
		is.Refs = Relabel(strings.Join(refs, "; "), in.Mode)
		is.Pages = Relabel(strings.Join(pages, " · "), in.Mode)
		is.Section = firstNonEmpty(g.volTgt, g.volRef)
		var b strings.Builder
		if len(where) > 0 {
			fmt.Fprintf(&b, "%s: %s", Relabel(strings.Join(where, " / "), in.Mode), intro)
		} else {
			b.WriteString(capitalize(intro))
		}
		for i, f := range g.items {
			text := cleanDescription(f.Description)
			if len(g.items) > 1 {
				fmt.Fprintf(&b, "%d) %s\n", i+1, text)
			} else {
				fmt.Fprintf(&b, "%s\n", text)
			}
			is.FindingIDs = append(is.FindingIDs, f.ID)
		}
		b.WriteString(align)
		is.Text = b.String()
		for _, d := range []string{g.docRef, g.docTgt} {
			if d != "" && !hasString(is.DocIDs, d) {
				is.DocIDs = append(is.DocIDs, d)
			}
		}
		if g.open == 0 && g.resolved > 0 {
			res.Fixed = append(res.Fixed, is)
		} else {
			res.Active = append(res.Active, is)
		}
	}
	return res
}

// Number gives the remarks of a list numbers "<prefix><batch>.<n>".
func Number(list []Issued, prefix string, batch int) {
	for i := range list {
		list[i].Number = fmt.Sprintf("%s%d.%d", prefix, batch, i+1)
	}
}

func atoiOr(s string, def int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return def
	}
	return n
}

func hasString(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func capitalize(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
