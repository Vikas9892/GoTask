package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/Vikas9892/GoTask/services/worker/internal/worker"
	"github.com/google/uuid"
)

func TestWorker_EndToEndProcessing(t *testing.T) {
	repo := repository.NewMockWorkerRepository()
	jobQueue := queue.NewQueue(10)
	execRegistry := executor.NewDefaultRegistry()

	processor := worker.NewDatabaseJobProcessor(repo, execRegistry, jobQueue, 5*time.Second)
	pool := worker.NewPool(2, jobQueue, processor)
	pool.Start()
	defer pool.Stop()

	// Seed job in repository and enqueue
	jobID := uuid.New()
	job := &model.Job{
		ID:          jobID,
		Type:        "email",
		Payload:     json.RawMessage(`{"to":"worker_e2e@example.com"}`),
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	repo.SaveJob(job)

	ctx := context.Background()
	if err := jobQueue.Enqueue(ctx, job); err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// Wait for processing
	deadline := time.Now().Add(5 * time.Second)
	var finalJob *model.Job
	for time.Now().Before(deadline) {
		j, err := repo.GetJob(ctx, jobID)
		if err == nil && j.Status == model.StatusCompleted {
			finalJob = j
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if finalJob == nil {
		t.Fatal("timed out waiting for worker to complete job")
	}

	if finalJob.Status != model.StatusCompleted {
		t.Errorf("expected status completed, got %s", finalJob.Status)
	}
	if finalJob.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", finalJob.Attempts)
	}
}
