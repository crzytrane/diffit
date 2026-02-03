package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/crzytrane/diffit/internal/repository"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/go-chi/chi/v5"
)

type ProjectHandlers struct {
	repo    *repository.ProjectRepository
	storage *storage.Storage
}

func NewProjectHandlers(repo *repository.ProjectRepository, storage *storage.Storage) *ProjectHandlers {
	return &ProjectHandlers{repo: repo, storage: storage}
}

// Create creates a new project
func (h *ProjectHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Slug == "" {
		respondError(w, http.StatusBadRequest, "Name and slug are required")
		return
	}

	project, err := h.repo.Create(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

// Get retrieves a project by ID
func (h *ProjectHandlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	project, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// GetBySlug retrieves a project by slug
func (h *ProjectHandlers) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "Slug is required")
		return
	}

	project, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// List lists all projects with pagination
func (h *ProjectHandlers) List(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)

	projects, total, err := h.repo.List(r.Context(), pagination)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list projects")
		return
	}

	if projects == nil {
		projects = []models.Project{}
	}

	respondJSON(w, http.StatusOK, paginatedResponse(projects, total, pagination))
}

// Update updates a project
func (h *ProjectHandlers) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	var req models.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.repo.Update(r.Context(), id, req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update project")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

// Delete deletes a project and all associated data
func (h *ProjectHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "projectID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Delete project files
	if err := h.storage.DeleteProjectFiles(id); err != nil {
		// Log but don't fail - files might not exist
	}

	// Delete project from database (cascades to builds, snapshots, baselines)
	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
