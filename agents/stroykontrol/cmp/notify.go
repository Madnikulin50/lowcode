package cmp

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/idcheck"
)

// Уведомления участников (ТЗ 3.7.5): о назначении исполнителя, о смене статуса замечания, о
// приближении срока и о его нарушении. События агент находит сам: назначения и смены статуса — по
// истории замечаний (запись истории — одно событие), сроки — проходом по реестру. Каждое уведомление
// имеет ключ (запись истории, или замечание + срок + вид), поэтому об одном и том же дважды не
// сообщается. Получатель — по справочнику (ФИО исполнителя → адрес; руководители проекта получают
// просрочки и, по подписке, смены статуса). Уведомление пишется в модуль notifications (лента в
// рабочем месте) и, если задан канал доставки, отправляется по нему.

const (
	ModNotifications = "notifications"
	ModRecipients    = "notify_recipients"
)

// Events of a notification.
const (
	EvAssigned = "assigned" // the remark was given to the recipient
	EvStatus   = "status"   // the working status of the remark changed
	EvDueSoon  = "due_soon" // the due date is near
	EvOverdue  = "overdue"  // the due date has passed and the remark is not resolved
)

// EventLabels are the Russian names of the events.
var EventLabels = map[string]string{
	EvAssigned: "назначение", EvStatus: "смена статуса", EvDueSoon: "срок приближается", EvOverdue: "срок нарушен",
}

// Roles of a recipient.
const (
	RoleAssignee = "assignee" // gets what concerns the remarks assigned to them
	RoleManager  = "manager"  // the project manager: overdue remarks of the object, and what else they subscribe to
)

// Delivery states of a notification.
const (
	DeliveryNone    = "none"    // the feed only: no channel, or the recipient has no address
	DeliveryPending = "pending" // to be sent
	DeliverySent    = "sent"
	DeliveryFailed  = "failed" // the last try failed; tried again on the next pass
)

// DefaultEvents are the events a recipient gets when the directory names none.
var DefaultEvents = map[string][]string{
	RoleAssignee: {EvAssigned, EvStatus, EvDueSoon, EvOverdue},
	RoleManager:  {EvOverdue},
}

// Recipient is one row of the directory of the recipients.
type Recipient struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"` // as the assignee is written in a remark
	Email    string   `json:"email,omitempty"`
	Telegram string   `json:"telegram,omitempty"` // chat id
	Role     string   `json:"role"`
	Object   string   `json:"object,omitempty"` // "" — every object
	Events   []string `json:"events"`
}

func (r Recipient) wants(ev string) bool {
	for _, e := range r.Events {
		if e == ev {
			return true
		}
	}
	return false
}

// NotifyRemark is what the scan needs to know of a remark.
type NotifyRemark struct {
	ID, Number, Object string
	Status, Assignee   string
	DueDate            string // YYYY-MM-DD…
	Text               string
	Archived           bool
}

// NotifyHist is one history entry of a remark.
type NotifyHist struct {
	ID, Remark string
	Kind       string
	Status     string
	Source     string
	Note       string
	Author     string
	At         string // RFC 3339
}

// Notification is one message to one recipient.
type Notification struct {
	ID        string `json:"id,omitempty"`
	Key       string `json:"key"`
	Event     string `json:"event"`
	Object    string `json:"object"`
	Remark    string `json:"remark"`
	Number    string `json:"number"`
	Recipient string `json:"recipient"`
	Email     string `json:"email,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
	Title     string `json:"title"`
	Text      string `json:"text"`
	Created   string `json:"created"`
	Read      bool   `json:"read"`
	Delivery  string `json:"delivery"`
	Error     string `json:"error,omitempty"`
}

// NotifyOptions tune a scan.
type NotifyOptions struct {
	SoonDays int           // «срок приближается» this many days before the due date (default 3)
	Lookback time.Duration // history entries older than this are not announced (default 3 days): no flood on the first run
	Channel  bool          // a delivery channel is configured: new notifications with an address are «pending»
}

func (o NotifyOptions) norm() NotifyOptions {
	if o.SoonDays <= 0 {
		o.SoonDays = 3
	}
	if o.Lookback <= 0 {
		o.Lookback = 72 * time.Hour
	}
	return o
}

// NormName folds a name for matching an assignee to the directory: case, ё, spaces and dots.
func NormName(s string) string {
	s = strings.ToLower(strings.ReplaceAll(s, "ё", "е"))
	s = strings.NewReplacer(".", " ", ",", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

type directory struct {
	byName   map[string]Recipient
	managers []Recipient
}

func newDirectory(rs []Recipient, object string) directory {
	d := directory{byName: map[string]Recipient{}}
	for _, r := range rs {
		if r.Object != "" && r.Object != object {
			continue
		}
		if len(r.Events) == 0 {
			r.Events = DefaultEvents[r.Role]
		}
		if r.Role == RoleManager {
			d.managers = append(d.managers, r)
			continue
		}
		d.byName[NormName(r.Name)] = r
		if r.Email != "" {
			d.byName[strings.ToLower(r.Email)] = r
		}
	}
	return d
}

// assignee is the recipient for the remark's assignee: from the directory, else by the name alone (feed only;
// an assignee written as an e-mail address gets mail without a directory row).
func (d directory) assignee(name string) (Recipient, bool) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Recipient{}, false
	}
	if r, ok := d.byName[NormName(name)]; ok {
		return r, true
	}
	if r, ok := d.byName[strings.ToLower(name)]; ok {
		return r, true
	}
	r := Recipient{Name: name, Role: RoleAssignee, Events: DefaultEvents[RoleAssignee]}
	if strings.Contains(name, "@") && !strings.ContainsAny(name, " ,;") {
		r.Email = name
	}
	return r, true
}

// NotifyEvents lists the notifications due now for the remarks of one object. known holds the keys already
// made («key|recipient»); they are not made again.
func NotifyEvents(object string, remarks []NotifyRemark, hist []NotifyHist, dir []Recipient, known map[string]bool, now time.Time, opt NotifyOptions) []Notification {
	opt = opt.norm()
	d := newDirectory(dir, object)
	byID := map[string]NotifyRemark{}
	for _, r := range remarks {
		byID[r.ID] = r
	}
	var out []Notification
	add := func(to Recipient, ev, key string, r NotifyRemark, title, text string) {
		k := key + "|" + NormName(to.Name)
		if known[k] || !to.wants(ev) {
			return
		}
		known[k] = true
		n := Notification{Key: key, Event: ev, Object: object, Remark: r.ID, Number: r.Number, Recipient: to.Name, Email: to.Email, Telegram: to.Telegram,
			Title: title, Text: text, Created: now.UTC().Format(time.RFC3339), Delivery: DeliveryNone}
		if opt.Channel && (to.Email != "" || to.Telegram != "") {
			n.Delivery = DeliveryPending
		}
		out = append(out, n)
	}
	toAssignee := func(r NotifyRemark, ev, key, title, text string) {
		if to, ok := d.assignee(r.Assignee); ok {
			add(to, ev, key, r, title, text)
		}
	}
	toManagers := func(r NotifyRemark, ev, key, title, text string) {
		for _, m := range d.managers {
			add(m, ev, key, r, title, text)
		}
	}
	label := func(r NotifyRemark) string {
		t := strings.TrimSpace(r.Text)
		if rs := []rune(t); len(rs) > 160 {
			t = string(rs[:160]) + "…"
		}
		if t == "" {
			return "Замечание № " + r.Number
		}
		return fmt.Sprintf("Замечание № %s: %s", r.Number, t)
	}

	// the history: in order, per remark; a status that differs from the entry before is a change
	sort.SliceStable(hist, func(i, j int) bool { return hist[i].At < hist[j].At })
	last := map[string]string{}
	since := now.Add(-opt.Lookback).UTC().Format(time.RFC3339)
	for _, h := range hist {
		prev, seen := last[h.Remark]
		if h.Status != "" {
			last[h.Remark] = h.Status
		}
		r, ok := byID[h.Remark]
		if !ok || r.Archived || h.At < since {
			continue
		}
		key := "hist:" + h.ID
		who := ""
		if h.Author != "" {
			who = " (" + h.Author + ")"
		}
		switch {
		case h.Kind == HistAssignee && strings.HasPrefix(h.Source, "назначен: "):
			name := strings.TrimPrefix(h.Source, "назначен: ")
			if strings.TrimSpace(name) != strings.TrimSpace(r.Assignee) {
				continue // given to someone else since
			}
			due := ""
			if len(r.DueDate) >= 10 {
				due = ", срок " + r.DueDate[:10]
			}
			toAssignee(r, EvAssigned, key, "Вам назначено замечание № "+r.Number, label(r)+"\nНазначено"+who+due+".")
		case seen && h.Status != "" && h.Status != prev:
			title := fmt.Sprintf("Замечание № %s: %s", r.Number, statusLabel(h.Status))
			text := fmt.Sprintf("%s\nСтатус: %s → %s%s.", label(r), statusLabel(prev), statusLabel(h.Status), who)
			if n := strings.TrimSpace(h.Note); n != "" {
				text += "\nКомментарий: " + n
			}
			toAssignee(r, EvStatus, key, title, text)
			toManagers(r, EvStatus, key, title, text)
		}
	}

	// the due dates
	today := now.UTC().Format("2006-01-02")
	soon := now.UTC().AddDate(0, 0, opt.SoonDays).Format("2006-01-02")
	for _, r := range remarks {
		if r.Archived || r.Status == RemarkResolved || len(r.DueDate) < 10 {
			continue
		}
		due := r.DueDate[:10]
		switch {
		case due < today:
			key := "overdue:" + r.ID + ":" + due
			title := fmt.Sprintf("Просрочено замечание № %s", r.Number)
			who := strings.TrimSpace(r.Assignee)
			if who == "" {
				who = "не назначен"
			}
			text := fmt.Sprintf("%s\nСрок %s нарушен, статус: %s, исполнитель: %s.", label(r), due, statusLabel(r.Status), who)
			toAssignee(r, EvOverdue, key, title, text)
			toManagers(r, EvOverdue, key, title, text)
		case due <= soon:
			key := "due_soon:" + r.ID + ":" + due
			days := daysBetween(today, due)
			title := fmt.Sprintf("Срок по замечанию № %s — %s", r.Number, dueIn(days))
			text := fmt.Sprintf("%s\nСрок: %s (%s), статус: %s.", label(r), due, dueIn(days), statusLabel(r.Status))
			toAssignee(r, EvDueSoon, key, title, text)
			toManagers(r, EvDueSoon, key, title, text)
		}
	}
	return out
}

func daysBetween(a, b string) int {
	ta, err1 := time.Parse("2006-01-02", a)
	tb, err2 := time.Parse("2006-01-02", b)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(tb.Sub(ta).Hours() / 24)
}

func dueIn(days int) string {
	switch {
	case days <= 0:
		return "сегодня"
	case days == 1:
		return "завтра"
	case days%10 == 1 && days%100 != 11:
		return fmt.Sprintf("через %d день", days)
	case days%10 >= 2 && days%10 <= 4 && (days%100 < 12 || days%100 > 14):
		return fmt.Sprintf("через %d дня", days)
	}
	return fmt.Sprintf("через %d дней", days)
}

// ---- persistence (Compose) ----

func recipientFrom(r idcheck.Record) Recipient {
	var ev []string
	for _, e := range r.Values["events"] {
		for _, x := range strings.Split(e, ",") {
			if x = strings.TrimSpace(x); x != "" {
				ev = append(ev, x)
			}
		}
	}
	role := r.Get("role")
	if role == "" {
		role = RoleAssignee
	}
	return Recipient{ID: r.ID, Name: r.Get("name"), Email: strings.TrimSpace(r.Get("email")), Telegram: strings.TrimSpace(r.Get("telegram")),
		Role: role, Object: r.Get("object"), Events: ev}
}

// Recipients reads the directory (inactive rows left out).
func Recipients(b Backend, ns string) ([]Recipient, error) {
	recs, err := b.Search(ns, ModRecipients, "")
	if err != nil {
		return nil, err
	}
	var out []Recipient
	for _, r := range recs {
		if a := r.Get("active"); a == "0" || a == "false" {
			continue
		}
		out = append(out, recipientFrom(r))
	}
	return out, nil
}

func notificationFrom(r idcheck.Record) Notification {
	return Notification{ID: r.ID, Key: r.Get("key"), Event: r.Get("event"), Object: r.Get("object"), Remark: r.Get("remark"), Number: r.Get("number"),
		Recipient: r.Get("recipient"), Email: r.Get("email"), Telegram: r.Get("telegram"), Title: r.Get("title"), Text: r.Get("text"),
		Created: r.Get("created_at"), Read: r.Get("read") == "1" || r.Get("read") == "true", Delivery: r.Get("delivery"), Error: r.Get("error")}
}

// ScanNotifications finds the new notifications of an object and stores them; it returns those it made.
func ScanNotifications(b Backend, ns, objectID string, dir []Recipient, now time.Time, opt NotifyOptions) ([]Notification, error) {
	recs, err := b.Search(ns, ModExpertise, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	var remarks []NotifyRemark
	var hist []NotifyHist
	for _, r := range recs {
		remarks = append(remarks, NotifyRemark{ID: r.ID, Number: r.Get("number"), Object: objectID, Status: r.Get("status"), Assignee: r.Get("assignee"),
			DueDate: r.Get("due_date"), Text: r.Get("text"), Archived: r.Get("archived") != ""})
		hs, err := b.Search(ns, ModRemarkHistory, fmt.Sprintf("remark = '%s'", r.ID))
		if err != nil {
			return nil, err
		}
		for _, h := range hs {
			hist = append(hist, NotifyHist{ID: h.ID, Remark: r.ID, Kind: h.Get("kind"), Status: h.Get("status"), Source: h.Get("source"),
				Note: h.Get("note"), Author: h.Get("author"), At: h.Get("at")})
		}
	}
	old, err := b.Search(ns, ModNotifications, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	known := map[string]bool{}
	for _, o := range old {
		known[o.Get("key")+"|"+NormName(o.Get("recipient"))] = true
	}
	made := NotifyEvents(objectID, remarks, hist, dir, known, now, opt)
	for i, n := range made {
		v := idcheck.V("object", objectID, "remark", n.Remark, "number", n.Number, "key", n.Key, "event", n.Event, "recipient", n.Recipient,
			"email", n.Email, "telegram", n.Telegram, "title", n.Title, "text", n.Text, "created_at", n.Created, "delivery", n.Delivery)
		v["read"] = []string{"0"}
		id, err := b.Create(ns, ModNotifications, v)
		if err != nil {
			return made[:i], err
		}
		made[i].ID = id
	}
	return made, nil
}

// ListNotifications returns the notifications of an object, newest first; for one recipient when named.
func ListNotifications(b Backend, ns, objectID, recipient string) ([]Notification, error) {
	recs, err := b.Search(ns, ModNotifications, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return nil, err
	}
	out := []Notification{}
	who := NormName(recipient)
	for _, r := range recs {
		n := notificationFrom(r)
		if who != "" && NormName(n.Recipient) != who && !strings.EqualFold(n.Email, recipient) {
			continue
		}
		out = append(out, n)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Created > out[j].Created })
	return out, nil
}

// MarkNotificationsRead marks notifications as read.
func MarkNotificationsRead(b Backend, ns string, ids []string) error {
	for _, id := range ids {
		if err := b.Update(ns, ModNotifications, id, idcheck.Values{"read": {"1"}}); err != nil {
			return err
		}
	}
	return nil
}

// Sender delivers one notification over a channel (e-mail, messenger).
type Sender interface {
	Send(n Notification) error
}

// DeliverNotifications sends the pending (and earlier failed) notifications of an object and records the
// outcome. It returns how many were sent.
func DeliverNotifications(b Backend, ns, objectID string, s Sender, now time.Time) (int, error) {
	if s == nil {
		return 0, nil
	}
	recs, err := b.Search(ns, ModNotifications, fmt.Sprintf("object = '%s'", objectID))
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, r := range recs {
		n := notificationFrom(r)
		if n.Delivery != DeliveryPending && n.Delivery != DeliveryFailed {
			continue
		}
		tries, _ := strconv.Atoi(r.Get("tries"))
		if tries >= maxDeliveryTries {
			continue
		}
		v := idcheck.Values{"tries": {strconv.Itoa(tries + 1)}}
		if err := s.Send(n); err != nil {
			v["delivery"], v["error"] = []string{DeliveryFailed}, []string{trunc(err.Error(), 300)}
		} else {
			v["delivery"], v["error"], v["delivered_at"] = []string{DeliverySent}, []string{""}, []string{now.UTC().Format(time.RFC3339)}
			sent++
		}
		if err := b.Update(ns, ModNotifications, n.ID, v); err != nil {
			return sent, err
		}
	}
	return sent, nil
}

const maxDeliveryTries = 5
