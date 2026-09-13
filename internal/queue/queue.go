package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/Vikas9892/GoTask/internal/model"
)

var (
	ErrQueueFull   = errors.New("job queue is full")
	ErrQueueClosed = errors.New("job queue is closed")
)

// Queue is a bounded in-memory work queue backed by a buffered Go channel.
type Queue struct {
	jobs   chan *model.Job
	closed bool
	mu     sync.RWMutex
}

// NewQueue initializes a bounded queue with the specified capacity.
func NewQueue(capacity int) *Queue {
	if capacity <= 0 {
		capacity = 100
	}
	return &Queue{
		jobs: make(chan *model.Job, capacity),
	}
}

// Enqueue pushes a job onto the bounded channel.
// It respects context cancellation, rejects enqueues when closed,
// and immediately returns ErrQueueFull if buffer is full rather than blocking indefinitely.
func (q *Queue) Enqueue(ctx context.Context, job *model.Job) error {
	q.mu.RLock()
	if q.closed {
		q.mu.RUnlock()
		return ErrQueueClosed
	}
	q.mu.RUnlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.jobs <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

// Jobs returns the receive-only channel for worker goroutines to consume.
func (q *Queue) Jobs() <-chan *model.Job {
	return q.jobs
}

// Close gracefully closes the queue channel.
// No further enqueues will succeed, but consumers can drain remaining buffered jobs.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !q.closed {
		q.closed = true
		close(q.jobs)
	}
}

// Size returns the count of items currently in the buffer.
func (q *Queue) Size() int {
	return len(q.jobs)
}

// Capacity returns the maximum capacity of the queue buffer.
func (q *Queue) Capacity() int {
	return cap(q.jobs)
}
