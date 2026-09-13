package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
)

var (
	ErrUnknownJobType    = errors.New("unknown job type")
	ErrInvalidJobPayload = errors.New("invalid job payload")
	ErrSimulatedFailure  = errors.New("simulated execution failure")
)

type JobExecutor interface {
	Execute(ctx context.Context, job *model.Job) error
}

type Registry struct {
	executors map[string]JobExecutor
}

func NewDefaultRegistry() *Registry {
	r := &Registry{executors: make(map[string]JobExecutor)}
	r.Register("email", &EmailExecutor{})
	r.Register("webhook", &WebhookExecutor{})
	r.Register("report", &ReportExecutor{})
	return r
}

func (r *Registry) Register(jobType string, exec JobExecutor) {
	r.executors[jobType] = exec
}

func (r *Registry) Execute(ctx context.Context, job *model.Job) error {
	exec, ok := r.executors[job.Type]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownJobType, job.Type)
	}
	return exec.Execute(ctx, job)
}

// EmailExecutor simulates sending transactional emails.
type EmailExecutor struct{}

type emailPayload struct {
	To              string `json:"to"`
	Subject         string `json:"subject,omitempty"`
	SimulateFailure bool   `json:"simulate_failure,omitempty"`
}

func (e *EmailExecutor) Execute(ctx context.Context, job *model.Job) error {
	var p emailPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil || p.To == "" {
		return fmt.Errorf("%w: email payload requires 'to' field", ErrInvalidJobPayload)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}

	if p.SimulateFailure {
		return ErrSimulatedFailure
	}

	return nil
}

// WebhookExecutor simulates delivering webhooks to subscriber URLs.
type WebhookExecutor struct{}

type webhookPayload struct {
	URL             string `json:"url"`
	SimulateFailure bool   `json:"simulate_failure,omitempty"`
}

func (w *WebhookExecutor) Execute(ctx context.Context, job *model.Job) error {
	var p webhookPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil || p.URL == "" {
		return fmt.Errorf("%w: webhook payload requires 'url' field", ErrInvalidJobPayload)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(20 * time.Millisecond):
	}

	if p.SimulateFailure {
		return ErrSimulatedFailure
	}

	return nil
}

// ReportExecutor simulates compiling background reports.
type ReportExecutor struct{}

type reportPayload struct {
	Format          string `json:"format,omitempty"`
	SimulateFailure bool   `json:"simulate_failure,omitempty"`
}

func (r *ReportExecutor) Execute(ctx context.Context, job *model.Job) error {
	var p reportPayload
	if err := json.Unmarshal(job.Payload, &p); err != nil {
		return fmt.Errorf("%w: report payload must be JSON", ErrInvalidJobPayload)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(30 * time.Millisecond):
	}

	if p.SimulateFailure {
		return ErrSimulatedFailure
	}

	return nil
}
