package cache

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Options configures the cache behavior
type Options struct {
	// CacheDir is the directory to store cached data
	CacheDir string

	// DefaultTTL is the default time-to-live for cached items
	DefaultTTL time.Duration

	// MaxSize is the maximum size of the cache in bytes (0 = unlimited)
	MaxSize int64

	// MemoryLimit is the maximum memory usage in bytes (0 = unlimited)
	MemoryLimit int64

	// ReadOnly opens the cache in read-only mode
	ReadOnly bool

	// EnablePrefetching enables background prefetching
	EnablePrefetching bool

	// PrefetchConcurrency is the number of concurrent prefetch operations
	PrefetchConcurrency int

	// EnableCompression enables compression for cached values
	EnableCompression bool

	// EnableEncryption enables encryption for cached values
	EnableEncryption bool

	// EncryptionKey is the key used for encryption
	EncryptionKey []byte

	// EnableMetrics enables collection of cache metrics
	EnableMetrics bool
}

// DefaultOptions returns the default cache options
func DefaultOptions() *Options {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &Options{
		CacheDir:            filepath.Join(homeDir, ".gh-notif-cache"),
		DefaultTTL:          1 * time.Hour,
		MaxSize:             1 * 1024 * 1024 * 1024, // 1GB
		MemoryLimit:         100 * 1024 * 1024,      // 100MB
		ReadOnly:            false,
		EnablePrefetching:   true,
		PrefetchConcurrency: 2,
		EnableCompression:   true,
		EnableEncryption:    false,
		EnableMetrics:       true,
	}
}

// Metrics tracks cache performance metrics
type Metrics struct {
	// Gets is the number of Get operations
	Gets int64

	// Sets is the number of Set operations
	Sets int64

	// Hits is the number of cache hits
	Hits int64

	// Misses is the number of cache misses
	Misses int64

	// Deletes is the number of Delete operations
	Deletes int64

	// Clears is the number of Clear operations
	Clears int64

	// Errors is the number of errors
	Errors int64

	// Prefetches is the number of prefetch requests
	Prefetches int64

	// PrefetchesProcessed is the number of processed prefetch requests
	PrefetchesProcessed int64

	// Size is the current size of the cache in bytes
	Size int64

	// mu protects the metrics
	mu sync.Mutex
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Inc increments a metric
func (m *Metrics) Inc(metric *int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*metric++
}

// Add adds a value to a metric
func (m *Metrics) Add(metric *int64, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	*metric += value
}

// Get gets a metric value
func (m *Metrics) Get(metric *int64) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return *metric
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gets = 0
	m.Sets = 0
	m.Hits = 0
	m.Misses = 0
	m.Deletes = 0
	m.Clears = 0
	m.Errors = 0
	m.Prefetches = 0
	m.PrefetchesProcessed = 0
	m.Size = 0
}

// HitRatio returns the cache hit ratio
func (m *Metrics) HitRatio() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	hits := m.Hits
	misses := m.Misses
	total := hits + misses

	if total == 0 {
		return 0
	}

	return float64(hits) / float64(total)
}
