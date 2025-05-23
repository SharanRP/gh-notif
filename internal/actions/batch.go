package actions

import (
	"context"
	"sync"
	"time"
)

// Task represents a task to be executed in a batch
type Task func() (Action, error)

// BatchProcessor processes tasks in batches with concurrency
type BatchProcessor struct {
	// ctx is the context for cancellation
	ctx context.Context
	// tasks is the list of tasks to process
	tasks []Task
	// options contains the batch options
	options *BatchOptions
	// mu protects the tasks list
	mu sync.RWMutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(ctx context.Context, options *BatchOptions) *BatchProcessor {
	if options == nil {
		options = DefaultBatchOptions()
	}
	return &BatchProcessor{
		ctx:     ctx,
		tasks:   make([]Task, 0),
		options: options,
	}
}

// AddTask adds a task to the processor
func (p *BatchProcessor) AddTask(task Task) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tasks = append(p.tasks, task)
}

// Process processes all tasks with concurrency
func (p *BatchProcessor) Process() *BatchResult {
	p.mu.RLock()
	tasks := make([]Task, len(p.tasks))
	copy(tasks, p.tasks)
	p.mu.RUnlock()

	totalCount := len(tasks)
	if totalCount == 0 {
		return &BatchResult{
			TotalCount:   0,
			SuccessCount: 0,
			FailureCount: 0,
			Results:      []ActionResult{},
			Duration:     0,
			Errors:       []error{},
		}
	}

	// Create channels for tasks and results
	taskCh := make(chan Task, totalCount)
	resultCh := make(chan ActionResult, totalCount)
	doneCh := make(chan struct{})

	// Start worker goroutines
	var wg sync.WaitGroup
	concurrency := p.options.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				// Check if the context is canceled
				select {
				case <-p.ctx.Done():
					return
				default:
					// Continue processing
				}

				// Execute the task
				action, err := task()
				success := err == nil

				// Send the result
				resultCh <- ActionResult{
					Action:  action,
					Success: success,
					Error:   err,
				}
			}
		}()
	}

	// Start a goroutine to close the result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultCh)
		close(doneCh)
	}()

	// Send tasks to workers
	go func() {
		for _, task := range tasks {
			select {
			case taskCh <- task:
				// Task sent successfully
			case <-p.ctx.Done():
				// Context canceled, stop sending tasks
				break
			}
		}
		close(taskCh)
	}()

	// Collect results
	startTime := time.Now()
	results := make([]ActionResult, 0, totalCount)
	errors := make([]error, 0)
	successCount := 0
	failureCount := 0
	completed := 0

	for result := range resultCh {
		results = append(results, result)
		completed++

		if result.Success {
			successCount++
		} else {
			failureCount++
			if result.Error != nil {
				errors = append(errors, result.Error)
			}

			// Call the error callback if provided
			if p.options.ErrorCallback != nil {
				p.options.ErrorCallback(result.Action.NotificationID, result.Error)
			}
		}

		// Call the progress callback if provided
		if p.options.ProgressCallback != nil {
			p.options.ProgressCallback(completed, totalCount)
		}
	}

	duration := time.Since(startTime)

	return &BatchResult{
		TotalCount:   totalCount,
		SuccessCount: successCount,
		FailureCount: failureCount,
		Results:      results,
		Duration:     duration,
		Errors:       errors,
	}
}
