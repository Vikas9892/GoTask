package recovery

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/google/uuid"
)

func TestRecoverer_PendingJobsRecovery(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	q := queue.NewQueue(10)
	rec := NewRecoverer(repo, q, 5*time.Minute)

	// Seed 3 pending jobs
	for i := 0; i < 3; i++ {
		repo.SaveJob(&model.Job{
			ID:        uuid.New(),
			Type:      "email",
			Payload:   json.RawMessage(`{"to":"test@example.com"}`),
			Status:    model.StatusPending,
			CreatedAt: time.Now().UTC().Add(-time.Duration(i) * time.Minute),
			UpdatedAt: time.Now().UTC(),
		})
	}

	ctx := context.Background()
	enqueued, err := rec.Recover(ctx)
	if err != nil {
		t.Fatalf("unexpected recovery error: %v", err)
	}

	if enqueued != 3 {
		t.Errorf("expected 3 recovered jobs, got %d", enqueued)
	}
	if q.Size() != 3 {
		t.Errorf("expected queue size 3, got %d", q.Size())
	}
}

func TestRecoverer_StaleProcessingRecovery(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	q := queue.NewQueue(10)
	staleThreshold := 2 * time.Minute
	rec := NewRecoverer(repo, q, staleThreshold)

	// Stale job (started 10 minutes ago, status processing)
	staleID := uuid.New()
	staleJob := &model.Job{
		ID:        staleID,
		Type:      "webhook",
		Payload:   json.RawMessage(`{"url":"https://example.com"}`),
		Status:    model.StatusProcessing,
		UpdatedAt: time.Now().UTC().Add(-10 * time.Minute),
	}
	repo.SaveJob(staleJob)

	// Fresh processing job (started 30 seconds ago)
	freshID := uuid.New()
	freshJob := &model.Job{
		ID:        freshID,
		Type:      "report",
		Payload:   json.RawMessage(`{}`),
		Status:    model.StatusProcessing,
		UpdatedAt: time.Now().UTC().Add(-30 * time.Second),
	}
	repo.SaveJob(freshJob)

	ctx := context.Background()
	enqueued, err := rec.Recover(ctx)
	if err != nil {
		t.Fatalf("unexpected recovery error: %v", err)
	}

	// The stale job should have been reset to pending and enqueued
	if enqueued != 1 {
		t.Errorf("expected 1 recovered job (the reset stale job), got %d", enqueued)
	}

	// Check repository states
	checkStale, _ := repo.GetJob(ctx, staleID)
	if checkStale.Status != model.StatusPending {
		t.Errorf("expected stale job status pending, got %s", checkStale.Status)
	}

	checkFresh, _ := repo.GetJob(ctx, freshID)
	if checkFresh.Status != model.StatusProcessing {
		t.Errorf("expected fresh job status processing, got %s", checkFresh.Status)
	}
}
