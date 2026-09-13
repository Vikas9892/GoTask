package database

import (
	"context"
	"testing"
	"time"
)

func TestConnectPool_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// An invalid schema or bad connection string should return an immediate error
	_, err := ConnectPool(ctx, "invalid-postgres-url://bad:bad@bad:9999/none")
	if err == nil {
		t.Fatal("expected error for invalid database URL, got nil")
	}
}
