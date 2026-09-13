package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/Vikas9892/GoTask/services/api/internal/service"
	"github.com/google/uuid"
)

func setupTestServer() (*http.ServeMux, *service.JobService) {
	repo := repository.NewMockJobRepository()
	svc := service.NewJobService(repo)
	h := NewJobHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux, svc
}

func TestJobHandler_CreateJob(t *testing.T) {
	mux, _ := setupTestServer()

	// Valid POST
	body := []byte(`{"type":"email","payload":{"to":"dev@example.com"}}`)
	req := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var created model.Job
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.Type != "email" {
		t.Errorf("expected type email, got %s", created.Type)
	}

	// Malformed JSON
	badReq := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader([]byte(`{bad-json`)))
	badRec := httptest.NewRecorder()
	mux.ServeHTTP(badRec, badReq)
	if badRec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for malformed json, got %d", badRec.Code)
	}

	// Unsupported type
	unsupportedReq := httptest.NewRequest(http.MethodPost, "/api/jobs", bytes.NewReader([]byte(`{"type":"unknown","payload":{"a":1}}`)))
	unsupportedRec := httptest.NewRecorder()
	mux.ServeHTTP(unsupportedRec, unsupportedReq)
	if unsupportedRec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for unsupported type, got %d", unsupportedRec.Code)
	}
}

func TestJobHandler_GetAndListJobs(t *testing.T) {
	mux, svc := setupTestServer()

	job, err := svc.CreateJob(t.Context(), service.CreateJobInput{
		Type:    "report",
		Payload: json.RawMessage(`{"report_id":123}`),
	})
	if err != nil {
		t.Fatalf("failed to seed job: %v", err)
	}

	// GET by ID
	req := httptest.NewRequest(http.MethodGet, "/api/jobs/"+job.ID.String(), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	// GET Not Found
	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/jobs/"+uuid.New().String(), nil)
	notFoundRec := httptest.NewRecorder()
	mux.ServeHTTP(notFoundRec, notFoundReq)
	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", notFoundRec.Code)
	}

	// GET Invalid UUID
	invalidReq := httptest.NewRequest(http.MethodGet, "/api/jobs/not-a-uuid", nil)
	invalidRec := httptest.NewRecorder()
	mux.ServeHTTP(invalidRec, invalidReq)
	if invalidRec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", invalidRec.Code)
	}

	// LIST jobs
	listReq := httptest.NewRequest(http.MethodGet, "/api/jobs?limit=10&offset=0", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", listRec.Code)
	}

	var listResp ListJobsResponse
	if err := json.NewDecoder(listRec.Body).Decode(&listResp); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if listResp.Total != 1 {
		t.Errorf("expected total 1, got %d", listResp.Total)
	}
}

func TestJobHandler_DeleteJob(t *testing.T) {
	mux, svc := setupTestServer()

	job, err := svc.CreateJob(t.Context(), service.CreateJobInput{
		Type:    "webhook",
		Payload: json.RawMessage(`{"url":"https://example.com"}`),
	})
	if err != nil {
		t.Fatalf("failed to seed job: %v", err)
	}

	// DELETE
	req := httptest.NewRequest(http.MethodDelete, "/api/jobs/"+job.ID.String(), nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}

	// DELETE Not Found
	notFoundReq := httptest.NewRequest(http.MethodDelete, "/api/jobs/"+job.ID.String(), nil)
	notFoundRec := httptest.NewRecorder()
	mux.ServeHTTP(notFoundRec, notFoundReq)
	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for deleted job, got %d", notFoundRec.Code)
	}
}

func TestJobHandler_GetJobAttempts(t *testing.T) {
	mux, svc := setupTestServer()

	job, err := svc.CreateJob(t.Context(), service.CreateJobInput{
		Type:    "report",
		Payload: json.RawMessage(`{"report_id":999}`),
	})
	if err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/jobs/"+job.ID.String()+"/attempts", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestJobHandler_PaginationValidation(t *testing.T) {
	mux, _ := setupTestServer()

	// Invalid negative limit
	req := httptest.NewRequest(http.MethodGet, "/api/jobs?limit=-5", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative limit, got %d", rec.Code)
	}

	// Invalid offset
	req2 := httptest.NewRequest(http.MethodGet, "/api/jobs?offset=-1", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative offset, got %d", rec2.Code)
	}

	// Limit capping at 100
	req3 := httptest.NewRequest(http.MethodGet, "/api/jobs?limit=500", nil)
	rec3 := httptest.NewRecorder()
	mux.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec3.Code)
	}

	var resp ListJobsResponse
	_ = json.NewDecoder(rec3.Body).Decode(&resp)
	if resp.Limit != 100 {
		t.Errorf("expected limit capped at 100, got %d", resp.Limit)
	}
}
