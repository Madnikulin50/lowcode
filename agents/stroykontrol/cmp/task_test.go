package cmp

import (
	"strings"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

var (
	tNow     = time.Date(2026, 10, 22, 10, 0, 0, 0, time.UTC)
	qcActor  = Actor{Known: true, Name: "Иванова", Roles: []string{SpaceQC}}
	desActor = Actor{Known: true, Name: "Петров", Roles: []string{SpaceDesigner}}
	mgrActor = Actor{Known: true, Name: "Сидоров", Roles: []string{SpaceManager}}
	obsActor = Actor{Known: true, Name: "Наблюдатель", Roles: []string{SpaceObserver}}
)

func taskFixture(t *testing.T) (*fakeBackend, string, string) {
	t.Helper()
	f := newFake()
	obj := f.put(ModObjects, idcheck.V("name", "Объект"))
	rem := f.put(ModExpertise, idcheck.V("object", obj, "number", "30.1", "text", "Замечание"))
	return f, obj, rem
}

func s(v string) *string { return &v }

func TestTaskTransitions(t *testing.T) {
	ok := [][2]string{{TaskOpen, TaskInProgress}, {TaskInProgress, TaskReview}, {TaskReview, TaskDone}, {TaskReview, TaskReturned}, {TaskReturned, TaskInProgress}, {TaskDone, TaskOpen}, {TaskInProgress, TaskOpen}}
	for _, c := range ok {
		comment := "ok"
		if err := CheckTaskTransition(c[0], c[1], comment); err != nil {
			t.Errorf("%s → %s: %v", c[0], c[1], err)
		}
	}
	for _, c := range [][2]string{{TaskOpen, TaskDone}, {TaskOpen, TaskReview}, {TaskInProgress, TaskDone}, {TaskDone, TaskReview}, {TaskCancelled, TaskDone}} {
		if err := CheckTaskTransition(c[0], c[1], "причина"); err == nil || !strings.Contains(err.Error(), "taskStatus") {
			t.Errorf("%s → %s must not be a move of the flow: %v", c[0], c[1], err)
		}
	}
	// a decision needs a reason, work does not
	for _, to := range []string{TaskDone, TaskReturned, TaskCancelled} {
		from := TaskReview
		if to == TaskCancelled {
			from = TaskOpen
		}
		if err := CheckTaskTransition(from, to, " "); err == nil || !strings.Contains(err.Error(), "statusComment") {
			t.Errorf("%s → %s needs a comment: %v", from, to, err)
		}
	}
	if err := CheckTaskTransition(TaskOpen, TaskInProgress, ""); err != nil {
		t.Errorf("taking a task into work needs no reason: %v", err)
	}
	if err := CheckTaskTransition(TaskDone, TaskOpen, ""); err == nil {
		t.Error("reopening a done task is a decision and needs a reason")
	}
	if err := CheckTaskTransition(TaskOpen, "bogus", "x"); err == nil {
		t.Error("unknown status")
	}
	if err := CheckTaskTransition(TaskOpen, TaskOpen, ""); err != nil {
		t.Error("no move, nothing to check")
	}
}

func TestCreateTaskValidatesAndRecordsHistory(t *testing.T) {
	f, obj, rem := taskFixture(t)
	in := TaskInput{Object: obj, Remark: rem, Requirement: 2, Title: " Загрузить исправленный том ", Assignee: "Петров", DueDate: "2026-11-01", EstimateH: "2,5"}
	task, err := CreateTask(f, "ns", in, mgrActor, tNow)
	if err != nil {
		t.Fatal(err)
	}
	if task.Status != TaskOpen || task.Title != "Загрузить исправленный том" || task.EstimateH != "2.5" || task.Requirement != 2 || task.CreatedBy != "Сидоров" || task.Overdue {
		t.Errorf("%+v", task)
	}
	hist, err := TaskHistory(f, "ns", obj, task.ID)
	if err != nil || len(hist) != 2 || hist[0].Kind != TaskHistCreated || hist[1].Source != "назначен: Петров" {
		t.Errorf("history: %v %+v", err, hist)
	}
	// a task without a remark is allowed; one with a foreign remark, a bad date, an empty title is not
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "Согласовать с ГИПом"}, qcActor, tNow); err != nil {
		t.Errorf("a task without a remark: %v", err)
	}
	other := f.put(ModObjects, idcheck.V("name", "Другой"))
	foreign := f.put(ModExpertise, idcheck.V("object", other, "number", "1"))
	for name, bad := range map[string]TaskInput{
		"foreign remark": {Object: obj, Remark: foreign, Title: "x"}, "no title": {Object: obj, Title: "  "}, "bad date": {Object: obj, Title: "x", DueDate: "01.11.2026"},
		"bad estimate": {Object: obj, Title: "x", EstimateH: "много"}, "no object": {Title: "x"},
	} {
		if _, err := CreateTask(f, "ns", bad, mgrActor, tNow); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
	// roles: an observer creates nothing, a designer cannot give a task to someone, a specialist without the right to assign neither
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "x"}, obsActor, tNow); err == nil {
		t.Error("an observer creates no task")
	}
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "x", Assignee: "Иванова"}, desActor, tNow); err == nil {
		t.Error("a designer does not assign")
	}
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "для себя"}, desActor, tNow); err != nil {
		t.Errorf("a designer may make a task without an assignee: %v", err)
	}
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "x", Assignee: "Иванова"}, qcActor, tNow); err == nil {
		t.Error("the quality specialist does not assign either (only the manager)")
	}
}

func TestTaskFlowWithRoles(t *testing.T) {
	f, obj, rem := taskFixture(t)
	task, err := CreateTask(f, "ns", TaskInput{Object: obj, Remark: rem, Title: "Исправить лист 5", Assignee: "Петров", DueDate: "2026-10-20"}, mgrActor, tNow)
	if err != nil {
		t.Fatal(err)
	}
	if !task.Overdue {
		t.Error("the due date has passed")
	}
	up := func(id string, p TaskPatch, a Actor) (*Task, error) {
		return UpdateTask(f, "ns", obj, id, p, a, tNow.Add(time.Hour))
	}
	// the assignee works on the task and hands it over
	if _, err := up(task.ID, TaskPatch{Status: s(TaskInProgress)}, desActor); err != nil {
		t.Fatalf("take into work: %v", err)
	}
	if got, err := up(task.ID, TaskPatch{Status: s(TaskReview)}, desActor); err != nil || got.Status != TaskReview {
		t.Fatalf("hand over: %v %+v", err, got)
	}
	// but cannot accept it himself, even with a reason, and a designer has no right to
	if _, err := up(task.ID, TaskPatch{Status: s(TaskDone), StatusComment: s("сделано")}, desActor); err == nil {
		t.Error("the designer does not accept tasks")
	}
	self := Actor{Known: true, Name: "Петров", Roles: []string{SpaceQC}}
	if _, err := up(task.ID, TaskPatch{Status: s(TaskDone), StatusComment: s("сделано")}, self); err == nil || !strings.Contains(err.Error(), "исполнитель задачи") {
		t.Errorf("the assignee does not accept his own task: %v", err)
	}
	// no reason, no decision
	if _, err := up(task.ID, TaskPatch{Status: s(TaskReturned)}, qcActor); err == nil || !strings.Contains(err.Error(), "statusComment") {
		t.Errorf("a return needs a reason: %v", err)
	}
	back, err := up(task.ID, TaskPatch{Status: s(TaskReturned), StatusComment: s("лист 5 не исправлен")}, qcActor)
	if err != nil || back.Status != TaskReturned || back.Comment != "лист 5 не исправлен" || back.ClosedBy != "Иванова" {
		t.Fatalf("return: %v %+v", err, back)
	}
	for _, st := range []string{TaskInProgress, TaskReview} {
		if _, err := up(task.ID, TaskPatch{Status: s(st)}, desActor); err != nil {
			t.Fatalf("%s: %v", st, err)
		}
	}
	done, err := up(task.ID, TaskPatch{Status: s(TaskDone), StatusComment: s("проверено по листу 5")}, qcActor)
	if err != nil || done.Status != TaskDone || done.DoneAt == "" || done.Overdue {
		t.Fatalf("accept: %v %+v", err, done)
	}
	// fields: only the manager moves the due date and the assignee; any specialist edits the title
	if _, err := up(task.ID, TaskPatch{DueDate: s("2026-12-01")}, qcActor); err == nil {
		t.Error("the due date is the manager's")
	}
	if got, err := up(task.ID, TaskPatch{DueDate: s("2026-12-01"), Assignee: s("Козлов")}, mgrActor); err != nil || got.DueDate != "2026-12-01" || got.Assignee != "Козлов" {
		t.Errorf("manager: %v %+v", err, got)
	}
	if got, err := up(task.ID, TaskPatch{Title: s("Исправить лист 5 и 6")}, desActor); err != nil || got.Title != "Исправить лист 5 и 6" {
		t.Errorf("title: %v %+v", err, got)
	}
	if again, err := up(task.ID, TaskPatch{Title: s("Исправить лист 5 и 6")}, obsActor); err != nil || again.Title == "" {
		t.Errorf("a patch that changes nothing needs no right and is no history: %v", err)
	}
	// reopening a done task is a decision
	if _, err := up(task.ID, TaskPatch{Status: s(TaskOpen)}, qcActor); err == nil {
		t.Error("a reopening needs a reason")
	}
	if got, err := up(task.ID, TaskPatch{Status: s(TaskOpen), StatusComment: s("нашли ещё замечание")}, qcActor); err != nil || got.DoneAt != "" || got.Status != TaskOpen {
		t.Errorf("reopen: %v %+v", err, got)
	}
	hist, _ := TaskHistory(f, "ns", obj, task.ID)
	kinds := map[string]int{}
	for _, h := range hist {
		kinds[h.Kind]++
	}
	if kinds[TaskHistStatus] != 7 || kinds[TaskHistAssignee] < 2 || kinds[TaskHistDue] != 1 || kinds[TaskHistEdit] != 1 {
		t.Errorf("every change is a history entry: %v", kinds)
	}
	if _, err := UpdateTask(f, "ns", "other-object", task.ID, TaskPatch{Title: s("x")}, mgrActor, tNow); err == nil {
		t.Error("a task of another object is not reachable")
	}
}

func TestListAndCountTasks(t *testing.T) {
	f, obj, rem := taskFixture(t)
	mk := func(title, assignee, due string, remark string) string {
		task, err := CreateTask(f, "ns", TaskInput{Object: obj, Remark: remark, Title: title, Assignee: assignee, DueDate: due}, mgrActor, tNow)
		if err != nil {
			t.Fatal(err)
		}
		return task.ID
	}
	a := mk("А", "Петров", "2026-10-30", rem)
	b := mk("Б", "Петров", "2026-10-01", rem) // overdue
	c := mk("В", "", "", rem)
	_ = mk("Без замечания", "Иванова", "2026-10-25", "")
	for _, st := range [][2]string{{TaskInProgress, ""}, {TaskReview, ""}, {TaskDone, "принято"}} {
		if _, err := UpdateTask(f, "ns", obj, a, TaskPatch{Status: s(st[0]), StatusComment: s(st[1])}, qcActor, tNow); err != nil {
			t.Fatal(err)
		}
	}
	all, err := ListTasks(f, "ns", TaskFilter{Object: obj}, tNow)
	if err != nil || len(all) != 4 || all[len(all)-1].ID != a {
		t.Fatalf("done goes last: %v %+v", err, all)
	}
	if all[0].ID != b {
		t.Errorf("the overdue one first: %+v", all[0])
	}
	if got, _ := ListTasks(f, "ns", TaskFilter{Object: obj, Assignee: "петров"}, tNow); len(got) != 2 {
		t.Errorf("by assignee (case aside): %d", len(got))
	}
	if got, _ := ListTasks(f, "ns", TaskFilter{Object: obj, Assignee: "—"}, tNow); len(got) != 1 || got[0].ID != c {
		t.Errorf("not assigned: %+v", got)
	}
	if got, _ := ListTasks(f, "ns", TaskFilter{Object: obj, OverdueOnly: true}, tNow); len(got) != 1 || got[0].ID != b {
		t.Errorf("overdue: %+v", got)
	}
	if got, _ := ListTasks(f, "ns", TaskFilter{Object: obj, OpenOnly: true}, tNow); len(got) != 3 {
		t.Errorf("open: %d", len(got))
	}
	if got, _ := ListTasks(f, "ns", TaskFilter{Object: obj, Remark: rem, Status: map[string]bool{TaskDone: true}}, tNow); len(got) != 1 || got[0].ID != a {
		t.Errorf("by remark and status: %+v", got)
	}
	cnt := CountTasks(all)[rem]
	if cnt.Total != 3 || cnt.Done != 1 || cnt.Open != 2 || cnt.Overdue != 1 || cnt.AllDone {
		t.Errorf("count: %+v", cnt)
	}
	// when the rest is done or cancelled, the remark's tasks are all done — a hint, nothing else
	for _, id := range []string{b, c} {
		st := TaskCancelled
		if id == b {
			for _, to := range []string{TaskInProgress, TaskReview, TaskDone} {
				comment := ""
				if to == TaskDone {
					comment = "ok"
				}
				if _, err := UpdateTask(f, "ns", obj, id, TaskPatch{Status: s(to), StatusComment: s(comment)}, qcActor, tNow); err != nil {
					t.Fatal(err)
				}
			}
			continue
		}
		if _, err := UpdateTask(f, "ns", obj, id, TaskPatch{Status: s(st), StatusComment: s("не нужна")}, qcActor, tNow); err != nil {
			t.Fatal(err)
		}
	}
	all, _ = ListTasks(f, "ns", TaskFilter{Object: obj}, tNow)
	if c := CountTasks(all)[rem]; !c.AllDone || c.Open != 0 || c.Done != 2 {
		t.Errorf("all done: %+v", c)
	}
	if st := f.recs[ModExpertise][rem].Get("status"); st != "" {
		t.Errorf("the tasks never change the status of the remark: %q", st)
	}
}

func TestTaskNotifications(t *testing.T) {
	f, obj, rem := taskFixture(t)
	dir := []Recipient{{Name: "Петров", Role: RoleAssignee}, {Name: "Сидоров", Role: RoleManager, Events: []string{EvOverdue, EvStatus}}}
	t1, err := CreateTask(f, "ns", TaskInput{Object: obj, Remark: rem, Title: "Исправить лист 5", Assignee: "Петров", DueDate: "2026-10-20"}, mgrActor, tNow)
	if err != nil {
		t.Fatal(err)
	}
	t2, _ := CreateTask(f, "ns", TaskInput{Object: obj, Title: "Согласовать с ГИПом", Assignee: "Петров", DueDate: "2026-10-24"}, mgrActor, tNow)
	if _, err := CreateTask(f, "ns", TaskInput{Object: obj, Title: "Без исполнителя"}, mgrActor, tNow); err != nil {
		t.Fatal(err)
	}
	made, err := ScanNotifications(f, "ns", obj, dir, tNow, NotifyOptions{})
	if err != nil {
		t.Fatal(err)
	}
	by := map[string][]Notification{}
	for _, n := range made {
		if n.Task != "" {
			by[n.Event+"|"+n.Recipient] = append(by[n.Event+"|"+n.Recipient], n)
		}
	}
	// both tasks were given to Петров; the first is overdue (Петров and the manager hear of it), the second is due in two days
	if len(by["assigned|Петров"]) != 2 || len(by["overdue|Петров"]) != 1 || len(by["overdue|Сидоров"]) != 1 || len(by["due_soon|Петров"]) != 1 {
		t.Fatalf("notifications: %v", len(made))
	}
	if n := by["overdue|Петров"][0]; n.Task != t1.ID || n.Remark != rem || !strings.Contains(n.Title, "Исправить лист 5") {
		t.Errorf("the notification names its task and remark: %+v", n)
	}
	if n := by["due_soon|Петров"][0]; n.Task != t2.ID {
		t.Errorf("due soon: %+v", n)
	}
	// the same pass again makes nothing new
	if again, _ := ScanNotifications(f, "ns", obj, dir, tNow, NotifyOptions{}); len(again) != 0 {
		t.Errorf("each event is announced once: %d", len(again))
	}
	// the assignee hands the task over: the managers hear of it; the specialist accepts: the assignee hears of it
	later := tNow.Add(time.Hour)
	for _, st := range []string{TaskInProgress, TaskReview} {
		if _, err := UpdateTask(f, "ns", obj, t1.ID, TaskPatch{Status: s(st)}, desActor, later); err != nil {
			t.Fatal(err)
		}
	}
	made, _ = ScanNotifications(f, "ns", obj, dir, later, NotifyOptions{})
	if len(made) != 1 || made[0].Recipient != "Сидоров" || made[0].Event != EvStatus || !strings.Contains(made[0].Title, "на проверке") {
		t.Fatalf("a task on the check goes to the manager only: %+v", made)
	}
	if _, err := UpdateTask(f, "ns", obj, t1.ID, TaskPatch{Status: s(TaskDone), StatusComment: s("проверено")}, qcActor, later.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	made, _ = ScanNotifications(f, "ns", obj, dir, later.Add(time.Minute), NotifyOptions{})
	got := map[string]bool{}
	for _, n := range made {
		got[n.Event+"|"+n.Recipient] = true
	}
	if !got["status|Петров"] || got["overdue|Петров"] {
		t.Errorf("the assignee hears of the acceptance; a done task is no longer overdue: %+v", made)
	}
	// the list of the notifications carries the task
	list, _ := ListNotifications(f, "ns", obj, "Петров")
	withTask := 0
	for _, n := range list {
		if n.Task != "" {
			withTask++
		}
	}
	if withTask == 0 {
		t.Error("the notification keeps its task")
	}
}
