package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/crzytrane/diffit/internal/repository"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/go-chi/chi/v5"
)

type BuildHandlers struct {
	repo        *repository.BuildRepository
	projectRepo *repository.ProjectRepository
	storage     *storage.Storage
}

func NewBuildHandlers(repo *repository.BuildRepository, projectRepo *repository.ProjectRepository, storage *storage.Storage) *BuildHandlers {
	return &BuildHandlers{repo: repo, projectRepo: projectRepo, storage: storage}
}

// Create creates a new build
func (h *BuildHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Branch == "" {
		respondError(w, http.StatusBadRequest, "Branch is required")
		return
	}

	// Verify project exists
	_, err := h.projectRepo.GetByID(r.Context(), req.ProjectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	build, err := h.repo.Create(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create build")
		return
	}

	respondJSON(w, http.StatusCreated, build)
}

// Get retrieves a build by ID
func (h *BuildHandlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "buildID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	build, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Build not found")
		return
	}

	respondJSON(w, http.StatusOK, build)
}

// ListByProject lists builds for a project
func (h *BuildHandlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := parseUUID(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	pagination := parsePagination(r)
	branch := r.URL.Query().Get("branch")

	var builds []models.Build
	var total int

	if branch != "" {
		builds, total, err = h.repo.ListByBranch(r.Context(), projectID, branch, pagination)
	} else {
		builds, total, err = h.repo.ListByProject(r.Context(), projectID, pagination)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list builds")
		return
	}

	if builds == nil {
		builds = []models.Build{}
	}

	respondJSON(w, http.StatusOK, paginatedResponse(builds, total, pagination))
}

// UpdateStatus updates the status of a build
func (h *BuildHandlers) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "buildID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	var req struct {
		Status models.BuildStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update build status")
		return
	}

	// Update stats
	if err := h.repo.UpdateStats(r.Context(), id); err != nil {
		// Log but don't fail
	}

	build, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated build")
		return
	}

	respondJSON(w, http.StatusOK, build)
}

// Finalize finalizes a build and updates stats
func (h *BuildHandlers) Finalize(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "buildID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	// Update stats
	if err := h.repo.UpdateStats(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update build stats")
		return
	}

	// Set status to completed
	if err := h.repo.UpdateStatus(r.Context(), id, models.BuildStatusCompleted); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to finalize build")
		return
	}

	build, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get finalized build")
		return
	}

	respondJSON(w, http.StatusOK, build)
}

// GetLatest gets the latest build for a branch
func (h *BuildHandlers) GetLatest(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := parseUUID(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	branch := r.URL.Query().Get("branch")
	if branch == "" {
		// Get default branch from project
		project, err := h.projectRepo.GetByID(r.Context(), projectID)
		if err != nil {
			respondError(w, http.StatusNotFound, "Project not found")
			return
		}
		branch = project.DefaultBranch
	}

	build, err := h.repo.GetLatestByBranch(r.Context(), projectID, branch)
	if err != nil {
		respondError(w, http.StatusNotFound, "No builds found")
		return
	}

	respondJSON(w, http.StatusOK, build)
}

// Delete deletes a build
func (h *BuildHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "buildID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete build")
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
