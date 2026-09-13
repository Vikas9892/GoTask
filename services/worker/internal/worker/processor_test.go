package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/google/uuid"
)

func TestProcessor_ExecutionLifecycle_Success(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	registry := executor.NewDefaultRegistry()
	proc := NewDatabaseJobProcessor(repo, registry)

	jobID := uuid.New()
	initialJob := &model.Job{
		ID:          jobID,
		Type:        "email",
		Payload:     json.RawMessage(`{"to":"dev@example.com"}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.SaveJob(initialJob)

	ctx := context.Background()
	err := proc(ctx, initialJob)
	if err != nil {
		t.Fatalf("expected successful execution, got error: %v", err)
	}

	finalJob, err := repo.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("failed to get job from repo: %v", err)
	}
	if finalJob.Status != model.StatusCompleted {
		t.Errorf("expected status completed, got %s", finalJob.Status)
	}
	if finalJob.Attempts != 1 {
		t.Errorf("expected attempts 1, got %d", finalJob.Attempts)
	}
	if finalJob.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestProcessor_ExecutionLifecycle_Failure(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	registry := executor.NewDefaultRegistry()
	proc := NewDatabaseJobProcessor(repo, registry)

	jobID := uuid.New()
	initialJob := &model.Job{
		ID:          jobID,
		Type:        "email",
		Payload:     json.RawMessage(`{"to":"dev@example.com","simulate_failure":true}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.SaveJob(initialJob)

	ctx := context.Background()
	err := proc(ctx, initialJob)
	if err == nil {
		t.Fatal("expected execution error, got nil")
	}

	finalJob, err := repo.GetJob(ctx, jobID)
	if err != nil {
		t.Fatalf("failed to get job from repo: %v", err)
	}
	if finalJob.Status != model.StatusFailed {
		t.Errorf("expected status failed, got %s", finalJob.Status)
	}
	if finalJob.LastError == nil || *finalJob.LastError == "" {
		t.Error("expected last_error to be populated")
	}
}
