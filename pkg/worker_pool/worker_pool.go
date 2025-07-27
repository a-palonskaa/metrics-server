package workerpool

import (
	"context"
	"sync"
)

type WorkerPool[job any, res any] struct {
	jobs      chan job
	results   chan res
	workerNum int

	wg  sync.WaitGroup
	ctx context.Context

	startOnce sync.Once
	stopOnce  sync.Once
}

func New[job any, res any](rateLimit int, ctx context.Context) *WorkerPool[job, res] {
	if rateLimit < 1 {
		rateLimit = 1
	}

	return &WorkerPool[job, res]{
		jobs:      make(chan job, rateLimit),
		results:   make(chan res, rateLimit),
		workerNum: rateLimit,
		ctx:       ctx,
	}
}

func (w *WorkerPool[job, res]) Close() {
	w.stopOnce.Do(func() {
		w.wg.Wait()
		close(w.jobs)
		close(w.results)
	})
}

func (w *WorkerPool[job, res]) Start(fn func(context.Context, job) res) {
	w.startOnce.Do(func() {
		for i := 0; i < w.workerNum; i++ {
			w.wg.Add(1)
			go func() {
				defer w.wg.Done()
				for {
					select {
					case job := <-w.jobs:
						w.results <- fn(w.ctx, job)
					case <-w.ctx.Done():
						return
					}
				}
			}()
		}
	})
}

func (w *WorkerPool[job, res]) AddTask(j job) {
	select {
	case w.jobs <- j:
	case <-w.ctx.Done():
	}

}

func (w *WorkerPool[job, res]) Result() <-chan res {
	return w.results
}
