package cmp

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

var notifyNow = time.Date(2026, 10, 23, 9, 0, 0, 0, time.UTC)

func evKeys(ns []Notification) []string {
	var out []string
	for _, n := range ns {
		out = append(out, n.Event+"→"+n.Recipient+"→"+n.Number)
	}
	sort.Strings(out)
	return out
}

func TestNotifyEvents(t *testing.T) {
	remarks := []NotifyRemark{
		{ID: "r1", Number: "1", Status: RemarkBeingFix, Assignee: "Иванов И.И.", DueDate: "2026-10-25", Text: "Уточнить армирование"},
		{ID: "r2", Number: "2", Status: RemarkRework, Assignee: "Петров", DueDate: "2026-10-20"},
		{ID: "r3", Number: "3", Status: RemarkResolved, Assignee: "Петров", DueDate: "2026-10-01"},                  // resolved: no deadline notices
		{ID: "r4", Number: "4", Status: RemarkAwaiting, DueDate: "2026-11-30"},                                      // far away, unassigned
		{ID: "r5", Number: "5", Status: RemarkAwaiting, Assignee: "Сидоров", DueDate: "2026-10-01", Archived: true}, // archived
	}
	hist := []NotifyHist{
		{ID: "h1", Remark: "r1", Kind: "", Status: RemarkAwaiting, At: "2026-10-01T10:00:00Z"},
		{ID: "h2", Remark: "r1", Kind: HistAssignee, Status: RemarkAwaiting, Source: "назначен: Иванов И.И.", Author: "специалист ПТО", At: "2026-10-22T10:00:00Z"},
		{ID: "h3", Remark: "r1", Kind: HistStatus, Status: RemarkBeingFix, Author: "специалист ПТО", At: "2026-10-22T11:00:00Z"},
		{ID: "h4", Remark: "r2", Kind: "", Status: RemarkAnswerRev, At: "2026-10-02T10:00:00Z"},
		{ID: "h5", Remark: "r2", Kind: HistStatus, Status: RemarkRework, Note: "расчёт не приложен", At: "2026-10-03T10:00:00Z"}, // too old
		{ID: "h6", Remark: "r2", Kind: HistAssignee, Source: "назначен: Сидоров", At: "2026-10-22T09:00:00Z"},                    // reassigned since
	}
	dir := []Recipient{
		{Name: "Иванов Иван Иванович", Email: "ivanov@example.ru", Role: RoleAssignee},
		{Name: "Иванов И. И.", Email: "ivanov@example.ru", Role: RoleAssignee},
		{Name: "Руководитель", Email: "rp@example.ru", Role: RoleManager},
	}
	known := map[string]bool{}
	got := NotifyEvents("o1", remarks, hist, dir, known, notifyNow, NotifyOptions{Channel: true})
	want := []string{
		"assigned→Иванов И. И.→1",
		"due_soon→Иванов И. И.→1",
		"overdue→Петров→2",
		"overdue→Руководитель→2",
		"status→Иванов И. И.→1",
	}
	if g := evKeys(got); strings.Join(g, ";") != strings.Join(want, ";") {
		t.Fatalf("events:\n got %v\nwant %v", g, want)
	}
	for _, n := range got {
		switch n.Recipient {
		case "Петров":
			if n.Delivery != DeliveryNone || n.Email != "" {
				t.Errorf("no address — the feed only: %+v", n)
			}
		default:
			if n.Delivery != DeliveryPending || n.Email == "" {
				t.Errorf("with an address — to be sent: %+v", n)
			}
		}
		if n.Event == EvDueSoon && !strings.Contains(n.Title, "через 2 дня") {
			t.Errorf("due soon title: %q", n.Title)
		}
		if n.Event == EvStatus && !strings.Contains(n.Text, "Ответ не предоставлен → Устраняется заявителем") {
			t.Errorf("status text: %q", n.Text)
		}
	}
	// the same pass again makes nothing: every notification has its key
	if again := NotifyEvents("o1", remarks, hist, dir, known, notifyNow, NotifyOptions{Channel: true}); len(again) != 0 {
		t.Errorf("repeated: %v", evKeys(again))
	}
	// a new due date is a new deadline
	remarks[1].DueDate = "2026-10-21"
	if again := NotifyEvents("o1", remarks, hist, dir, known, notifyNow, NotifyOptions{}); len(again) != 2 {
		t.Errorf("new due date: %v", evKeys(again))
	}
}

func TestNotifyManagerSubscriptionAndObject(t *testing.T) {
	remarks := []NotifyRemark{{ID: "r1", Number: "7", Status: RemarkResolved}}
	hist := []NotifyHist{
		{ID: "h1", Remark: "r1", Status: RemarkAnswerRev, At: "2026-10-22T08:00:00Z"},
		{ID: "h2", Remark: "r1", Kind: HistStatus, Status: RemarkResolved, At: "2026-10-22T09:00:00Z"},
	}
	dir := []Recipient{
		{Name: "РП 1", Role: RoleManager, Events: []string{EvStatus, EvOverdue}},
		{Name: "РП другого объекта", Role: RoleManager, Object: "o2", Events: []string{EvStatus}},
		{Name: "РП 2", Role: RoleManager}, // default: overdue only
	}
	got := NotifyEvents("o1", remarks, hist, dir, map[string]bool{}, notifyNow, NotifyOptions{})
	if g := evKeys(got); len(g) != 1 || g[0] != "status→РП 1→7" {
		t.Fatalf("got %v", g)
	}
}

func TestDueIn(t *testing.T) {
	for d, want := range map[int]string{0: "сегодня", 1: "завтра", 2: "через 2 дня", 5: "через 5 дней", 21: "через 21 день", 12: "через 12 дней"} {
		if got := dueIn(d); got != want {
			t.Errorf("%d: %q, want %q", d, got, want)
		}
	}
}

type fakeSender struct {
	sent []Notification
	fail bool
}

func (s *fakeSender) Send(n Notification) error {
	if s.fail {
		return errors.New("smtp: 451 try later")
	}
	s.sent = append(s.sent, n)
	return nil
}

func TestScanAndDeliverNotifications(t *testing.T) {
	b := newFake()
	r := b.put(ModExpertise, idcheck.V("object", "o1", "number", "12", "status", RemarkAwaiting, "assignee", "a@example.ru", "due_date", "2026-10-24", "text", "Нет расчёта"))
	b.put(ModRemarkHistory, idcheck.V("remark", r, "status", RemarkAwaiting, "at", "2026-10-01T10:00:00Z"))
	b.put(ModRemarkHistory, idcheck.V("remark", r, "kind", HistAssignee, "status", RemarkAwaiting, "source", "назначен: a@example.ru", "at", "2026-10-23T08:00:00Z"))
	b.put(ModRecipients, idcheck.V("name", "Руководитель", "email", "rp@example.ru", "role", RoleManager, "events", "overdue,due_soon"))
	b.put(ModRecipients, idcheck.V("name", "Уволен", "email", "x@example.ru", "role", RoleManager, "active", "0"))

	dir, err := Recipients(b, "ns")
	if err != nil || len(dir) != 1 {
		t.Fatalf("directory: %+v %v", dir, err)
	}
	made, err := ScanNotifications(b, "ns", "o1", dir, notifyNow, NotifyOptions{Channel: true})
	if err != nil {
		t.Fatal(err)
	}
	if g := evKeys(made); strings.Join(g, ";") != "assigned→a@example.ru→12;due_soon→a@example.ru→12;due_soon→Руководитель→12" {
		t.Fatalf("made %v", g)
	}
	if again, _ := ScanNotifications(b, "ns", "o1", dir, notifyNow, NotifyOptions{Channel: true}); len(again) != 0 {
		t.Fatalf("stored keys are not made again: %v", evKeys(again))
	}

	s := &fakeSender{fail: true}
	if n, _ := DeliverNotifications(b, "ns", "o1", s, notifyNow); n != 0 {
		t.Fatalf("a failure sends nothing")
	}
	s.fail = false
	if n, err := DeliverNotifications(b, "ns", "o1", s, notifyNow); n != 3 || err != nil {
		t.Fatalf("sent %d %v", n, err)
	}
	if n, _ := DeliverNotifications(b, "ns", "o1", s, notifyNow); n != 0 {
		t.Fatalf("sent twice")
	}

	list, err := ListNotifications(b, "ns", "o1", "Руководитель")
	if err != nil || len(list) != 1 || list[0].Delivery != DeliverySent || list[0].Read {
		t.Fatalf("list: %+v %v", list, err)
	}
	if err := MarkNotificationsRead(b, "ns", []string{list[0].ID}); err != nil {
		t.Fatal(err)
	}
	if list, _ = ListNotifications(b, "ns", "o1", "руководитель"); !list[0].Read {
		t.Errorf("not marked read")
	}
}
