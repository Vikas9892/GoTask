package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Message string `json:"message,omitempty"`
}

// LivenessHandler returns HTTP 200 when the service process is alive.
func LivenessHandler(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CheckResponse{
			Status:  "ok",
			Service: serviceName,
		})
	}
}

// ReadinessHandler verifies required dependencies (PostgreSQL).
// Returns HTTP 200 if connected, HTTP 503 if unavailable.
func ReadinessHandler(serviceName string, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if pool == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(CheckResponse{
				Status:  "unavailable",
				Service: serviceName,
				Message: "database connection is uninitialized",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(CheckResponse{
				Status:  "unavailable",
				Service: serviceName,
				Message: "database ping failed: " + err.Error(),
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CheckResponse{
			Status:  "ready",
			Service: serviceName,
		})
	}
}
