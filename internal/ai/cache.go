package ai

import (
	"sync"
	"time"
)

// cacheEntry holds a cached result with expiration time.
type cacheEntry struct {
	value      []byte
	expiration time.Time
}

// Cache provides a thread‑safe in‑memory cache with TTL.
type Cache struct {
	mu    sync.RWMutex
	store map[string]cacheEntry
	ttl   time.Duration
}

// NewCache creates a new cache with a given TTL (e.g., 5 minutes).
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		store: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// Get returns the cached value for a key, or nil if not found or expired.
func (c *Cache) Get(key string) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok {
		return nil
	}
	if time.Now().After(entry.expiration) {
		// expired; we could delete it here, but we'll let Set overwrite.
		return nil
	}
	return entry.value
}

// Set stores a value in the cache with the configured TTL.
func (c *Cache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Clear removes all entries (useful for testing).
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]cacheEntry)
}
