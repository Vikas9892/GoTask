package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/Vikas9892/GoTask/internal/metrics"
	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
)

// NewDatabaseJobProcessor returns a JobProcessor that executes jobs with a timeout, tracks metrics, and synchronizes state with PostgreSQL.
func NewDatabaseJobProcessor(repo repository.WorkerRepository, execRegistry *executor.Registry, q *queue.Queue, jobTimeout time.Duration) JobProcessor {
	if jobTimeout <= 0 {
		jobTimeout = 30 * time.Second
	}

	return func(ctx context.Context, queuedJob *model.Job) error {
		// 1. Load fresh job record from PostgreSQL
		job, err := repo.GetJob(ctx, queuedJob.ID)
		if err != nil {
			return fmt.Errorf("failed to load job: %w", err)
		}

		// 2. Verify the job can be processed
		if job.Status != model.StatusPending && job.Status != model.StatusProcessing {
			return nil
		}

		// 3. Mark it processing (increments attempts)
		if err := repo.MarkProcessing(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to transition job to processing: %w", err)
		}
		job.Attempts++

		// Record initial job_attempt
		attemptRecord := &model.JobAttempt{
			JobID:     job.ID,
			Attempt:   job.Attempts,
			Status:    model.StatusProcessing,
			StartedAt: time.Now().UTC(),
		}
		attemptID, _ := repo.CreateAttempt(ctx, attemptRecord)

		// Metrics tracking
		metrics.ActiveWorkers.Inc()
		defer metrics.ActiveWorkers.Dec()
		start := time.Now()

		// 4. Execute the job with timeout
		execCtx, cancel := context.WithTimeout(ctx, jobTimeout)
		execErr := execRegistry.Execute(execCtx, job)
		cancel()

		metrics.JobProcessingDuration.Observe(time.Since(start).Seconds())

		if execErr != nil {
			errStr := execErr.Error()
			if attemptID > 0 {
				_ = repo.UpdateAttempt(ctx, attemptID, model.StatusFailed, &errStr)
			}

			// Check if attempts remain for retry
			if job.Attempts < job.MaxAttempts {
				metrics.JobsRetriedTotal.Inc()
				if err := repo.MarkRetry(ctx, job.ID, errStr); err != nil {
					return fmt.Errorf("failed to mark job for retry: %w", err)
				}
				if q != nil {
					_ = q.Enqueue(ctx, job)
				}
				return execErr
			}

			// Permanently failed after maximum attempts
			metrics.JobsFailedTotal.Inc()
			if err := repo.MarkFailed(ctx, job.ID, errStr); err != nil {
				return fmt.Errorf("failed to mark job failed: %w", err)
			}
			return execErr
		}

		// 5. Mark completed on success
		metrics.JobsCompletedTotal.Inc()
		if attemptID > 0 {
			_ = repo.UpdateAttempt(ctx, attemptID, model.StatusCompleted, nil)
		}
		if err := repo.MarkCompleted(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to mark job completed: %w", err)
		}

		return nil
	}
}
