package agent

import (
	"fmt"
	"time"
)

func CollectAlerts(docs []Document, items []WBSItem, now time.Time, cpiThreshold float64) []Alert {
	if cpiThreshold <= 0 {
		cpiThreshold = 0.9
	}
	var out []Alert
	for _, d := range docs {
		if d.Status != "in_review" && d.Status != "draft" {
			continue
		}
		if d.DueDate.IsZero() || !d.DueDate.Before(now) {
			continue
		}
		out = append(out, Alert{
			Kind:    "overdue_document",
			Title:   "Просрочено согласование: " + d.Title,
			Project: d.Project,
			Detail:  fmt.Sprintf("Срок %s, статус %s", d.DueDate.Format("2006-01-02"), d.Status),
		})
	}
	for _, it := range items {
		if it.CPI == 0 {
			continue
		}
		if it.CPI < cpiThreshold {
			out = append(out, Alert{
				Kind:    "low_cpi",
				Title:   "CPI ниже порога: " + it.Name,
				Project: it.ProjectID,
				WBS:     formatUint(it.ID),
				Detail:  fmt.Sprintf("CPI=%.3f порог=%.2f код=%s", it.CPI, cpiThreshold, it.Code),
			})
		}
	}
	return out
}
