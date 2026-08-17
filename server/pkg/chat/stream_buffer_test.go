package chat

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBufferStreamCoalescesUntilInterval(t *testing.T) {
	var mu sync.Mutex
	var events [][2]string
	fn := func(token, reason string, done bool) error {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, [2]string{token, reason})
		if done {
			events[len(events)-1][0] += "|done"
		}
		return nil
	}

	buffered, flush := BufferStreamInterval(fn, 40*time.Millisecond)
	_ = buffered("hel", "", false)
	_ = buffered("lo", "", false)
	_ = buffered("", "think", false)
	_ = buffered("", "ing", false)

	mu.Lock()
	if len(events) != 0 {
		t.Fatalf("expected no events before interval, got %v", events)
	}
	mu.Unlock()

	time.Sleep(60 * time.Millisecond)

	mu.Lock()
	if len(events) != 1 {
		t.Fatalf("expected 1 coalesced event, got %d: %v", len(events), events)
	}
	if events[0][0] != "hello" || events[0][1] != "thinking" {
		t.Fatalf("got token=%q reason=%q", events[0][0], events[0][1])
	}
	mu.Unlock()

	if err := flush(); err != nil {
		t.Fatal(err)
	}
}

func TestBufferStreamDoneFlushesImmediately(t *testing.T) {
	var gotToken, gotReason string
	var gotDone bool
	fn := func(token, reason string, done bool) error {
		gotToken += token
		gotReason += reason
		gotDone = done
		return nil
	}

	buffered, flush := BufferStreamInterval(fn, time.Hour)
	_ = buffered("abc", "xyz", false)
	if err := buffered("", "", true); err != nil {
		t.Fatal(err)
	}
	if gotToken != "abc" || gotReason != "xyz" || !gotDone {
		t.Fatalf("token=%q reason=%q done=%v", gotToken, gotReason, gotDone)
	}
	if err := flush(); err != nil {
		t.Fatal(err)
	}
}

func TestBufferStreamFlushWithoutDone(t *testing.T) {
	var mu sync.Mutex
	var tokens []string
	var anyDone bool
	fn := func(token, reason string, done bool) error {
		mu.Lock()
		defer mu.Unlock()
		if token != "" {
			tokens = append(tokens, token)
		}
		if done {
			anyDone = true
		}
		return nil
	}

	buffered, flush := BufferStreamInterval(fn, time.Hour)
	_ = buffered("partial", "", false)
	if err := flush(); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(tokens, "") != "partial" {
		t.Fatalf("tokens=%v", tokens)
	}
	if anyDone {
		t.Fatal("flush without done must not emit done=true")
	}
}
