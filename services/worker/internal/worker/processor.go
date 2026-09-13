package worker

import (
	"context"
	"fmt"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
)

// NewDatabaseJobProcessor returns a JobProcessor that executes jobs and synchronizes state with PostgreSQL and retries.
func NewDatabaseJobProcessor(repo repository.WorkerRepository, execRegistry *executor.Registry, q *queue.Queue) JobProcessor {
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

		// 4. Execute the job
		execErr := execRegistry.Execute(ctx, job)
		if execErr != nil {
			// Check if attempts remain for retry
			if job.Attempts < job.MaxAttempts {
				if err := repo.MarkRetry(ctx, job.ID, execErr.Error()); err != nil {
					return fmt.Errorf("failed to mark job for retry: %w", err)
				}
				if q != nil {
					_ = q.Enqueue(ctx, job)
				}
				return execErr
			}

			// Permanently failed after maximum attempts
			if err := repo.MarkFailed(ctx, job.ID, execErr.Error()); err != nil {
				return fmt.Errorf("failed to mark job failed: %w", err)
			}
			return execErr
		}

		// 5. Mark completed on success
		if err := repo.MarkCompleted(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to mark job completed: %w", err)
		}

		return nil
	}
}
