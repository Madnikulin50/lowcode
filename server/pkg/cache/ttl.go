package cache

import (
	"sync"
	"time"
)

type (
	entry[V any] struct {
		val       V
		expiresAt time.Time
	}

	Cache[K comparable, V any] struct {
		mu      sync.RWMutex
		entries map[K]*entry[V]
		ttl     time.Duration
		stop    chan struct{}
	}
)

func New[K comparable, V any](ttl, cleanupInterval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		entries: make(map[K]*entry[V]),
		ttl:     ttl,
		stop:    make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go c.cleanupLoop(cleanupInterval)
	}

	return c
}

func (c *Cache[K, V]) Stop() {
	close(c.stop)
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	if time.Now().After(e.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		var zero V
		return zero, false
	}

	return e.val, true
}

func (c *Cache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	c.entries[key] = &entry[V]{
		val:       val,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}

func (c *Cache[K, V]) Purge() {
	c.mu.Lock()
	c.entries = make(map[K]*entry[V])
	c.mu.Unlock()
}

func (c *Cache[K, V]) DeleteBy(fn func(K) bool) {
	c.mu.Lock()
	for k := range c.entries {
		if fn(k) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stop:
			return
		}
	}
}

func (c *Cache[K, V]) deleteExpired() {
	now := time.Now()
	c.mu.Lock()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}
