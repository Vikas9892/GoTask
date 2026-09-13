package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/google/uuid"
)

const MaxPayloadSize = 64 * 1024 // 64 KB

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrPayloadTooLarge    = errors.New("payload exceeds maximum size of 64KB")
	ErrInvalidJSONPayload = errors.New("payload must be valid JSON")
)

type CreateJobInput struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JobService struct {
	repo repository.JobRepository
}

func NewJobService(repo repository.JobRepository) *JobService {
	return &JobService{repo: repo}
}

func (s *JobService) CreateJob(ctx context.Context, input CreateJobInput) (*model.Job, error) {
	if !model.IsSupportedJobType(input.Type) {
		return nil, fmt.Errorf("%w: unsupported job type '%s' (supported: email, webhook, report)", ErrInvalidInput, input.Type)
	}

	if len(input.Payload) == 0 {
		return nil, fmt.Errorf("%w: payload cannot be empty", ErrInvalidInput)
	}

	if len(input.Payload) > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	if !json.Valid(input.Payload) {
		return nil, ErrInvalidJSONPayload
	}

	now := time.Now().UTC()
	job := &model.Job{
		ID:          uuid.New(),
		Type:        input.Type,
		Payload:     input.Payload,
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

func (s *JobService) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func (s *JobService) ListJobs(ctx context.Context, limit, offset int) ([]*model.Job, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListJobs(ctx, limit, offset)
}

func (s *JobService) DeleteJob(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteJob(ctx, id)
}
