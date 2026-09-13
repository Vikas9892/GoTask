package worker

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/google/uuid"
)

type fastBenchmarkExecutor struct {
	processedCount int32
}

func (f *fastBenchmarkExecutor) Execute(ctx context.Context, job *model.Job) error {
	// Simulate lightweight CPU/IO work: 1ms per job
	time.Sleep(1 * time.Millisecond)
	atomic.AddInt32(&f.processedCount, 1)
	return nil
}

func TestWorkerPool_Throughput(t *testing.T) {
	workerCounts := []int{2, 5, 10}
	totalJobs := 1000

	for _, wc := range workerCounts {
		q := queue.NewQueue(1000)
		repo := repository.NewMockWorkerRepository()
		execRegistry := executor.NewDefaultRegistry()
		benchExec := &fastBenchmarkExecutor{}
		execRegistry.Register("benchmark", benchExec)

		proc := NewDatabaseJobProcessor(repo, execRegistry, q, 5*time.Second)
		pool := NewPool(wc, q, proc)
		pool.Start()

		ctx := context.Background()

		// 1. Measure submission
		subStart := time.Now()
		for i := 0; i < totalJobs; i++ {
			job := &model.Job{
				ID:          uuid.New(),
				Type:        "benchmark",
				Payload:     json.RawMessage(`{}`),
				Status:      model.StatusPending,
				MaxAttempts: 3,
			}
			repo.SaveJob(job)
			_ = q.Enqueue(ctx, job)
		}
		subDuration := time.Since(subStart)

		// 2. Measure processing
		procStart := time.Now()
		for atomic.LoadInt32(&benchExec.processedCount) < int32(totalJobs) {
			time.Sleep(5 * time.Millisecond)
		}
		procDuration := time.Since(procStart)
		pool.Stop()

		subThroughput := float64(totalJobs) / subDuration.Seconds()
		procThroughput := float64(totalJobs) / procDuration.Seconds()

		t.Logf("Workers: %2d | Submit: %.1f jobs/sec (%v) | Process: %.1f jobs/sec (%v)",
			wc, subThroughput, subDuration, procThroughput, procDuration)
	}
}
