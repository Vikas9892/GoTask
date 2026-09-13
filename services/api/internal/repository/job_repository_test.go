package repository

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

// InMemoryJobRepository is a test implementation of JobRepository.
type InMemoryJobRepository struct {
	mu   sync.RWMutex
	jobs map[uuid.UUID]*model.Job
}

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		jobs: make(map[uuid.UUID]*model.Job),
	}
}

func (m *InMemoryJobRepository) CreateJob(ctx context.Context, job *model.Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *InMemoryJobRepository) GetJob(ctx context.Context, id uuid.UUID) (*model.Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	job, ok := m.jobs[id]
	if !ok {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (m *InMemoryJobRepository) ListJobs(ctx context.Context, limit, offset int) ([]*model.Job, int, error) {
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

func (m *InMemoryJobRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.jobs[id]; !ok {
		return ErrJobNotFound
	}
	delete(m.jobs, id)
	return nil
}

func TestJobRepository_Operations(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryJobRepository()

	id := uuid.New()
	job := &model.Job{
		ID:          id,
		Type:        "email",
		Payload:     json.RawMessage(`{"to":"test@example.com"}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Create
	if err := repo.CreateJob(ctx, job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	// Get
	found, err := repo.GetJob(ctx, id)
	if err != nil {
		t.Fatalf("failed to get job: %v", err)
	}
	if found.ID != id {
		t.Errorf("expected ID %v, got %v", id, found.ID)
	}

	// List
	list, total, err := repo.ListJobs(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list jobs: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Errorf("expected 1 job, got total=%d len=%d", total, len(list))
	}

	// Delete
	if err := repo.DeleteJob(ctx, id); err != nil {
		t.Fatalf("failed to delete job: %v", err)
	}

	// Get Not Found
	_, err = repo.GetJob(ctx, id)
	if err != ErrJobNotFound {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}
