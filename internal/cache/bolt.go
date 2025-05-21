package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// boltCache implements a persistent cache using BoltDB
type boltCache struct {
	db           *bbolt.DB
	bucketName   []byte
	prefetchChan chan string
	prefetchWg   sync.WaitGroup
	mu           sync.RWMutex
	options      *Options
	metrics      *Metrics
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewBoltCache creates a new BoltDB-backed cache
func NewBoltCache(opts *Options) (Cache, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Configure BoltDB options
	boltOpts := &bbolt.Options{
		Timeout:      1 * time.Second,
		NoGrowSync:   false,
		ReadOnly:     opts.ReadOnly,
		FreelistType: bbolt.FreelistMapType,
	}

	// Open the database
	dbPath := filepath.Join(opts.CacheDir, "bolt-cache.db")
	db, err := bbolt.Open(dbPath, 0600, boltOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BoltDB: %w", err)
	}

	// Create the bucket if it doesn't exist
	bucketName := []byte("cache")
	if !opts.ReadOnly {
		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(bucketName)
			return err
		})
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	// Create context for background operations
	ctx, cancel := context.WithCancel(context.Background())

	// Create the cache
	cache := &boltCache{
		db:           db,
		bucketName:   bucketName,
		prefetchChan: make(chan string, 100),
		options:      opts,
		metrics:      NewMetrics(),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background processes
	if !opts.ReadOnly {
		go cache.processPrefetchQueue(ctx)
	}

	return cache, nil
}

// Close closes the cache
func (c *boltCache) Close() error {
	c.cancel()
	c.prefetchWg.Wait()
	return c.db.Close()
}

// Get retrieves a value from the cache
func (c *boltCache) Get(key string) (interface{}, bool) {
	c.metrics.Inc(&c.metrics.Gets)

	var value []byte
	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(c.bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		val := bucket.Get([]byte(key))
		if val == nil {
			return fmt.Errorf("key not found")
		}

		// Copy the value since it's only valid during the transaction
		value = make([]byte, len(val))
		copy(value, val)
		return nil
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Misses)
		return nil, false
	}

	// Unmarshal the value
	var item struct {
		Value      interface{} `json:"v"`
		Expiration int64       `json:"e"`
	}
	if err := json.Unmarshal(value, &item); err != nil {
		c.metrics.Inc(&c.metrics.Errors)
		return nil, false
	}

	// Check if the item has expired
	if item.Expiration > 0 && time.Now().Unix() > item.Expiration {
		c.Delete(key) // Clean up expired item
		c.metrics.Inc(&c.metrics.Misses)
		return nil, false
	}

	c.metrics.Inc(&c.metrics.Hits)
	return item.Value, true
}

// Set adds a value to the cache with the given TTL
func (c *boltCache) Set(key string, value interface{}, ttl time.Duration) {
	c.metrics.Inc(&c.metrics.Sets)

	// Calculate expiration time
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).Unix()
	}

	// Marshal the value with expiration
	item := struct {
		Value      interface{} `json:"v"`
		Expiration int64       `json:"e"`
	}{
		Value:      value,
		Expiration: expiration,
	}

	data, err := json.Marshal(item)
	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
		return
	}

	// Set the value in the database
	err = c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(c.bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.Put([]byte(key), data)
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Delete removes a value from the cache
func (c *boltCache) Delete(key string) {
	c.metrics.Inc(&c.metrics.Deletes)

	err := c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(c.bucketName)
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}
		return bucket.Delete([]byte(key))
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Clear removes all items from the cache
func (c *boltCache) Clear() {
	c.metrics.Inc(&c.metrics.Clears)

	err := c.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(c.bucketName); err != nil {
			return err
		}
		_, err := tx.CreateBucket(c.bucketName)
		return err
	})

	if err != nil {
		c.metrics.Inc(&c.metrics.Errors)
	}
}

// Prefetch queues a key for background prefetching
func (c *boltCache) Prefetch(key string) {
	select {
	case c.prefetchChan <- key:
		c.metrics.Inc(&c.metrics.Prefetches)
	default:
		// Channel is full, skip prefetching
	}
}

// processPrefetchQueue processes the prefetch queue
func (c *boltCache) processPrefetchQueue(ctx context.Context) {
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

// GetMetrics returns cache metrics
func (c *boltCache) GetMetrics() *Metrics {
	return c.metrics
}
