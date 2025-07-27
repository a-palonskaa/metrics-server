package workerpool

import (
	"context"
	"sync"
)

type WorkerPool struct {
	jobs    chan func(ctx context.Context) interface{}
	results chan interface{}

	wg  sync.WaitGroup
	ctx context.Context

	startOnce sync.Once
	stopOnce  sync.Once
}

func New(rateLimit int, ctx context.Context) *WorkerPool {
	if rateLimit < 1 {
		rateLimit = 1
	}

	w := &WorkerPool{
		jobs:    make(chan func(ctx context.Context) interface{}, rateLimit),
		results: make(chan interface{}, rateLimit),
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

func (w *WorkerPool) Close() {
	w.stopOnce.Do(func() {
		w.wg.Wait()
		close(w.jobs)
		close(w.results)
	})
}

func (w *WorkerPool) AddTask(j func(ctx context.Context) interface{}) {
	select {
	case w.jobs <- j:
	case <-w.ctx.Done():
	}

}

func (w *WorkerPool) Result() <-chan interface{} {
	return w.results
}
