package repository

import (
	"context"
	"sync"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

type MockWorkerRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*model.Job
}

func NewMockWorkerRepository() *MockWorkerRepository {
	return &MockWorkerRepository{
		jobs: make(map[uuid.UUID]*model.Job),
	}
}

func (m *MockWorkerRepository) SaveJob(job *model.Job) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
}

func (m *MockWorkerRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	// Return copy
	clone := *job
	return &clone, nil
}

func (m *MockWorkerRepository) MarkProcessing(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return ErrJobNotFound
	}
	if job.Status != model.StatusPending && job.Status != model.StatusProcessing {
		return ErrInvalidStatusState
	}
	job.Status = model.StatusProcessing
	now := time.Now().UTC()
	if job.StartedAt == nil {
		job.StartedAt = &now
	}
	job.UpdatedAt = now
	job.Attempts++
	return nil
}

func (m *MockWorkerRepository) MarkCompleted(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return ErrJobNotFound
	}
	job.Status = model.StatusCompleted
	now := time.Now().UTC()
	job.CompletedAt = &now
	job.UpdatedAt = now
	return nil
}

func (m *MockWorkerRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		return ErrJobNotFound
	}
	job.Status = model.StatusFailed
	now := time.Now().UTC()
	job.CompletedAt = &now
	job.UpdatedAt = now
	job.LastError = &errMsg
	return nil
}

func (m *MockWorkerRepository) FindPendingJobs(ctx context.Context, limit int) ([]*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var pending []*model.Job
	for _, j := range m.jobs {
		if j.Status == model.StatusPending {
			clone := *j
			pending = append(pending, &clone)
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *MockWorkerRepository) ResetStaleProcessingJobs(ctx context.Context, threshold time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().UTC().Add(-threshold)
	var count int64
	for _, j := range m.jobs {
		if j.Status == model.StatusProcessing && j.UpdatedAt.Before(cutoff) {
			j.Status = model.StatusPending
			j.UpdatedAt = time.Now().UTC()
			count++
		}
	}
	return count, nil
}
