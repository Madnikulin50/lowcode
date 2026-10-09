package agent

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DefaultApprovalRoles is the PMO → investor chain (FR-019).
var DefaultApprovalRoles = []string{"pmo", "investor"}

func nextVersion(existing []float64) int {
	max := 0
	for _, v := range existing {
		n := int(v)
		if n > max {
			max = n
		}
	}
	return max + 1
}

func membersByRole(members []Member, role string) []Member {
	var out []Member
	for _, m := range members {
		if m.Role == role && m.User != "" && m.User != "0" {
			out = append(out, m)
		}
	}
	return out
}

// BuildApprovalSteps returns ordered (step, role, user) tuples.
// Author is step 1 when set; then PMO, then investor from project_members.
func BuildApprovalSteps(author string, members []Member, hasContract bool) []Approval {
	seen := map[string]bool{}
	var out []Approval
	step := 1.0
	add := func(user, role string) {
		if user == "" || user == "0" || seen[user] {
			return
		}
		seen[user] = true
		out = append(out, Approval{Approver: user, Role: role, Step: step, Decision: "pending"})
		step++
	}
	if author != "" {
		add(author, "pmo")
	}
	roles := append([]string{}, DefaultApprovalRoles...)
	if hasContract {
		roles = append([]string{"pmo"}, roles...)
	}
	for _, role := range roles {
		for _, m := range membersByRole(members, role) {
			add(m.User, role)
		}
	}
	return out
}

func pendingApprovals(all []Approval) []Approval {
	var out []Approval
	for _, a := range all {
		if a.Decision == "" || a.Decision == "pending" {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Step < out[j].Step })
	return out
}

func pickApproval(pending []Approval, userID string) *Approval {
	if len(pending) == 0 {
		return nil
	}
	if userID != "" {
		for i := range pending {
			if pending[i].Approver == userID {
				return &pending[i]
			}
		}
	}
	return &pending[0]
}

func fmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

type WorkflowResult struct {
	DocumentID  string `json:"documentID"`
	Status      string `json:"status"`
	Approvals   int    `json:"approvals"`
	Version     int    `json:"version,omitempty"`
	NotifyEmail string `json:"notifyEmail,omitempty"`
	Title       string `json:"title,omitempty"`
	Assignee    string `json:"assignee,omitempty"`
	Message     string `json:"message"`
}

func (e *Engine) SubmitApproval(ctx context.Context, req JobRequest) (*WorkflowResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.DocumentRef())
	if id == 0 {
		return nil, fmt.Errorf("recordID (document) required")
	}
	doc, err := cz.LoadDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("document %d: %w", id, err)
	}
	members, _ := cz.LoadMembers(ctx, doc.Project)
	existing, _ := cz.LoadApprovals(ctx, formatUint(doc.ID))
	pending := pendingApprovals(existing)

	ver := 0
	versions, _ := cz.ListVersionNumbers(ctx, formatUint(doc.ID))
	ver = nextVersion(versions)
	verVals := compactValues(map[string]string{
		"document":   formatUint(doc.ID),
		"version":    strconv.Itoa(ver),
		"file":       doc.File,
		"author":     firstNonEmpty(req.UserID, doc.Author),
		"comment":    "Автоверсия при отправке на согласование",
		"created_on": fmtDate(time.Now()),
	})
	if _, err := cz.CreateValues(ctx, "document_versions", verVals); err != nil {
		delete(verVals, "file")
		_, _ = cz.CreateValues(ctx, "document_versions", verVals)
	}

	created := 0
	if len(pending) == 0 {
		author := firstNonEmpty(doc.Author, req.UserID)
		steps := BuildApprovalSteps(author, members, doc.Contract != "" && doc.Contract != "0")
		if len(steps) == 0 {
			fallback := firstNonEmpty(req.UserID, doc.Assignee, author)
			if fallback != "" {
				steps = []Approval{{Approver: fallback, Role: "pmo", Step: 1, Decision: "pending"}}
			}
		}
		due := fmtDate(doc.DueDate)
		for _, s := range steps {
			vals := compactValues(map[string]string{
				"document": formatUint(doc.ID),
				"approver": s.Approver,
				"decision": "pending",
				"due_date": due,
				"role":     s.Role,
				"step":     strconv.FormatFloat(s.Step, 'f', 0, 64),
			})
			if _, err := cz.CreateValues(ctx, "approvals", vals); err != nil {
				return nil, fmt.Errorf("create approval: %w", err)
			}
			created++
		}
		pending = steps
	}

	assignee := doc.Assignee
	if len(pending) > 0 && pending[0].Approver != "" {
		assignee = pending[0].Approver
	}
	if err := cz.UpdateValues(ctx, "documents", doc.ID, compactValues(map[string]string{
		"status":   "in_review",
		"assignee": assignee,
	})); err != nil {
		return nil, err
	}
	email, _ := cz.UserEmail(ctx, assignee)
	return &WorkflowResult{
		DocumentID:  formatUint(doc.ID),
		Status:      "in_review",
		Approvals:   created,
		Version:     ver,
		NotifyEmail: email,
		Title:       doc.Title,
		Assignee:    assignee,
		Message:     "Документ отправлен на согласование",
	}, nil
}

func (e *Engine) DecideApproval(ctx context.Context, req JobRequest) (*WorkflowResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.DocumentRef())
	if id == 0 {
		return nil, fmt.Errorf("recordID (document) required")
	}
	decision := strings.TrimSpace(req.Decision)
	if decision == "" {
		decision = "approved"
	}
	if decision != "approved" && decision != "rejected" {
		return nil, fmt.Errorf("decision must be approved or rejected")
	}
	doc, err := cz.LoadDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("document %d: %w", id, err)
	}
	all, err := cz.LoadApprovals(ctx, formatUint(doc.ID))
	if err != nil {
		return nil, err
	}
	pending := pendingApprovals(all)
	cur := pickApproval(pending, req.UserID)
	if cur == nil || cur.ID == 0 {
		return nil, fmt.Errorf("маршрута нет — сначала нажмите «Отправить»")
	}
	now := fmtDate(time.Now())
	if err := cz.UpdateValues(ctx, "approvals", cur.ID, compactValues(map[string]string{
		"decision":   decision,
		"decided_at": now,
		"comment":    req.Comment,
	})); err != nil {
		return nil, err
	}
	status := doc.Status
	assignee := doc.Assignee
	if decision == "rejected" {
		status = "rejected"
	} else {
		left := pendingApprovals(all)
		var rest []Approval
		for _, a := range left {
			if a.ID != cur.ID {
				rest = append(rest, a)
			}
		}
		if len(rest) == 0 {
			status = "approved"
			assignee = ""
		} else {
			status = "in_review"
			assignee = rest[0].Approver
		}
	}
	if err := cz.UpdateValues(ctx, "documents", doc.ID, compactValues(map[string]string{
		"status":   status,
		"assignee": assignee,
	})); err != nil {
		return nil, err
	}
	email, _ := cz.UserEmail(ctx, assignee)
	msg := "Шаг согласован"
	if decision == "rejected" {
		msg = "Документ отклонён"
	} else if status == "approved" {
		msg = "Все шаги пройдены, документ утверждён"
	}
	return &WorkflowResult{
		DocumentID:  formatUint(doc.ID),
		Status:      status,
		NotifyEmail: email,
		Title:       doc.Title,
		Assignee:    assignee,
		Message:     msg,
	}, nil
}

func (e *Engine) EscalateApproval(ctx context.Context, req JobRequest) (*WorkflowResult, error) {
	req.Normalize()
	cz := e.client(req)
	id := ParseID(req.DocumentRef())
	if id == 0 {
		return nil, fmt.Errorf("recordID (document) required")
	}
	doc, err := cz.LoadDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("document %d: %w", id, err)
	}
	all, _ := cz.LoadApprovals(ctx, formatUint(doc.ID))
	pending := pendingApprovals(all)
	cur := pickApproval(pending, req.UserID)
	if cur != nil && cur.ID != 0 {
		_ = cz.UpdateValues(ctx, "approvals", cur.ID, compactValues(map[string]string{
			"decision":   "escalated",
			"decided_at": fmtDate(time.Now()),
			"comment":    firstNonEmpty(req.Comment, "Эскалация"),
		}))
	}
	members, _ := cz.LoadMembers(ctx, doc.Project)
	pmo := membersByRole(members, "pmo")
	assignee := doc.Assignee
	if len(pmo) > 0 {
		assignee = pmo[0].User
	}
	maxStep := 0.0
	for _, a := range all {
		if a.Step > maxStep {
			maxStep = a.Step
		}
	}
	if assignee != "" {
		_, _ = cz.CreateValues(ctx, "approvals", compactValues(map[string]string{
			"document": formatUint(doc.ID),
			"approver": assignee,
			"decision": "pending",
			"due_date": fmtDate(doc.DueDate),
			"role":     "pmo",
			"step":     strconv.FormatFloat(maxStep+1, 'f', 0, 64),
			"comment":  "Эскалировано на PMO",
		}))
	}
	notes := firstNonEmpty(doc.Number, "")
	_ = cz.UpdateValues(ctx, "documents", doc.ID, compactValues(map[string]string{
		"status":   "in_review",
		"assignee": assignee,
		"notes":    firstNonEmpty(req.Comment, "Эскалация: "+notes),
	}))
	email, _ := cz.UserEmail(ctx, assignee)
	return &WorkflowResult{
		DocumentID:  formatUint(doc.ID),
		Status:      "in_review",
		NotifyEmail: email,
		Title:       doc.Title,
		Assignee:    assignee,
		Message:     "Согласование эскалировано на PMO",
	}, nil
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func compactValues(m map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			continue
		}
		out[k] = v
	}
	return out
}
