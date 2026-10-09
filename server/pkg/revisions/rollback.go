package revisions

import (
	"encoding/json"
	"fmt"
	"sort"
)

// Rollback reconstructs the field values a resource had at the target
// revision.
//
// It starts from the current values (field -> values in order) and walks the
// revisions newer than target from the latest down, replacing each changed
// field with its previous values. Revisions that carry no changes (soft
// delete, undelete) are skipped. The input map is not modified.
//
// It returns false when the history cannot answer for target: there is no
// revision recorded with that number, or target is not older than the
// newest recorded revision.
func Rollback(current map[string][]string, revs []*Revision, target int) (map[string][]string, bool) {
	var (
		found  bool
		newest int
	)

	for _, r := range revs {
		if r.Revision == target {
			found = true
		}
		if r.Revision > newest {
			newest = r.Revision
		}
	}

	if !found || target >= newest {
		return nil, false
	}

	sorted := make([]*Revision, len(revs))
	copy(sorted, revs)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Revision > sorted[j].Revision })

	out := make(map[string][]string, len(current))
	for k, v := range current {
		out[k] = append([]string(nil), v...)
	}

	for _, r := range sorted {
		if r.Revision <= target {
			break
		}

		if r.Operation != Updated {
			continue
		}

		for _, ch := range r.Changes {
			vals := make([]string, 0, len(ch.Old))
			for _, v := range ch.Old {
				if v == nil {
					continue
				}
				vals = append(vals, changeValueString(v))
			}

			if len(vals) == 0 {
				delete(out, ch.Key)
				continue
			}

			out[ch.Key] = vals
		}
	}

	return out, true
}

func changeValueString(v any) string {
	switch aux := v.(type) {
	case string:
		return aux
	case json.Number:
		return aux.String()
	default:
		return fmt.Sprint(aux)
	}
}
