package main

import (
	"bufio"
	"net"
	"strings"
	"testing"

	"github.com/madnikulin50/lowcode/agents/stroykontrol/cmp"
)

// a minimal SMTP server: accepts one letter and hands its envelope and data back
func fakeSMTP(t *testing.T) (addr string, got chan string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got = make(chan string, 1)
	go func() {
		defer ln.Close()
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		r, w := bufio.NewReader(c), bufio.NewWriter(c)
		say := func(s string) { w.WriteString(s + "\r\n"); w.Flush() }
		say("220 fake")
		var b strings.Builder
		data := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			if data {
				if line == ".\r\n" {
					data = false
					say("250 queued")
					continue
				}
				b.WriteString(line)
				continue
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				say("250 fake")
			case strings.HasPrefix(cmd, "MAIL FROM"), strings.HasPrefix(cmd, "RCPT TO"):
				b.WriteString(strings.TrimSpace(line) + "\n")
				say("250 ok")
			case cmd == "DATA":
				data = true
				say("354 go")
			case cmd == "QUIT":
				say("221 bye")
				got <- b.String()
				return
			default:
				say("250 ok")
			}
		}
	}()
	return ln.Addr().String(), got
}

func TestSMTPSender(t *testing.T) {
	addr, got := fakeSMTP(t)
	s := smtpSender{smtpConfig{Addr: addr, From: "Стройконтроль <noreply@example.ru>", TLS: "none", Link: "https://sk.example.ru/"}}
	n := cmp.Notification{Email: "ivanov@example.ru", Title: "Вам назначено замечание № 12", Text: "Замечание № 12: нет расчёта", Object: "o1", Remark: "r 1"}
	if err := s.Send(n); err != nil {
		t.Fatal(err)
	}
	mail := <-got
	for _, want := range []string{"MAIL FROM:<noreply@example.ru>", "RCPT TO:<ivanov@example.ru>", "Subject: =?utf-8?q?", "Замечание № 12: нет расчёта",
		"https://sk.example.ru/?view=remarks&objectID=o1&remark=r+1", "charset=utf-8"} {
		if !strings.Contains(mail, want) {
			t.Errorf("letter lacks %q:\n%s", want, mail)
		}
	}
	if err := (smtpSender{smtpConfig{Addr: addr, From: "x@y"}}).Send(cmp.Notification{}); err == nil {
		t.Error("no address must fail")
	}
}
