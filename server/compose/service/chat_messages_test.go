package service

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestBuildMessagesIncludesCurrentPrompt(t *testing.T) {
	c := &chatService{}
	msgs := c.buildMessages(context.Background(), &ChatPromptArguments{
		Prompt: "Привет, сколько модулей?",
		Messages: []ChatMessage{
			{Role: "user", Content: "раньше"},
			{Role: "assistant", Content: "ответ"},
		},
	}, false)

	var lastUser string
	for _, m := range msgs {
		if m.Role == schema.User {
			lastUser = m.Content
		}
	}
	if lastUser != "Привет, сколько модулей?" {
		t.Fatalf("current prompt missing from messages, last user=%q", lastUser)
	}

	var sawHistory bool
	for _, m := range msgs {
		if m.Role == schema.User && m.Content == "раньше" {
			sawHistory = true
		}
	}
	if !sawHistory {
		t.Fatal("history user message dropped")
	}
}

func TestBuildMessagesAttachesFilesToCurrentPrompt(t *testing.T) {
	c := &chatService{}
	msgs := c.buildMessages(context.Background(), &ChatPromptArguments{
		Prompt: "разбери файл",
		Files:  []ChatFile{{Name: "a.csv", Content: "a,b\n1,2"}},
	}, false)

	if len(msgs) == 0 {
		t.Fatal("no messages")
	}
	last := msgs[len(msgs)-1]
	if last.Role != schema.User {
		t.Fatalf("last role=%s want user", last.Role)
	}
	if !strings.Contains(last.Content, "a.csv") || !strings.Contains(last.Content, "разбери файл") {
		t.Fatalf("files/prompt not on last user message: %q", last.Content)
	}
}

func TestSplitPromptKeepsPlainTextAsUser(t *testing.T) {
	c := &chatService{}
	msgs := c.splitPrompt("просто текст")
	if len(msgs) == 0 {
		t.Fatal("empty split")
	}
	var user string
	for _, m := range msgs {
		if m.Role == schema.User {
			user = m.Content
		}
		if m.Role == schema.System && strings.Contains(m.Content, "просто текст") {
			t.Fatalf("plain text should not become a system message: %#v", m)
		}
	}
	if user != "просто текст" {
		t.Fatalf("user=%q", user)
	}
}
