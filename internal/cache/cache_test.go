package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	c := New(100, 3600)
	c.Set("k", "v", 10)
	if v, ok := c.Get("k"); !ok || v != "v" {
		t.Fatal("cache miss")
	}
}

func TestGetExpired(t *testing.T) {
	c := New(100, 3600)
	c.Set("k", "v", 0)
	time.Sleep(10 * time.Millisecond)
	if _, ok := c.Get("k"); ok {
		t.Error("expected miss for expired")
	}
}

func TestGetStale(t *testing.T) {
	c := New(100, 3600)
	c.Set("k", "v", 0)
	time.Sleep(10 * time.Millisecond)
	if v, _, ok := c.GetStale("k"); !ok || v != "v" {
		t.Fatal("expected stale hit")
	}
}

func TestRunOnceDedup(t *testing.T) {
	c := New(100, 3600)
	var count int
	var mu sync.Mutex

	compute := func() (any, error) {
		mu.Lock()
		count++
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return "r", nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.RunOnce("x", compute)
		}()
	}
	wg.Wait()

	if count != 1 {
		t.Errorf("expected 1 call, got %d", count)
	}
}

func TestEvictOverflow(t *testing.T) {
	c := New(3, 3600)
	for i := 0; i < 5; i++ {
		c.Set(string(rune('a'+i)), i, 60)
	}
	if c.Entries() != 3 {
		t.Errorf("expected 3, got %d", c.Entries())
	}
}
