package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrJobNotFound = errors.New("job not found")

type JobRepository interface {
	CreateJob(ctx context.Context, job *model.Job) error
	GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error)
	ListJobs(ctx context.Context, limit, offset int) ([]*model.Job, int, error)
	DeleteJob(ctx context.Context, id uuid.UUID) error
	GetJobAttempts(ctx context.Context, jobID uuid.UUID) ([]*model.JobAttempt, error)
}

type PostgresJobRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresJobRepository(pool *pgxpool.Pool) *PostgresJobRepository {
	return &PostgresJobRepository{pool: pool}
}

func (r *PostgresJobRepository) CreateJob(ctx context.Context, job *model.Job) error {
	query := `
		INSERT INTO jobs (id, type, payload, status, attempts, max_attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(
		ctx, query,
		job.ID,
		job.Type,
		job.Payload,
		job.Status,
		job.Attempts,
		job.MaxAttempts,
		job.CreatedAt,
		job.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}
	return nil
}

func (r *PostgresJobRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
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

func (r *PostgresJobRepository) ListJobs(ctx context.Context, limit, offset int) ([]*model.Job, int, error) {
	countQuery := `SELECT COUNT(*) FROM jobs`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	query := `
		SELECT id, type, payload, status, attempts, max_attempts, last_error, created_at, updated_at, started_at, completed_at
		FROM jobs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*model.Job, 0)
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
			return nil, 0, fmt.Errorf("failed to scan job row: %w", err)
		}
		jobs = append(jobs, &job)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return jobs, total, nil
}

func (r *PostgresJobRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM jobs WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

func (r *PostgresJobRepository) GetJobAttempts(ctx context.Context, jobID uuid.UUID) ([]*model.JobAttempt, error) {
	query := `
		SELECT id, job_id, attempt, status, error, started_at, completed_at
		FROM job_attempts
		WHERE job_id = $1
		ORDER BY attempt ASC
	`
	rows, err := r.pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to query job attempts: %w", err)
	}
	defer rows.Close()

	attempts := make([]*model.JobAttempt, 0)
	for rows.Next() {
		var a model.JobAttempt
		err := rows.Scan(
			&a.ID,
			&a.JobID,
			&a.Attempt,
			&a.Status,
			&a.Error,
			&a.StartedAt,
			&a.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attempt row: %w", err)
		}
		attempts = append(attempts, &a)
	}
	return attempts, nil
}
