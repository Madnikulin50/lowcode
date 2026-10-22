package cmp

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Идентификация редакции документа (ТЗ 3.1.6, 4.2.5): том определяется по шифру, редакция — по
// номеру изменения, а не по имени PDF-файла. Номер изменения читается там, где его пишут:
//   - на титульном листе рядом с шифром — «03102022/РД-3-ПИК1-НК2 (Изм. 2)» (cover);
//   - в имени файла — «…КЖ6-1 изм. 2.pdf» (name; только для файла из одного тома).
// Номер, заданный вручную в записи документа, агент не перезаписывает (пока файл тот же).

// Where a change number was read from.
const (
	ChangeFromCover = "cover"
	ChangeFromName  = "name"
)

var (
	changeCoverRe = regexp.MustCompile(`(?i)\(\s*(?:изм|изменение)\.?\s*№?\s*(\d{1,2})\s*\)`)
	changeLineRe  = regexp.MustCompile(`(?i)^(?:изм|изменение)\.?\s*№?\s*(\d{1,2})$`)
	changeNameRe  = regexp.MustCompile(`(?i)(?:^|[\s_\-(.])(?:изм|изменение)\.?\s*№?\s*(\d{1,2})(?:[\s_.)\-]|$)`)

	bookRe       = regexp.MustCompile(`(?i)(?:^|[^\p{L}])книга\s+(\d{1,2})(?:[^\p{L}\d]|$)`)
	partRe       = regexp.MustCompile(`(?i)(?:^|[^\p{L}])часть\s+(\d{1,2})(?:[^\p{L}\d]|$)`)
	subsectionRe = regexp.MustCompile(`(?i)(?:^|[^\p{L}])подраздел\s+(\d{1,2}(?:\.\d{1,2})*)(?:[^\p{L}\d]|$)`)
)

// DetectChange reads the change number of a volume: from the title page of the volume (its first
// three pages), then — for a file that is one volume — from the file name. 0 is the original edition
// (nothing was found).
func DetectChange(d *Document, from, to int, single bool) (int, string) {
	for pn := from; pn <= to && pn < from+3 && pn <= len(d.Pages); pn++ {
		for _, l := range d.Pages[pn-1].Lines {
			if l.Stamp {
				continue
			}
			if m := changeCoverRe.FindStringSubmatch(l.Text); m != nil {
				n, _ := strconv.Atoi(m[1])
				return n, ChangeFromCover
			}
			if m := changeLineRe.FindStringSubmatch(strings.TrimSpace(l.Text)); m != nil {
				n, _ := strconv.Atoi(m[1])
				return n, ChangeFromCover
			}
		}
	}
	if single {
		if m := changeNameRe.FindStringSubmatch(strings.TrimSuffix(d.Name, ".pdf")); m != nil {
			n, _ := strconv.Atoi(m[1])
			return n, ChangeFromName
		}
	}
	return 0, ""
}

// DetectClass reads the part («Часть 2»), the book («Книга 1») and the subsection («Подраздел 3») a
// volume names on its title pages; "" when it names none.
func DetectClass(d *Document, from, to int) (part, book, subsection string) {
	for pn := from; pn <= to && pn < from+3 && pn <= len(d.Pages); pn++ {
		for _, l := range d.Pages[pn-1].Lines {
			if l.Stamp || len([]rune(l.Text)) > 120 {
				continue
			}
			if m := partRe.FindStringSubmatch(l.Text); m != nil && part == "" {
				part = m[1]
			}
			if m := bookRe.FindStringSubmatch(l.Text); m != nil && book == "" {
				book = m[1]
			}
			if m := subsectionRe.FindStringSubmatch(l.Text); m != nil && subsection == "" {
				subsection = m[1]
			}
		}
	}
	return
}

// ChangeLabel is "изм. 2" ("" for the original edition).
func ChangeLabel(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("изм. %d", n)
}

// supersede drops, from the volumes of one side, the older editions of a volume that is there
// again with a higher change number: a «КЖ6-1» and a «КЖ6-1 изм. 2» in one set are one volume, and only
// the newest is compared. Volumes with the same change number are all kept (they may be parts).
func supersede(vols []*Volume) (keep, dropped []*Volume) {
	newest := map[string]int{}
	for _, v := range vols {
		if v.Code != "" && v.Change > newest[v.Code] {
			newest[v.Code] = v.Change
		}
	}
	for _, v := range vols {
		if v.Code != "" && v.Change < newest[v.Code] {
			dropped = append(dropped, v)
			continue
		}
		keep = append(keep, v)
	}
	return
}

// editionNotes says what the change numbers of a pair tell about the choice of the compared editions.
// For two editions of one stage («До» → «После») the second must be newer; for a ПД → РД pair the
// numbers are the editions of two different stages and are not ordered against each other.
func editionNotes(pairs []*Pair, mode string) []string {
	var out []string
	if mode != ModeVersions {
		return nil
	}
	seen := map[string]bool{}
	for _, p := range pairs {
		for _, t := range p.Targets {
			if p.Ref.Change == 0 || t.Change == 0 {
				continue
			}
			var s string
			switch {
			case p.Ref.Change > t.Change:
				s = fmt.Sprintf("«%s»: до — %s, после — %s: «после» старше «до». Возможно, редакции выбраны в обратном порядке.", trunc(volLabel(p.Ref), 60), ChangeLabel(p.Ref.Change), ChangeLabel(t.Change))
			case p.Ref.Change == t.Change:
				s = fmt.Sprintf("«%s»: у обеих редакций один номер изменения (%s). Сравнивается один и тот же выпуск?", trunc(volLabel(p.Ref), 60), ChangeLabel(p.Ref.Change))
			}
			if s != "" && !seen[s] {
				seen[s] = true
				out = append(out, s)
			}
		}
	}
	sort.Strings(out)
	return out
}

// supersededNote lists the older editions that were left out of the comparison.
func supersededNote(side string, dropped []*Volume) string {
	if len(dropped) == 0 {
		return ""
	}
	var parts []string
	for _, v := range dropped {
		label := volLabel(v)
		if c := ChangeLabel(v.Change); c != "" {
			label += " (" + c + ")"
		} else {
			label += " (исходная редакция)"
		}
		parts = append(parts, label)
	}
	name := "эталона"
	if side != SideRef {
		name = "проверяемой стороны"
	}
	return fmt.Sprintf("Не сравнивались старые редакции томов %s — в загруженных файлах есть более новые: %s.", name, strings.Join(parts, "; "))
}

// EditionValues are the fields of a document record the agent fills from the volumes found in the
// file: the hash of the file (the version of the uploaded PDF), the change number (the edition) and
// the part, book and subsection. A file that is the same as the one read before (same hash) keeps
// whatever the specialist typed: only empty fields are filled. A new file replaces all of them — what
// was typed belonged to the old one.
func EditionValues(existing *idcheck.Record, hash string, vs []VolumeInfo) idcheck.Values {
	same := existing != nil && existing.Get("file_hash") == hash && hash != ""
	auto := map[string]string{}
	change, mixed := -1, false
	for _, v := range vs {
		if change >= 0 && v.Change != change {
			mixed = true
		}
		change = v.Change
		for k, x := range map[string]string{"part": v.Part, "book": v.Book, "subsection": v.Subsection} {
			if auto[k] == "" {
				auto[k] = x
			}
		}
	}
	if change > 0 && !mixed {
		auto["change_no"] = strconv.Itoa(change)
	}
	out := idcheck.Values{"file_hash": []string{hash}}
	for _, k := range []string{"change_no", "part", "book", "subsection"} {
		if same && existing.Get(k) != "" {
			continue
		}
		if same && auto[k] == "" {
			continue
		}
		out[k] = []string{auto[k]}
	}
	return out
}
