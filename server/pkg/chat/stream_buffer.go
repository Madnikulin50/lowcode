package chat

import (
	"strings"
	"sync"
	"time"
)

// StreamBufferInterval is how often buffered token/reason chunks are flushed
// to the SSE client. Status events are not buffered.
const StreamBufferInterval = 2 * time.Second

// BufferStream wraps fn so token and reason chunks are coalesced and emitted
// at most once per interval. A done=true call (or the returned flush) sends
// whatever is still buffered immediately. Status must stay on StatusFunc.
func BufferStream(fn StreamFunc) (buffered StreamFunc, flush func() error) {
	return BufferStreamInterval(fn, StreamBufferInterval)
}

// BufferStreamInterval is BufferStream with a custom flush period (tests).
func BufferStreamInterval(fn StreamFunc, interval time.Duration) (buffered StreamFunc, flush func() error) {
	if fn == nil {
		return func(string, string, bool) error { return nil }, func() error { return nil }
	}
	if interval <= 0 {
		interval = StreamBufferInterval
	}

	b := &streamBuffer{
		fn:     fn,
		ticker: time.NewTicker(interval),
		stop:   make(chan struct{}),
	}
	b.wg.Add(1)
	go b.loop()

	wrapped := func(token, reason string, done bool) error {
		b.mu.Lock()
		if b.closed {
			b.mu.Unlock()
			return nil
		}
		if token != "" {
			b.token.WriteString(token)
		}
		if reason != "" {
			b.reason.WriteString(reason)
		}
		b.mu.Unlock()
		if done {
			return b.close(true)
		}
		return nil
	}

	return wrapped, func() error { return b.close(false) }
}

type streamBuffer struct {
	fn     StreamFunc
	mu     sync.Mutex
	token  strings.Builder
	reason strings.Builder
	closed bool
	ticker *time.Ticker
	stop   chan struct{}
	wg     sync.WaitGroup
	once   sync.Once
}

func (b *streamBuffer) loop() {
	defer b.wg.Done()
	for {
		select {
		case <-b.ticker.C:
			_ = b.flushTick()
		case <-b.stop:
			return
		}
	}
}

func (b *streamBuffer) take() (token, reason string) {
	token = b.token.String()
	reason = b.reason.String()
	b.token.Reset()
	b.reason.Reset()
	return token, reason
}

func (b *streamBuffer) flushTick() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	token, reason := b.take()
	b.mu.Unlock()
	if token == "" && reason == "" {
		return nil
	}
	return b.fn(token, reason, false)
}

func (b *streamBuffer) close(done bool) error {
	var err error
	b.once.Do(func() {
		b.mu.Lock()
		b.closed = true
		b.mu.Unlock()

		b.ticker.Stop()
		close(b.stop)
		b.wg.Wait()

		b.mu.Lock()
		token, reason := b.take()
		b.mu.Unlock()
		if token == "" && reason == "" && !done {
			return
		}
		err = b.fn(token, reason, done)
	})
	return err
}
