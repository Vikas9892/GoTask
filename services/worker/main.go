package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/Vikas9892/GoTask/internal/config"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "worker",
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadWorkerConfig()
	if err != nil {
		slog.Error("failed to load Worker configuration", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	slog.Info("starting Worker service", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		slog.Error("Worker service stopped", "error", err)
		os.Exit(1)
	}
}
