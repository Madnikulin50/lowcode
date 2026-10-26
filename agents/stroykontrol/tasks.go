package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Задачи (ТЗ 3.7.1): отдельная сущность — модуль tasks (в режиме --fixtures лежит в памяти). Правила —
// cmp/task.go; здесь маршруты и права вызывающего.

// taskB is where the tasks live: the namespace (Compose) or the memory (--fixtures).
var taskB idcheck.Backend

// initTasks chooses the place of the tasks once the store is known.
func initTasks() {
	if fx, ok := compose.(*FixtureStore); ok {
		mem := newMemBackend()
		mem.remarksOf = func(objectID string) []idcheck.Record {
			b, err := fx.Board("", objectID)
			if err != nil {
				return nil
			}
			var out []idcheck.Record
			for _, r := range b.Remarks {
				out = append(out, idcheck.Record{ID: r.ID, Values: map[string][]string{"object": {objectID}, "number": {r.Number}}})
			}
			return out
		}
		taskB = mem
		return
	}
	if cmpRunner != nil {
		taskB = cmpRunner.Backend
	}
}

func taskError(w http.ResponseWriter, err error) {
	code := http.StatusBadGateway
	switch msg := err.Error(); {
	case strings.Contains(msg, "недостаточно прав") || strings.Contains(msg, "исполнитель задачи"):
		code = http.StatusForbidden
	case strings.Contains(msg, "task") || strings.Contains(msg, "statusComment") || strings.Contains(msg, "estimateH") || strings.Contains(msg, "unknown"):
		code = http.StatusBadRequest
	}
	var fb forbiddenError
	if errors.As(err, &fb) {
		code = http.StatusForbidden
	}
	writeErr(w, code, err)
}

type tasksOut struct {
	Tasks  []cmp.Task               `json:"tasks"`
	Counts map[string]cmp.TaskCount `json:"counts"` // by remark
}

func tasksBackend(w http.ResponseWriter) bool {
	if taskB == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("задачи недоступны"))
		return false
	}
	return true
}

// GET /api/tasks?objectID=&remarkID=&assignee=&status=a,b&overdue=1&open=1
func handleTasks(w http.ResponseWriter, r *http.Request) {
	ns, objectID, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if !tasksBackend(w) {
		return
	}
	q := r.URL.Query()
	f := cmp.TaskFilter{Object: objectID, Remark: q.Get("remarkID"), Assignee: strings.TrimSpace(q.Get("assignee")), OverdueOnly: q.Get("overdue") == "1", OpenOnly: q.Get("open") == "1"}
	if s := q.Get("status"); s != "" {
		f.Status = map[string]bool{}
		for _, c := range strings.Split(s, ",") {
			if _, ok := cmp.TaskStatusLabels[strings.TrimSpace(c)]; !ok {
				writeErr(w, http.StatusBadRequest, fmt.Errorf("unknown task status %q", c))
				return
			}
			f.Status[strings.TrimSpace(c)] = true
		}
	}
	all, err := cmp.ListTasks(taskB, ns, cmp.TaskFilter{Object: objectID}, time.Now())
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	list, err := cmp.ListTasks(taskB, ns, f, time.Now())
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, tasksOut{Tasks: list, Counts: cmp.CountTasks(all)})
}

// POST /api/tasks {namespaceID, objectID, remark, requirement, title, description, assignee, dueDate, estimateH}
func handleTaskCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		ObjectID    string `json:"objectID"`
		cmp.TaskInput
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if body.Object == "" {
		body.Object = body.ObjectID
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("namespaceID is required"))
		return
	}
	if !tasksBackend(w) {
		return
	}
	by, err := actorOf(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	t, err := cmp.CreateTask(taskB, ns, body.TaskInput, by, time.Now())
	if err != nil {
		taskError(w, err)
		return
	}
	notif.soon(ns, t.Object) // the assignee hears of it now, not at the next pass
	writeJSON(w, http.StatusOK, t)
}

// POST /api/tasks/update {namespaceID, objectID, id, title?, description?, assignee?, dueDate?, estimateH?, status?, statusComment?}
func handleTaskUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string `json:"namespaceID"`
		ObjectID    string `json:"objectID"`
		ID          string `json:"id"`
		cmp.TaskPatch
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	ns := firstNonEmpty(body.NamespaceID, defaultNamespaceID)
	if ns == "" || body.ObjectID == "" || body.ID == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("objectID and id are required"))
		return
	}
	if !tasksBackend(w) {
		return
	}
	by, err := actorOf(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	t, err := cmp.UpdateTask(taskB, ns, body.ObjectID, body.ID, body.TaskPatch, by, time.Now())
	if err != nil {
		taskError(w, err)
		return
	}
	if body.Status != nil || body.Assignee != nil || body.DueDate != nil {
		notif.soon(ns, body.ObjectID)
	}
	writeJSON(w, http.StatusOK, t)
}

// GET /api/tasks/history?objectID=&id=
func handleTaskHistory(w http.ResponseWriter, r *http.Request) {
	ns, objectID, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErr(w, http.StatusBadRequest, fmt.Errorf("id is required"))
		return
	}
	if !tasksBackend(w) {
		return
	}
	h, err := cmp.TaskHistory(taskB, ns, objectID, id)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, h)
}
