package executor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

func TestExecutors_Success(t *testing.T) {
	registry := NewDefaultRegistry()
	ctx := context.Background()

	tests := []struct {
		name    string
		jobType string
		payload string
	}{
		{"email", "email", `{"to":"user@domain.com"}`},
		{"webhook", "webhook", `{"url":"https://example.com/api"}`},
		{"report", "report", `{"format":"pdf"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &model.Job{
				ID:      uuid.New(),
				Type:    tt.jobType,
				Payload: json.RawMessage(tt.payload),
			}
			if err := registry.Execute(ctx, job); err != nil {
				t.Fatalf("expected success for %s, got %v", tt.name, err)
			}
		})
	}
}

func TestExecutors_FailuresAndValidation(t *testing.T) {
	registry := NewDefaultRegistry()
	ctx := context.Background()

	// Invalid email
	job := &model.Job{ID: uuid.New(), Type: "email", Payload: json.RawMessage(`{}`)}
	if err := registry.Execute(ctx, job); !errors.Is(err, ErrInvalidJobPayload) {
		t.Errorf("expected ErrInvalidJobPayload, got %v", err)
	}

	// Simulated failure
	failJob := &model.Job{ID: uuid.New(), Type: "email", Payload: json.RawMessage(`{"to":"a@b.com","simulate_failure":true}`)}
	if err := registry.Execute(ctx, failJob); !errors.Is(err, ErrSimulatedFailure) {
		t.Errorf("expected ErrSimulatedFailure, got %v", err)
	}

	// Unknown job type
	unknownJob := &model.Job{ID: uuid.New(), Type: "unknown", Payload: json.RawMessage(`{}`)}
	if err := registry.Execute(ctx, unknownJob); !errors.Is(err, ErrUnknownJobType) {
		t.Errorf("expected ErrUnknownJobType, got %v", err)
	}
}

func TestExecutors_ContextCancellation(t *testing.T) {
	registry := NewDefaultRegistry()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	job := &model.Job{
		ID:      uuid.New(),
		Type:    "report",
		Payload: json.RawMessage(`{"format":"pdf"}`),
	}

	err := registry.Execute(ctx, job)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Errorf("expected context timeout/cancellation, got %v", err)
	}
}
