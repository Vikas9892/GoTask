package repository

import (
	"context"
	"sync"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

// MockJobRepository is an in-memory implementation of JobRepository for testing.
type MockJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*model.Job
}

func NewMockJobRepository() *MockJobRepository {
	return &MockJobRepository{
		jobs: make(map[uuid.UUID]*model.Job),
	}
}

func (m *MockJobRepository) CreateJob(ctx context.Context, job *model.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (m *MockJobRepository) ListJobs(ctx context.Context, limit, offset int) ([]*model.Job, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]*model.Job, 0, len(m.jobs))
	for _, j := range m.jobs {
		all = append(all, j)
	}

	total := len(all)
	if offset >= total {
		return []*model.Job{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *MockJobRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[id]; !ok {
		return ErrJobNotFound
	}
	delete(m.jobs, id)
	return nil
}

func (m *MockJobRepository) GetJobAttempts(ctx context.Context, jobID uuid.UUID) ([]*model.JobAttempt, error) {
	return []*model.JobAttempt{}, nil
}
