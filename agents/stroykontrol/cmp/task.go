package cmp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Задача (ТЗ 3.7.1): отдельная сущность — у замечания их может быть несколько (по требованию, по тому,
// по согласованию), бывают и задачи без замечания («загрузить исправленный том», «согласовать с ГИПом»).
// У задачи свои исполнитель, срок, оценка трудоёмкости и статус; поля самого замечания остаются «общим»
// ответственным и сроком. Выполнение задач на статус замечания не влияет: когда выполнены все задачи
// замечания, агент лишь подсказывает об этом (TaskCount.AllDone). «Выполнена» и «Возвращена на доработку»
// ставит не исполнитель, а другой специалист, и только с комментарием.

// Modules of the tasks.
const (
	ModTasks       = "tasks"
	ModTaskHistory = "task_history"
)

// Statuses of a task.
const (
	TaskOpen       = "open"        // Открыта
	TaskInProgress = "in_progress" // В работе
	TaskReview     = "review"      // На проверке
	TaskReturned   = "returned"    // Возвращена на доработку
	TaskDone       = "done"        // Выполнена
	TaskCancelled  = "cancelled"   // Отменена
)

// TaskStatusLabels are the Russian names of the statuses.
var TaskStatusLabels = map[string]string{
	TaskOpen: "Открыта", TaskInProgress: "В работе", TaskReview: "На проверке", TaskReturned: "Возвращена на доработку",
	TaskDone: "Выполнена", TaskCancelled: "Отменена",
}

// TaskStatusOrder is the order the statuses are shown in.
var TaskStatusOrder = []string{TaskOpen, TaskInProgress, TaskReview, TaskReturned, TaskDone, TaskCancelled}

// taskNext are the moves from a status.
var taskNext = map[string][]string{
	TaskOpen:       {TaskInProgress, TaskCancelled},
	TaskInProgress: {TaskReview, TaskOpen, TaskCancelled},
	TaskReview:     {TaskDone, TaskReturned, TaskCancelled},
	TaskReturned:   {TaskInProgress, TaskCancelled},
	TaskDone:       {TaskOpen},
	TaskCancelled:  {TaskOpen},
}

// NextTaskStatuses are the statuses a task may move to from the given one.
func NextTaskStatuses(from string) []string { return append([]string(nil), taskNext[from]...) }

// taskClosing: moves that need the specialist's decision (a reason is written down): the task is accepted,
// sent back, cancelled, or reopened after that.
func taskDecision(from, to string) bool {
	return to == TaskDone || to == TaskReturned || to == TaskCancelled || from == TaskDone || from == TaskCancelled
}

// CheckTaskTransition validates a move of the status: it must be a move the flow has, and a decision
// (accepted, sent back, cancelled, reopened) needs a comment.
func CheckTaskTransition(from, to, comment string) error {
	if _, ok := TaskStatusLabels[to]; !ok {
		return fmt.Errorf("unknown task status %q", to)
	}
	if from == to {
		return nil
	}
	ok := false
	for _, n := range taskNext[from] {
		ok = ok || n == to
	}
	if !ok {
		return fmt.Errorf("taskStatus: из «%s» нельзя перейти в «%s»", TaskStatusLabels[from], TaskStatusLabels[to])
	}
	if taskDecision(from, to) && strings.TrimSpace(comment) == "" {
		return fmt.Errorf("statusComment is required to set the task status «%s»", TaskStatusLabels[to])
	}
	return nil
}

// Task is a task as the interface and the reports see it.
type Task struct {
	ID          string `json:"id"`
	Object      string `json:"object"`
	Remark      string `json:"remark,omitempty"`
	Requirement int    `json:"requirement,omitempty"` // the number of the requirement of the remark the task is about
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Assignee    string `json:"assignee,omitempty"`
	DueDate     string `json:"dueDate,omitempty"`
	EstimateH   string `json:"estimateH,omitempty"`
	Status      string `json:"status"`
	Comment     string `json:"comment,omitempty"` // the reason of the last decision
	CreatedBy   string `json:"createdBy,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	DoneAt      string `json:"doneAt,omitempty"`
	ClosedBy    string `json:"closedBy,omitempty"`
	Overdue     bool   `json:"overdue"`
}

// TaskInput is what creates a task.
type TaskInput struct {
	Object      string `json:"object"`
	Remark      string `json:"remark"`
	Requirement int    `json:"requirement"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Assignee    string `json:"assignee"`
	DueDate     string `json:"dueDate"`
	EstimateH   string `json:"estimateH"`
}

// TaskPatch is what the specialist changes in a task; nil = leave.
type TaskPatch struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	Assignee      *string `json:"assignee"`
	DueDate       *string `json:"dueDate"`
	EstimateH     *string `json:"estimateH"`
	Status        *string `json:"status"`
	StatusComment *string `json:"statusComment"`
}

// TaskHist is one entry of the history of a task.
type TaskHist struct {
	At     string `json:"at"`
	Kind   string `json:"kind"`
	Source string `json:"source"`
	Note   string `json:"note,omitempty"`
	Author string `json:"author"`
	Status string `json:"status"`
}

// Kinds of an entry of the history of a task.
const (
	TaskHistCreated  = "created"
	TaskHistStatus   = "status"
	TaskHistAssignee = "assignee"
	TaskHistDue      = "due"
	TaskHistEstimate = "estimate"
	TaskHistEdit     = "edit"
)

const (
	maxTaskTitle = 200
	maxTaskText  = 4000
)

func taskFrom(r idcheck.Record, today string) Task {
	req, _ := strconv.Atoi(r.Get("requirement"))
	t := Task{
		ID: r.ID, Object: r.Get("object"), Remark: r.Get("remark"), Requirement: req, Title: r.Get("title"), Description: r.Get("description"),
		Assignee: r.Get("assignee"), DueDate: dateOnly10(r.Get("due_date")), EstimateH: r.Get("estimate_h"), Status: firstNonEmpty(r.Get("status"), TaskOpen),
		Comment: r.Get("comment"), CreatedBy: r.Get("created_by"), CreatedAt: r.Get("created_at"), DoneAt: r.Get("done_at"), ClosedBy: r.Get("closed_by"),
	}
	t.Overdue = t.DueDate != "" && t.DueDate < today && t.Status != TaskDone && t.Status != TaskCancelled
	return t
}

func dateOnly10(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func validTaskDate(s string) bool {
	if s == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func taskStamp(now time.Time) string { return now.UTC().Format("2006-01-02T15:04:05Z") }

func addTaskHist(b Backend, ns, taskID, kind, source, note, status string, actor Actor, now time.Time) error {
	_, err := b.Create(ns, ModTaskHistory, idcheck.V("task", taskID, "kind", kind, "source", source, "note", note, "author", actor.Label(), "status", status, "at", taskStamp(now)))
	return err
}

// CreateTask makes a task. The remark, when named, must belong to the object. The task starts «Открыта».
func CreateTask(b Backend, ns string, in TaskInput, actor Actor, now time.Time) (*Task, error) {
	in.Title = strings.TrimSpace(in.Title)
	switch {
	case in.Object == "":
		return nil, fmt.Errorf("task: objectID is required")
	case in.Title == "":
		return nil, fmt.Errorf("task: title is required")
	case len([]rune(in.Title)) > maxTaskTitle || len([]rune(in.Description)) > maxTaskText:
		return nil, fmt.Errorf("task: the title is at most %d characters and the description %d", maxTaskTitle, maxTaskText)
	case !validTaskDate(in.DueDate):
		return nil, fmt.Errorf("task: dueDate must be YYYY-MM-DD")
	}
	est, err := ParseEstimate(in.EstimateH)
	if err != nil {
		return nil, err
	}
	if in.Remark != "" && !remarkOfObject(b, ns, in.Remark, in.Object) {
		return nil, fmt.Errorf("task: the remark %s does not belong to the object %s", in.Remark, in.Object)
	}
	if err := actor.CheckTaskChange(TaskChange{Create: true, AssignOrDue: strings.TrimSpace(in.Assignee) != "" || in.DueDate != ""}); err != nil {
		return nil, err
	}
	v := idcheck.V("object", in.Object, "remark", in.Remark, "title", in.Title, "description", strings.TrimSpace(in.Description),
		"assignee", strings.TrimSpace(in.Assignee), "due_date", in.DueDate, "estimate_h", est, "status", TaskOpen,
		"created_by", actor.Label(), "created_at", taskStamp(now))
	if in.Requirement > 0 {
		v["requirement"] = []string{strconv.Itoa(in.Requirement)}
	}
	if o, err := b.GetRecord(ns, ModObjects, in.Object); err == nil && o.Get("data_kind") != "" {
		v["data_kind"] = []string{o.Get("data_kind")}
	}
	id, err := b.Create(ns, ModTasks, v)
	if err != nil {
		return nil, err
	}
	_ = addTaskHist(b, ns, id, TaskHistCreated, "задача создана", in.Title, TaskOpen, actor, now)
	if a := strings.TrimSpace(in.Assignee); a != "" {
		_ = addTaskHist(b, ns, id, TaskHistAssignee, "назначен: "+a, "", TaskOpen, actor, now)
	}
	rec, err := b.GetRecord(ns, ModTasks, id)
	if err != nil {
		return nil, err
	}
	t := taskFrom(*rec, now.UTC().Format("2006-01-02"))
	return &t, nil
}

// remarkOfObject says whether the remark is one of the object: by the record, else by the list of the object's remarks.
func remarkOfObject(b Backend, ns, remarkID, objectID string) bool {
	if rem, err := b.GetRecord(ns, ModExpertise, remarkID); err == nil {
		return rem.Get("object") == objectID
	}
	recs, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return false
	}
	for _, r := range recs {
		if r.ID == remarkID {
			return true
		}
	}
	return false
}

// UpdateTask applies the specialist's changes to a task: the fields, the assignee and the due date, the status
// (along the flow, a decision with a reason). Each change is an entry of the history.
func UpdateTask(b Backend, ns, objectID, taskID string, p TaskPatch, actor Actor, now time.Time) (*Task, error) {
	cur, err := b.GetRecord(ns, ModTasks, taskID)
	if err != nil {
		return nil, err
	}
	if cur.Get("object") != objectID {
		return nil, fmt.Errorf("task %s does not belong to object %s", taskID, objectID)
	}
	old := taskFrom(*cur, now.UTC().Format("2006-01-02"))
	v := idcheck.Values{}
	var hist []TaskHist
	change := TaskChange{Assignee: old.Assignee, FromStatus: old.Status}

	if p.Title != nil {
		t := strings.TrimSpace(*p.Title)
		if t == "" || len([]rune(t)) > maxTaskTitle {
			return nil, fmt.Errorf("task: title is required (at most %d characters)", maxTaskTitle)
		}
		if t != old.Title {
			v["title"] = []string{t}
			hist = append(hist, TaskHist{Kind: TaskHistEdit, Source: "название изменено", Note: t})
			change.WorkFields = true
		}
	}
	if p.Description != nil {
		d := strings.TrimSpace(*p.Description)
		if len([]rune(d)) > maxTaskText {
			return nil, fmt.Errorf("task: the description is at most %d characters", maxTaskText)
		}
		if d != old.Description {
			v["description"] = []string{d}
			hist = append(hist, TaskHist{Kind: TaskHistEdit, Source: "описание изменено"})
			change.WorkFields = true
		}
	}
	if p.EstimateH != nil {
		est, err := ParseEstimate(*p.EstimateH)
		if err != nil {
			return nil, err
		}
		if est != old.EstimateH {
			v["estimate_h"] = []string{est}
			src := "оценка: " + est + " ч"
			if est == "" {
				src = "оценка снята"
			}
			hist = append(hist, TaskHist{Kind: TaskHistEstimate, Source: src})
			change.WorkFields = true
		}
	}
	if p.Assignee != nil {
		a := strings.TrimSpace(*p.Assignee)
		if a != old.Assignee {
			v["assignee"] = []string{a}
			if a == "" {
				hist = append(hist, TaskHist{Kind: TaskHistAssignee, Source: "назначение снято", Note: "был: " + old.Assignee})
			} else {
				hist = append(hist, TaskHist{Kind: TaskHistAssignee, Source: "назначен: " + a})
			}
			change.AssignOrDue = true
		}
	}
	if p.DueDate != nil {
		if !validTaskDate(*p.DueDate) {
			return nil, fmt.Errorf("task: dueDate must be YYYY-MM-DD")
		}
		if *p.DueDate != old.DueDate {
			v["due_date"] = []string{*p.DueDate}
			src := "срок: " + *p.DueDate
			if *p.DueDate == "" {
				src = "срок снят"
			}
			hist = append(hist, TaskHist{Kind: TaskHistDue, Source: src})
			change.AssignOrDue = true
		}
	}
	newStatus := old.Status
	if p.Status != nil && *p.Status != old.Status {
		comment := ""
		if p.StatusComment != nil {
			comment = strings.TrimSpace(*p.StatusComment)
		}
		if err := CheckTaskTransition(old.Status, *p.Status, comment); err != nil {
			return nil, err
		}
		newStatus = *p.Status
		change.ToStatus = newStatus
		v["status"] = []string{newStatus}
		v["comment"] = []string{comment}
		switch newStatus {
		case TaskDone:
			v["done_at"], v["closed_by"] = []string{taskStamp(now)}, []string{actor.Label()}
		case TaskOpen:
			v["done_at"], v["closed_by"] = []string{""}, []string{""}
		case TaskReturned, TaskCancelled:
			v["closed_by"] = []string{actor.Label()}
		}
		hist = append(hist, TaskHist{Kind: TaskHistStatus, Source: "статус: " + TaskStatusLabels[old.Status] + " → " + TaskStatusLabels[newStatus], Note: comment})
	}
	if len(v) == 0 {
		return &old, nil
	}
	if err := actor.CheckTaskChange(change); err != nil {
		return nil, err
	}
	if err := b.Update(ns, ModTasks, taskID, v); err != nil {
		return nil, err
	}
	for _, h := range hist {
		if err := addTaskHist(b, ns, taskID, h.Kind, h.Source, h.Note, newStatus, actor, now); err != nil {
			return nil, err
		}
	}
	rec, err := b.GetRecord(ns, ModTasks, taskID)
	if err != nil {
		return nil, err
	}
	t := taskFrom(*rec, now.UTC().Format("2006-01-02"))
	return &t, nil
}

// TaskFilter selects tasks of an object.
type TaskFilter struct {
	Object, Remark, Assignee string
	Status                   map[string]bool
	OverdueOnly, OpenOnly    bool
}

// ListTasks returns the tasks of an object (filtered), the ones not done first, then by due date.
func ListTasks(b Backend, ns string, f TaskFilter, now time.Time) ([]Task, error) {
	q := ""
	if f.Object != "" {
		q = fmt.Sprintf("object = '%s'", f.Object)
	}
	recs, err := b.Search(ns, ModTasks, q)
	if err != nil {
		return nil, err
	}
	today := now.UTC().Format("2006-01-02")
	out := []Task{}
	who := NormName(f.Assignee)
	for _, r := range recs {
		t := taskFrom(r, today)
		switch {
		case f.Object != "" && t.Object != f.Object,
			f.Remark != "" && t.Remark != f.Remark,
			who == "—" && t.Assignee != "",
			who != "" && who != "—" && NormName(t.Assignee) != who,
			len(f.Status) > 0 && !f.Status[t.Status],
			f.OverdueOnly && !t.Overdue,
			f.OpenOnly && (t.Status == TaskDone || t.Status == TaskCancelled):
			continue
		}
		out = append(out, t)
	}
	rank := func(s string) int {
		if s == TaskDone || s == TaskCancelled {
			return 1
		}
		return 0
	}
	sort.SliceStable(out, func(i, j int) bool {
		if rank(out[i].Status) != rank(out[j].Status) {
			return rank(out[i].Status) < rank(out[j].Status)
		}
		a, c := out[i].DueDate, out[j].DueDate
		if (a == "") != (c == "") {
			return a != ""
		}
		if a != c {
			return a < c
		}
		return out[i].CreatedAt > out[j].CreatedAt
	})
	return out, nil
}

// TaskHistory returns the history of a task, oldest first.
func TaskHistory(b Backend, ns, objectID, taskID string) ([]TaskHist, error) {
	t, err := b.GetRecord(ns, ModTasks, taskID)
	if err != nil {
		return nil, err
	}
	if t.Get("object") != objectID {
		return nil, fmt.Errorf("task %s does not belong to object %s", taskID, objectID)
	}
	recs, err := b.Search(ns, ModTaskHistory, fmt.Sprintf("task = '%s'", taskID))
	if err != nil {
		return nil, err
	}
	// entries of the same second come in the order they were made: by the number of the record
	sort.SliceStable(recs, func(i, j int) bool { return recordNo(recs[i].ID) < recordNo(recs[j].ID) })
	out := make([]TaskHist, 0, len(recs))
	for _, r := range recs {
		out = append(out, TaskHist{At: r.Get("at"), Kind: r.Get("kind"), Source: r.Get("source"), Note: r.Get("note"), Author: r.Get("author"), Status: r.Get("status")})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At < out[j].At })
	return out, nil
}

// recordNo is the number a record id ends with (platform ids are numbers, the in-memory ones "module-7"); 0 if none.
func recordNo(id string) uint64 {
	i := len(id)
	for i > 0 && id[i-1] >= '0' && id[i-1] <= '9' {
		i--
	}
	n, _ := strconv.ParseUint(id[i:], 10, 64)
	return n
}

// TaskCount is how many tasks a remark has and what became of them.
type TaskCount struct {
	Total   int `json:"total"`
	Done    int `json:"done"`    // done
	Open    int `json:"open"`    // not done and not cancelled
	Overdue int `json:"overdue"` // open and past the due date
	// AllDone: there are tasks and every one that was not cancelled is done — a hint that the remark can go to the control
	AllDone bool `json:"allDone"`
}

// CountTasks counts the tasks by remark (tasks without a remark are not counted).
func CountTasks(tasks []Task) map[string]TaskCount {
	out := map[string]TaskCount{}
	live := map[string]int{}
	for _, t := range tasks {
		if t.Remark == "" {
			continue
		}
		c := out[t.Remark]
		c.Total++
		switch t.Status {
		case TaskDone:
			c.Done++
			live[t.Remark]++
		case TaskCancelled:
		default:
			c.Open++
			live[t.Remark]++
			if t.Overdue {
				c.Overdue++
			}
		}
		out[t.Remark] = c
	}
	for id, c := range out {
		c.AllDone = c.Done > 0 && c.Open == 0
		out[id] = c
	}
	return out
}
