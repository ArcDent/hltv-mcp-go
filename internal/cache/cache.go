package cache

import (
	"sync"
	"time"
)

// StaleMeta carries metadata about a stale cache hit
type StaleMeta struct {
	StaleAgeSec int
}

// Cache is a thread-safe in-memory cache with TTL, stale window, and concurrent request merging
type Cache struct {
	mu         sync.RWMutex
	store      map[string]*entry
	inFlight   map[string]*inflightEntry
	maxEntries int
	maxStale   time.Duration
}

type entry struct {
	value     any
	createdAt time.Time
	expiresAt time.Time
}

type inflightEntry struct {
	ch  chan struct{}
	val any
	err error
}

// New creates a new Cache with the given maximum entries and stale window in seconds
func New(maxEntries, maxStaleSec int) *Cache {
	return &Cache{
		store:      make(map[string]*entry),
		inFlight:   make(map[string]*inflightEntry),
		maxEntries: maxEntries,
		maxStale:   time.Duration(maxStaleSec) * time.Second,
	}
}

// Get returns the cached value if it exists and hasn't expired
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// GetStale returns the cached value even if expired, as long as it's within the stale window
func (c *Cache) GetStale(key string) (any, StaleMeta, bool) {
	c.mu.RLock()
	e, ok := c.store[key]
	c.mu.RUnlock()
	if !ok {
		return nil, StaleMeta{}, false
	}
	now := time.Now()
	if now.Sub(e.expiresAt) > c.maxStale {
		c.mu.Lock()
		delete(c.store, key)
		c.mu.Unlock()
		return nil, StaleMeta{}, false
	}
	return e.value, StaleMeta{StaleAgeSec: int(now.Sub(e.expiresAt).Seconds())}, true
}

// Set stores a value in the cache with the given TTL in seconds
func (c *Cache) Set(key string, value any, ttlSec int) {
	ttl := time.Duration(ttlSec) * time.Second
	if ttlSec <= 0 {
		ttl = 0
	}
	now := time.Now()

	c.mu.Lock()
	c.store[key] = &entry{
		value:     value,
		createdAt: now,
		expiresAt: now.Add(ttl),
	}
	// FIFO eviction when over capacity
	for len(c.store) > c.maxEntries {
		var oldestK string
		var oldestT time.Time
		for k, e := range c.store {
			if oldestK == "" || e.createdAt.Before(oldestT) {
				oldestK, oldestT = k, e.createdAt
			}
		}
		delete(c.store, oldestK)
	}
	c.mu.Unlock()
}

// RunOnce executes compute only once per key, merging concurrent callers onto the same inflight computation
func (c *Cache) RunOnce(key string, compute func() (any, error)) (any, error) {
	c.mu.Lock()
	if inf, ok := c.inFlight[key]; ok {
		c.mu.Unlock()
		<-inf.ch
		return inf.val, inf.err
	}
	inf := &inflightEntry{ch: make(chan struct{})}
	c.inFlight[key] = inf
	c.mu.Unlock()

	inf.val, inf.err = compute()
	close(inf.ch)

	c.mu.Lock()
	delete(c.inFlight, key)
	c.mu.Unlock()

	return inf.val, inf.err
}

// Clear removes all cached entries
func (c *Cache) Clear() {
	c.mu.Lock()
	c.store = make(map[string]*entry)
	c.mu.Unlock()
}

// Entries returns the current number of cached entries
func (c *Cache) Entries() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}
