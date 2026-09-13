package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrJobNotFound        = errors.New("job not found")
	ErrInvalidStatusState = errors.New("job is not in a valid state to process")
)

type WorkerRepository interface {
	GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error)
	MarkProcessing(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
	MarkRetry(ctx context.Context, id uuid.UUID, errMsg string) error
	CreateAttempt(ctx context.Context, attempt *model.JobAttempt) (int64, error)
	UpdateAttempt(ctx context.Context, attemptID int64, status model.JobStatus, errMsg *string) error
	FindPendingJobs(ctx context.Context, limit int) ([]*model.Job, error)
	ResetStaleProcessingJobs(ctx context.Context, threshold time.Duration) (int64, error)
}

type PostgresWorkerRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWorkerRepository(pool *pgxpool.Pool) *PostgresWorkerRepository {
	return &PostgresWorkerRepository{pool: pool}
}

func (r *PostgresWorkerRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	query := `
		SELECT id, type, payload, status, attempts, max_attempts, last_error, created_at, updated_at, started_at, completed_at
		FROM jobs
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var job model.Job
	err := row.Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.Attempts,
		&job.MaxAttempts,
		&job.LastError,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to query job: %w", err)
	}
	return &job, nil
}

func (r *PostgresWorkerRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = 'processing',
		    started_at = COALESCE(started_at, NOW()),
		    updated_at = NOW(),
		    attempts = attempts + 1
		WHERE id = $1 AND (status = 'pending' OR status = 'processing')
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark job processing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInvalidStatusState
	}
	return nil
}

func (r *PostgresWorkerRepository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE jobs
		SET status = 'completed',
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark job completed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (r *PostgresWorkerRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE jobs
		SET status = 'failed',
		    last_error = $2,
		    completed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, errMsg)
	if err != nil {
		return fmt.Errorf("failed to mark job failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (r *PostgresWorkerRepository) MarkRetry(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE jobs
		SET status = 'pending',
		    last_error = $2,
		    updated_at = NOW()
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, errMsg)
	if err != nil {
		return fmt.Errorf("failed to mark job for retry: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (r *PostgresWorkerRepository) CreateAttempt(ctx context.Context, attempt *model.JobAttempt) (int64, error) {
	query := `
		INSERT INTO job_attempts (job_id, attempt, status, started_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id int64
	err := r.pool.QueryRow(ctx, query, attempt.JobID, attempt.Attempt, attempt.Status, attempt.StartedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert job attempt: %w", err)
	}
	return id, nil
}

func (r *PostgresWorkerRepository) UpdateAttempt(ctx context.Context, attemptID int64, status model.JobStatus, errMsg *string) error {
	query := `
		UPDATE job_attempts
		SET status = $2,
		    error = $3,
		    completed_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, attemptID, status, errMsg)
	if err != nil {
		return fmt.Errorf("failed to update job attempt: %w", err)
	}
	return nil
}

func (r *PostgresWorkerRepository) FindPendingJobs(ctx context.Context, limit int) ([]*model.Job, error) {
	query := `
		SELECT id, type, payload, status, attempts, max_attempts, last_error, created_at, updated_at, started_at, completed_at
		FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*model.Job
	for rows.Next() {
		var job model.Job
		err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Payload,
			&job.Status,
			&job.Attempts,
			&job.MaxAttempts,
			&job.LastError,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending job: %w", err)
		}
		jobs = append(jobs, &job)
	}
	return jobs, nil
}

func (r *PostgresWorkerRepository) ResetStaleProcessingJobs(ctx context.Context, threshold time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-threshold)
	query := `
		UPDATE jobs
		SET status = 'pending',
		    updated_at = NOW()
		WHERE status = 'processing' AND updated_at < $1
	`
	tag, err := r.pool.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to reset stale processing jobs: %w", err)
	}
	return tag.RowsAffected(), nil
}
