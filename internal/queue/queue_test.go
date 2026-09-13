package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/google/uuid"
)

func TestQueue_SuccessfulEnqueueAndReceive(t *testing.T) {
	q := NewQueue(5)
	ctx := context.Background()

	job := &model.Job{
		ID:   uuid.New(),
		Type: "email",
	}

	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("expected successful enqueue, got: %v", err)
	}

	if q.Size() != 1 {
		t.Errorf("expected size 1, got %d", q.Size())
	}

	select {
	case received := <-q.Jobs():
		if received.ID != job.ID {
			t.Errorf("expected job %v, got %v", job.ID, received.ID)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job from queue")
	}

	if q.Size() != 0 {
		t.Errorf("expected size 0 after receive, got %d", q.Size())
	}
}

func TestQueue_QueueFull(t *testing.T) {
	capacity := 2
	q := NewQueue(capacity)
	ctx := context.Background()

	for i := 0; i < capacity; i++ {
		err := q.Enqueue(ctx, &model.Job{ID: uuid.New(), Type: "report"})
		if err != nil {
			t.Fatalf("unexpected error enqueuing item %d: %v", i, err)
		}
	}

	// Next enqueue should immediately return ErrQueueFull
	err := q.Enqueue(ctx, &model.Job{ID: uuid.New(), Type: "report"})
	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestQueue_Cancellation(t *testing.T) {
	q := NewQueue(1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := q.Enqueue(ctx, &model.Job{ID: uuid.New(), Type: "webhook"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestQueue_Shutdown(t *testing.T) {
	q := NewQueue(5)
	ctx := context.Background()

	job := &model.Job{ID: uuid.New(), Type: "email"}
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("failed to enqueue: %v", err)
	}

	q.Close()

	// Subsequent enqueues must fail with ErrQueueClosed
	err := q.Enqueue(ctx, &model.Job{ID: uuid.New(), Type: "webhook"})
	if !errors.Is(err, ErrQueueClosed) {
		t.Errorf("expected ErrQueueClosed, got %v", err)
	}

	// Existing buffered job can still be received
	received, ok := <-q.Jobs()
	if !ok || received.ID != job.ID {
		t.Errorf("expected to drain queued job, got %v, ok=%v", received, ok)
	}

	// After drain, channel is closed
	_, ok = <-q.Jobs()
	if ok {
		t.Error("expected channel to be closed after drain")
	}
}
