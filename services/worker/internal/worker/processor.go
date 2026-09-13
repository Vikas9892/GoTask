package worker

import (
	"context"
	"fmt"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
)

// NewDatabaseJobProcessor returns a JobProcessor that synchronizes state with PostgreSQL.
func NewDatabaseJobProcessor(repo repository.WorkerRepository, execRegistry *executor.Registry) JobProcessor {
	return func(ctx context.Context, queuedJob *model.Job) error {
		// 1. Load the job from PostgreSQL
		job, err := repo.GetJob(ctx, queuedJob.ID)
		if err != nil {
			return fmt.Errorf("failed to load job: %w", err)
		}

		// 2. Verify the job can be processed
		if job.Status != model.StatusPending && job.Status != model.StatusProcessing {
			return nil
		}

		// 3. Mark it processing
		if err := repo.MarkProcessing(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to transition job to processing: %w", err)
		}

		// 4. Execute the job
		execErr := execRegistry.Execute(ctx, job)
		if execErr != nil {
			// 6. Record failure on error
			_ = repo.MarkFailed(ctx, job.ID, execErr.Error())
			return execErr
		}

		// 5. Mark completed on success
		if err := repo.MarkCompleted(ctx, job.ID); err != nil {
			return fmt.Errorf("failed to mark job completed: %w", err)
		}

		return nil
	}
}
