package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kouji/nvpm/pkg/core/config"
)

// entry represents a single cached value
type entry struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// expired reports whether the entry has passed its TTL
func (e *entry) expired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// Cache provides key-value storage with TTL and disk persistence
type Cache struct {
	Config *config.Config

	mu      sync.Mutex
	entries map[string]*entry
}

// Stats contains cache statistics
type Stats struct {
	Entries   int
	TotalSize int64
}

// NewCache creates a new cache instance
func NewCache(cfg *config.Config) *Cache {
	return &Cache{
		Config:  cfg,
		entries: make(map[string]*entry),
	}
}

// Enable creates the cache directory (if configured) and loads persisted entries
func (c *Cache) Enable() error {
	if !c.Config.Performance.Cache.Enabled {
		return nil
	}

	if err := os.MkdirAll(c.Config.Performance.Cache.Path, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return c.load()
}

// Set stores a value under key with the given time-to-live
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	c.entries[key] = &entry{
		Data:      data,
		ExpiresAt: expiresAt,
	}

	c.save()
}

// Get retrieves a value by key. The second return value reports whether
// the key was found and has not expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if e.expired() {
		delete(c.entries, key)
		return nil, false
	}

	return e.Data, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	c.save()
}

// Cleanup removes all expired entries
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, e := range c.entries {
		if e.expired() {
			delete(c.entries, key)
		}
	}

	c.save()
}

// GetStats returns statistics about the cache
func (c *Cache) GetStats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalSize int64
	for _, e := range c.entries {
		if b, err := json.Marshal(e.Data); err == nil {
			totalSize += int64(len(b))
		}
	}

	return Stats{
		Entries:   len(c.entries),
		TotalSize: totalSize,
	}
}

// filePath returns the path to the on-disk cache file
func (c *Cache) filePath() string {
	return filepath.Join(c.Config.Performance.Cache.Path, "cache.json")
}

// load reads persisted entries from disk, discarding any that have expired
func (c *Cache) load() error {
	data, err := os.ReadFile(c.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var entries map[string]*entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to parse cache file: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, e := range entries {
		if !e.expired() {
			c.entries[key] = e
		}
	}

	return nil
}

// save persists entries to disk. Caller must hold c.mu.
func (c *Cache) save() {
	if !c.Config.Performance.Cache.Enabled {
		return
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(c.filePath(), data, 0644)
}
