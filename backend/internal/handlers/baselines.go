package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/crzytrane/diffit/internal/repository"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/go-chi/chi/v5"
)

type BaselineHandlers struct {
	repo         *repository.BaselineRepository
	projectRepo  *repository.ProjectRepository
	snapshotRepo *repository.SnapshotRepository
	storage      *storage.Storage
}

func NewBaselineHandlers(
	repo *repository.BaselineRepository,
	projectRepo *repository.ProjectRepository,
	snapshotRepo *repository.SnapshotRepository,
	storage *storage.Storage,
) *BaselineHandlers {
	return &BaselineHandlers{
		repo:         repo,
		projectRepo:  projectRepo,
		snapshotRepo: snapshotRepo,
		storage:      storage,
	}
}

// Create creates a new baseline from an uploaded image
func (h *BaselineHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	projectIDStr := r.FormValue("project_id")
	name := r.FormValue("name")
	branch := r.FormValue("branch")

	if projectIDStr == "" || name == "" || branch == "" {
		respondError(w, http.StatusBadRequest, "project_id, name, and branch are required")
		return
	}

	projectID, err := parseUUID(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Verify project exists
	_, err = h.projectRepo.GetByID(r.Context(), projectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Project not found")
		return
	}

	// Get optional fields
	var width, height *int
	if w := r.FormValue("width"); w != "" {
		var wVal int
		fmt.Sscanf(w, "%d", &wVal)
		width = &wVal
	}
	if h := r.FormValue("height"); h != "" {
		var hVal int
		fmt.Sscanf(h, "%d", &hVal)
		height = &hVal
	}

	browser := r.FormValue("browser")
	viewport := r.FormValue("viewport")

	var browserPtr, viewportPtr *string
	if browser != "" {
		browserPtr = &browser
	}
	if viewport != "" {
		viewportPtr = &viewport
	}

	// Handle image upload
	file, _, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Image file is required")
		return
	}
	defer file.Close()

	// Save baseline image
	imagePath, err := h.storage.SaveFile(projectID, storage.StorageTypeBaseline, fmt.Sprintf("%s.png", name), file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	// Create or update baseline
	baseline, err := h.repo.Upsert(r.Context(), repository.CreateBaselineParams{
		ProjectID: projectID,
		Name:      name,
		Branch:    branch,
		ImagePath: imagePath,
		Width:     width,
		Height:    height,
		Browser:   browserPtr,
		Viewport:  viewportPtr,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create baseline")
		return
	}

	respondJSON(w, http.StatusCreated, baseline)
}

// CreateFromSnapshot creates a baseline from an existing snapshot
func (h *BaselineHandlers) CreateFromSnapshot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SnapshotID string `json:"snapshot_id"`
		Branch     string `json:"branch"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	snapshotID, err := parseUUID(req.SnapshotID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid snapshot ID")
		return
	}

	snapshot, err := h.snapshotRepo.GetByID(r.Context(), snapshotID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Snapshot not found")
		return
	}

	if snapshot.ComparisonImagePath == nil {
		respondError(w, http.StatusBadRequest, "Snapshot has no comparison image")
		return
	}

	// Get build to get project ID
	build, err := h.projectRepo.GetByID(r.Context(), snapshot.BuildID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get build")
		return
	}

	branch := req.Branch
	if branch == "" {
		// Use build's branch
		buildData, _ := h.snapshotRepo.GetByID(r.Context(), snapshot.BuildID)
		if buildData != nil {
			branch = "main" // Default fallback
		}
	}

	// Copy image to baseline storage
	newImagePath, err := h.storage.CopyFile(*snapshot.ComparisonImagePath, build.ID, storage.StorageTypeBaseline, fmt.Sprintf("%s.png", snapshot.Name))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to copy image")
		return
	}

	// Create baseline
	baseline, err := h.repo.Upsert(r.Context(), repository.CreateBaselineParams{
		ProjectID:        build.ID,
		Name:             snapshot.Name,
		Branch:           branch,
		ImagePath:        newImagePath,
		Width:            snapshot.Width,
		Height:           snapshot.Height,
		Browser:          snapshot.Browser,
		Viewport:         snapshot.Viewport,
		SourceSnapshotID: &snapshot.ID,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create baseline")
		return
	}

	respondJSON(w, http.StatusCreated, baseline)
}

// Get retrieves a baseline by ID
func (h *BaselineHandlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "baselineID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid baseline ID")
		return
	}

	baseline, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Baseline not found")
		return
	}

	respondJSON(w, http.StatusOK, baseline)
}

// ListByProject lists baselines for a project
func (h *BaselineHandlers) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := parseUUID(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid project ID")
		return
	}

	pagination := parsePagination(r)
	branch := r.URL.Query().Get("branch")

	var baselines []models.Baseline
	var total int

	if branch != "" {
		baselines, total, err = h.repo.ListByProjectAndBranch(r.Context(), projectID, branch, pagination)
	} else {
		baselines, total, err = h.repo.ListByProject(r.Context(), projectID, pagination)
	}

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list baselines")
		return
	}

	if baselines == nil {
		baselines = []models.Baseline{}
	}

	respondJSON(w, http.StatusOK, paginatedResponse(baselines, total, pagination))
}

// GetImage serves a baseline image
func (h *BaselineHandlers) GetImage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "baselineID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid baseline ID")
		return
	}

	baseline, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Baseline not found")
		return
	}

	file, err := h.storage.GetFile(baseline.ImagePath)
	if err != nil {
		respondError(w, http.StatusNotFound, "Image file not found")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	io.Copy(w, file)
}

// Delete deletes a baseline
func (h *BaselineHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "baselineID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid baseline ID")
		return
	}

	baseline, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Baseline not found")
		return
	}

	// Delete image
	h.storage.DeleteFile(baseline.ImagePath)

	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete baseline")
		return
	}

	respondJSON(w, http.StatusNoContent, nil)
}
