package cache

import (
	"context"
	"sync"
	"time"
)

// Manager manages the cache with advanced features
type Manager struct {
	// Cache is the underlying cache implementation
	Cache Cache

	// Options are the cache manager options
	Options *ManagerOptions

	// prefetchQueue is the queue for prefetch operations
	prefetchQueue chan PrefetchRequest

	// invalidationPatterns are patterns for cache invalidation
	invalidationPatterns []InvalidationPattern

	// ctx is the context for background operations
	ctx context.Context

	// cancel is the function to cancel the context
	cancel context.CancelFunc

	// wg is used to wait for background operations
	wg sync.WaitGroup
}

// ManagerOptions configures the cache manager
type ManagerOptions struct {
	// PrefetchConcurrency is the number of concurrent prefetch operations
	PrefetchConcurrency int

	// PrefetchQueueSize is the size of the prefetch queue
	PrefetchQueueSize int

	// EnablePrefetching enables background prefetching
	EnablePrefetching bool

	// EnableInvalidation enables cache invalidation
	EnableInvalidation bool

	// DefaultTTL is the default time-to-live for cached items
	DefaultTTL time.Duration

	// RefreshBeforeExpiry is the duration before expiry to refresh items
	RefreshBeforeExpiry time.Duration

	// RefreshCallback is called to refresh a cached item
	RefreshCallback func(ctx context.Context, key string) (interface{}, error)
}

// DefaultManagerOptions returns the default manager options
func DefaultManagerOptions() *ManagerOptions {
	return &ManagerOptions{
		PrefetchConcurrency: 2,
		PrefetchQueueSize:   100,
		EnablePrefetching:   true,
		EnableInvalidation:  true,
		DefaultTTL:          1 * time.Hour,
		RefreshBeforeExpiry: 5 * time.Minute,
	}
}

// PrefetchRequest represents a request to prefetch a cache item
type PrefetchRequest struct {
	// Key is the cache key
	Key string

	// Priority is the priority of the request (higher = more important)
	Priority int

	// Callback is called to fetch the item if not in cache
	Callback func(ctx context.Context) (interface{}, error)
}

// InvalidationPattern represents a pattern for cache invalidation
type InvalidationPattern struct {
	// Pattern is the pattern to match cache keys
	Pattern string

	// Action is the action to take when the pattern matches
	Action InvalidationAction

	// Condition is the condition for invalidation
	Condition InvalidationCondition
}

// InvalidationAction is the action to take when invalidating cache items
type InvalidationAction string

const (
	// InvalidateDelete deletes matching items
	InvalidateDelete InvalidationAction = "delete"

	// InvalidateRefresh refreshes matching items
	InvalidateRefresh InvalidationAction = "refresh"

	// InvalidateExpire expires matching items
	InvalidateExpire InvalidationAction = "expire"
)

// InvalidationCondition is the condition for invalidation
type InvalidationCondition string

const (
	// OnWrite invalidates on write operations
	OnWrite InvalidationCondition = "write"

	// OnRead invalidates on read operations
	OnRead InvalidationCondition = "read"

	// OnTime invalidates based on time
	OnTime InvalidationCondition = "time"
)

// NewManager creates a new cache manager
func NewManager(cache Cache, opts *ManagerOptions) *Manager {
	if opts == nil {
		opts = DefaultManagerOptions()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		Cache:   cache,
		Options: opts,
		ctx:     ctx,
		cancel:  cancel,
	}

	if opts.EnablePrefetching {
		manager.prefetchQueue = make(chan PrefetchRequest, opts.PrefetchQueueSize)
		for i := 0; i < opts.PrefetchConcurrency; i++ {
			manager.wg.Add(1)
			go manager.prefetchWorker(ctx)
		}
	}

	return manager
}

// Close closes the cache manager
func (m *Manager) Close() error {
	m.cancel()
	m.wg.Wait()
	return m.Cache.Close()
}

// Get retrieves a value from the cache
func (m *Manager) Get(key string) (interface{}, bool) {
	value, found := m.Cache.Get(key)

	// If not found and prefetching is enabled, try to prefetch
	if !found && m.Options.EnablePrefetching {
		m.Cache.Prefetch(key)
	}

	return value, found
}

// Set adds a value to the cache with the given TTL
func (m *Manager) Set(key string, value interface{}, ttl time.Duration) {
	if ttl <= 0 {
		ttl = m.Options.DefaultTTL
	}

	m.Cache.Set(key, value, ttl)

	// Apply invalidation patterns
	if m.Options.EnableInvalidation {
		m.applyInvalidationPatterns(key, OnWrite)
	}
}

// Delete removes a value from the cache
func (m *Manager) Delete(key string) {
	m.Cache.Delete(key)
}

// Prefetch queues a key for background prefetching
func (m *Manager) Prefetch(req PrefetchRequest) {
	if !m.Options.EnablePrefetching {
		return
	}

	select {
	case m.prefetchQueue <- req:
		// Successfully queued
	default:
		// Queue is full, skip prefetching
	}
}

// prefetchWorker processes prefetch requests
func (m *Manager) prefetchWorker(ctx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-m.prefetchQueue:
			// Skip if already in cache
			if _, found := m.Cache.Get(req.Key); found {
				continue
			}

			// Fetch the item
			if req.Callback != nil {
				value, err := req.Callback(ctx)
				if err == nil {
					m.Cache.Set(req.Key, value, m.Options.DefaultTTL)
				}
			}
		}
	}
}

// applyInvalidationPatterns applies invalidation patterns
func (m *Manager) applyInvalidationPatterns(key string, condition InvalidationCondition) {
	for _, pattern := range m.invalidationPatterns {
		if pattern.Condition == condition {
			// TODO: Implement pattern matching
			// For now, just a simple exact match
			if pattern.Pattern == key {
				switch pattern.Action {
				case InvalidateDelete:
					m.Cache.Delete(key)
				case InvalidateRefresh:
					// TODO: Implement refresh
				case InvalidateExpire:
					// TODO: Implement expire
				}
			}
		}
	}
}

// AddInvalidationPattern adds an invalidation pattern
func (m *Manager) AddInvalidationPattern(pattern InvalidationPattern) {
	m.invalidationPatterns = append(m.invalidationPatterns, pattern)
}

// GetMetrics returns cache metrics
func (m *Manager) GetMetrics() *Metrics {
	return m.Cache.GetMetrics()
}
