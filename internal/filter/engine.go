package filter

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

// Engine is a high-performance notification filtering engine
type Engine struct {
	// Concurrency controls the number of worker goroutines
	Concurrency int
	// BatchSize controls the size of notification batches
	BatchSize int
	// Timeout controls the maximum time to spend filtering
	Timeout time.Duration
	// FilterExpr is the filter to apply
	FilterExpr Filter
	// IndexEnabled controls whether to use indexing
	IndexEnabled bool
	// indexes stores precomputed indexes for faster filtering
	indexes map[string]interface{}
	// indexMu protects the indexes map
	indexMu sync.RWMutex
}

// NewEngine creates a new filtering engine
func NewEngine() *Engine {
	return &Engine{
		Concurrency:  runtime.NumCPU(),
		BatchSize:    100,
		Timeout:      5 * time.Second,
		IndexEnabled: true,
		indexes:      make(map[string]interface{}),
	}
}

// WithFilter sets the filter for the engine
func (e *Engine) WithFilter(filter Filter) *Engine {
	e.FilterExpr = filter
	return e
}

// WithConcurrency sets the concurrency for the engine
func (e *Engine) WithConcurrency(concurrency int) *Engine {
	if concurrency > 0 {
		e.Concurrency = concurrency
	}
	return e
}

// WithBatchSize sets the batch size for the engine
func (e *Engine) WithBatchSize(batchSize int) *Engine {
	if batchSize > 0 {
		e.BatchSize = batchSize
	}
	return e
}

// WithTimeout sets the timeout for the engine
func (e *Engine) WithTimeout(timeout time.Duration) *Engine {
	if timeout > 0 {
		e.Timeout = timeout
	}
	return e
}

// WithIndexing enables or disables indexing
func (e *Engine) WithIndexing(enabled bool) *Engine {
	e.IndexEnabled = enabled
	return e
}

// buildIndexes builds indexes for faster filtering
func (e *Engine) buildIndexes(notifications []*github.Notification) {
	if !e.IndexEnabled {
		return
	}

	e.indexMu.Lock()
	defer e.indexMu.Unlock()

	// Clear existing indexes
	e.indexes = make(map[string]interface{})

	// Build repository index
	repoIndex := make(map[string][]*github.Notification)
	for _, n := range notifications {
		repo := n.GetRepository().GetFullName()
		repoIndex[repo] = append(repoIndex[repo], n)
	}
	e.indexes["repository"] = repoIndex

	// Build organization index
	orgIndex := make(map[string][]*github.Notification)
	for _, n := range notifications {
		fullName := n.GetRepository().GetFullName()
		parts := strings.Split(fullName, "/")
		if len(parts) >= 2 {
			org := parts[0]
			orgIndex[org] = append(orgIndex[org], n)
		}
	}
	e.indexes["organization"] = orgIndex

	// Build type index
	typeIndex := make(map[string][]*github.Notification)
	for _, n := range notifications {
		typ := n.GetSubject().GetType()
		typeIndex[typ] = append(typeIndex[typ], n)
	}
	e.indexes["type"] = typeIndex

	// Build status index
	statusIndex := make(map[bool][]*github.Notification)
	for _, n := range notifications {
		unread := n.GetUnread()
		statusIndex[unread] = append(statusIndex[unread], n)
	}
	e.indexes["status"] = statusIndex
}

// Filter filters notifications using the configured filter
func (e *Engine) Filter(ctx context.Context, notifications []*github.Notification) ([]*github.Notification, error) {
	if len(notifications) == 0 {
		return nil, nil
	}

	if e.FilterExpr == nil {
		return notifications, nil
	}

	// Build indexes for faster filtering
	e.buildIndexes(notifications)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, e.Timeout)
	defer cancel()

	// For small sets, don't bother with concurrency
	if len(notifications) < e.BatchSize {
		return e.filterSequential(notifications), nil
	}

	return e.filterConcurrent(ctx, notifications)
}

// filterSequential filters notifications sequentially
func (e *Engine) filterSequential(notifications []*github.Notification) []*github.Notification {
	var result []*github.Notification
	for _, n := range notifications {
		if e.FilterExpr.Apply(n) {
			result = append(result, n)
		}
	}
	return result
}

// filterConcurrent filters notifications concurrently
func (e *Engine) filterConcurrent(ctx context.Context, notifications []*github.Notification) ([]*github.Notification, error) {
	// Create channels for input and output
	input := make(chan *github.Notification, e.BatchSize)
	output := make(chan *github.Notification, e.BatchSize)
	done := make(chan struct{})

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < e.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range input {
				if e.FilterExpr.Apply(n) {
					select {
					case output <- n:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Start a goroutine to close the output channel when all workers are done
	go func() {
		wg.Wait()
		close(output)
		close(done)
	}()

	// Feed input channel with notifications
	go func() {
		defer close(input)
		for _, n := range notifications {
			select {
			case input <- n:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	var result []*github.Notification
	for {
		select {
		case n, ok := <-output:
			if !ok {
				return result, nil
			}
			result = append(result, n)
		case <-ctx.Done():
			return result, ctx.Err()
		case <-done:
			return result, nil
		}
	}
}

// OptimizeFilter optimizes a filter for better performance
func (e *Engine) OptimizeFilter(filter Filter) Filter {
	// This is a placeholder for a more sophisticated optimization
	// In a real implementation, this would analyze the filter and
	// rewrite it for better performance
	return filter
}

// FilterBuilder helps build complex filters
type FilterBuilder struct {
	filters []Filter
	ops     []Operator
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{}
}

// And adds a filter with AND logic
func (b *FilterBuilder) And(filter Filter) *FilterBuilder {
	b.filters = append(b.filters, filter)
	if len(b.filters) > 1 {
		b.ops = append(b.ops, And)
	}
	return b
}

// Or adds a filter with OR logic
func (b *FilterBuilder) Or(filter Filter) *FilterBuilder {
	b.filters = append(b.filters, filter)
	if len(b.filters) > 1 {
		b.ops = append(b.ops, Or)
	}
	return b
}

// Not adds a filter with NOT logic
func (b *FilterBuilder) Not(filter Filter) *FilterBuilder {
	notFilter := &CompositeFilter{
		Filters:  []Filter{filter},
		Operator: Not,
	}
	b.filters = append(b.filters, notFilter)
	if len(b.filters) > 1 {
		b.ops = append(b.ops, And) // Default to AND for NOT filters
	}
	return b
}

// Build builds the final filter
func (b *FilterBuilder) Build() Filter {
	if len(b.filters) == 0 {
		return nil
	}

	if len(b.filters) == 1 {
		return b.filters[0]
	}

	// Build the filter tree
	return b.buildTree(0, len(b.filters)-1)
}

// buildTree builds a filter tree from the given range
func (b *FilterBuilder) buildTree(start, end int) Filter {
	if start == end {
		return b.filters[start]
	}

	// Find the lowest precedence operator
	lowestPrecedenceIdx := start
	lowestPrecedence := b.ops[start]
	for i := start + 1; i < end; i++ {
		if b.ops[i] == Or && lowestPrecedence == And {
			lowestPrecedenceIdx = i
			lowestPrecedence = Or
		}
	}

	// Split at the lowest precedence operator
	left := b.buildTree(start, lowestPrecedenceIdx)
	right := b.buildTree(lowestPrecedenceIdx+1, end)

	return &CompositeFilter{
		Filters:  []Filter{left, right},
		Operator: b.ops[lowestPrecedenceIdx],
	}
}
