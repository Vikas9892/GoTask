package worker

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/google/uuid"
)

func TestProcessor_SuccessOnFirstAttempt(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	registry := executor.NewDefaultRegistry()
	q := queue.NewQueue(10)
	proc := NewDatabaseJobProcessor(repo, registry, q)

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
		t.Fatalf("expected success on first attempt, got error: %v", err)
	}

	finalJob, _ := repo.GetJob(ctx, jobID)
	if finalJob.Status != model.StatusCompleted {
		t.Errorf("expected status completed, got %s", finalJob.Status)
	}
	if finalJob.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", finalJob.Attempts)
	}
}

func TestProcessor_SuccessAfterRetry(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	registry := executor.NewDefaultRegistry()
	q := queue.NewQueue(10)

	var executionCount int32
	mockExec := &mockFailingExecutor{
		failUntilAttempt: 2,
		count:            &executionCount,
	}
	registry.Register("custom", mockExec)

	proc := NewDatabaseJobProcessor(repo, registry, q)

	jobID := uuid.New()
	job := &model.Job{
		ID:          jobID,
		Type:        "custom",
		Payload:     json.RawMessage(`{}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.SaveJob(job)

	ctx := context.Background()

	// Attempt 1: Should fail and be queued for retry
	err := proc(ctx, job)
	if err == nil {
		t.Fatal("expected attempt 1 to fail, got nil")
	}

	jobAfter1, _ := repo.GetJob(ctx, jobID)
	if jobAfter1.Status != model.StatusPending {
		t.Errorf("expected job to be pending for retry, got %s", jobAfter1.Status)
	}
	if q.Size() != 1 {
		t.Errorf("expected job to be re-enqueued to queue, got size %d", q.Size())
	}

	// Attempt 2: Should succeed
	retriedJob := <-q.Jobs()
	err = proc(ctx, retriedJob)
	if err != nil {
		t.Fatalf("expected attempt 2 to succeed, got error: %v", err)
	}

	finalJob, _ := repo.GetJob(ctx, jobID)
	if finalJob.Status != model.StatusCompleted {
		t.Errorf("expected status completed, got %s", finalJob.Status)
	}
	if finalJob.Attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", finalJob.Attempts)
	}
}

func TestProcessor_PermanentFailureAfterMaxAttempts(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	registry := executor.NewDefaultRegistry()
	q := queue.NewQueue(10)
	proc := NewDatabaseJobProcessor(repo, registry, q)

	jobID := uuid.New()
	// Job with max_attempts = 1
	job := &model.Job{
		ID:          jobID,
		Type:        "email",
		Payload:     json.RawMessage(`{"to":"dev@example.com","simulate_failure":true}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 1,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.SaveJob(job)

	ctx := context.Background()
	err := proc(ctx, job)
	if err == nil {
		t.Fatal("expected failure, got nil")
	}

	finalJob, _ := repo.GetJob(ctx, jobID)
	if finalJob.Status != model.StatusFailed {
		t.Errorf("expected status failed, got %s", finalJob.Status)
	}
	if q.Size() != 0 {
		t.Errorf("expected queue to be empty (no retry), got size %d", q.Size())
	}
}

type mockFailingExecutor struct {
	failUntilAttempt int32
	count            *int32
}

func (m *mockFailingExecutor) Execute(ctx context.Context, job *model.Job) error {
	curr := atomic.AddInt32(m.count, 1)
	if curr < m.failUntilAttempt {
		return executor.ErrSimulatedFailure
	}
	return nil
}
