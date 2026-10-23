package cmp

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Автоматическая ведомость томов: агент предлагает состав комплекта, а человек принимает или
// отклоняет. Источники:
//   - contents — таблица «Состав проектной документации» / «Содержание» / «Перечень томов» на
//     начальных листах загруженных файлов (шифр тома + наименование);
//   - set — состав самого раннего комплекта стадии: что выпущено первым, то и ожидается в следующих
//     версиях (дальнейшие комплекты сверяются с ним на комплектность).
// Предложения хранятся в object_volumes с признаком proposed; на комплектность влияют только принятые.

// Sources of a proposed volume.
const (
	SrcManual   = "manual"
	SrcContents = "contents"
	SrcSet      = "set"
)

// ContentsEntry is a volume named in a contents table.
type ContentsEntry struct {
	Code  string
	Title string
	Page  int
}

var (
	volContentsHeadRe = regexp.MustCompile(`(?i)(состав проектной документации|состав рабочей документации|состав раздела|состав тома|содержание тома|содержание|перечень томов|ведомость томов)`)
	tailPageRe        = regexp.MustCompile(`(?i)\s*(?:стр\.?|лист)\s*\d{1,3}\s*$`)
)

// ParseContents reads the volumes named in the contents table of a document. ownCode is the шифр of
// the volume the document is: codes that extend it name parts of this volume, not other volumes.
func ParseContents(pages []Page, ownCode string) []ContentsEntry {
	own := NormCode(ownCode)
	head := 0
	for i := 0; i < len(pages) && i < 8; i++ {
		for _, l := range pages[i].Lines {
			if !l.Stamp && len([]rune(l.Text)) < 60 && volContentsHeadRe.MatchString(l.Key) {
				head = i + 1
				break
			}
		}
		if head != 0 {
			break
		}
	}
	if head == 0 {
		return nil
	}
	var out []ContentsEntry
	seen := map[string]bool{}
	for pn := head; pn <= len(pages) && pn <= head+1; pn++ {
		var lines []Line
		for _, l := range pages[pn-1].Lines {
			if !l.Stamp {
				lines = append(lines, l)
			}
		}
		for i := 0; i < len(lines); i++ {
			text := stationRe.ReplaceAllString(lines[i].Text, " ")
			// a шифр wrapped over two lines: "…-ОП-" / "КЖ1.13-…"
			if strings.HasSuffix(strings.TrimSpace(text), "-") && i+1 < len(lines) {
				text = strings.TrimSpace(text) + strings.TrimSpace(lines[i+1].Text)
				i++
			}
			for _, m := range codeRe.FindAllStringIndex(text, -1) {
				raw := text[m[0]:m[1]]
				n := NormCode(raw)
				if !plausibleCode(n) || len(n) < 8 || seen[n] || longestLetterRun(raw) > 6 {
					continue
				}
				if own != "" && strings.HasPrefix(n, own) { // this volume itself or a part of it
					continue
				}
				title := strings.TrimSpace(strings.Trim(tailPageRe.ReplaceAllString(text[m[1]:], ""), " -–—:;."))
				if title == "" && i+1 < len(lines) && !codeRe.MatchString(lines[i+1].Text) {
					title = strings.TrimSpace(tailPageRe.ReplaceAllString(lines[i+1].Text, ""))
				}
				seen[n] = true
				out = append(out, ContentsEntry{Code: raw, Title: trunc(title, 160), Page: pn})
			}
		}
	}
	return out
}

// VolumeProposal is a suggested row of the statement of volumes.
type VolumeProposal struct {
	Stage    string `json:"stage"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Evidence string `json:"evidence"`
}

// proposeFromSets: the first (earliest) set of a stage that has no statement yet gives its volumes.
func proposeFromSets(od *ObjectDocs, have map[string][]string) []VolumeProposal {
	var out []VolumeProposal
	for _, stage := range []string{SideRef, SideTarget} {
		if len(have[stage]) > 0 {
			continue // the stage already has a statement
		}
		var first *SetInfo
		for i := range od.Sets {
			s := &od.Sets[i]
			if s.Stage == stage && (first == nil) { // Sets are ordered by date, then ID
				first = s
			}
		}
		if first == nil {
			continue
		}
		seen := map[string]bool{}
		for _, d := range first.Docs {
			for _, c := range d.Codes {
				if n := NormCode(c); !seen[n] {
					seen[n] = true
					out = append(out, VolumeProposal{Stage: stage, Code: c, Title: volumeTitleOf(d, c), Source: SrcSet,
						Evidence: fmt.Sprintf("есть в первом комплекте %s %s", StageName(stage), first.Version)})
				}
			}
		}
	}
	return out
}

func volumeTitleOf(d SetDoc, code string) string {
	if len(d.Codes) == 1 {
		return d.Title
	}
	return ""
}

func hasVolume(codes []string, code string) bool {
	for _, c := range codes {
		if volumesMatch(c, code) {
			return true
		}
	}
	return false
}

// SuggestVolumes reads the contents tables of the documents of the object's sets and proposes the
// statement: new rows are written to object_volumes with proposed = yes; volumes the statement
// already has (confirmed or proposed) are not proposed again.
func (r *Runner) SuggestVolumes(ctx context.Context, ns, objectID string) ([]VolumeProposal, error) {
	b := r.Backend
	if _, err := r.CheckSets(ctx, ns, objectID); err != nil { // reads new files, so that every document has its codes
		return nil, err
	}
	od, err := LoadObjectDocs(b, ns, objectID)
	if err != nil {
		return nil, err
	}
	existing, err := b.Search(ns, ModVolumes, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	have := map[string][]string{} // stage → codes already in the statement (a row without a stage counts for both)
	for _, v := range existing {
		st, c := v.Get("stage"), v.Get("volume_code")
		if c == "" || (v.Get("kind") != "" && v.Get("kind") != UnitVolume) {
			continue // a section, an appendix or a reference is not a volume
		}
		if st == "" {
			have[SideRef], have[SideTarget] = append(have[SideRef], c), append(have[SideTarget], c)
		} else {
			have[st] = append(have[st], c)
		}
	}

	var props []VolumeProposal
	// 1. the contents tables of the documents
	for _, s := range od.Sets {
		for _, d := range s.Docs {
			rec, err := b.GetRecord(ns, ModDocuments, d.ID)
			if err != nil || rec.Get("file") == "" {
				continue
			}
			data, name, err := b.Download(ns, rec.Get("file"))
			if err != nil || !IsPDF(data) {
				continue
			}
			pages, _, _, err := ExtractPagesOCR(ctx, d.ID, data, r.CacheDir, r.OCR)
			if err != nil {
				continue
			}
			own := ""
			if len(d.Codes) > 0 {
				own = d.Codes[0]
			}
			for _, e := range ParseContents(pages, own) {
				props = append(props, VolumeProposal{Stage: s.Stage, Code: e.Code, Title: e.Title, Source: SrcContents,
					Evidence: fmt.Sprintf("в содержании файла «%s», стр. %d", name, e.Page)})
			}
		}
	}
	// 2. the first set of a stage that has no statement
	haveWithContents := map[string][]string{SideRef: append([]string{}, have[SideRef]...), SideTarget: append([]string{}, have[SideTarget]...)}
	for _, p := range props {
		haveWithContents[p.Stage] = append(haveWithContents[p.Stage], p.Code)
	}
	props = append(props, proposeFromSets(od, haveWithContents)...)

	var created []VolumeProposal
	for _, p := range props {
		if hasVolume(have[p.Stage], p.Code) {
			continue
		}
		have[p.Stage] = append(have[p.Stage], p.Code)
		v := idcheck.V("object", objectID, "stage", p.Stage, "volume_code", p.Code, "title", p.Title, "source", p.Source, "evidence", p.Evidence, "proposed", "1")
		if _, err := b.Create(ns, ModVolumes, v); err != nil {
			return created, err
		}
		created = append(created, p)
	}
	sort.SliceStable(created, func(i, j int) bool { return created[i].Code < created[j].Code })
	return created, nil
}

// ResolveVolumes accepts (clears the proposed mark) or rejects (deletes) proposed rows of the object.
// accept = nil and rejectAll = false does nothing; acceptAll takes every proposal.
func (r *Runner) ResolveVolumes(ns, objectID string, accept, reject []string, acceptAll bool) (accepted, rejected int, err error) {
	b := r.Backend
	rows, err := b.Search(ns, ModVolumes, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return 0, 0, err
	}
	acc, rej := map[string]bool{}, map[string]bool{}
	for _, id := range accept {
		acc[id] = true
	}
	for _, id := range reject {
		rej[id] = true
	}
	for _, v := range rows {
		if v.Get("proposed") == "" || v.Get("proposed") == "0" {
			continue // only proposals are touched: a confirmed row is never deleted here
		}
		switch {
		case rej[v.ID]:
			if err := b.Delete(ns, ModVolumes, v.ID); err != nil {
				return accepted, rejected, err
			}
			rejected++
		case acc[v.ID] || acceptAll:
			if err := b.Update(ns, ModVolumes, v.ID, idcheck.V("proposed", "0")); err != nil {
				return accepted, rejected, err
			}
			accepted++
		}
	}
	return accepted, rejected, nil
}

// longestLetterRun is the length of the longest run of letters: a шифр is made of short marks
// («ИССО», «ТКР», «КЖ»), a long word in a "code" means a name with dashes, not a volume.
func longestLetterRun(s string) int {
	best, cur := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) {
			cur++
			if cur > best {
				best = cur
			}
		} else {
			cur = 0
		}
	}
	return best
}
