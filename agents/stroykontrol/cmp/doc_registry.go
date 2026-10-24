package cmp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Учёт документов объекта (ТЗ 2.2.4, 3.1.2, 3.1.4, 3.1.7): раздел документа, сооружения и участки
// объекта как отдельный справочник, журнал загрузок и дерево «сооружение → раздел → том → редакции».
// Раздел берётся из шифра тома (буквенная марка: «2024-15-КЖ2» → «КЖ»), если его не задали вручную.
// Сооружение связывается с документом по полю «Сооружение / участок» или по наименованию тома, когда
// в них есть наименование или шифр сооружения из справочника; ручной выбор не перезаписывается.
// Журнал загрузок пишет агент, когда впервые читает файл: кто и когда приложил его (системные поля
// записи платформы), хэш, число страниц и что это за файл для объекта — новый документ, новая редакция
// тома или повтор уже загруженного файла.

// Modules.
const (
	ModStructures = "structures"
	ModUploads    = "upload_log"
)

// System values of a record (filled by the backend from the record's metadata).
const (
	SysCreatedAt = "_created_at"
	SysCreatedBy = "_created_by"
)

// Kinds of an upload.
const (
	UploadNew     = "new"      // a volume the object did not have
	UploadEdition = "edition"  // another file of a volume the object has: a new edition
	UploadRepeat  = "repeat"   // the same file (by hash) is already loaded
	UploadNoCode  = "unparsed" // the volume code was not read
)

// UploadLabels name the kinds of an upload.
var UploadLabels = map[string]string{
	UploadNew: "новый том", UploadEdition: "новая редакция тома", UploadRepeat: "повтор уже загруженного файла", UploadNoCode: "шифр тома не прочитан",
}

// UserNamer is a backend that can name a user of the platform (the author of an upload).
type UserNamer interface {
	UserName(id string) string
}

// SectionOf is the section (the letter mark) of the volume codes: "2024-15-КЖ2, 2024-15-КЖ3" → "КЖ".
func SectionOf(codes string) string {
	var out []string
	seen := map[string]bool{}
	for _, c := range strings.Split(codes, ",") {
		if s := Series(strings.TrimSpace(c)); s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return strings.Join(out, ", ")
}

// Structure is a structure or a section of the route of an object (сооружение / участок).
type Structure struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Kind string `json:"kind"`
}

func structuresOf(b Backend, ns, objectID string) []Structure {
	recs, err := b.Search(ns, ModStructures, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil // the module is optional: an older space has none
	}
	var out []Structure
	for _, r := range recs {
		out = append(out, Structure{ID: r.ID, Name: r.Get("name"), Code: r.Get("code"), Kind: r.Get("kind")})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// MatchStructure finds the structure a document belongs to by its texts (the field «Сооружение /
// участок», the volume title): the code of a structure or its whole name. Only one match counts.
func MatchStructure(texts []string, structs []Structure) string {
	hay := " " + NormKey(strings.Join(texts, " ")) + " "
	found := ""
	for _, s := range structs {
		hit := false
		if c := NormKey(s.Code); len([]rune(c)) >= 2 && strings.Contains(hay, " "+c+" ") {
			hit = true
		}
		if n := NormKey(s.Name); len([]rune(n)) >= 4 && strings.Contains(hay, n) {
			hit = true
		}
		if hit {
			if found != "" && found != s.ID {
				return "" // two structures named: the person decides
			}
			found = s.ID
		}
	}
	return found
}

// docRegistryValues are the section and the structure of a document read now (manual values stay).
func docRegistryValues(d idcheck.Record, codes, title string, structs []Structure) idcheck.Values {
	v := idcheck.Values{}
	if d.Get("section") == "" {
		if s := SectionOf(codes); s != "" {
			v["section"] = []string{s}
		}
	}
	if d.Get("structure") == "" && len(structs) > 0 {
		if id := MatchStructure([]string{d.Get("structure_object"), title, d.Get("object_name")}, structs); id != "" {
			v["structure"] = []string{id}
		}
	}
	return v
}

// uploadKind says what a newly read file is for the object.
func uploadKind(hash, codes string, docID string, others []idcheck.Record) (kind, note string) {
	if codes == "" {
		return UploadNoCode, ""
	}
	for _, o := range others {
		if o.ID == docID {
			continue
		}
		if hash != "" && o.Get("file_hash") == hash {
			return UploadRepeat, "тот же файл: " + firstNonEmpty(o.Get("volume_code"), o.ID)
		}
	}
	for _, o := range others {
		if o.ID == docID {
			continue
		}
		for _, c := range strings.Split(codes, ",") {
			for _, oc := range strings.Split(o.Get("volume_code"), ",") {
				if c, oc = strings.TrimSpace(c), strings.TrimSpace(oc); c != "" && oc != "" && volumesMatch(c, oc) {
					n := "уже есть " + oc
					if ch := o.Get("change_no"); ch != "" && ch != "0" {
						n += " изм. " + ch
					}
					return UploadEdition, n
				}
			}
		}
	}
	return UploadNew, ""
}

// journalUpload writes the record of the upload journal for a document read for the first time.
func journalUpload(b Backend, ns, objectID, setID string, d idcheck.Record, name, hash string, pages int, codes, kind, note string) error {
	by := d.Get(SysCreatedBy)
	if un, ok := b.(UserNamer); ok && by != "" {
		if n := un.UserName(by); n != "" {
			by = n
		}
	}
	at := d.Get(SysCreatedAt)
	if at == "" {
		at = time.Now().UTC().Format(time.RFC3339)
	}
	v := idcheck.V("object", objectID, "set", setID, "document", d.ID, "file_name", name, "file_hash", hash, "volume_code", codes,
		"kind", kind, "note", note, "uploaded_at", at, "uploaded_by", by, "read_at", time.Now().UTC().Format(time.RFC3339))
	v["pages"] = []string{strconv.Itoa(pages)}
	_, err := b.Create(ns, ModUploads, v)
	return err
}

// Upload is a row of the journal.
type Upload struct {
	At    string `json:"at"`
	By    string `json:"by"`
	File  string `json:"file"`
	Codes string `json:"codes"`
	Kind  string `json:"kind"`
	Note  string `json:"note"`
	Set   string `json:"set"`
	Pages int    `json:"pages"`
}

// Uploads reads the journal of an object, newest first.
func Uploads(b Backend, ns, objectID string) []Upload {
	recs, err := b.Search(ns, ModUploads, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil
	}
	var out []Upload
	for _, r := range recs {
		p, _ := strconv.Atoi(r.Get("pages"))
		out = append(out, Upload{At: r.Get("uploaded_at"), By: r.Get("uploaded_by"), File: r.Get("file_name"), Codes: r.Get("volume_code"),
			Kind: r.Get("kind"), Note: r.Get("note"), Set: r.Get("set"), Pages: p})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At > out[j].At })
	return out
}

// The tree of the documents of an object (ТЗ 3.1.2): structure → section → volume → editions.

// TreeEdition is one file of a volume in one set.
type TreeEdition struct {
	DocID    string `json:"docID"`
	Set      string `json:"set"` // "ПД v2"
	Stage    string `json:"stage"`
	ChangeNo string `json:"changeNo,omitempty"`
	Latest   bool   `json:"latest"` // in the latest set of the stage
	Pages    int    `json:"pages"`
}

// TreeVolume is one volume with its editions.
type TreeVolume struct {
	Code     string        `json:"code"`
	Title    string        `json:"title"`
	Editions []TreeEdition `json:"editions"`
}

// TreeSection is a section of a structure.
type TreeSection struct {
	Name    string       `json:"name"`
	Volumes []TreeVolume `json:"volumes"`
}

// TreeStructure is a structure of the object ("" — the documents of no structure).
type TreeStructure struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Sections []TreeSection `json:"sections"`
}

// BuildTree groups the documents of the sets.
func BuildTree(sets []SetInfo, structs []Structure) []TreeStructure {
	name := map[string]string{}
	for _, s := range structs {
		name[s.ID] = strings.TrimSpace(s.Name + map[bool]string{true: " (" + s.Code + ")", false: ""}[s.Code != ""])
	}
	type key struct{ st, sec, vol string }
	vols := map[key]*TreeVolume{}
	var order []key
	for _, s := range sets {
		for _, d := range s.Docs {
			codes := d.Codes
			if len(codes) == 0 {
				codes = []string{""}
			}
			for _, c := range codes {
				sec := firstNonEmpty(d.Section, SectionOf(c), "раздел не определён")
				k := key{d.Structure, sec, c}
				v := vols[k]
				if v == nil {
					v = &TreeVolume{Code: firstNonEmpty(c, "шифр не прочитан"), Title: d.Title}
					vols[k] = v
					order = append(order, k)
				}
				if v.Title == "" {
					v.Title = d.Title
				}
				v.Editions = append(v.Editions, TreeEdition{DocID: d.ID, Set: s.Label(), Stage: s.Stage, ChangeNo: d.ChangeNo, Latest: s.Latest, Pages: d.Pages})
			}
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := order[i], order[j]
		if a.st != b.st {
			if a.st == "" || b.st == "" {
				return b.st == "" // the documents of no structure last
			}
			return name[a.st] < name[b.st]
		}
		if a.sec != b.sec {
			return a.sec < b.sec
		}
		return a.vol < b.vol
	})
	var out []TreeStructure
	for _, k := range order {
		if len(out) == 0 || out[len(out)-1].ID != k.st {
			out = append(out, TreeStructure{ID: k.st, Name: firstNonEmpty(name[k.st], map[bool]string{true: "сооружение не указано", false: k.st}[k.st == ""])})
		}
		st := &out[len(out)-1]
		if len(st.Sections) == 0 || st.Sections[len(st.Sections)-1].Name != k.sec {
			st.Sections = append(st.Sections, TreeSection{Name: k.sec})
		}
		sec := &st.Sections[len(st.Sections)-1]
		sec.Volumes = append(sec.Volumes, *vols[k])
	}
	return out
}
