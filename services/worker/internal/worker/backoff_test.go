package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 16 * time.Second},
		{6, 30 * time.Second}, // Capped at 30s
		{10, 30 * time.Second},
	}

	for _, tt := range tests {
		got := CalculateBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("attempt %d: expected %v, got %v", tt.attempt, tt.expected, got)
		}
	}
}

func TestSleepWithContext_Cancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := SleepWithContext(ctx, 1*time.Second)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}

	if elapsed > 200*time.Millisecond {
		t.Errorf("expected sleep to be interrupted quickly, took %v", elapsed)
	}
}

func TestSleepWithContext_NormalCompletion(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := SleepWithContext(ctx, 10*time.Millisecond)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if time.Since(start) < 10*time.Millisecond {
		t.Error("expected sleep to last at least 10ms")
	}
}
