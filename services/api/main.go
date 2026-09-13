package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/Vikas9892/GoTask/internal/config"
	"github.com/Vikas9892/GoTask/internal/database"
	"github.com/Vikas9892/GoTask/services/api/internal/handler"
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

	slog.Info("starting API service", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("API service stopped", "error", err)
		os.Exit(1)
	}
}
