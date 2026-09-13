package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID, _ = r.Context().Value(RequestIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Without incoming header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected generated request ID in context")
	}
	if rec.Header().Get("X-Request-ID") != capturedID {
		t.Errorf("expected header %s, got %s", capturedID, rec.Header().Get("X-Request-ID"))
	}

	// Case 2: With incoming header
	customID := "custom-request-id-123"
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Request-ID", customID)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if capturedID != customID {
		t.Errorf("expected propagated request ID %s, got %s", customID, capturedID)
	}
}

func TestRecoverer_HandlesPanic(t *testing.T) {
	panickingHandler := Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected runtime panic!")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	panickingHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
	if rec.Body.String() != `{"error":"internal server error"}` {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}
