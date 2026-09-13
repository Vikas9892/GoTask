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
		"service": "api",
	})
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("starting API service", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("API service stopped", "error", err)
		os.Exit(1)
	}
}
