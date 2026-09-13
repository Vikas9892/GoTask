package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLivenessHandler(t *testing.T) {
	handler := LivenessHandler("api")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var resp CheckResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if resp.Status != "ok" || resp.Service != "api" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestReadinessHandler_NilPool(t *testing.T) {
	handler := ReadinessHandler("worker", nil)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503 for nil pool, got %d", rec.Code)
	}

	var resp CheckResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if resp.Status != "unavailable" {
		t.Errorf("expected status unavailable, got %s", resp.Status)
	}
}
