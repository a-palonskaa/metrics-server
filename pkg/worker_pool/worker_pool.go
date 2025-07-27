package workerpool

import (
	"context"
	"sync"
)

type WorkerPool[res any] struct {
	jobs    chan func(ctx context.Context) res
	results chan res

	wg  sync.WaitGroup
	ctx context.Context

	startOnce sync.Once
	stopOnce  sync.Once
}

func New[res any](rateLimit int, ctx context.Context) *WorkerPool[res] {
	if rateLimit < 1 {
		rateLimit = 1
	}

	w := &WorkerPool[res]{
		jobs:    make(chan func(ctx context.Context) res, rateLimit),
		results: make(chan res, rateLimit),
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

func (w *WorkerPool[res]) Close() {
	w.stopOnce.Do(func() {
		w.wg.Wait()
		close(w.jobs)
		close(w.results)
	})
}

func (w *WorkerPool[res]) AddTask(j func(ctx context.Context) res) {
	select {
	case w.jobs <- j:
	case <-w.ctx.Done():
	}

}

func (w *WorkerPool[res]) Result() <-chan res {
	return w.results
}
