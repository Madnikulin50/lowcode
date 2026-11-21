package agent

import "math"

// ComputeCriticalPath runs a single-project CPM.
// DurationDays must be > 0; missing predecessors mean the activity is a start node.
func ComputeCriticalPath(acts []Activity) []Activity {
	if len(acts) == 0 {
		return acts
	}
	byID := map[string]int{}
	out := make([]Activity, len(acts))
	copy(out, acts)
	for i := range out {
		if out[i].DurationDays <= 0 {
			out[i].DurationDays = 1
		}
		byID[out[i].ID] = i
	}

	// Kahn-style forward: process in predecessor order.
	remaining := map[string]int{}
	succ := map[string][]string{}
	for i := range out {
		n := 0
		for _, p := range out[i].Predecessors {
			if _, ok := byID[p]; ok && p != out[i].ID {
				n++
				succ[p] = append(succ[p], out[i].ID)
			}
		}
		remaining[out[i].ID] = n
	}

	queue := make([]string, 0, len(out))
	for i := range out {
		if remaining[out[i].ID] == 0 {
			queue = append(queue, out[i].ID)
		}
	}
	order := make([]string, 0, len(out))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, s := range succ[id] {
			remaining[s]--
			if remaining[s] == 0 {
				queue = append(queue, s)
			}
		}
	}
	// Cycles / orphans: append leftover.
	seen := map[string]bool{}
	for _, id := range order {
		seen[id] = true
	}
	for i := range out {
		if !seen[out[i].ID] {
			order = append(order, out[i].ID)
		}
	}

	for _, id := range order {
		i := byID[id]
		es := 0.0
		for _, p := range out[i].Predecessors {
			j, ok := byID[p]
			if !ok {
				continue
			}
			if out[j].EF > es {
				es = out[j].EF
			}
		}
		out[i].ES = es
		out[i].EF = es + out[i].DurationDays
	}

	projectEnd := 0.0
	for i := range out {
		if out[i].EF > projectEnd {
			projectEnd = out[i].EF
		}
	}

	// Backward: reverse order.
	for i := range out {
		out[i].LF = projectEnd
	}
	for k := len(order) - 1; k >= 0; k-- {
		i := byID[order[k]]
		lf := projectEnd
		if children := succ[out[i].ID]; len(children) > 0 {
			lf = math.Inf(1)
			for _, s := range children {
				j := byID[s]
				if out[j].LS < lf {
					lf = out[j].LS
				}
			}
		}
		out[i].LF = lf
		out[i].LS = lf - out[i].DurationDays
		out[i].Float = out[i].LS - out[i].ES
		if out[i].Float < 0 {
			out[i].Float = 0
		}
		out[i].Critical = out[i].Float < 1e-6
	}
	return out
}

func WBSToActivities(items []WBSItem) []Activity {
	acts := make([]Activity, 0, len(items))
	for _, it := range items {
		dur := 1.0
		if !it.StartPlanned.IsZero() && it.EndPlanned.After(it.StartPlanned) {
			dur = it.EndPlanned.Sub(it.StartPlanned).Hours() / 24
			if dur < 1 {
				dur = 1
			}
		}
		preds := append([]string{}, it.Predecessors...)
		acts = append(acts, Activity{
			ID:           formatUint(it.ID),
			DurationDays: dur,
			Predecessors: preds,
		})
	}
	return acts
}

func ApplyCritical(items []WBSItem, acts []Activity) []WBSItem {
	byID := map[string]Activity{}
	for _, a := range acts {
		byID[a.ID] = a
	}
	out := make([]WBSItem, len(items))
	copy(out, items)
	for i := range out {
		a, ok := byID[formatUint(out[i].ID)]
		if !ok {
			continue
		}
		out[i].IsCritical = a.Critical
		out[i].TotalFloat = a.Float
	}
	return out
}
