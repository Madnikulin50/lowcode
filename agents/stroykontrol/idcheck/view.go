package idcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ViewFinding is a finding as the viewer shows it.
type ViewFinding struct {
	ID          string    `json:"id"`
	ActID       string    `json:"actID,omitempty"`
	Check       int       `json:"check"`
	CheckName   string    `json:"checkName"`
	Severity    string    `json:"severity"`
	Field       string    `json:"field,omitempty"`
	Expected    string    `json:"expected,omitempty"`
	Actual      string    `json:"actual,omitempty"`
	Description string    `json:"description"`
	Page        int       `json:"page,omitempty"`
	Box         []float64 `json:"box,omitempty"`
	RegistryRow int       `json:"registryRow,omitempty"`
	Position    string    `json:"position,omitempty"`
	Source      string    `json:"source,omitempty"`
}

// ViewAct is an act as the viewer shows it.
type ViewAct struct {
	ID          string `json:"id"`
	FileName    string `json:"fileName"`
	Number      string `json:"number"`
	Date        string `json:"date"`
	Positions   string `json:"positions"`
	CheckStatus string `json:"checkStatus"`
	OCRStatus   string `json:"ocrStatus"`
	Pages       []Page `json:"pages"`
	High        int    `json:"high"`
	Medium      int    `json:"medium"`
	Low         int    `json:"low"`
}

// ViewRow is a registry row for the in-place registry view.
type ViewRow struct {
	Row        int    `json:"row"`
	Kind       string `json:"kind"`
	Position   string `json:"position"`
	Name       string `json:"name"`
	DocNumber  string `json:"docNumber"`
	DocDate    string `json:"docDate"`
	Unit       string `json:"unit"`
	QtyKS2     string `json:"qtyKS2"`
	QtyProject string `json:"qtyProject"`
	QtyFact    string `json:"qtyFact"`
}

// ViewPosition is a check-8 summary row.
type ViewPosition struct {
	Position    string `json:"position"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	RegistrySum string `json:"registrySum"`
	RegistryKS2 string `json:"registryKS2"`
	KS2         string `json:"ks2"`
	Delta       string `json:"delta"`
	Status      string `json:"status"`
}

// View is everything the in-place viewer needs for one package.
type View struct {
	PackageID  string         `json:"packageID"`
	Title      string         `json:"title"`
	Status     string         `json:"status"`
	Progress   string         `json:"progress"`
	FocusActID string         `json:"focusActID,omitempty"`
	CheckNames map[int]string `json:"checkNames"`
	Acts       []ViewAct      `json:"acts"`
	Findings   []ViewFinding  `json:"findings"`
	Registry   []ViewRow      `json:"registry"`
	Positions  []ViewPosition `json:"positions"`
}

func atoi(s string) int {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return int(f)
}

func parseBox(s string) []float64 {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil
	}
	out := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return nil
		}
		out[i] = v
	}
	return out
}

func short(s string) string {
	if i := strings.Index(s, "T"); i > 0 {
		return s[:i]
	}
	return s
}

// BuildView assembles the viewer payload. Pass packageID, or actID alone
// (the act card), or findingID alone (the finding card) — the package is
// then taken from that record.
func BuildView(b Backend, ns, packageID, actID, findingID string) (*View, error) {
	if packageID == "" && actID == "" && findingID != "" {
		f, err := b.GetRecord(ns, ModFindings, findingID)
		if err != nil {
			return nil, fmt.Errorf("finding: %w", err)
		}
		packageID = f.Get("package")
	}
	v := &View{CheckNames: CheckNames, FocusActID: actID}
	if packageID == "" && actID != "" {
		act, err := b.GetRecord(ns, ModActs, actID)
		if err != nil {
			return nil, fmt.Errorf("act: %w", err)
		}
		packageID = act.Get("package")
	}
	if packageID == "" {
		return nil, fmt.Errorf("packageID or actID is required")
	}
	pkg, err := b.GetRecord(ns, ModPackages, packageID)
	if err != nil {
		return nil, fmt.Errorf("package: %w", err)
	}
	v.PackageID, v.Title, v.Status, v.Progress = packageID, pkg.Get("title"), pkg.Get("status"), pkg.Get("progress")
	filter := fmt.Sprintf("package = '%s'", packageID)

	acts, err := b.Search(ns, ModActs, filter)
	if err != nil {
		return nil, err
	}
	for _, a := range acts {
		va := ViewAct{
			ID: a.ID, FileName: a.Get("file_name"), Number: firstNonEmptyStr(a.Get("doc_number"), a.Get("number")),
			Date: short(firstNonEmptyStr(a.Get("doc_date"), a.Get("date"))), Positions: a.Get("positions"),
			CheckStatus: a.Get("check_status"), OCRStatus: a.Get("ocr_status"),
			High: atoi(a.Get("issues_high")), Medium: atoi(a.Get("issues_medium")), Low: atoi(a.Get("issues_low")),
		}
		_ = json.Unmarshal([]byte(a.Get("pages_json")), &va.Pages)
		v.Acts = append(v.Acts, va)
	}
	sort.Slice(v.Acts, func(i, j int) bool { return v.Acts[i].FileName < v.Acts[j].FileName })

	fs, err := b.Search(ns, ModFindings, filter)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		v.Findings = append(v.Findings, ViewFinding{
			ID: f.ID, ActID: f.Get("act"), Check: atoi(f.Get("check_no")), CheckName: f.Get("check_name"),
			Severity: f.Get("severity"), Field: f.Get("field"), Expected: f.Get("expected"), Actual: f.Get("actual"),
			Description: f.Get("description"), Page: atoi(f.Get("page")), Box: parseBox(f.Get("bbox")),
			RegistryRow: atoi(f.Get("registry_row_no")), Position: f.Get("position"), Source: f.Get("source"),
		})
	}
	sort.SliceStable(v.Findings, func(i, j int) bool {
		a, c := v.Findings[i], v.Findings[j]
		if sevRank(Severity(a.Severity)) != sevRank(Severity(c.Severity)) {
			return sevRank(Severity(a.Severity)) < sevRank(Severity(c.Severity))
		}
		if a.Check != c.Check {
			return a.Check < c.Check
		}
		return a.Page < c.Page
	})

	rows, err := b.Search(ns, ModRegistryRows, filter)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		v.Registry = append(v.Registry, ViewRow{
			Row: atoi(r.Get("row_no")), Kind: r.Get("kind"), Position: r.Get("position"),
			Name: firstNonEmptyStr(r.Get("doc_name"), r.Get("work_name")), DocNumber: r.Get("doc_number"),
			DocDate: firstNonEmptyStr(r.Get("doc_date_raw"), short(r.Get("doc_date"))), Unit: r.Get("unit"),
			QtyKS2: r.Get("qty_ks2"), QtyProject: r.Get("qty_project"), QtyFact: r.Get("qty_fact"),
		})
	}
	sort.Slice(v.Registry, func(i, j int) bool { return v.Registry[i].Row < v.Registry[j].Row })

	pos, err := b.Search(ns, ModPositions, filter)
	if err != nil {
		return nil, err
	}
	for _, p := range pos {
		v.Positions = append(v.Positions, ViewPosition{
			Position: p.Get("position"), Name: p.Get("name"), Unit: p.Get("unit"), RegistrySum: p.Get("registry_sum"),
			RegistryKS2: p.Get("registry_ks2"), KS2: p.Get("ks2_qty"), Delta: p.Get("delta"), Status: p.Get("status"),
		})
	}
	sort.Slice(v.Positions, func(i, j int) bool { return positionLess(v.Positions[i].Position, v.Positions[j].Position) })
	return v, nil
}

func firstNonEmptyStr(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

// ActPageImage returns the path of a rendered page of an act's PDF,
// downloading the attachment once into the cache.
func ActPageImage(ctx context.Context, b Backend, cacheDir, ns, actID string, page, dpi int) (string, int, error) {
	act, err := b.GetRecord(ns, ModActs, actID)
	if err != nil {
		return "", 0, err
	}
	attID := act.Get("attachment_id")
	if attID == "" {
		return "", 0, fmt.Errorf("act has no attachment")
	}
	local := filepath.Join(cacheDir, "att", attID+".pdf")
	data, err := os.ReadFile(local)
	if err != nil {
		data, _, err = b.Download(ns, attID)
		if err != nil {
			return "", 0, err
		}
		_ = os.MkdirAll(filepath.Dir(local), 0o755)
		_ = os.WriteFile(local, data, 0o644)
	}
	imgs, err := PageImages(ctx, cacheDir, data, dpi)
	if err != nil {
		return "", 0, err
	}
	if page < 1 || page > len(imgs) {
		return "", len(imgs), fmt.Errorf("page %d out of range 1..%d", page, len(imgs))
	}
	return imgs[page-1], len(imgs), nil
}
