// Package recovery handles restoring pending and abandoned jobs on worker startup.
//
// Failure Scenario Solved:
// Because the queue is an in-memory Go channel, if the Worker Service process crashes,
// is terminated (e.g. OOM, SIGKILL, server restart), or restarts during deployment:
// 1. In-flight jobs that were marked 'processing' are left orphaned in the database.
// 2. Unprocessed jobs stored in memory channels are lost from RAM.
//
// Recovery guarantees that no jobs are permanently lost by:
// 1. Identifying jobs stuck in 'processing' beyond a reasonable timeout threshold and resetting them to 'pending'.
// 2. Fetching all 'pending' jobs from PostgreSQL and re-enqueuing them into the bounded Go channel queue before starting workers.
package recovery

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
)

type Recoverer struct {
	repo           repository.WorkerRepository
	queue          *queue.Queue
	staleThreshold time.Duration
}

func NewRecoverer(repo repository.WorkerRepository, q *queue.Queue, staleThreshold time.Duration) *Recoverer {
	if staleThreshold <= 0 {
		staleThreshold = 5 * time.Minute
	}
	return &Recoverer{
		repo:           repo,
		queue:          q,
		staleThreshold: staleThreshold,
	}
}

// Recover resets stale in-flight jobs and enqueues all pending jobs into the in-process queue.
func (r *Recoverer) Recover(ctx context.Context) (int, error) {
	// 1. Reset stale processing jobs
	staleCount, err := r.repo.ResetStaleProcessingJobs(ctx, r.staleThreshold)
	if err != nil {
		return 0, fmt.Errorf("failed to reset stale processing jobs: %w", err)
	}
	if staleCount > 0 {
		slog.Info("reset stale processing jobs to pending", "count", staleCount)
	}

	// 2. Fetch pending jobs up to queue capacity
	pendingJobs, err := r.repo.FindPendingJobs(ctx, r.queue.Capacity())
	if err != nil {
		return 0, fmt.Errorf("failed to load pending jobs: %w", err)
	}

	// 3. Enqueue into bounded Go channel
	enqueued := 0
	for _, job := range pendingJobs {
		if err := r.queue.Enqueue(ctx, job); err != nil {
			slog.Warn("queue buffer full during startup recovery", "enqueued", enqueued, "error", err)
			break
		}
		enqueued++
	}

	return enqueued, nil
}
