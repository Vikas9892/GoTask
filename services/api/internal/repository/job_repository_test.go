package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

func TestJobRepository_Operations(t *testing.T) {
	ctx := context.Background()
	repo := NewMockJobRepository()

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
