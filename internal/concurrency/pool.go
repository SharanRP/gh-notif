package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
)

// Task represents a task to be executed by the worker pool
type Task func() (interface{}, error)

// Result represents the result of a task
type Result struct {
	// Value is the result of the task
	Value interface{}

	// Error is any error that occurred during task execution
	Error error

	// Index is the index of the task in the original task list
	Index int
}

// WorkerPool is a pool of workers for executing tasks concurrently
type WorkerPool struct {
	// workers is the number of workers in the pool
	workers int

	// taskQueue is the channel for sending tasks to workers
	taskQueue chan taskWrapper

	// resultQueue is the channel for receiving results from workers
	resultQueue chan Result

	// wg is used to wait for all workers to complete
	wg sync.WaitGroup

	// ctx is the context for the worker pool
	ctx context.Context

	// cancel is the function to cancel the context
	cancel context.CancelFunc

	// activeWorkers is the number of active workers
	activeWorkers atomic.Int32

	// queuedTasks is the number of queued tasks
	queuedTasks atomic.Int32

	// completedTasks is the number of completed tasks
	completedTasks atomic.Int32
}

// taskWrapper wraps a task with its index
type taskWrapper struct {
	task  Task
	index int
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	if workers <= 0 {
		workers = 1
	}

	if queueSize <= 0 {
		queueSize = workers * 2
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:     workers,
		taskQueue:   make(chan taskWrapper, queueSize),
		resultQueue: make(chan Result, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	return pool
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// Stop stops the worker pool
func (p *WorkerPool) Stop() {
	p.cancel()
	p.wg.Wait()
	close(p.resultQueue)
}

// Submit submits a task to the worker pool
func (p *WorkerPool) Submit(task Task) {
	p.queuedTasks.Add(1)
	p.taskQueue <- taskWrapper{task: task, index: int(p.queuedTasks.Load() - 1)}
}

// SubmitBatch submits a batch of tasks to the worker pool
func (p *WorkerPool) SubmitBatch(tasks []Task) {
	for _, task := range tasks {
		p.Submit(task)
	}
}

// Results returns a channel for receiving results
func (p *WorkerPool) Results() <-chan Result {
	return p.resultQueue
}

// Wait waits for all tasks to complete
func (p *WorkerPool) Wait() {
	close(p.taskQueue)
	p.wg.Wait()
	close(p.resultQueue)
}

// worker is the worker goroutine
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	p.activeWorkers.Add(1)
	defer p.activeWorkers.Add(-1)

	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}

			// Execute the task
			value, err := task.task()

			// Send the result
			select {
			case <-p.ctx.Done():
				return
			case p.resultQueue <- Result{Value: value, Error: err, Index: task.index}:
				p.completedTasks.Add(1)
			}
		}
	}
}

// ExecuteBatch executes a batch of tasks and returns the results
func ExecuteBatch(ctx context.Context, tasks []Task, workers int) ([]Result, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	// Create a worker pool
	pool := NewWorkerPool(workers, len(tasks))

	// Create a context that's canceled when the parent context is canceled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start the worker pool
	pool.Start()

	// Submit tasks
	go func() {
		defer pool.Wait()
		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			default:
				pool.Submit(task)
			}
		}
	}()

	// Collect results
	results := make([]Result, len(tasks))
	for result := range pool.Results() {
		results[result.Index] = result

		// Check if the context is canceled
		select {
		case <-ctx.Done():
			pool.Stop()
			return results, ctx.Err()
		default:
			// Continue
		}
	}

	return results, nil
}
