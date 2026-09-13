package model

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsSupportedJobType(t *testing.T) {
	valid := []string{"email", "webhook", "report"}
	for _, v := range valid {
		if !IsSupportedJobType(v) {
			t.Errorf("expected %s to be supported", v)
		}
	}

	invalid := []string{"", "sms", "crypto", "unknown"}
	for _, inv := range invalid {
		if IsSupportedJobType(inv) {
			t.Errorf("expected %s to NOT be supported", inv)
		}
	}
}

func TestJob_TransitionTo(t *testing.T) {
	job := &Job{
		ID:        uuid.New(),
		Type:      "email",
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// pending -> processing (valid)
	if err := job.TransitionTo(StatusProcessing); err != nil {
		t.Fatalf("expected pending -> processing to be valid, got %v", err)
	}
	if job.Status != StatusProcessing {
		t.Errorf("expected status processing, got %s", job.Status)
	}

	// processing -> completed (valid)
	if err := job.TransitionTo(StatusCompleted); err != nil {
		t.Fatalf("expected processing -> completed to be valid, got %v", err)
	}

	// completed -> processing (invalid terminal state)
	if err := job.TransitionTo(StatusProcessing); err == nil {
		t.Fatal("expected error transitioning from completed to processing, got nil")
	}

	// processing -> failed & failed -> pending
	retryJob := &Job{
		ID:     uuid.New(),
		Status: StatusProcessing,
	}
	if err := retryJob.TransitionTo(StatusFailed); err != nil {
		t.Fatalf("expected processing -> failed to be valid, got %v", err)
	}
	if err := retryJob.TransitionTo(StatusPending); err != nil {
		t.Fatalf("expected failed -> pending to be valid, got %v", err)
	}

	// processing -> pending (retry / crash recovery)
	recoverJob := &Job{
		ID:     uuid.New(),
		Status: StatusProcessing,
	}
	if err := recoverJob.TransitionTo(StatusPending); err != nil {
		t.Fatalf("expected processing -> pending to be valid, got %v", err)
	}
}
