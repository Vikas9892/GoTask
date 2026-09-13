package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8081"
	}

	slog.Info("starting Worker service", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("Worker service stopped", "error", err)
		os.Exit(1)
	}
}
