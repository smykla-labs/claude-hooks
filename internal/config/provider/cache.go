// Package provider provides multi-source configuration loading with precedence.
package provider

import (
	"sync"

	pkgconfig "github.com/smykla-labs/klaudiush/pkg/config"
)

// Cache provides thread-safe caching for configuration.
type Cache struct {
	mu     sync.RWMutex
	config *pkgconfig.Config
}

// NewCache creates a new Cache.
func NewCache() *Cache {
	return &Cache{}
}

// Get retrieves the cached configuration.
// Returns nil if no configuration is cached.
func (c *Cache) Get() *pkgconfig.Config {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config
}

// Set caches the given configuration.
func (c *Cache) Set(cfg *pkgconfig.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = cfg
}

// Clear clears the cached configuration.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = nil
}

// Has checks if a configuration is cached.
func (c *Cache) Has() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.config != nil
}
