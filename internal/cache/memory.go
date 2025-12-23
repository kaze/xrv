package cache

import (
	"context"
	"sync"
	"time"
)

type cacheEntry struct {
	value      []byte
	expiration time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	data  map[string]*cacheEntry
	done  chan struct{}
	wg    sync.WaitGroup
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		data: make(map[string]*cacheEntry),
		done: make(chan struct{}),
	}

	c.wg.Add(1)
	go c.cleanupExpired()

	return c
}

func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		return nil, &ErrCacheMiss{Key: key}
	}

	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		return nil, &ErrCacheMiss{Key: key}
	}

	return entry.value, nil
}

func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &cacheEntry{
		value: value,
	}

	if ttl > 0 {
		entry.expiration = time.Now().Add(ttl)
	}

	c.data[key] = entry
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
	return nil
}

func (c *MemoryCache) Close() error {
	close(c.done)
	c.wg.Wait()
	return nil
}

func (c *MemoryCache) cleanupExpired() {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.done:
			return
		}
	}
}

func (c *MemoryCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if !entry.expiration.IsZero() && now.After(entry.expiration) {
			delete(c.data, key)
		}
	}
}
