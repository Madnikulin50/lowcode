package cmp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Комплект документов объекта: ПД или РД одной версии. Документы (cmp_documents)
// принадлежат комплекту, сравнение (pd_rd_comparisons) ссылается на два комплекта —
// так всегда видно, какая версия с какой сравнивается и насколько комплект полон.

// Module handles of the sets layer.
const (
	ModSets    = "document_sets"
	ModVolumes = "object_volumes"
)

// Set statuses.
const (
	SetDraft      = "draft"
	SetAccepted   = "accepted"
	SetSuperseded = "superseded"
)

// Cell states of the volume matrix.
const (
	CellPresent  = "present"  // volume is in the set
	CellMissing  = "missing"  // expected by the statement of volumes, not in the set
	CellExtra    = "extra"    // in the set, not in the statement of volumes
	CellNoSet    = "no_set"   // no set of this stage yet
	CellUnparsed = "unparsed" // the document has not been read yet
)

// SetDoc is one file of a set.
type SetDoc struct {
	ID     string   `json:"id"`
	File   string   `json:"file"`
	Codes  []string `json:"codes"`
	Title  string   `json:"title"`
	Pages  int      `json:"pages"`
	Parsed bool     `json:"parsed"`
	// the section (from the code or by hand), the structure (structures record) and the edition
	Section   string `json:"section,omitempty"`
	Structure string `json:"structure,omitempty"`
	ChangeNo  string `json:"changeNo,omitempty"`
	// the appendices of the file and its references to other volumes of the project (ТЗ 3.2.4, 3.2.8)
	Headings []DocHeading `json:"-"`
	Refs     []VolRefAt   `json:"refs,omitempty"`
	// the volumes the file's contents table («Состав проекта») names, with their titles (ТЗ 3.2.5)
	Contents []ContentsEntry `json:"-"`
}

// SetInfo is a set with its completeness.
type SetInfo struct {
	ID         string   `json:"id"`
	Stage      string   `json:"stage"`
	Version    string   `json:"version"`
	Status     string   `json:"status"`
	IssuedAt   string   `json:"issuedAt"`
	Comment    string   `json:"comment"`
	Latest     bool     `json:"latest"`
	Newer      string   `json:"newer,omitempty"` // version of a newer set of the stage (strictly later date, or this one is marked replaced)
	Docs       []SetDoc `json:"docs"`
	Total      int      `json:"total"`   // volumes the statement expects (0: no statement)
	Present    int      `json:"present"` // of them present in the set
	Pct        int      `json:"pct"`
	Missing    []string `json:"missing"`
	Extra      []string `json:"extra"`
	Duplicates []string `json:"duplicates"`
	// Renamed: volumes the set has under the right шифр but with a title that shares no words with the
	// title of the statement of volumes (ТЗ 3.2.5)
	Renamed []string `json:"renamed"`
	// sections, appendices and referenced documents of the statement (ТЗ 3.2.1, 3.2.8) and the volumes the
	// documents refer to that the object does not have (ТЗ 3.2.4: for information)
	Units       []UnitState `json:"units"`
	RefsMissing []string    `json:"refsMissing"`
	Unparsed    int         `json:"unparsed"`
	Note        string      `json:"note"`
}

// Label is "ПД v2" / "РД v1".
func (s SetInfo) Label() string { return StageName(s.Stage) + " " + s.Version }

// StageName is the short name of a stage ("ПД" / "РД").
func StageName(stage string) string {
	if stage == SideRef {
		return "ПД"
	}
	return "РД"
}

// MatrixCell is the state of one volume in the latest set of a stage.
type MatrixCell struct {
	State   string   `json:"state"`
	SetID   string   `json:"setID,omitempty"`
	Version string   `json:"version,omitempty"`
	Docs    []string `json:"docs,omitempty"`
	Code    string   `json:"code,omitempty"` // the шифр as written in the document
}

// MatrixRow is one volume across the stages.
type MatrixRow struct {
	Code     string     `json:"code"`
	Title    string     `json:"title"`
	Expected bool       `json:"expected"`
	Ref      MatrixCell `json:"ref"`
	Target   MatrixCell `json:"target"`
}

// PairSide is a set as one side of a comparison.
type PairSide struct {
	SetID   string `json:"setID"`
	Label   string `json:"label"`
	Version string `json:"version"`
	Latest  bool   `json:"latest"`
	Present int    `json:"present"`
	Total   int    `json:"total"`
	Pct     int    `json:"pct"`
}

// SetPair is a comparison: which version of which set is compared with which.
type SetPair struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Iteration  int       `json:"iteration"`
	Kind       string    `json:"kind"` // stages | versions
	Status     string    `json:"status"`
	Similarity string    `json:"similarity"`
	Label      string    `json:"label"` // "ПД v2 ↔ РД v1"
	Ref        *PairSide `json:"ref,omitempty"`
	Target     *PairSide `json:"target,omitempty"`
	Warnings   []string  `json:"warnings"`
}

// VolumeRow is a proposed row of the statement of volumes.
type VolumeRow struct {
	ID       string `json:"id"`
	Stage    string `json:"stage"`
	Code     string `json:"code"`
	Title    string `json:"title"`
	Source   string `json:"source"`
	Evidence string `json:"evidence"`
}

// ObjectDocs is everything the «Документы» view of an object shows.
type ObjectDocs struct {
	ObjectID   string      `json:"objectID"`
	ObjectName string      `json:"objectName"`
	Sets       []SetInfo   `json:"sets"`
	Matrix     []MatrixRow `json:"matrix"`
	Pairs      []SetPair   `json:"pairs"`
	Proposals  []VolumeRow `json:"proposals"` // the statement the agent suggests, to accept or reject
	// the structures of the object, the documents grouped by them (ТЗ 3.1.2, 3.1.7) and the upload journal (ТЗ 3.1.4)
	Structures []Structure     `json:"structures"`
	Tree       []TreeStructure `json:"tree"`
	Uploads    []Upload        `json:"uploads"`
	Expected   struct {
		Ref    int `json:"ref"`
		Target int `json:"target"`
	} `json:"expected"`
}

type expectedVolume struct {
	code, title, stage string // stage: ref | target | "" (both)
	required           bool
	kind, parent       string // UnitVolume…; parent: the volume an appendix belongs to
}

func volumesMatch(a, b string) bool { return sameVolume(a, b) || sameVolume(b, a) }

func setDocOf(rec idcheck.Record) SetDoc {
	d := SetDoc{ID: rec.ID, File: rec.Get("file"), Title: rec.Get("volume_title"), Codes: []string{}}
	for _, c := range strings.Split(rec.Get("volume_code"), ",") {
		if c = strings.TrimSpace(c); c != "" {
			d.Codes = append(d.Codes, c)
		}
	}
	d.Pages, _ = strconv.Atoi(rec.Get("pages"))
	d.Section, d.Structure = rec.Get("section"), rec.Get("structure")
	d.Headings, d.Refs = ParseHeadings(rec.Get("headings_json")), parseVolRefs(rec.Get("volume_refs_json"))
	d.Contents = parseContentsJSON(rec.Get("contents_json"))
	if ch := rec.Get("change_no"); ch != "0" {
		d.ChangeNo = ch
	}
	d.Parsed = len(d.Codes) > 0 || d.Pages > 0
	return d
}

// LoadObjectDocs reads the sets, the statement of volumes and the comparisons
// of an object and works out completeness, the volume matrix and the pairs.
// Documents are taken as they were read earlier (CheckSets reads new ones).
func LoadObjectDocs(b Backend, ns, objectID string) (*ObjectDocs, error) {
	out := &ObjectDocs{ObjectID: objectID, Sets: []SetInfo{}, Matrix: []MatrixRow{}, Pairs: []SetPair{}, Proposals: []VolumeRow{}}
	if obj, err := b.GetRecord(ns, ModObjects, objectID); err == nil {
		out.ObjectName = obj.Get("name")
	}
	sets, err := b.Search(ns, ModSets, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	vols, err := b.Search(ns, ModVolumes, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	var expected, units []expectedVolume
	for _, v := range vols {
		if p := v.Get("proposed"); p != "" && p != "0" {
			out.Proposals = append(out.Proposals, VolumeRow{ID: v.ID, Stage: v.Get("stage"), Code: v.Get("volume_code"), Title: v.Get("title"), Source: v.Get("source"), Evidence: v.Get("evidence")})
			continue // a proposal is not part of the statement until a person accepts it
		}
		if c := v.Get("volume_code"); c != "" {
			st := v.Get("stage")
			if st != SideRef && st != SideTarget {
				st = ""
			}
			e := expectedVolume{code: c, title: v.Get("title"), stage: st, required: v.Get("required") != "0", kind: firstNonEmpty(v.Get("kind"), UnitVolume), parent: v.Get("parent_code")}
			if e.kind != UnitVolume {
				units = append(units, e)
				continue
			}
			expected = append(expected, e)
		}
	}
	sort.SliceStable(expected, func(i, j int) bool { return expected[i].code < expected[j].code })
	unitOrder := map[string]int{UnitSection: 0, UnitAppendix: 1, UnitReference: 2}
	sort.SliceStable(units, func(i, j int) bool {
		if units[i].kind != units[j].kind {
			return unitOrder[units[i].kind] < unitOrder[units[j].kind]
		}
		return units[i].code < units[j].code
	})

	for _, s := range sets {
		si := SetInfo{ID: s.ID, Stage: s.Get("stage"), Version: s.Get("version"), Status: s.Get("status"), IssuedAt: s.Get("issued_at"), Comment: s.Get("comment"),
			Docs: []SetDoc{}, Missing: []string{}, Extra: []string{}, Duplicates: []string{}, Renamed: []string{}, Units: []UnitState{}, RefsMissing: []string{}}
		if si.Stage != SideRef {
			si.Stage = SideTarget
		}
		if si.Version == "" {
			si.Version = "v?"
		}
		docs, err := b.Search(ns, ModDocuments, fmt.Sprintf("set = '%s'", s.ID))
		if err != nil {
			return nil, err
		}
		for _, d := range docs {
			if d.Get("file") != "" {
				si.Docs = append(si.Docs, setDocOf(d))
			}
		}
		sort.SliceStable(si.Docs, func(i, j int) bool { return si.Docs[i].ID < si.Docs[j].ID })
		out.Sets = append(out.Sets, si)
	}
	sort.SliceStable(out.Sets, func(i, j int) bool {
		a, c := out.Sets[i], out.Sets[j]
		if a.Stage != c.Stage {
			return a.Stage == SideRef
		}
		if a.IssuedAt != c.IssuedAt {
			return a.IssuedAt < c.IssuedAt
		}
		return a.ID < c.ID
	})
	// latest = the newest set of the stage that is not replaced
	latest := map[string]int{}
	for i, s := range out.Sets {
		if s.Status != SetSuperseded {
			latest[s.Stage] = i
		}
	}
	for _, i := range latest {
		out.Sets[i].Latest = true
	}
	// "not the latest" is claimed only when it is certain: the set is marked as replaced, or a
	// set of the same stage has a strictly later issue date (sets without dates are just packs)
	for i := range out.Sets {
		s := &out.Sets[i]
		l := latestOf(out.Sets, s.Stage)
		switch {
		case l == nil || l.ID == s.ID:
		case s.Status == SetSuperseded:
			s.Newer = l.Version
		case s.IssuedAt != "" && l.IssuedAt > s.IssuedAt:
			s.Newer = l.Version
		}
	}

	for _, e := range expected {
		if e.stage != SideTarget {
			out.Expected.Ref++
		}
		if e.stage != SideRef {
			out.Expected.Target++
		}
	}
	for i := range out.Sets {
		completeness(&out.Sets[i], expected, units)
	}
	for i := range out.Sets {
		if r := missingRefs(&out.Sets[i], out.Sets); len(r) > 0 {
			out.Sets[i].RefsMissing = r
			out.Sets[i].Note = setNote(&out.Sets[i])
		}
	}
	out.Matrix = buildMatrix(out.Sets, expected)

	byID := map[string]*SetInfo{}
	for i := range out.Sets {
		byID[out.Sets[i].ID] = &out.Sets[i]
	}
	comps, err := b.Search(ns, ModComparisons, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	for _, c := range comps {
		mode := ModeStages
		if c.Get("kind") == ModeVersions {
			mode = ModeVersions
		}
		refName, tgtName := SideNames(mode)
		p := SetPair{ID: c.ID, Title: c.Get("title"), Kind: mode, Status: c.Get("status"), Similarity: c.Get("similarity_percent"), Warnings: []string{}}
		p.Iteration, _ = strconv.Atoi(c.Get("iteration_no"))
		side := func(field, stage string) *PairSide {
			id := c.Get(field)
			if id == "" {
				return nil
			}
			s := byID[id]
			if s == nil {
				return nil
			}
			ps := &PairSide{SetID: s.ID, Label: s.Label(), Version: s.Version, Latest: s.Newer == "", Present: s.Present, Total: s.Total, Pct: s.Pct}
			if s.Newer != "" {
				p.Warnings = append(p.Warnings, fmt.Sprintf("%s: сравнивается %s, есть более новая — %s", StageName(s.Stage), s.Version, s.Newer))
			}
			if s.Total > 0 && s.Present < s.Total {
				p.Warnings = append(p.Warnings, fmt.Sprintf("%s %s неполон: %d из %d томов", StageName(s.Stage), s.Version, s.Present, s.Total))
			}
			if s.Unparsed > 0 {
				p.Warnings = append(p.Warnings, fmt.Sprintf("%s %s: не прочитано файлов — %d", StageName(s.Stage), s.Version, s.Unparsed))
			}
			return ps
		}
		p.Ref, p.Target = side("set_pd", SideRef), side("set_rd", SideTarget)
		switch {
		case p.Ref == nil && p.Target == nil:
			p.Label = "комплекты не указаны"
			p.Warnings = append(p.Warnings, "у сравнения не указаны комплекты ПД и РД")
		default:
			p.Label = pairLabel(p.Ref, p.Target, mode)
			if p.Ref == nil {
				p.Warnings = append(p.Warnings, "не указан комплект «"+refName+"»")
			}
			if p.Target == nil {
				p.Warnings = append(p.Warnings, "не указан комплект «"+tgtName+"»")
			}
		}
		out.Pairs = append(out.Pairs, p)
	}
	sort.SliceStable(out.Pairs, func(i, j int) bool { return out.Pairs[i].ID < out.Pairs[j].ID })
	out.Structures = structuresOf(b, ns, objectID)
	out.Tree = BuildTree(out.Sets, out.Structures)
	out.Uploads = Uploads(b, ns, objectID)
	if out.Structures == nil {
		out.Structures = []Structure{}
	}
	if out.Tree == nil {
		out.Tree = []TreeStructure{}
	}
	if out.Uploads == nil {
		out.Uploads = []Upload{}
	}
	return out, nil
}

func pairLabel(ref, target *PairSide, mode string) string {
	rn, tn := SideNames(mode)
	l := func(s *PairSide, name string) string {
		if s == nil {
			return name + " —"
		}
		return s.Label
	}
	arrow := " ↔ "
	if mode == ModeVersions {
		arrow = " → "
	}
	return l(ref, rn) + arrow + l(target, tn)
}

func latestOf(sets []SetInfo, stage string) *SetInfo {
	for i := range sets {
		if sets[i].Stage == stage && sets[i].Latest {
			return &sets[i]
		}
	}
	return nil
}

// completeness fills the counters of a set against the statement of volumes of its stage.
func completeness(s *SetInfo, expected, units []expectedVolume) {
	var codes []string
	seen := map[string]string{} // normalized code → first file
	for _, d := range s.Docs {
		if !d.Parsed {
			s.Unparsed++
			continue
		}
		for _, c := range d.Codes {
			n := NormCode(c)
			if _, dup := seen[n]; dup {
				s.Duplicates = append(s.Duplicates, c)
				continue
			}
			seen[n] = d.ID
			codes = append(codes, c)
		}
	}
	var want []expectedVolume
	for _, e := range expected {
		if e.stage == "" || e.stage == s.Stage {
			want = append(want, e)
		}
	}
	s.Total = len(want)
	matched := map[string]bool{}
	for _, e := range want {
		hit := false
		for _, c := range codes {
			if volumesMatch(c, e.code) {
				hit = true
				matched[NormCode(c)] = true
			}
		}
		if hit {
			s.Present++
			if r := renamedAgainst(s, e); r != "" {
				s.Renamed = append(s.Renamed, r)
			}
		} else {
			s.Missing = append(s.Missing, e.code)
		}
	}
	if s.Total > 0 {
		for _, c := range codes {
			if !matched[NormCode(c)] {
				s.Extra = append(s.Extra, c)
			}
		}
	}
	checkUnits(s, units)
	s.Renamed = append(s.Renamed, namesInSet(s)...)
	if s.Total > 0 {
		s.Pct = s.Present * 100 / s.Total
	}
	s.Note = setNote(s)
}

// stems are the content words of a title cut to their stems (the endings differ: «свайного» / «свайное»).
func stems(title string) map[string]bool {
	out := map[string]bool{}
	for _, w := range contentWords(title) {
		if r := []rune(w); len(r) > 6 {
			w = string(r[:len(r)-2])
		}
		out[w] = true
	}
	return out
}

// renamedAgainst compares the title the statement of volumes gives a volume with the title in the
// file's own title block. Judged only when the file is that one volume, both titles have content
// words and not one of them is shared — a different wording alone is not a mismatch.
func renamedAgainst(s *SetInfo, e expectedVolume) string {
	if e.title == "" {
		return ""
	}
	want := stems(e.title)
	if len(want) == 0 {
		return ""
	}
	for _, d := range s.Docs {
		if !d.Parsed || len(d.Codes) != 1 || d.Title == "" || !volumesMatch(d.Codes[0], e.code) {
			continue
		}
		have := stems(d.Title)
		if len(have) == 0 {
			continue
		}
		for w := range have {
			if want[w] {
				return ""
			}
		}
		return fmt.Sprintf("%s: в ведомости «%s», в документе «%s»", e.code, trunc(e.title, 60), trunc(d.Title, 60))
	}
	return ""
}

func setNote(s *SetInfo) string {
	var parts []string
	if s.Total == 0 {
		vols := 0
		for _, d := range s.Docs {
			vols += len(d.Codes)
		}
		parts = append(parts, fmt.Sprintf("ведомость состава не задана; в комплекте файлов: %d, томов: %d", len(s.Docs), vols))
	} else {
		parts = append(parts, fmt.Sprintf("%d из %d %s", s.Present, s.Total, map[bool]string{true: "позиций ведомости", false: "томов"}[hasRequiredUnits(s)]))
		if len(s.Missing) > 0 {
			parts = append(parts, "нет: "+strings.Join(s.Missing, ", "))
		}
		if len(s.Extra) > 0 {
			parts = append(parts, "лишние: "+strings.Join(s.Extra, ", "))
		}
	}
	if len(s.Duplicates) > 0 {
		parts = append(parts, "дубли: "+strings.Join(s.Duplicates, ", "))
	}
	if len(s.Renamed) > 0 {
		parts = append(parts, "наименования не согласованы — "+strings.Join(s.Renamed, "; "))
	}
	if s.Unparsed > 0 {
		parts = append(parts, fmt.Sprintf("не прочитано файлов: %d", s.Unparsed))
	}
	if len(s.RefsMissing) > 0 {
		parts = append(parts, "к сведению — ссылки на тома, которых нет у объекта: "+strings.Join(s.RefsMissing, "; "))
	}
	return strings.Join(parts, "; ")
}

// currentSets are the sets of a stage that are not certainly outdated: the newest dated version
// and the sets without dates (those are packs of one stage, e.g. different volumes).
func currentSets(sets []SetInfo, stage string) []*SetInfo {
	var out []*SetInfo
	for i := range sets {
		if sets[i].Stage == stage && sets[i].Status != SetSuperseded && sets[i].Newer == "" {
			out = append(out, &sets[i])
		}
	}
	return out
}

func versionsOf(sets []*SetInfo) string {
	var v []string
	for _, s := range sets {
		v = append(v, s.Version)
	}
	return strings.Join(v, ", ")
}

// buildMatrix lines the volumes of the current ПД and РД sets up against the
// statement of volumes: one row per expected volume, then the volumes found
// only in documents.
func buildMatrix(sets []SetInfo, expected []expectedVolume) []MatrixRow {
	ref, tgt := currentSets(sets, SideRef), currentSets(sets, SideTarget)
	cellOf := func(cur []*SetInfo, code string) MatrixCell {
		if len(cur) == 0 {
			return MatrixCell{State: CellNoSet}
		}
		c := MatrixCell{State: CellMissing, SetID: cur[0].ID, Version: versionsOf(cur)}
		for _, s := range cur {
			for _, d := range s.Docs {
				for _, dc := range d.Codes {
					if volumesMatch(dc, code) {
						c.State, c.Code, c.SetID, c.Version = CellPresent, dc, s.ID, s.Version
						c.Docs = append(c.Docs, d.ID)
					}
				}
			}
		}
		return c
	}
	rows := []MatrixRow{}
	used := map[string]bool{} // "<stage>:<normalized doc code>"
	for _, e := range expected {
		r := MatrixRow{Code: e.code, Title: e.title, Expected: true, Ref: MatrixCell{State: CellNoSet}, Target: MatrixCell{State: CellNoSet}}
		if e.stage != SideTarget {
			r.Ref = cellOf(ref, e.code) // expected at this stage
		}
		if e.stage != SideRef {
			r.Target = cellOf(tgt, e.code)
		}
		if r.Ref.State == CellPresent {
			used[SideRef+":"+NormCode(r.Ref.Code)] = true
		}
		if r.Target.State == CellPresent {
			used[SideTarget+":"+NormCode(r.Target.Code)] = true
		}
		rows = append(rows, r)
	}
	// volumes that are in documents only
	extra := map[string]*MatrixRow{}
	var order []string
	add := func(cur []*SetInfo, stage string) {
		for _, s := range cur {
			for _, d := range s.Docs {
				for _, c := range d.Codes {
					if used[stage+":"+NormCode(c)] {
						continue
					}
					key := NormCode(c)
					row := extra[key]
					if row == nil {
						// the same volume in the other stage may be written differently ("ПЗ5.1" / "ПЗ5-1")
						for _, k := range order {
							if volumesMatch(extra[k].Code, c) {
								row = extra[k]
								break
							}
						}
					}
					if row == nil {
						row = &MatrixRow{Code: c, Title: d.Title, Ref: missingOrNone(ref), Target: missingOrNone(tgt)}
						extra[key] = row
						order = append(order, key)
					}
					mc := MatrixCell{State: CellPresent, SetID: s.ID, Version: s.Version, Code: c, Docs: []string{d.ID}}
					if len(expected) > 0 {
						mc.State = CellExtra
					}
					if stage == SideRef {
						row.Ref = mc
					} else {
						row.Target = mc
					}
				}
			}
		}
	}
	add(ref, SideRef)
	add(tgt, SideTarget)
	sort.Strings(order)
	for _, k := range order {
		rows = append(rows, *extra[k])
	}
	return rows
}

func missingOrNone(cur []*SetInfo) MatrixCell {
	if len(cur) == 0 {
		return MatrixCell{State: CellNoSet}
	}
	return MatrixCell{State: CellMissing, SetID: cur[0].ID, Version: versionsOf(cur)}
}

// LoadObjectDocs is the package function on the runner's backend.
func (r *Runner) LoadObjectDocs(ns, objectID string) (*ObjectDocs, error) {
	return LoadObjectDocs(r.Backend, ns, objectID)
}

// CheckSets reads the documents of the object's sets that have not been read yet
// (volume code, title, pages), writes the completeness back into the sets and the
// "ПД v2 ↔ РД v1" label into the comparisons, and returns the fresh picture.
func (r *Runner) CheckSets(ctx context.Context, ns, objectID string) (*ObjectDocs, error) {
	b := r.Backend
	if err := r.readSetDocs(ctx, ns, objectID); err != nil {
		return nil, err
	}
	od, err := LoadObjectDocs(b, ns, objectID)
	if err != nil {
		return nil, err
	}
	for _, s := range od.Sets {
		v := idcheck.V(
			"volumes_total", strconv.Itoa(s.Total),
			"volumes_present", strconv.Itoa(s.Present),
			"completeness_pct", strconv.Itoa(s.Pct),
			"completeness_note", s.Note,
		)
		if err := b.Update(ns, ModSets, s.ID, v); err != nil {
			return nil, err
		}
	}
	for _, p := range od.Pairs {
		label := p.Label
		if p.Iteration > 1 {
			label += fmt.Sprintf(" (итерация %d)", p.Iteration)
		}
		if err := b.Update(ns, ModComparisons, p.ID, idcheck.V("sets_label", label)); err != nil {
			return nil, err
		}
	}
	return od, nil
}

// readSetDocs reads the PDFs of the object's sets that carry no volume code yet.
func (r *Runner) readSetDocs(ctx context.Context, ns, objectID string) error {
	b := r.Backend
	sets, err := b.Search(ns, ModSets, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return err
	}
	structs := structuresOf(b, ns, objectID)
	bySet := map[string][]idcheck.Record{}
	var all []idcheck.Record // every document of the object: a new file is compared with them for the journal
	for _, s := range sets {
		docs, err := b.Search(ns, ModDocuments, fmt.Sprintf("set = '%s'", s.ID))
		if err != nil {
			return err
		}
		bySet[s.ID] = docs
		all = append(all, docs...)
	}
	for _, s := range sets {
		for _, d := range bySet[s.ID] {
			if d.Get("file") == "" {
				continue
			}
			if d.Get("volume_code") != "" || d.Get("pages") != "" {
				// read earlier: only what the registry adds later (a structure added to the object) and the
				// references to other volumes of a file read before they were looked for
				v := docRegistryValues(d, d.Get("volume_code"), d.Get("volume_title"), structs)
				if (d.Get("volume_refs_json") == "" || d.Get("contents_json") == "") && d.Get("volume_code") != "" {
					if data, _, err := b.Download(ns, d.Get("file")); err == nil && IsPDF(data) {
						if pages, _, _, err := ExtractPagesOCR(ctx, d.ID, data, r.CacheDir, r.OCR); err == nil {
							codes := splitCodes(d.Get("volume_code"))
							v["volume_refs_json"] = []string{VolRefsJSON(ExtractVolumeRefs(pages, codes))}
							v["contents_json"] = []string{contentsJSON(ParseContents(pages, firstOf(codes)))}
						}
					}
				}
				if len(v) > 0 {
					if err := b.Update(ns, ModDocuments, d.ID, v); err != nil {
						return err
					}
				}
				continue
			}
			data, name, err := b.Download(ns, d.Get("file"))
			if err != nil {
				return fmt.Errorf("загрузка файла %s: %w", d.ID, err)
			}
			if !IsPDF(data) {
				_ = b.Update(ns, ModDocuments, d.ID, idcheck.V("pages", "0"))
				continue
			}
			pages, _, _, err := ExtractPagesOCR(ctx, d.ID, data, r.CacheDir, r.OCR)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			doc := &Document{ID: d.ID, Name: name, Pages: pages}
			vs := DetectVolumes(doc)
			v := idcheck.V("pages", strconv.Itoa(len(pages)), "detected_by", "auto")
			v["sheet_map"] = []string{FormatSheetMap(SheetMap(pages))}
			v["headings_json"] = []string{HeadingsJSON(ExtractHeadings(doc))}
			var infos []VolumeInfo
			for _, x := range vs {
				infos = append(infos, volInfo(x))
			}
			for k, x := range EditionValues(&d, hashKey(data), infos) {
				v[k] = x
			}
			var codes, titles []string
			for _, x := range vs {
				if x.Code != "" {
					codes = append(codes, x.Code)
				}
				titles = append(titles, x.Title)
			}
			switch {
			case len(vs) == 1:
				v["volume_code"] = []string{vs[0].Code}
				v["volume_title"] = []string{vs[0].Title}
			case len(vs) > 1:
				v["volume_code"] = []string{strings.Join(codes, ", ")}
				v["volume_title"] = []string{fmt.Sprintf("%d томов: %s", len(vs), strings.Join(titles, "; "))}
			}
			v["volume_refs_json"] = []string{VolRefsJSON(ExtractVolumeRefs(pages, splitCodes(firstVal(v, "volume_code"))))}
			v["contents_json"] = []string{contentsJSON(ParseContents(pages, firstOf(splitCodes(firstVal(v, "volume_code")))))}
			for k, x := range docRegistryValues(d, firstVal(v, "volume_code"), firstVal(v, "volume_title"), structs) {
				v[k] = x
			}
			if err := b.Update(ns, ModDocuments, d.ID, v); err != nil {
				return err
			}
			kind, note := uploadKind(hashKey(data), firstVal(v, "volume_code"), d.ID, all)
			if err := journalUpload(b, ns, objectID, s.ID, d, name, hashKey(data), len(pages), firstVal(v, "volume_code"), kind, note); err != nil {
				log.Printf("cmp: upload journal %s: %v", d.ID, err) // an older space has no journal: the reading must not fail
			}
			for i := range all { // the next files of this pass see this one as read
				if all[i].ID == d.ID {
					for k, x := range v {
						all[i].Values[k] = x
					}
				}
			}
		}
	}
	return nil
}

func firstVal(v idcheck.Values, k string) string {
	if x := v[k]; len(x) > 0 {
		return x[0]
	}
	return ""
}

func hasRequiredUnits(s *SetInfo) bool {
	for _, u := range s.Units {
		if u.Required {
			return true
		}
	}
	return false
}

func splitCodes(s string) []string {
	var out []string
	for _, c := range strings.Split(s, ",") {
		if c = strings.TrimSpace(c); c != "" {
			out = append(out, c)
		}
	}
	return out
}

func firstOf(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}

func contentsJSON(c []ContentsEntry) string {
	if len(c) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(c)
	return string(b)
}

func parseContentsJSON(s string) []ContentsEntry {
	var c []ContentsEntry
	_ = json.Unmarshal([]byte(s), &c)
	return c
}

// namesInSet checks the names of the volumes inside a set (ТЗ 3.2.5): the title a contents table of the
// set gives a volume against the title of the volume's own file, and two files of one volume with
// different titles. As with the statement, only titles without one shared word are a mismatch.
func namesInSet(s *SetInfo) []string {
	var out []string
	differ := func(a, b string) bool {
		x, y := stems(a), stems(b)
		if len(x) == 0 || len(y) == 0 {
			return false
		}
		for w := range x {
			if y[w] {
				return false
			}
		}
		return true
	}
	seen := map[string]bool{}
	for _, src := range s.Docs {
		for _, e := range src.Contents {
			for _, d := range s.Docs {
				if d.ID == src.ID || len(d.Codes) != 1 || d.Title == "" || !volumesMatch(d.Codes[0], e.Code) || !differ(e.Title, d.Title) {
					continue
				}
				k := "c|" + NormCode(e.Code)
				if !seen[k] {
					seen[k] = true
					out = append(out, fmt.Sprintf("%s: в составе проекта (%s, стр. %d) «%s», в документе «%s»", e.Code, firstNonEmpty(strings.Join(src.Codes, ", "), src.ID), e.Page, trunc(e.Title, 60), trunc(d.Title, 60)))
				}
			}
		}
	}
	for i, a := range s.Docs {
		for _, b := range s.Docs[i+1:] {
			if len(a.Codes) != 1 || len(b.Codes) != 1 || !volumesMatch(a.Codes[0], b.Codes[0]) || !differ(a.Title, b.Title) {
				continue
			}
			k := "d|" + NormCode(a.Codes[0])
			if !seen[k] {
				seen[k] = true
				out = append(out, fmt.Sprintf("%s: два файла тома с разными наименованиями — «%s» и «%s»", a.Codes[0], trunc(a.Title, 60), trunc(b.Title, 60)))
			}
		}
	}
	return out
}
