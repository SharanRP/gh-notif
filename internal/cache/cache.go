package cache

import (
	"fmt"
	"time"
)

// Cache is the interface for cache implementations
type Cache interface {
	// Get retrieves a value from the cache
	Get(key string) (interface{}, bool)

	// Set adds a value to the cache with the given TTL
	Set(key string, value interface{}, ttl time.Duration)

	// Delete removes a value from the cache
	Delete(key string)

	// Clear removes all items from the cache
	Clear()

	// Prefetch queues a key for background prefetching
	Prefetch(key string)

	// Close closes the cache
	Close() error

	// GetMetrics returns cache metrics
	GetMetrics() *Metrics
}

// CacheType represents the type of cache
type CacheType string

const (
	// MemoryCacheType is an in-memory cache
	MemoryCacheType CacheType = "memory"

	// BadgerCacheType is a BadgerDB-backed cache
	BadgerCacheType CacheType = "badger"

	// BoltCacheType is a BoltDB-backed cache
	BoltCacheType CacheType = "bolt"

	// NullCacheType is a no-op cache
	NullCacheType CacheType = "null"
)

// NewCache creates a new cache of the specified type
func NewCache(cacheType CacheType, opts *Options) (Cache, error) {
	switch cacheType {
	case MemoryCacheType:
		return NewMemoryCache(opts), nil
	case BadgerCacheType:
		return NewBadgerCache(opts)
	case BoltCacheType:
		return NewBoltCache(opts)
	case NullCacheType:
		return NewNullCache(), nil
	default:
		return nil, fmt.Errorf("unsupported cache type: %s", cacheType)
	}
}

// nullCache is a no-op cache implementation
type nullCache struct{}

// NewNullCache creates a new null cache
func NewNullCache() Cache {
	return &nullCache{}
}

// Get always returns not found
func (c *nullCache) Get(key string) (interface{}, bool) {
	return nil, false
}

// Set does nothing
func (c *nullCache) Set(key string, value interface{}, ttl time.Duration) {
	// No-op
}

// Delete does nothing
func (c *nullCache) Delete(key string) {
	// No-op
}

// Clear does nothing
func (c *nullCache) Clear() {
	// No-op
}

// Prefetch does nothing
func (c *nullCache) Prefetch(key string) {
	// No-op
}

// Close does nothing
func (c *nullCache) Close() error {
	return nil
}

// GetMetrics returns empty metrics
func (c *nullCache) GetMetrics() *Metrics {
	return NewMetrics()
}

// memoryCache is an in-memory cache implementation
type memoryCache struct {
	items   map[string]*cacheItem
	options *Options
	metrics *Metrics
}

// cacheItem represents a cached item
type cacheItem struct {
	Value      interface{}
	Expiration time.Time
	Size       int64
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(opts *Options) Cache {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &memoryCache{
		items:   make(map[string]*cacheItem),
		options: opts,
		metrics: NewMetrics(),
	}
}

// Get retrieves a value from the cache
func (c *memoryCache) Get(key string) (interface{}, bool) {
	c.metrics.Inc(&c.metrics.Gets)

	item, found := c.items[key]
	if !found {
		c.metrics.Inc(&c.metrics.Misses)
		return nil, false
	}

	// Check if the item has expired
	if time.Now().After(item.Expiration) {
		delete(c.items, key)
		c.metrics.Inc(&c.metrics.Misses)
		return nil, false
	}

	c.metrics.Inc(&c.metrics.Hits)
	return item.Value, true
}

// Set adds a value to the cache with the given TTL
func (c *memoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.metrics.Inc(&c.metrics.Sets)

	c.items[key] = &cacheItem{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Delete removes a value from the cache
func (c *memoryCache) Delete(key string) {
	c.metrics.Inc(&c.metrics.Deletes)
	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *memoryCache) Clear() {
	c.metrics.Inc(&c.metrics.Clears)
	c.items = make(map[string]*cacheItem)
}

// Prefetch does nothing for memory cache
func (c *memoryCache) Prefetch(key string) {
	// No-op for memory cache
}

// Close does nothing for memory cache
func (c *memoryCache) Close() error {
	return nil
}

// GetMetrics returns cache metrics
func (c *memoryCache) GetMetrics() *Metrics {
	return c.metrics
}
