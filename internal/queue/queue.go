// Package queue provides a job queue with worker pool.
package queue

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// Job represents a unit of work.
type Job struct {
	ID      string
	Handler func(ctx context.Context) error
	Created time.Time
}

// Queue manages background job processing.
type Queue struct {
	jobs    chan *Job
	workers int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	results map[string]error
	mu      sync.RWMutex
}

// New creates a new job queue.
func New(workers int) *Queue {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Queue{
		jobs:    make(chan *Job, workers*10),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
		results: make(map[string]error),
	}
}

// Start starts the worker pool.
func (q *Queue) Start() {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// worker processes jobs from the queue.
func (q *Queue) worker(id int) {
	defer q.wg.Done()
	
	for {
		select {
		case <-q.ctx.Done():
			return
		case job, ok := <-q.jobs:
			if !ok {
				return
			}
			err := job.Handler(q.ctx)
			q.mu.Lock()
			q.results[job.ID] = err
			q.mu.Unlock()
		}
	}
}

// Submit adds a job to the queue.
func (q *Queue) Submit(job *Job) error {
	select {
	case <-q.ctx.Done():
		return q.ctx.Err()
	case q.jobs <- job:
		return nil
	}
}

// Result returns the result of a job.
func (q *Queue) Result(jobID string) (error, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	err, ok := q.results[jobID]
	return err, ok
}

// Stop stops the queue and waits for workers to finish.
func (q *Queue) Stop() {
	q.cancel()
	close(q.jobs)
	q.wg.Wait()
}

// Pending returns the number of pending jobs.
func (q *Queue) Pending() int {
	return len(q.jobs)
}

// Workers returns the number of workers.
func (q *Queue) Workers() int {
	return q.workers
}
