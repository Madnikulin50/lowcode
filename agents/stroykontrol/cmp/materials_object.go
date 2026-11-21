package cmp

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Обосновывающие материалы замечания (ТЗ 3.4.9, 3.4.11): файлы, приложенные к замечанию (поле
// materials_files) и к пакетам исправления объекта, вместе с загруженными томами образуют то, среди
// чего ищутся материалы, заявленные в ответе проектировщика.

// supportingFile is a supporting file as the check needs it; read once per attachment.
type supportingFile struct {
	name     string
	pages    int
	headings []DocHeading
}

var supportingCache sync.Map // attachment ID → supportingFile

// matDocsFrom reads the loaded documents of the comparisons of an object.
func matDocsFrom(recs []idcheck.Record) []MatDoc {
	var out []MatDoc
	for _, d := range recs {
		pages, _ := strconv.Atoi(d.Get("pages"))
		out = append(out, MatDoc{
			ID: d.ID, Name: firstNonEmpty(strings.TrimSpace(d.Get("volume_code")+" "+d.Get("volume_title")), d.ID), Codes: d.Get("volume_code"), Title: d.Get("volume_title"),
			Pages: pages, Sheets: ParseSheetMap(d.Get("sheet_map")), Headings: ParseHeadings(d.Get("headings_json")),
		})
	}
	return out
}

// supportingDocs reads the supporting files (attachment IDs): their names and, for a PDF, its headings.
func (r *Runner) supportingDocs(ns string, atts []string) []MatDoc {
	var out []MatDoc
	for _, id := range atts {
		if id == "" {
			continue
		}
		var f supportingFile
		if v, ok := supportingCache.Load(id); ok {
			f = v.(supportingFile)
		} else {
			data, name, err := r.Backend.Download(ns, id)
			if err != nil {
				log.Printf("cmp: supporting file %s: %v", id, err)
				continue
			}
			f.name = name
			if IsPDF(data) {
				if pages, err := ExtractPages("att:"+id, data, r.CacheDir); err == nil {
					f.pages = len(pages)
					f.headings = ExtractHeadings(&Document{ID: "att:" + id, Name: name, Pages: pages})
				}
			}
			supportingCache.Store(id, f)
		}
		out = append(out, MatDoc{ID: "att:" + id, Name: f.name, Title: f.name, Pages: f.pages, Headings: f.headings, Supporting: true})
	}
	return out
}

// packageMaterials lists the supporting files attached to the correction packages of an object.
func (r *Runner) packageMaterials(ns, objectID string) []string {
	pkgs, err := r.Backend.Search(ns, ModCorrection, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		log.Printf("cmp: correction packages of object %s: %v", objectID, err)
		return nil
	}
	var out []string
	for _, p := range pkgs {
		out = append(out, nonBlank(p.Values["materials_files"])...)
	}
	return out
}

// MaterialValues are the fields the check of the materials writes on a remark.
func MaterialValues(ms []Material) idcheck.Values {
	// set explicitly: an empty value clears what an earlier answer declared
	return idcheck.Values{"materials_json": {MaterialsJSON(ms)}, "materials_state": {MaterialsSummary(ms)}}
}
