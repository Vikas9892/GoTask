package worker

import (
	"context"
	"sync"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
)

type JobProcessor func(ctx context.Context, job *model.Job) error

type Pool struct {
	workerCount int
	queue       *queue.Queue
	processor   JobProcessor
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewPool(workerCount int, q *queue.Queue, processor JobProcessor) *Pool {
	if workerCount <= 0 {
		workerCount = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		workerCount: workerCount,
		queue:       q,
		processor:   processor,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start spawns worker goroutines to consume and process jobs concurrently.
func (p *Pool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.workerLoop(i + 1)
	}
}

func (p *Pool) workerLoop(workerID int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.queue.Jobs():
			if !ok {
				return
			}
			if job == nil {
				continue
			}

			_ = p.processor(p.ctx, job)
		}
	}
}

// Stop signals all workers to finish and waits for goroutines to exit.
func (p *Pool) Stop() {
	p.cancel()
	p.wg.Wait()
}
