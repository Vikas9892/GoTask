package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type APIConfig struct {
	Port            string
	DatabaseURL     string
	QueueSize       int
	ShutdownTimeout time.Duration
}

type WorkerConfig struct {
	Port            string
	DatabaseURL     string
	QueueSize       int
	WorkerCount     int
	MaxRetries      int
	JobTimeout      time.Duration
	ShutdownTimeout time.Duration
}

func LoadAPIConfig() (*APIConfig, error) {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gotask?sslmode=disable")

	queueSize, err := getEnvInt("QUEUE_SIZE", 100, 1)
	if err != nil {
		return nil, fmt.Errorf("invalid QUEUE_SIZE: %w", err)
	}

	shutdownTimeout, err := getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	return &APIConfig{
		Port:            port,
		DatabaseURL:     dbURL,
		QueueSize:       queueSize,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func LoadWorkerConfig() (*WorkerConfig, error) {
	port := getEnv("WORKER_PORT", "8081")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/gotask?sslmode=disable")

	queueSize, err := getEnvInt("QUEUE_SIZE", 100, 1)
	if err != nil {
		return nil, fmt.Errorf("invalid QUEUE_SIZE: %w", err)
	}

	workerCount, err := getEnvInt("WORKER_COUNT", 5, 1)
	if err != nil {
		return nil, fmt.Errorf("invalid WORKER_COUNT: %w", err)
	}

	maxRetries, err := getEnvInt("MAX_RETRIES", 3, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_RETRIES: %w", err)
	}

	jobTimeout, err := getEnvDuration("JOB_TIMEOUT", 30*time.Second, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid JOB_TIMEOUT: %w", err)
	}

	shutdownTimeout, err := getEnvDuration("SHUTDOWN_TIMEOUT", 15*time.Second, 1*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	return &WorkerConfig{
		Port:            port,
		DatabaseURL:     dbURL,
		QueueSize:       queueSize,
		WorkerCount:     workerCount,
		MaxRetries:      maxRetries,
		JobTimeout:      jobTimeout,
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal, minVal int) (int, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return defaultVal, nil
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("must be an integer: %w", err)
	}
	if val < minVal {
		return 0, fmt.Errorf("must be at least %d, got %d", minVal, val)
	}
	return val, nil
}

func getEnvDuration(key string, defaultVal, minVal time.Duration) (time.Duration, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return defaultVal, nil
	}
	dur, err := time.ParseDuration(valStr)
	if err != nil {
		return 0, fmt.Errorf("must be a valid duration (e.g. 10s, 1m): %w", err)
	}
	if dur < minVal {
		return 0, fmt.Errorf("must be at least %v, got %v", minVal, dur)
	}
	return dur, nil
}
