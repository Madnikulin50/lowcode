package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// Уведомления (ТЗ 3.7.5): агент периодически проходит по объектам, находит новые события (назначение,
// смена статуса, срок близко, срок нарушен — cmp/notify.go), пишет их в ленту (модуль notifications) и
// отправляет письма через SMTP, если он настроен. Смена статуса или исполнителя вручную запускает проход
// по объекту сразу. В режиме --fixtures лента считается на лету по реестру и истории в памяти, писем нет.

// smtpConfig is the mail channel; Addr == "" — no mail, the feed only.
type smtpConfig struct {
	Addr     string // host:port
	User     string
	Password string
	From     string
	TLS      string // starttls (default) | tls | none
	Link     string // base URL of the agent's UI, for the link to the remark in a letter
}

func (c smtpConfig) enabled() bool { return c.Addr != "" && c.From != "" }

// smtpSender sends a notification as a plain-text letter.
type smtpSender struct{ cfg smtpConfig }

func (s smtpSender) Send(n cmp.Notification) error {
	if n.Email == "" {
		return fmt.Errorf("у получателя нет адреса")
	}
	body := n.Text
	if s.cfg.Link != "" {
		body += fmt.Sprintf("\n\nОткрыть: %s/?view=remarks&objectID=%s&remark=%s", strings.TrimRight(s.cfg.Link, "/"), url.QueryEscape(n.Object), url.QueryEscape(n.Remark))
	}
	body += "\n\n— Стройконтроль. Письмо отправлено автоматически, отвечать на него не нужно."
	msg := buildMail(s.cfg.From, n.Email, n.Title, body)
	return smtpSend(s.cfg, n.Email, msg)
}

func buildMail(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))
	b.WriteString("\r\n")
	return []byte(b.String())
}

func smtpSend(cfg smtpConfig, to string, msg []byte) error {
	host, _, err := net.SplitHostPort(cfg.Addr)
	if err != nil {
		return err
	}
	var c *smtp.Client
	if cfg.TLS == "tls" {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 20 * time.Second}, "tcp", cfg.Addr, &tls.Config{ServerName: host})
		if err != nil {
			return err
		}
		if c, err = smtp.NewClient(conn, host); err != nil {
			return err
		}
	} else {
		conn, err := net.DialTimeout("tcp", cfg.Addr, 20*time.Second)
		if err != nil {
			return err
		}
		if c, err = smtp.NewClient(conn, host); err != nil {
			return err
		}
		if cfg.TLS != "none" {
			if ok, _ := c.Extension("STARTTLS"); ok {
				if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
					c.Close()
					return err
				}
			}
		}
	}
	defer c.Close()
	if cfg.User != "" {
		if err := c.Auth(smtp.PlainAuth("", cfg.User, cfg.Password, host)); err != nil {
			return err
		}
	}
	if err := c.Mail(addrOnly(cfg.From)); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// addrOnly takes the address out of «Имя <a@b>».
func addrOnly(s string) string {
	if i := strings.LastIndex(s, "<"); i >= 0 {
		if j := strings.LastIndex(s, ">"); j > i {
			return s[i+1 : j]
		}
	}
	return strings.TrimSpace(s)
}

// notifier runs the passes.
type notifier struct {
	b      cmp.Backend
	cfg    smtpConfig
	opt    cmp.NotifyOptions
	mu     sync.Mutex // one pass at a time
	lastAt time.Time
	lastNo int
	lastEr string
}

var notif *notifier

func (n *notifier) sender() cmp.Sender {
	if !n.cfg.enabled() {
		return nil
	}
	return smtpSender{n.cfg}
}

// scanObject makes the new notifications of one object and sends what is pending.
func (n *notifier) scanObject(ns, objectID string) (int, int, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	dir, err := cmp.Recipients(n.b, ns)
	if err != nil {
		return 0, 0, err
	}
	now := time.Now()
	made, err := cmp.ScanNotifications(n.b, ns, objectID, dir, now, n.opt)
	if err != nil {
		return len(made), 0, err
	}
	sent, err := cmp.DeliverNotifications(n.b, ns, objectID, n.sender(), now)
	return len(made), sent, err
}

// pass goes over every object of the namespace.
func (n *notifier) pass(ns string) {
	objs, err := n.b.Search(ns, cmp.ModObjects, "")
	if err != nil {
		log.Printf("notify: objects: %v", err)
		return
	}
	total := 0
	var errs []string
	for _, o := range objs {
		if o.Get("data_kind") == "demo" {
			continue
		}
		made, sent, err := n.scanObject(ns, o.ID)
		total += made
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", o.ID, err))
		}
		if made+sent > 0 {
			log.Printf("notify: object %s — new %d, sent %d", o.ID, made, sent)
		}
	}
	n.mu.Lock()
	n.lastAt, n.lastNo, n.lastEr = time.Now(), total, strings.Join(errs, "; ")
	n.mu.Unlock()
	if len(errs) > 0 {
		log.Printf("notify: %s", n.lastEr)
	}
}

// run passes over the namespace every interval (the first pass a minute after the start).
func (n *notifier) run(ns string, every time.Duration) {
	if ns == "" || every <= 0 {
		return
	}
	time.Sleep(time.Minute)
	for {
		n.pass(ns)
		time.Sleep(every)
	}
}

// soon starts a pass over one object in the background (after a hand-made change).
func (n *notifier) soon(ns, objectID string) {
	if n == nil || ns == "" || objectID == "" {
		return
	}
	go func() {
		if _, _, err := n.scanObject(ns, objectID); err != nil {
			log.Printf("notify: object %s: %v", objectID, err)
		}
	}()
}

type notificationsOut struct {
	Items   []cmp.Notification `json:"items"`
	Unread  int                `json:"unread"`
	Mail    bool               `json:"mail"`    // the mail channel is configured
	Fixture bool               `json:"fixture"` // computed on the fly, nothing stored
}

// GET /api/notifications?objectID=&recipient=
func handleNotifications(w http.ResponseWriter, r *http.Request) {
	ns, objectID, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	who := strings.TrimSpace(r.URL.Query().Get("recipient"))
	out := notificationsOut{Items: []cmp.Notification{}}
	if fx, ok := compose.(*FixtureStore); ok {
		items, err := fx.Notifications(objectID, who)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err)
			return
		}
		out.Items, out.Fixture = items, true
	} else if notif != nil {
		items, err := cmp.ListNotifications(notif.b, ns, objectID, who)
		if err != nil {
			writeErr(w, http.StatusBadGateway, err)
			return
		}
		out.Items, out.Mail = items, notif.cfg.enabled()
	}
	for _, it := range out.Items {
		if !it.Read {
			out.Unread++
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /api/notifications/read {namespaceID, ids}
func handleNotificationsRead(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NamespaceID string   `json:"namespaceID"`
		IDs         []string `json:"ids"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if fx, ok := compose.(*FixtureStore); ok {
		fx.MarkNotificationsRead(body.IDs)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if notif == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("уведомления недоступны"))
		return
	}
	if err := cmp.MarkNotificationsRead(notif.b, firstNonEmpty(body.NamespaceID, defaultNamespaceID), body.IDs); err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/notifications/scan?objectID= — a pass over one object now (the button and the rule chain).
func handleNotificationsScan(w http.ResponseWriter, r *http.Request) {
	ns, objectID, err := objectFromQuery(r)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if _, ok := compose.(*FixtureStore); ok {
		writeJSON(w, http.StatusOK, map[string]any{"made": 0, "sent": 0, "fixture": true})
		return
	}
	if notif == nil {
		writeErr(w, http.StatusServiceUnavailable, fmt.Errorf("уведомления недоступны"))
		return
	}
	made, sent, err := notif.scanObject(ns, objectID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"made": made, "sent": sent, "mail": notif.cfg.enabled()})
}
