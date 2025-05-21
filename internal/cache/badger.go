package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

// badgerCache implements a persistent cache using BadgerDB
type badgerCache struct {
	db           *badger.DB
	prefetchChan chan string
	prefetchWg   sync.WaitGroup
	mu           sync.RWMutex
	options      *Options
	metrics      *Metrics
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewBadgerCache creates a new BadgerDB-backed cache
func NewBadgerCache(opts *Options) (Cache, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Configure BadgerDB options
	badgerOpts := badger.DefaultOptions(opts.CacheDir)
	badgerOpts.Logger = nil // Disable default logging

	// Optimize for performance
	badgerOpts.SyncWrites = false
	badgerOpts.NumVersionsToKeep = 1

	// Memory optimization
	if opts.MemoryLimit > 0 {
		badgerOpts.MemTableSize = int64(opts.MemoryLimit / 4)
	}

	// Compression settings
	badgerOpts.Compression = options.Snappy

	// Set appropriate values based on expected usage
	if opts.ReadOnly {
		badgerOpts.ReadOnly = true
	}

	// Open the database
	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}

	// Create context for background operations
	ctx, cancel := context.WithCancel(context.Background())

	// Create the cache
	cache := &badgerCache{
		db:           db,
		prefetchChan: make(chan string, 100),
		options:      opts,
		metrics:      NewMetrics(),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background processes
	if !opts.ReadOnly {
		go cache.runGC(ctx)
		go cache.processPrefetchQueue(ctx)
	}

	return cache, nil
}

// Close closes the cache
func (c *badgerCache) Close() error {
	c.cancel()
	c.prefetchWg.Wait()
	return c.db.Close()
}

// Get retrieves a value from the cache
func (c *badgerCache) Get(key string) (interface{}, bool) {
	c.metrics.Inc(&c.metrics.Gets)

	var value []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Misses)
		return nil, false
	}

	// Unmarshal the value
	var result interface{}
	if err := json.Unmarshal(value, &result); err != nil {
		c.metrics.Inc(&c.metrics.Errors)
		return nil, false
	}

	c.metrics.Inc(&c.metrics.Hits)
	return result, true
}

// Set adds a value to the cache with the given TTL
func (c *badgerCache) Set(key string, value interface{}, ttl time.Duration) {
	c.metrics.Inc(&c.metrics.Sets)

	// Marshal the value
	data, err := json.Marshal(value)
	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
		return
	}

	// Set the value in the database
	err = c.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), data).WithTTL(ttl)
		return txn.SetEntry(entry)
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Delete removes a value from the cache
func (c *badgerCache) Delete(key string) {
	c.metrics.Inc(&c.metrics.Deletes)

	err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Clear removes all items from the cache
func (c *badgerCache) Clear() {
	c.metrics.Inc(&c.metrics.Clears)

	// Drop all data
	err := c.db.DropAll()
	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Prefetch queues a key for background prefetching
func (c *badgerCache) Prefetch(key string) {
	select {
	case c.prefetchChan <- key:
		c.metrics.Inc(&c.metrics.Prefetches)
	default:
		// Channel is full, skip prefetching
	}
}

// processPrefetchQueue processes the prefetch queue
func (c *badgerCache) processPrefetchQueue(ctx context.Context) {
	c.prefetchWg.Add(1)
	defer c.prefetchWg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.prefetchChan:
			// The actual prefetching would be implemented by the client
			// This just provides the mechanism for queueing prefetch requests
			c.metrics.Inc(&c.metrics.PrefetchesProcessed)
		}
	}
}

// runGC runs garbage collection periodically
func (c *badgerCache) runGC(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := c.db.RunValueLogGC(0.5)
			if err != nil && err != badger.ErrNoRewrite {
				// Log error but continue
				fmt.Printf("BadgerDB GC error: %v\n", err)
			}
		}
	}
}

// GetMetrics returns cache metrics
func (c *badgerCache) GetMetrics() *Metrics {
	return c.metrics
}
