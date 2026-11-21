package cmp

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// ModSnapshots keeps one StatusSnapshot per object and day: the dynamics of the corrections
// (ТЗ 3.9.4). Written whenever the registry changes (import, check against the documents, hand-made
// status change, list of remarks issued from a comparison); the last write of a day wins.
const ModSnapshots = "expertise_snapshots"

func snapshotField(status string) string { return "n_" + status }

// RecordSnapshot counts the object's live remarks and stores today's snapshot (creating or
// replacing it). It returns the snapshot.
func RecordSnapshot(b Backend, ns, objectID string, now time.Time) (*StatusSnapshot, error) {
	recs, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	var rows []StatusRow
	for _, r := range recs {
		if r.Get("archived") != "" {
			continue
		}
		rows = append(rows, StatusRow{Status: r.Get("status"), DueDate: r.Get("due_date")})
	}
	day := now.UTC().Format("2006-01-02")
	snap := CountStatuses(rows, day)

	v := idcheck.V("object", objectID, "day", day)
	v["total"] = []string{strconv.Itoa(snap.Total)}
	v["overdue"] = []string{strconv.Itoa(snap.Overdue)}
	for _, code := range RemarkStatusOrder {
		v[snapshotField(code)] = []string{strconv.Itoa(snap.ByStatus[code])}
	}
	old, err := b.Search(ns, ModSnapshots, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	for _, o := range old {
		if o.Get("day") == day {
			return &snap, b.Update(ns, ModSnapshots, o.ID, v)
		}
	}
	_, err = b.Create(ns, ModSnapshots, v)
	return &snap, err
}

// Dynamics returns the object's snapshots, oldest first.
func Dynamics(b Backend, ns, objectID string) ([]StatusSnapshot, error) {
	recs, err := b.Search(ns, ModSnapshots, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	out := make([]StatusSnapshot, 0, len(recs))
	for _, r := range recs {
		n := func(k string) int { x, _ := strconv.Atoi(r.Get(k)); return x }
		s := StatusSnapshot{Date: r.Get("day"), Total: n("total"), Overdue: n("overdue"), ByStatus: map[string]int{}}
		for _, code := range RemarkStatusOrder {
			s.ByStatus[code] = n(snapshotField(code))
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date < out[j].Date })
	return out, nil
}
