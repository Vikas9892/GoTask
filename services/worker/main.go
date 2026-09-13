package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vikas9892/GoTask/internal/config"
	"github.com/Vikas9892/GoTask/internal/database"
	"github.com/Vikas9892/GoTask/internal/health"
	"github.com/Vikas9892/GoTask/internal/queue"
	"github.com/Vikas9892/GoTask/services/worker/internal/executor"
	"github.com/Vikas9892/GoTask/services/worker/internal/recovery"
	"github.com/Vikas9892/GoTask/services/worker/internal/repository"
	"github.com/Vikas9892/GoTask/services/worker/internal/worker"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		slog.Error("failed to load Worker configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL pool
	pool, err := database.ConnectPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("postgres connection failed on startup, proceeding in degraded mode", "error", err)
	} else {
		defer pool.Close()
		slog.Info("connected to PostgreSQL")
	}

	// Initialize Queue and Executor Registry
	jobQueue := queue.NewQueue(cfg.QueueSize)
	execRegistry := executor.NewDefaultRegistry()

	var workerPool *worker.Pool
	if pool != nil {
		workerRepo := repository.NewPostgresWorkerRepository(pool)

		// 1. Startup Recovery: reset stale in-flight jobs and recover pending jobs
		recoverer := recovery.NewRecoverer(workerRepo, jobQueue, 5*time.Minute)
		if recovered, recErr := recoverer.Recover(ctx); recErr != nil {
			slog.Error("startup recovery encountered error", "error", recErr)
		} else {
			slog.Info("startup recovery finished", "recovered_jobs", recovered)
		}

		// 2. Build JobProcessor with database tracking, retries and job timeout
		processor := worker.NewDatabaseJobProcessor(workerRepo, execRegistry, jobQueue, cfg.JobTimeout)

		// 3. Initialize and start worker pool
		workerPool = worker.NewPool(cfg.WorkerCount, jobQueue, processor)
		workerPool.Start()
		slog.Info("worker pool started", "workers", cfg.WorkerCount)

		// 4. Background dispatcher: periodically polls PostgreSQL for new pending jobs
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// Fetch pending jobs up to available queue space
					avail := jobQueue.Capacity() - jobQueue.Size()
					if avail <= 0 {
						continue
					}
					pendingJobs, pErr := workerRepo.FindPendingJobs(ctx, avail)
					if pErr != nil {
						continue
					}
					for _, j := range pendingJobs {
						if err := jobQueue.Enqueue(ctx, j); err != nil {
							break
						}
					}
				}
			}
		}()
	}

	// Health & readiness HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", health.LivenessHandler("worker"))
	mux.HandleFunc("GET /ready", health.ReadinessHandler("worker", pool))

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("starting Worker HTTP health server", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Worker health server stopped", "error", err)
		}
	}()

	// Listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down Worker service gracefully...")

	// 1. Stop accepting new HTTP health requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	// 2. Stop background dispatcher
	cancel()

	// 3. Close the queue so no new jobs are accepted
	jobQueue.Close()

	// 4. Stop worker pool and wait for all workers to finish active jobs
	if workerPool != nil {
		slog.Info("waiting for workers to finish active jobs...")
		workerPool.Stop()
	}

	// 5. Close PostgreSQL connection pool
	if pool != nil {
		pool.Close()
	}

	slog.Info("Worker service shutdown complete")
}
