package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/google/uuid"
)

func TestPool_ConcurrentProcessing(t *testing.T) {
	q := queue.NewQueue(20)
	var processedCount int32
	var maxConcurrent int32
	var activeNow int32

	processor := func(ctx context.Context, job *model.Job) error {
		current := atomic.AddInt32(&activeNow, 1)
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt32(&maxConcurrent, max, current) {
				break
			}
		}

		time.Sleep(30 * time.Millisecond) // simulate work

		atomic.AddInt32(&activeNow, -1)
		atomic.AddInt32(&processedCount, 1)
		return nil
	}

	workerCount := 4
	pool := NewPool(workerCount, q, processor)
	pool.Start()

	totalJobs := 12
	ctx := context.Background()
	for i := 0; i < totalJobs; i++ {
		err := q.Enqueue(ctx, &model.Job{
			ID:   uuid.New(),
			Type: "email",
		})
		if err != nil {
			t.Fatalf("failed to enqueue job %d: %v", i, err)
		}
	}

	// Wait for jobs to finish
	deadline := time.Now().Add(2 * time.Second)
	for atomic.LoadInt32(&processedCount) < int32(totalJobs) {
		if time.Now().After(deadline) {
			t.Fatalf("timed out: processed %d of %d jobs", atomic.LoadInt32(&processedCount), totalJobs)
		}
		time.Sleep(10 * time.Millisecond)
	}

	pool.Stop()

	if atomic.LoadInt32(&maxConcurrent) < 2 {
		t.Errorf("expected concurrent execution (maxConcurrent >= 2), got %d", atomic.LoadInt32(&maxConcurrent))
	}
}

func TestPool_GracefulShutdown(t *testing.T) {
	q := queue.NewQueue(10)
	var processedCount int32
	var wg sync.WaitGroup
	wg.Add(1)

	processor := func(ctx context.Context, job *model.Job) error {
		atomic.AddInt32(&processedCount, 1)
		wg.Done()
		return nil
	}

	pool := NewPool(2, q, processor)
	pool.Start()

	_ = q.Enqueue(context.Background(), &model.Job{ID: uuid.New(), Type: "report"})
	wg.Wait()

	pool.Stop()

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("expected 1 processed job, got %d", atomic.LoadInt32(&processedCount))
	}
}
