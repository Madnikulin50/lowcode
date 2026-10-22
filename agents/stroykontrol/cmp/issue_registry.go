package cmp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// OriginComparison marks a remark made from a ПД/РД comparison (as opposed to one
// that came in an expertise export).
const OriginComparison = "comparison"

const issuePrefix = "ПТО-"

// IssueRegistered says what putting a list into the registry did.
type IssueRegistered struct {
	Created  int      `json:"created"`
	Updated  int      `json:"updated"`
	Closed   int      `json:"closed"`   // the findings were fixed in a later iteration
	Reopened int      `json:"reopened"` // closed before, the difference is back
	Numbers  []string `json:"numbers"`  // numbers of the new remarks
}

// numberParts reads "ПТО-3.12" → (3, 12).
func numberParts(n string) (batch, idx int, ok bool) {
	if !strings.HasPrefix(n, issuePrefix) {
		return 0, 0, false
	}
	a, b, found := strings.Cut(strings.TrimPrefix(n, issuePrefix), ".")
	if !found {
		return 0, 0, false
	}
	x, e1 := strconv.Atoi(a)
	y, e2 := strconv.Atoi(b)
	return x, y, e1 == nil && e2 == nil
}

func issuedValues(objectID string, in IssueInput, is Issued) idcheck.Values {
	v := idcheck.V(
		"object", objectID, "remark_key", is.GroupKey, "group_key", is.GroupKey,
		"text", is.Text, "refs_text", is.Refs, "section", is.Section, "basis", is.Basis,
		"criticality", is.Criticality, "origin", OriginComparison, "source_comparison", in.ComparisonID,
		"source_file", "сравнение: "+in.ComparisonTitle, "verdict", VerdictStillDiffers,
		"match_note", fmt.Sprintf("Сформировано по сравнению «%s»: подтверждённых находок — %d (%s).", in.ComparisonTitle, is.Count, is.Pages),
	)
	v["matched_discrepancies"] = is.FindingIDs
	v["linked_documents"] = nonBlank(is.DocIDs)
	// the places of the findings: the viewer opens the sheets and marks the regions (ТЗ 3.5.3)
	byID := map[string]FindingIn{}
	for _, f := range in.Findings {
		byID[f.ID] = f
	}
	var at []FoundAt
	for _, id := range is.FindingIDs {
		if f, ok := byID[id]; ok {
			at = append(at, FoundAt{ID: f.ID, Comparison: in.ComparisonID, Text: cleanDescription(f.Description),
				RefDoc: f.DocRef, TargetDoc: f.DocTarget, RefPage: atoi(f.PageRef), TargetPage: atoi(f.PageTarget), RefBox: f.BoxRef, TargetBox: f.BoxTarget})
		}
	}
	v["links_json"] = []string{LocationsJSON(LocateRemark(nil, nil, at))}
	return v
}

func nonBlank(in []string) []string {
	var out []string
	for _, s := range in {
		if s != "" && !strings.HasPrefix(s, "legacy:") { // legacy documents are not records
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return []string{""}
	}
	return out
}

func issueSnapshot(remarkID, status, note string) idcheck.Values {
	v := idcheck.V("remark", remarkID, "status", status, "source", note, "author", "агент", "kind", HistComparison)
	v["iteration"] = []string{"1"}
	v["at"] = []string{dateValue(time.Now().UTC())}
	return v
}

// IssueToRegistry puts a list into the registry of the object's remarks. A
// sheet pair that is already there (same group key) is updated and keeps the
// specialist's status, assignee and notes; a new one gets the next number of
// the comparison's batch ("ПТО-1.1"…). Remarks whose confirmed findings were
// all fixed in a later iteration of the comparison are closed.
func IssueToRegistry(b Backend, ns, objectID string, in IssueInput, res IssueResult) (*IssueRegistered, error) {
	if objectID == "" {
		return nil, fmt.Errorf("у сравнения не указан объект: замечания некуда вносить")
	}
	recs, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	byGroup := map[string]idcheck.Record{}
	maxBatch, myBatch, maxIdx := 0, 0, 0
	for _, r := range recs {
		if r.Get("origin") != OriginComparison {
			continue
		}
		bt, idx, ok := numberParts(r.Get("number"))
		if !ok {
			continue
		}
		if bt > maxBatch {
			maxBatch = bt
		}
		if r.Get("source_comparison") == in.ComparisonID {
			byGroup[r.Get("group_key")] = r
			myBatch = bt
			if idx > maxIdx {
				maxIdx = idx
			}
		}
	}
	if myBatch == 0 {
		myBatch = maxBatch + 1
	}

	out := &IssueRegistered{}
	kind := ""
	if obj, err := b.GetRecord(ns, ModObjects, objectID); err == nil {
		kind = obj.Get("data_kind")
	}
	for _, is := range res.Active {
		v := issuedValues(objectID, in, is)
		if kind != "" {
			v["data_kind"] = []string{kind}
		}
		if ex, ok := byGroup[is.GroupKey]; ok {
			if ex.Get("status") == RemarkResolved {
				v["status"] = []string{RemarkAwaiting}
				out.Reopened++
				if _, err := b.Create(ns, ModRemarkHistory, issueSnapshot(ex.ID, RemarkAwaiting, "возвращено: расхождение снова есть в сравнении «"+in.ComparisonTitle+"»")); err != nil {
					return nil, err
				}
			}
			if err := b.Update(ns, ModExpertise, ex.ID, v); err != nil {
				return nil, err
			}
			out.Updated++
			continue
		}
		maxIdx++
		number := fmt.Sprintf("%s%d.%d", issuePrefix, myBatch, maxIdx)
		v["number"] = []string{number}
		v["batch"] = []string{fmt.Sprintf("%s%d", issuePrefix, myBatch)}
		v["status"] = []string{RemarkAwaiting}
		v["iterations"] = []string{"1"}
		v["source_changed_at"] = []string{dateValue(time.Now().UTC())}
		v["source_status"] = []string{RemarkStatusLabels[RemarkAwaiting]}
		id, err := b.Create(ns, ModExpertise, v)
		if err != nil {
			return nil, err
		}
		if _, err := b.Create(ns, ModRemarkHistory, issueSnapshot(id, RemarkAwaiting, "сформировано по сравнению «"+in.ComparisonTitle+"»")); err != nil {
			return nil, err
		}
		out.Created++
		out.Numbers = append(out.Numbers, number)
	}
	for _, is := range res.Fixed {
		ex, ok := byGroup[is.GroupKey]
		if !ok || ex.Get("status") == RemarkResolved {
			continue
		}
		if err := b.Update(ns, ModExpertise, ex.ID, idcheck.V("status", RemarkResolved, "match_note", "Расхождения устранены в сравнении «"+in.ComparisonTitle+"» (повторная проверка).")); err != nil {
			return nil, err
		}
		if _, err := b.Create(ns, ModRemarkHistory, issueSnapshot(ex.ID, RemarkResolved, "устранено: расхождений в сравнении «"+in.ComparisonTitle+"» больше нет")); err != nil {
			return nil, err
		}
		out.Closed++
	}
	sort.Strings(out.Numbers)
	_, _ = RecordSnapshot(b, ns, objectID, time.Now()) // best effort: the dynamics must not fail the registration
	return out, nil
}
