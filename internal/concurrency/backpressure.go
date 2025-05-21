package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// BackpressureController manages backpressure for concurrent operations
type BackpressureController struct {
	// maxConcurrent is the maximum number of concurrent operations
	maxConcurrent int32

	// currentConcurrent is the current number of concurrent operations
	currentConcurrent atomic.Int32

	// maxQueueSize is the maximum size of the queue
	maxQueueSize int32

	// currentQueueSize is the current size of the queue
	currentQueueSize atomic.Int32

	// rejectionCount is the number of rejected operations
	rejectionCount atomic.Int32

	// successCount is the number of successful operations
	successCount atomic.Int32

	// failureCount is the number of failed operations
	failureCount atomic.Int32

	// lastAdjustment is the time of the last adjustment
	lastAdjustment time.Time

	// adjustmentInterval is the interval between adjustments
	adjustmentInterval time.Duration

	// mu is used to protect shared state
	mu sync.RWMutex

	// dynamicAdjustment enables dynamic adjustment of concurrency limits
	dynamicAdjustment bool

	// minConcurrent is the minimum number of concurrent operations
	minConcurrent int32

	// maxConcurrentHardLimit is the hard limit for maxConcurrent
	maxConcurrentHardLimit int32
}

// NewBackpressureController creates a new backpressure controller
func NewBackpressureController(maxConcurrent, maxQueueSize int) *BackpressureController {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	if maxQueueSize <= 0 {
		maxQueueSize = maxConcurrent * 2
	}

	return &BackpressureController{
		maxConcurrent:         int32(maxConcurrent),
		maxQueueSize:          int32(maxQueueSize),
		lastAdjustment:        time.Now(),
		adjustmentInterval:    5 * time.Second,
		dynamicAdjustment:     true,
		minConcurrent:         1,
		maxConcurrentHardLimit: int32(maxConcurrent * 2),
	}
}

// Acquire attempts to acquire a permit for an operation
func (c *BackpressureController) Acquire(ctx context.Context) bool {
	// Check if we can start a new operation
	if c.currentConcurrent.Load() >= c.maxConcurrent {
		// Check if we can queue the operation
		if c.currentQueueSize.Load() >= c.maxQueueSize {
			// Reject the operation
			c.rejectionCount.Add(1)
			return false
		}

		// Queue the operation
		c.currentQueueSize.Add(1)

		// Wait for an available slot or context cancellation
		for {
			select {
			case <-ctx.Done():
				// Context canceled, remove from queue
				c.currentQueueSize.Add(-1)
				return false
			case <-time.After(10 * time.Millisecond):
				// Check if we can start the operation now
				if c.currentConcurrent.Load() < c.maxConcurrent {
					// Remove from queue and start the operation
					c.currentQueueSize.Add(-1)
					c.currentConcurrent.Add(1)
					return true
				}
			}
		}
	}

	// Start the operation
	c.currentConcurrent.Add(1)
	return true
}

// Release releases a permit after an operation completes
func (c *BackpressureController) Release(success bool) {
	c.currentConcurrent.Add(-1)

	if success {
		c.successCount.Add(1)
	} else {
		c.failureCount.Add(1)
	}

	// Adjust concurrency limits if needed
	if c.dynamicAdjustment && time.Since(c.lastAdjustment) > c.adjustmentInterval {
		c.adjustConcurrencyLimits()
	}
}

// adjustConcurrencyLimits adjusts concurrency limits based on success/failure rates
func (c *BackpressureController) adjustConcurrencyLimits() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Only adjust if enough time has passed
	if time.Since(c.lastAdjustment) <= c.adjustmentInterval {
		return
	}

	// Calculate success rate
	total := c.successCount.Load() + c.failureCount.Load()
	if total == 0 {
		return
	}

	successRate := float64(c.successCount.Load()) / float64(total)
	rejectionRate := float64(c.rejectionCount.Load()) / float64(total+c.rejectionCount.Load())

	// Reset counters
	c.successCount.Store(0)
	c.failureCount.Store(0)
	c.rejectionCount.Store(0)

	// Adjust concurrency limits
	currentMax := c.maxConcurrent
	if successRate > 0.95 && rejectionRate > 0.1 {
		// High success rate but many rejections, increase concurrency
		newMax := currentMax + 1
		if newMax <= c.maxConcurrentHardLimit {
			c.maxConcurrent = newMax
		}
	} else if successRate < 0.8 {
		// Low success rate, decrease concurrency
		newMax := currentMax - 1
		if newMax >= c.minConcurrent {
			c.maxConcurrent = newMax
		}
	}

	c.lastAdjustment = time.Now()
}

// GetStats returns statistics about the controller
func (c *BackpressureController) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"maxConcurrent":     c.maxConcurrent,
		"currentConcurrent": c.currentConcurrent.Load(),
		"maxQueueSize":      c.maxQueueSize,
		"currentQueueSize":  c.currentQueueSize.Load(),
		"rejectionCount":    c.rejectionCount.Load(),
		"successCount":      c.successCount.Load(),
		"failureCount":      c.failureCount.Load(),
	}
}

// WithBackpressure executes a function with backpressure control
func WithBackpressure(ctx context.Context, controller *BackpressureController, fn func() (interface{}, error)) (interface{}, error) {
	// Acquire a permit
	if !controller.Acquire(ctx) {
		return nil, context.DeadlineExceeded
	}

	// Execute the function
	result, err := fn()

	// Release the permit
	controller.Release(err == nil)

	return result, err
}
