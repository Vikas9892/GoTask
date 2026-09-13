package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Vikas9892/GoTask/internal/metrics"
	"github.com/Vikas9892/GoTask/services/api/internal/repository"
	"github.com/Vikas9892/GoTask/services/api/internal/service"
	"github.com/google/uuid"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListJobsResponse struct {
	Jobs   any `json:"jobs"`
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type JobHandler struct {
	service *service.JobService
}

func NewJobHandler(service *service.JobService) *JobHandler {
	return &JobHandler{service: service}
}

func (h *JobHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/jobs", h.CreateJob)
	mux.HandleFunc("GET /api/jobs", h.ListJobs)
	mux.HandleFunc("GET /api/jobs/{id}", h.GetJob)
	mux.HandleFunc("DELETE /api/jobs/{id}", h.DeleteJob)
	mux.HandleFunc("GET /api/jobs/{id}/attempts", h.GetJobAttempts)
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit

	var input service.CreateJobInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "malformed JSON request body")
		return
	}

	job, err := h.service.CreateJob(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) || errors.Is(err, service.ErrInvalidJSONPayload) || errors.Is(err, service.ErrPayloadTooLarge) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		slog.Error("failed to create job in repository", "type", input.Type, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error creating job: %v", err))
		return
	}

	metrics.JobsSubmittedTotal.Inc()
	writeJSON(w, http.StatusCreated, job)
}

func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job UUID format")
		return
	}

	job, err := h.service.GetJob(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		slog.Error("failed to retrieve job from repository", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error retrieving job: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.Atoi(l)
		if err != nil || val <= 0 {
			writeError(w, http.StatusBadRequest, "limit must be a positive integer between 1 and 100")
			return
		}
		if val > 100 {
			limit = 100
		} else {
			limit = val
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		val, err := strconv.Atoi(o)
		if err != nil || val < 0 {
			writeError(w, http.StatusBadRequest, "offset must be a non-negative integer")
			return
		}
		offset = val
	}

	jobs, total, err := h.service.ListJobs(r.Context(), limit, offset)
	if err != nil {
		slog.Error("failed to list jobs from repository", "limit", limit, "offset", offset, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error listing jobs: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, ListJobsResponse{
		Jobs:   jobs,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job UUID format")
		return
	}

	if err := h.service.DeleteJob(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		slog.Error("failed to delete job from repository", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error deleting job: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *JobHandler) GetJobAttempts(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job UUID format")
		return
	}

	attempts, err := h.service.GetJobAttempts(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrJobNotFound) {
			writeError(w, http.StatusNotFound, "job not found")
			return
		}
		slog.Error("failed to retrieve job attempts from repository", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("database error retrieving job attempts: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, attempts)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
