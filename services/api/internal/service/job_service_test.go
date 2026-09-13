package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/google/uuid"
)

func TestJobService_CreateJob_Valid(t *testing.T) {
	repo := repository.NewMockJobRepository()
	svc := NewJobService(repo)
	ctx := context.Background()

	input := CreateJobInput{
		Type:    "email",
		Payload: json.RawMessage(`{"to":"alice@example.com","subject":"Welcome"}`),
	}

	job, err := svc.CreateJob(ctx, input)
	if err != nil {
		t.Fatalf("expected successful job creation, got error: %v", err)
	}

	if job.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if job.Status != model.StatusPending {
		t.Errorf("expected pending status, got %s", job.Status)
	}
	if job.Attempts != 0 {
		t.Errorf("expected 0 attempts, got %d", job.Attempts)
	}
}

func TestJobService_CreateJob_Validations(t *testing.T) {
	repo := repository.NewMockJobRepository()
	svc := NewJobService(repo)
	ctx := context.Background()

	// Unsupported type
	_, err := svc.CreateJob(ctx, CreateJobInput{
		Type:    "unsupported-type",
		Payload: json.RawMessage(`{"key":"value"}`),
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}

	// Empty payload
	_, err = svc.CreateJob(ctx, CreateJobInput{
		Type:    "webhook",
		Payload: nil,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput for empty payload, got %v", err)
	}

	// Invalid JSON payload
	_, err = svc.CreateJob(ctx, CreateJobInput{
		Type:    "report",
		Payload: json.RawMessage(`{not-valid-json`),
	})
	if !errors.Is(err, ErrInvalidJSONPayload) {
		t.Errorf("expected ErrInvalidJSONPayload, got %v", err)
	}

	// Payload too large (> 64KB)
	largeData := strings.Repeat("x", MaxPayloadSize+100)
	largeJSON, _ := json.Marshal(map[string]string{"data": largeData})
	_, err = svc.CreateJob(ctx, CreateJobInput{
		Type:    "email",
		Payload: largeJSON,
	})
	if !errors.Is(err, ErrPayloadTooLarge) {
		t.Errorf("expected ErrPayloadTooLarge, got %v", err)
	}
}

func TestJobService_Get_List_Delete(t *testing.T) {
	repo := repository.NewMockJobRepository()
	svc := NewJobService(repo)
	ctx := context.Background()

	created, err := svc.CreateJob(ctx, CreateJobInput{
		Type:    "webhook",
		Payload: json.RawMessage(`{"url":"https://example.com/hook"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get
	found, err := svc.GetJob(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if found.ID != created.ID {
		t.Errorf("expected ID %v, got %v", created.ID, found.ID)
	}

	// List
	jobs, total, err := svc.ListJobs(ctx, 10, 0)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if total != 1 || len(jobs) != 1 {
		t.Errorf("expected 1 job, got total=%d len=%d", total, len(jobs))
	}

	// Delete
	if err := svc.DeleteJob(ctx, created.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}

	// Get non-existent
	_, err = svc.GetJob(ctx, created.ID)
	if !errors.Is(err, repository.ErrJobNotFound) {
		t.Errorf("expected ErrJobNotFound, got %v", err)
	}
}
