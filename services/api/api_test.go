package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vikas9892/GoTask/internal/model"
	"github.com/Vikas9892/GoTask/services/api/internal/handler"
	"github.com/Vikas9892/GoTask/services/api/internal/middleware"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/Vikas9892/GoTask/services/api/internal/service"
)

func TestAPI_LifecycleIntegration(t *testing.T) {
	repo := repository.NewMockJobRepository()
	jobService := service.NewJobService(repo)
	jobHandler := handler.NewJobHandler(jobService)

	mux := http.NewServeMux()
	jobHandler.RegisterRoutes(mux)
	app := middleware.Chain(mux, middleware.Recoverer, middleware.RequestID)

	ts := httptest.NewServer(app)
	defer ts.Close()

	// 1. POST /api/jobs
	body := []byte(`{"type":"email","payload":{"to":"integration@gotask.dev"}}`)
	res, err := http.Post(ts.URL+"/api/jobs", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/jobs failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", res.StatusCode)
	}

	var created model.Job
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode created job: %v", err)
	}

	// 2. GET /api/jobs/{id}
	getRes, err := http.Get(ts.URL + "/api/jobs/" + created.ID.String())
	if err != nil {
		t.Fatalf("GET /api/jobs/{id} failed: %v", err)
	}
	defer getRes.Body.Close()

	if getRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", getRes.StatusCode)
	}

	// 3. GET /api/jobs
	listRes, err := http.Get(ts.URL + "/api/jobs?limit=10&offset=0")
	if err != nil {
		t.Fatalf("GET /api/jobs failed: %v", err)
	}
	defer listRes.Body.Close()

	if listRes.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", listRes.StatusCode)
	}

	// 4. DELETE /api/jobs/{id}
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/jobs/"+created.ID.String(), nil)
	delRes, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE /api/jobs/{id} failed: %v", err)
	}
	defer delRes.Body.Close()

	if delRes.StatusCode != http.StatusNoContent {
		t.Errorf("expected status 204 No Content, got %d", delRes.StatusCode)
	}
}
