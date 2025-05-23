# Performance Optimization Guide

This document covers performance optimization strategies, monitoring, and best practices for gh-notif.

## Performance Overview

gh-notif is designed for high performance with the following key optimizations:

1. **Concurrent Operations**: Parallel processing of API requests and data operations
2. **Intelligent Caching**: Multi-layer caching with smart invalidation
3. **Memory Efficiency**: Object pooling and streaming for reduced memory usage
4. **Network Optimization**: Request batching and conditional requests
5. **Algorithm Optimization**: Efficient data structures and algorithms

## Caching Strategy

### Multi-Layer Caching

gh-notif implements a sophisticated caching system:

```go
// Cache hierarchy
type CacheManager struct {
    l1Cache *sync.Map           // In-memory cache (fastest)
    l2Cache *badger.DB          // Persistent cache (fast)
    l3Cache *bolt.DB            // Long-term storage (slower but persistent)
}

func (cm *CacheManager) Get(key string) (interface{}, bool) {
    // Try L1 cache first
    if value, ok := cm.l1Cache.Load(key); ok {
        return value, true
    }
    
    // Try L2 cache
    if value, err := cm.getFromL2(key); err == nil {
        cm.l1Cache.Store(key, value) // Promote to L1
        return value, true
    }
    
    // Try L3 cache
    if value, err := cm.getFromL3(key); err == nil {
        cm.l1Cache.Store(key, value) // Promote to L1
        cm.storeInL2(key, value)     // Store in L2
        return value, true
    }
    
    return nil, false
}
```

### Cache Invalidation

Smart cache invalidation based on GitHub's ETag headers:

```go
// ETag-based cache invalidation
type CacheEntry struct {
    Data     interface{}
    ETag     string
    LastMod  time.Time
    TTL      time.Duration
}

func (c *Client) GetWithCache(url string) (*Response, error) {
    cacheKey := generateCacheKey(url)
    
    // Check cache
    if entry, exists := c.cache.Get(cacheKey); exists {
        // Use conditional request
        req.Header.Set("If-None-Match", entry.ETag)
        req.Header.Set("If-Modified-Since", entry.LastMod.Format(time.RFC1123))
        
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        
        // 304 Not Modified - use cached data
        if resp.StatusCode == http.StatusNotModified {
            return entry.Data.(*Response), nil
        }
        
        // Update cache with new data
        newEntry := &CacheEntry{
            Data:    newData,
            ETag:    resp.Header.Get("ETag"),
            LastMod: parseLastModified(resp.Header.Get("Last-Modified")),
            TTL:     determineTTL(resp),
        }
        c.cache.Set(cacheKey, newEntry)
    }
    
    return newData, nil
}
```

### Cache Configuration

Configurable cache settings for different use cases:

```yaml
# Performance-optimized cache configuration
cache:
  type: "badger"              # Options: memory, badger, bolt, hybrid
  max_size: 1073741824        # 1GB max cache size
  memory_limit: 104857600     # 100MB memory limit
  ttl: 3600                   # Default TTL in seconds
  cleanup_interval: 300       # Cleanup interval in seconds
  compression: true           # Enable compression for stored data
  
  # Cache policies
  policies:
    notifications:
      ttl: 300                # 5 minutes for notifications
      max_entries: 10000
    repositories:
      ttl: 3600               # 1 hour for repository data
      max_entries: 1000
    users:
      ttl: 7200               # 2 hours for user data
      max_entries: 500
```

## Concurrent Operations

### Worker Pool Pattern

Efficient concurrent processing using worker pools:

```go
// Worker pool for concurrent API requests
type WorkerPool struct {
    workers    int
    jobs       chan Job
    results    chan Result
    wg         sync.WaitGroup
    semaphore  chan struct{}
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers:   workers,
        jobs:      make(chan Job, workers*2),
        results:   make(chan Result, workers*2),
        semaphore: make(chan struct{}, workers),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    defer wp.wg.Done()
    
    for job := range wp.jobs {
        wp.semaphore <- struct{}{} // Acquire semaphore
        
        result := job.Execute()
        wp.results <- result
        
        <-wp.semaphore // Release semaphore
    }
}

// Usage for fetching notifications concurrently
func (c *Client) FetchNotificationsConcurrently(repos []string) ([]Notification, error) {
    pool := NewWorkerPool(10) // 10 concurrent workers
    pool.Start()
    defer pool.Stop()
    
    // Submit jobs
    for _, repo := range repos {
        job := &FetchNotificationsJob{
            client: c,
            repo:   repo,
        }
        pool.Submit(job)
    }
    
    // Collect results
    var allNotifications []Notification
    for i := 0; i < len(repos); i++ {
        result := <-pool.results
        if result.Error == nil {
            allNotifications = append(allNotifications, result.Notifications...)
        }
    }
    
    return allNotifications, nil
}
```

### Rate Limiting with Backoff

Intelligent rate limiting with exponential backoff:

```go
// Adaptive rate limiter
type AdaptiveRateLimiter struct {
    tokens     chan struct{}
    refillRate time.Duration
    maxTokens  int
    backoff    *ExponentialBackoff
}

func (rl *AdaptiveRateLimiter) Wait(ctx context.Context) error {
    select {
    case <-rl.tokens:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(rl.backoff.Next()):
        // Adaptive backoff based on API response
        return rl.Wait(ctx)
    }
}

// Exponential backoff with jitter
type ExponentialBackoff struct {
    baseDelay  time.Duration
    maxDelay   time.Duration
    multiplier float64
    jitter     float64
    attempts   int
}

func (eb *ExponentialBackoff) Next() time.Duration {
    if eb.attempts == 0 {
        eb.attempts++
        return eb.baseDelay
    }
    
    delay := float64(eb.baseDelay) * math.Pow(eb.multiplier, float64(eb.attempts))
    if delay > float64(eb.maxDelay) {
        delay = float64(eb.maxDelay)
    }
    
    // Add jitter to prevent thundering herd
    jitter := delay * eb.jitter * (rand.Float64()*2 - 1)
    delay += jitter
    
    eb.attempts++
    return time.Duration(delay)
}
```

## Memory Optimization

### Object Pooling

Reduce garbage collection pressure with object pools:

```go
// Object pools for frequently allocated objects
var (
    notificationPool = sync.Pool{
        New: func() interface{} {
            return &Notification{}
        },
    }
    
    requestPool = sync.Pool{
        New: func() interface{} {
            return &http.Request{}
        },
    }
    
    bufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 0, 4096)
        },
    }
)

// Usage
func processNotifications(data []byte) ([]*Notification, error) {
    // Get buffer from pool
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0]) // Reset and return to pool
    
    var notifications []*Notification
    
    // Parse notifications
    for _, item := range parseItems(data) {
        notif := notificationPool.Get().(*Notification)
        defer notificationPool.Put(notif) // Return to pool when done
        
        // Populate notification
        notif.Reset() // Reset to clean state
        notif.ID = item.ID
        notif.Repository = item.Repository
        // ... other fields
        
        notifications = append(notifications, notif)
    }
    
    return notifications, nil
}
```

### Streaming and Chunking

Process large datasets efficiently:

```go
// Streaming JSON parser for large responses
func (c *Client) StreamNotifications(ctx context.Context) (<-chan *Notification, error) {
    resp, err := c.makeRequest(ctx, "GET", "/notifications")
    if err != nil {
        return nil, err
    }
    
    notifications := make(chan *Notification, 100) // Buffered channel
    
    go func() {
        defer close(notifications)
        defer resp.Body.Close()
        
        decoder := json.NewDecoder(resp.Body)
        
        // Read opening bracket
        token, err := decoder.Token()
        if err != nil || token != json.Delim('[') {
            return
        }
        
        // Stream individual notifications
        for decoder.More() {
            notif := notificationPool.Get().(*Notification)
            
            if err := decoder.Decode(notif); err != nil {
                notificationPool.Put(notif)
                continue
            }
            
            select {
            case notifications <- notif:
            case <-ctx.Done():
                notificationPool.Put(notif)
                return
            }
        }
    }()
    
    return notifications, nil
}
```

## Algorithm Optimization

### Efficient Filtering

Optimized filtering with tries and indices:

```go
// Trie-based repository filtering
type RepositoryTrie struct {
    root *TrieNode
}

type TrieNode struct {
    children map[rune]*TrieNode
    isEnd    bool
    repos    []*Repository
}

func (rt *RepositoryTrie) Insert(repo *Repository) {
    node := rt.root
    for _, char := range repo.FullName {
        if node.children[char] == nil {
            node.children[char] = &TrieNode{
                children: make(map[rune]*TrieNode),
            }
        }
        node = node.children[char]
    }
    node.isEnd = true
    node.repos = append(node.repos, repo)
}

func (rt *RepositoryTrie) Search(prefix string) []*Repository {
    node := rt.root
    for _, char := range prefix {
        if node.children[char] == nil {
            return nil
        }
        node = node.children[char]
    }
    
    return rt.collectRepos(node)
}

// Indexed filtering for fast lookups
type NotificationIndex struct {
    byRepo   map[string][]*Notification
    byType   map[string][]*Notification
    byReason map[string][]*Notification
    byUser   map[string][]*Notification
}

func (ni *NotificationIndex) Filter(filter *Filter) []*Notification {
    var candidates []*Notification
    
    // Start with most selective index
    if filter.Repository != "" {
        candidates = ni.byRepo[filter.Repository]
    } else if filter.Type != "" {
        candidates = ni.byType[filter.Type]
    } else if filter.Reason != "" {
        candidates = ni.byReason[filter.Reason]
    } else {
        // Fallback to all notifications
        for _, notifications := range ni.byRepo {
            candidates = append(candidates, notifications...)
        }
    }
    
    // Apply additional filters
    return ni.applyFilters(candidates, filter)
}
```

### Parallel Sorting

Efficient parallel sorting for large datasets:

```go
// Parallel merge sort for notifications
func ParallelSort(notifications []*Notification, compareFn func(a, b *Notification) bool) {
    if len(notifications) < 1000 {
        // Use standard sort for small datasets
        sort.Slice(notifications, func(i, j int) bool {
            return compareFn(notifications[i], notifications[j])
        })
        return
    }
    
    // Parallel sort for large datasets
    numWorkers := runtime.NumCPU()
    chunkSize := len(notifications) / numWorkers
    
    var wg sync.WaitGroup
    
    // Sort chunks in parallel
    for i := 0; i < numWorkers; i++ {
        start := i * chunkSize
        end := start + chunkSize
        if i == numWorkers-1 {
            end = len(notifications)
        }
        
        wg.Add(1)
        go func(chunk []*Notification) {
            defer wg.Done()
            sort.Slice(chunk, func(i, j int) bool {
                return compareFn(chunk[i], chunk[j])
            })
        }(notifications[start:end])
    }
    
    wg.Wait()
    
    // Merge sorted chunks
    mergeChunks(notifications, chunkSize, compareFn)
}
```

## Performance Monitoring

### Built-in Profiling

Comprehensive profiling capabilities:

```go
// Performance profiler
type Profiler struct {
    cpuProfile    *os.File
    memProfile    *os.File
    httpServer    *http.Server
    metrics       *Metrics
}

func (p *Profiler) StartCPUProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    
    p.cpuProfile = f
    return pprof.StartCPUProfile(f)
}

func (p *Profiler) StopCPUProfile() {
    if p.cpuProfile != nil {
        pprof.StopCPUProfile()
        p.cpuProfile.Close()
        p.cpuProfile = nil
    }
}

func (p *Profiler) WriteMemProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    runtime.GC() // Force GC before memory profile
    return pprof.WriteHeapProfile(f)
}

// HTTP profiling server
func (p *Profiler) StartHTTPServer(port int) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
    
    p.httpServer = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: mux,
    }
    
    return p.httpServer.ListenAndServe()
}
```

### Metrics Collection

Real-time performance metrics:

```go
// Performance metrics
type Metrics struct {
    APICallDuration    *prometheus.HistogramVec
    CacheHitRate      *prometheus.CounterVec
    MemoryUsage       prometheus.Gauge
    GoroutineCount    prometheus.Gauge
    RequestsPerSecond prometheus.Counter
}

func NewMetrics() *Metrics {
    return &Metrics{
        APICallDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "api_call_duration_seconds",
                Help: "Duration of API calls",
            },
            []string{"endpoint", "method"},
        ),
        CacheHitRate: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "cache_operations_total",
                Help: "Cache operations",
            },
            []string{"type", "result"},
        ),
        MemoryUsage: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "memory_usage_bytes",
                Help: "Current memory usage",
            },
        ),
    }
}

// Usage
func (c *Client) makeAPICall(endpoint string) (*Response, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        c.metrics.APICallDuration.WithLabelValues(endpoint, "GET").Observe(duration.Seconds())
    }()
    
    // Make API call
    resp, err := c.httpClient.Get(endpoint)
    return resp, err
}
```

## Performance Tuning

### Configuration Optimization

Performance-tuned configuration:

```yaml
# High-performance configuration
performance:
  # Concurrency settings
  max_concurrent_requests: 20
  worker_pool_size: 10
  batch_size: 50
  
  # Cache settings
  cache_size: "1GB"
  cache_ttl: 300
  enable_compression: true
  
  # Network settings
  timeout: 30
  keep_alive: true
  max_idle_connections: 100
  
  # Memory settings
  gc_percent: 100
  max_memory: "512MB"
  enable_object_pooling: true
```

### Benchmarking

Comprehensive benchmarking suite:

```go
// Benchmark tests
func BenchmarkFilterNotifications(b *testing.B) {
    notifications := generateTestNotifications(10000)
    filter := &Filter{Repository: "owner/repo"}
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        FilterNotifications(notifications, filter)
    }
}

func BenchmarkConcurrentAPICall(b *testing.B) {
    client := NewClient("test-token")
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            client.ListNotifications(context.Background())
        }
    })
}
```

## Performance Best Practices

### For Users

1. **Configuration Tuning**:
   - Adjust cache size based on available memory
   - Tune concurrency settings for your network
   - Enable compression for slow connections

2. **Usage Patterns**:
   - Use filters to reduce data transfer
   - Batch operations when possible
   - Avoid frequent polling

3. **System Optimization**:
   - Ensure adequate memory
   - Use SSD storage for cache
   - Optimize network settings

### For Developers

1. **Code Optimization**:
   - Use object pooling for frequently allocated objects
   - Implement efficient algorithms
   - Minimize memory allocations

2. **Profiling**:
   - Regular performance profiling
   - Benchmark critical paths
   - Monitor memory usage

3. **Testing**:
   - Performance regression tests
   - Load testing
   - Memory leak detection
