package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vikas9892/GoTask/internal/config"
	"github.com/Vikas9892/GoTask/internal/database"
	"github.com/Vikas9892/GoTask/services/api/internal/handler"
	"github.com/Vikas9892/GoTask/services/api/internal/middleware"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/Vikas9892/GoTask/services/api/internal/service"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "api",
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadAPIConfig()
	if err != nil {
		slog.Error("failed to load API configuration", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.ConnectPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Warn("postgres connection failed on startup, proceeding in degraded mode", "error", err)
	} else {
		defer pool.Close()
		slog.Info("connected to PostgreSQL")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	if pool != nil {
		repo := repository.NewPostgresJobRepository(pool)
		jobService := service.NewJobService(repo)
		jobHandler := handler.NewJobHandler(jobService)
		jobHandler.RegisterRoutes(mux)
	}

	appHandler := middleware.Chain(mux,
		middleware.Recoverer,
		middleware.Logger,
		middleware.RequestID,
	)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      appHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run HTTP server in background goroutine
	go func() {
		slog.Info("starting API service", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("API HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down API service gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("error during API server shutdown", "error", err)
	}

	if pool != nil {
		pool.Close()
	}

	slog.Info("API service shutdown complete")
}
