package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadAPIConfig_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadAPIConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.QueueSize != 100 {
		t.Errorf("expected queue size 100, got %d", cfg.QueueSize)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("expected shutdown timeout 15s, got %v", cfg.ShutdownTimeout)
	}
}

func TestLoadWorkerConfig_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := LoadWorkerConfig()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.WorkerCount != 5 {
		t.Errorf("expected worker count 5, got %d", cfg.WorkerCount)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", cfg.MaxRetries)
	}
	if cfg.JobTimeout != 30*time.Second {
		t.Errorf("expected job timeout 30s, got %v", cfg.JobTimeout)
	}
}

func TestLoadAPIConfig_ValidationErrors(t *testing.T) {
	os.Clearenv()
	t.Setenv("QUEUE_SIZE", "-5")

	_, err := LoadAPIConfig()
	if err == nil {
		t.Fatal("expected error for negative QUEUE_SIZE, got nil")
	}

	t.Setenv("QUEUE_SIZE", "invalid")
	_, err = LoadAPIConfig()
	if err == nil {
		t.Fatal("expected error for non-integer QUEUE_SIZE, got nil")
	}

	t.Setenv("QUEUE_SIZE", "50")
	t.Setenv("SHUTDOWN_TIMEOUT", "invalid-duration")
	_, err = LoadAPIConfig()
	if err == nil {
		t.Fatal("expected error for invalid SHUTDOWN_TIMEOUT, got nil")
	}
}
