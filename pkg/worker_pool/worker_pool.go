// Package workerpool provides a worker pool implementation that can process concurrent tasks with rate limiting.
package workerpool

import (
	"context"
	"sync"
)

// WorkerPool manages a fixed-size pool of goroutines that execute tasks concurrently.
type WorkerPool struct {
	jobs    chan func(ctx context.Context) error
	results chan error

	wg  sync.WaitGroup
	ctx context.Context

	startOnce sync.Once
	stopOnce  sync.Once
}

// New creates and starts a new WorkerPool with the given rateLimit (number of concurrent workers)
func New(rateLimit int, ctx context.Context) *WorkerPool {
	if rateLimit < 1 {
		rateLimit = 1
	}

	w := &WorkerPool{
		jobs:    make(chan func(ctx context.Context) error, rateLimit),
		results: make(chan error, rateLimit),
		ctx:     ctx,
	}

	w.startOnce.Do(func() {
		for i := 0; i < rateLimit; i++ {
			w.wg.Add(1)
			go func() {
				defer w.wg.Done()
				for {
					select {
					case job := <-w.jobs:
						w.results <- job(w.ctx)
					case <-w.ctx.Done():
						return
					}
				}
			}()
		}
	})
	return w
}

// Close shuts down the worker pool, waits for all workers to finish,
// and closes the jobs and results channels.
func (w *WorkerPool) Close() {
	w.stopOnce.Do(func() {
		w.wg.Wait()
		close(w.jobs)
		close(w.results)
	})
}

// AddTask submits a new job to the worker pool. The job must be a function that
// accepts a context and returns an error. If the context is done, the job is discarded.
func (w *WorkerPool) AddTask(j func(ctx context.Context) error) {
	select {
	case w.jobs <- j:
	case <-w.ctx.Done():
	}

}

// Result returns a receive-only channel through which job results (errors) can be collected.
func (w *WorkerPool) Result() <-chan error {
	return w.results
}
