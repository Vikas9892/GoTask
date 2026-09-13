package worker

import (
	"context"
	"math"
	"time"
)

const (
	BaseBackoff = 1 * time.Second
	MaxBackoff  = 30 * time.Second
)

// CalculateBackoff calculates exponential backoff duration based on attempt number:
// Attempt 1: 1s
// Attempt 2: 2s
// Attempt 3: 4s
// Attempt 4: 8s
// Capped at 30s max.
func CalculateBackoff(attempt int) time.Duration {
	if attempt <= 1 {
		return BaseBackoff
	}

	factor := math.Pow(2, float64(attempt-1))
	dur := time.Duration(factor) * BaseBackoff
	if dur > MaxBackoff || dur <= 0 {
		return MaxBackoff
	}
	return dur
}

// SleepWithContext waits for the specified backoff duration or until context is cancelled.
func SleepWithContext(ctx context.Context, dur time.Duration) error {
	timer := time.NewTimer(dur)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
