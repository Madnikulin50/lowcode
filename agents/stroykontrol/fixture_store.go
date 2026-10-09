package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FixtureStore serves the comparison viewer's data from a local directory
// instead of a live Compose instance — no Compose server, network, or API
// token needed at all. Meant for UI development and demos: point --fixtures
// at a directory laid out like
//
//	<dir>/<caseID>/meta.json           {"title","status","similarityPercent","comment"}
//	<dir>/<caseID>/discrepancies.json  [{"type","severity","description","pageNumber"}, ...]
//	<dir>/<caseID>/pd.<ext>            (optional; any extension, content is sniffed)
//	<dir>/<caseID>/rd.<ext>            (optional)
//
// The recordID a request carries is just the caseID (namespaceID is
// accepted, since the viewer always sends one, but ignored). See
// fixtures/README.md; fixtures/demo-1 ships a ready-to-run example.
type FixtureStore struct {
	dir string
}

func NewFixtureStore(dir string) (*FixtureStore, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("fixtures dir: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("fixtures dir %q is not a directory", dir)
	}
	return &FixtureStore{dir: dir}, nil
}

// caseDir resolves a caseID to its directory, filepath.Base'd so a
// recordID can never escape the fixtures dir via "..".
func (f *FixtureStore) caseDir(caseID string) string {
	return filepath.Join(f.dir, filepath.Base(caseID))
}

// findSideFile locates <dir>/<side>.* (side is "pd" or "rd"); the extension
// is whatever the fixture author used, since IsPDF/IsDOCX (rasterize.go)
// sniff actual content rather than trust it.
func (f *FixtureStore) findSideFile(dir, side string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, side+".*"))
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func mimetypeFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

type fixtureMeta struct {
	Title             string `json:"title"`
	Status            string `json:"status"`
	SimilarityPercent string `json:"similarityPercent"`
	Comment           string `json:"comment"`
}

func (f *FixtureStore) Comparison(_, recordID string) (*comparisonRecord, error) {
	dir := f.caseDir(recordID)
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("no such fixture case %q under %s", recordID, f.dir)
	}

	var meta fixtureMeta
	b, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	switch {
	case err == nil:
		if err := json.Unmarshal(b, &meta); err != nil {
			return nil, fmt.Errorf("fixture %s: meta.json: %w", recordID, err)
		}
	case os.IsNotExist(err):
		meta.Title = recordID // meta.json is optional; still usable with just files
	default:
		return nil, err
	}

	rec := &comparisonRecord{
		Title:             meta.Title,
		Status:            meta.Status,
		SimilarityPercent: meta.SimilarityPercent,
		Comment:           meta.Comment,
	}
	// PdFileID/RdFileID are opaque to callers — encode caseID+side so
	// AttachmentName (which only receives this string, not recordID) can
	// find the file back.
	if f.findSideFile(dir, "pd") != "" {
		rec.PdFileID = recordID + ":pd"
	}
	if f.findSideFile(dir, "rd") != "" {
		rec.RdFileID = recordID + ":rd"
	}
	return rec, nil
}

func (f *FixtureStore) Discrepancies(_, recordID string) ([]discrepancyRecord, error) {
	b, err := os.ReadFile(filepath.Join(f.caseDir(recordID), "discrepancies.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []discrepancyRecord
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("fixture %s: discrepancies.json: %w", recordID, err)
	}
	return out, nil
}

// AttachmentName returns the fixture file's own basename as its "display
// name" — there's no separate attachment metadata in this backend, the
// file on disk is the attachment.
func (f *FixtureStore) AttachmentName(_, attachmentID string) string {
	recordID, side, ok := strings.Cut(attachmentID, ":")
	if !ok {
		return ""
	}
	path := f.findSideFile(f.caseDir(recordID), side)
	if path == "" {
		return ""
	}
	return filepath.Base(path)
}

func (f *FixtureStore) FetchFile(_, recordID, side string) ([]byte, string, error) {
	if side != "pd" && side != "rd" {
		return nil, "", fmt.Errorf("side must be 'pd' or 'rd'")
	}
	path := f.findSideFile(f.caseDir(recordID), side)
	if path == "" {
		return nil, "", fmt.Errorf("fixture case %q has no %s file", recordID, side)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return data, mimetypeFor(path), nil
}
