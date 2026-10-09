package agent

import (
	"fmt"
	"time"
)

var riskLevelScore = map[string]float64{
	"low": 1, "medium": 2, "high": 3, "critical": 4,
}

func RiskScore(probability, impact string) float64 {
	p := riskLevelScore[probability]
	i := riskLevelScore[impact]
	if p == 0 {
		p = 1
	}
	if i == 0 {
		i = 1
	}
	return p * i
}

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

func CollectReserveAlerts(lines []BudgetLine) []Alert {
	var out []Alert
	for _, l := range lines {
		if l.Reserve > 0 {
			continue
		}
		if l.Planned == 0 && l.Actual == 0 {
			continue
		}
		out = append(out, Alert{
			Kind:    "reserve_exhausted",
			Title:   "Резерв исчерпан: " + l.Article,
			Project: l.Project,
			Detail:  fmt.Sprintf("Статья %s: план %.0f факт %.0f резерв %.0f", l.Article, l.Planned, l.Actual, l.Reserve),
		})
	}
	return out
}

func CollectRFCOverdue(rfcs []RFC, now time.Time) []Alert {
	var out []Alert
	for _, r := range rfcs {
		if r.Status != "in_review" {
			continue
		}
		if r.EndAfter.IsZero() || !r.EndAfter.Before(now) {
			continue
		}
		out = append(out, Alert{
			Kind:    "overdue_rfc",
			Title:   "Просрочен RFC: " + r.Title,
			Project: r.Project,
			Detail:  "RFC на согласовании, плановый финиш после изменения уже в прошлом",
		})
	}
	return out
}
