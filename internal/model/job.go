package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrUnsupportedJobType      = errors.New("unsupported job type")
	ErrEmptyPayload            = errors.New("payload cannot be empty")
)

type JobType string

const (
	JobTypeEmail   JobType = "email"
	JobTypeWebhook JobType = "webhook"
	JobTypeReport  JobType = "report"
)

// IsSupportedJobType checks if the given type is one of email, webhook, or report.
func IsSupportedJobType(t string) bool {
	switch JobType(t) {
	case JobTypeEmail, JobTypeWebhook, JobTypeReport:
		return true
	default:
		return false
	}
}

type Job struct {
	ID          uuid.UUID       `json:"id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	LastError   *string         `json:"last_error,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

type JobAttempt struct {
	ID          int64      `json:"id"`
	JobID       uuid.UUID  `json:"job_id"`
	Attempt     int        `json:"attempt"`
	Status      JobStatus  `json:"status"`
	Error       *string    `json:"error,omitempty"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TransitionTo validates and applies a state transition to the Job.
// Allowed transitions:
//
//	pending    -> processing
//	processing -> completed, failed, pending (for retry or stale crash recovery)
//	failed     -> pending (manual re-queue)
//	completed  -> none (terminal)
func (j *Job) TransitionTo(target JobStatus) error {
	valid := false
	switch j.Status {
	case StatusPending:
		valid = (target == StatusProcessing)
	case StatusProcessing:
		valid = (target == StatusCompleted || target == StatusFailed || target == StatusPending)
	case StatusFailed:
		valid = (target == StatusPending)
	case StatusCompleted:
		valid = false
	}

	if !valid {
		return fmt.Errorf("%w: from %s to %s", ErrInvalidStatusTransition, j.Status, target)
	}

	j.Status = target
	j.UpdatedAt = time.Now().UTC()
	return nil
}
