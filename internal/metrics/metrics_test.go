package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandler(t *testing.T) {
	JobsSubmittedTotal.Inc()
	JobsCompletedTotal.Inc()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "jobs_submitted_total") {
		t.Error("expected metrics output to contain 'jobs_submitted_total'")
	}
	if !strings.Contains(body, "jobs_completed_total") {
		t.Error("expected metrics output to contain 'jobs_completed_total'")
	}
}
