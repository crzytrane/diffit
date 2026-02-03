package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/crzytrane/diffit/internal/repository"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/n7olkachev/imgdiff/pkg/imgdiff"
)

type SnapshotHandlers struct {
	repo         *repository.SnapshotRepository
	buildRepo    *repository.BuildRepository
	baselineRepo *repository.BaselineRepository
	storage      *storage.Storage
}

func NewSnapshotHandlers(
	repo *repository.SnapshotRepository,
	buildRepo *repository.BuildRepository,
	baselineRepo *repository.BaselineRepository,
	storage *storage.Storage,
) *SnapshotHandlers {
	return &SnapshotHandlers{
		repo:         repo,
		buildRepo:    buildRepo,
		baselineRepo: baselineRepo,
		storage:      storage,
	}
}

// Create creates a new snapshot and processes the comparison image
func (h *SnapshotHandlers) Create(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get required fields
	buildIDStr := r.FormValue("build_id")
	name := r.FormValue("name")

	if buildIDStr == "" || name == "" {
		respondError(w, http.StatusBadRequest, "build_id and name are required")
		return
	}

	buildID, err := parseUUID(buildIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	// Get build to get project ID
	build, err := h.buildRepo.GetByID(r.Context(), buildID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Build not found")
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

	// Create snapshot record
	snapshot, err := h.repo.Create(r.Context(), models.CreateSnapshotRequest{
		BuildID:  buildID,
		Name:     name,
		Width:    width,
		Height:   height,
		Browser:  browserPtr,
		Viewport: viewportPtr,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create snapshot")
		return
	}

	// Handle image upload
	file, _, err := r.FormFile("image")
	if err != nil {
		// No image uploaded - just return the snapshot
		respondJSON(w, http.StatusCreated, snapshot)
		return
	}
	defer file.Close()

	// Save comparison image
	comparisonPath, err := h.storage.SaveFile(build.ProjectID, storage.StorageTypeComparison, fmt.Sprintf("%s.png", snapshot.ID.String()), file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save image")
		return
	}

	// Find baseline for comparison
	baseline, err := h.baselineRepo.FindByKey(r.Context(), build.ProjectID, name, build.Branch, browserPtr, viewportPtr)
	if err != nil {
		// Error finding baseline - continue without it
		baseline = nil
	}

	// If no baseline on this branch, try default branch
	if baseline == nil {
		// Get project to get default branch
		baseline, _ = h.baselineRepo.FindByKey(r.Context(), build.ProjectID, name, "main", browserPtr, viewportPtr)
	}

	var baseImagePath, diffImagePath *string
	var diffPercentage *float64

	if baseline != nil {
		// Perform diff
		baseImagePath = &baseline.ImagePath

		diffPct, diffPath, err := h.performDiff(build.ProjectID, snapshot.ID, baseline.ImagePath, comparisonPath)
		if err == nil {
			diffPercentage = &diffPct
			if diffPath != "" {
				diffImagePath = &diffPath
			}
		}

		// Link snapshot to baseline
		h.repo.SetBaseline(r.Context(), snapshot.ID, baseline.ID)
	} else {
		// New snapshot - no baseline yet
		zeroPercent := float64(0)
		diffPercentage = &zeroPercent
	}

	// Update snapshot with image paths
	if err := h.repo.UpdateImagePaths(r.Context(), snapshot.ID, baseImagePath, &comparisonPath, diffImagePath, diffPercentage); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update snapshot")
		return
	}

	// Update snapshot status
	h.repo.UpdateStatus(r.Context(), snapshot.ID, models.SnapshotStatusCompleted)

	// Refresh snapshot
	snapshot, _ = h.repo.GetByID(r.Context(), snapshot.ID)

	respondJSON(w, http.StatusCreated, snapshot)
}

// performDiff compares two images and returns diff percentage and diff image path
func (h *SnapshotHandlers) performDiff(projectID, snapshotID uuid.UUID, basePath, comparisonPath string) (float64, string, error) {
	baseFile, err := h.storage.GetFile(basePath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to open base image: %w", err)
	}
	defer baseFile.Close()

	comparisonFile, err := h.storage.GetFile(comparisonPath)
	if err != nil {
		return 0, "", fmt.Errorf("failed to open comparison image: %w", err)
	}
	defer comparisonFile.Close()

	baseImage, _, err := image.Decode(baseFile)
	if err != nil {
		return 0, "", fmt.Errorf("failed to decode base image: %w", err)
	}

	comparisonImage, _, err := image.Decode(comparisonFile)
	if err != nil {
		return 0, "", fmt.Errorf("failed to decode comparison image: %w", err)
	}

	// Perform diff
	result := imgdiff.Diff(baseImage, comparisonImage, &imgdiff.Options{
		Threshold: 0.1,
		DiffImage: true,
	})

	if result.Equal {
		return 0, "", nil
	}

	// Calculate diff percentage based on different pixels
	bounds := baseImage.Bounds()
	totalPixels := bounds.Dx() * bounds.Dy()
	diffPercentage := float64(result.DiffPixelsCount) / float64(totalPixels) * 100

	// Save diff image
	diffFilename := fmt.Sprintf("%s.png", snapshotID.String())
	diffDir := filepath.Join(h.storage.GetFullPath(""), projectID.String(), string(storage.StorageTypeDiff))
	os.MkdirAll(diffDir, 0755)
	diffFullPath := filepath.Join(diffDir, diffFilename)

	diffFile, err := os.Create(diffFullPath)
	if err != nil {
		return diffPercentage, "", fmt.Errorf("failed to create diff file: %w", err)
	}
	defer diffFile.Close()

	writer := bufio.NewWriter(diffFile)
	enc := &png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(writer, result.Image); err != nil {
		return diffPercentage, "", fmt.Errorf("failed to encode diff image: %w", err)
	}
	writer.Flush()

	diffPath := filepath.Join(projectID.String(), string(storage.StorageTypeDiff), diffFilename)
	return diffPercentage, diffPath, nil
}

// Get retrieves a snapshot by ID
func (h *SnapshotHandlers) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "snapshotID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid snapshot ID")
		return
	}

	snapshot, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Snapshot not found")
		return
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// ListByBuild lists snapshots for a build
func (h *SnapshotHandlers) ListByBuild(w http.ResponseWriter, r *http.Request) {
	buildIDStr := chi.URLParam(r, "buildID")
	buildID, err := parseUUID(buildIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	pagination := parsePagination(r)

	// Check for filter
	reviewStatusStr := r.URL.Query().Get("review_status")
	var reviewStatus *models.ReviewStatus
	if reviewStatusStr != "" {
		rs := models.ReviewStatus(reviewStatusStr)
		reviewStatus = &rs
	}

	snapshots, total, err := h.repo.ListByBuildWithFilter(r.Context(), buildID, reviewStatus, pagination)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list snapshots")
		return
	}

	if snapshots == nil {
		snapshots = []models.Snapshot{}
	}

	respondJSON(w, http.StatusOK, paginatedResponse(snapshots, total, pagination))
}

// GetChanged gets snapshots with changes
func (h *SnapshotHandlers) GetChanged(w http.ResponseWriter, r *http.Request) {
	buildIDStr := chi.URLParam(r, "buildID")
	buildID, err := parseUUID(buildIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid build ID")
		return
	}

	snapshots, err := h.repo.GetChangedSnapshots(r.Context(), buildID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get changed snapshots")
		return
	}

	if snapshots == nil {
		snapshots = []models.Snapshot{}
	}

	respondJSON(w, http.StatusOK, snapshots)
}

// Review updates the review status of a snapshot
func (h *SnapshotHandlers) Review(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "snapshotID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid snapshot ID")
		return
	}

	var req models.ReviewSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ReviewedBy == "" {
		respondError(w, http.StatusBadRequest, "reviewed_by is required")
		return
	}

	if err := h.repo.UpdateReviewStatus(r.Context(), id, req); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update review status")
		return
	}

	// If approved and this is on the default branch, update baseline
	if req.ReviewStatus == models.ReviewStatusApproved {
		snapshot, _ := h.repo.GetByID(r.Context(), id)
		if snapshot != nil {
			build, _ := h.buildRepo.GetByID(r.Context(), snapshot.BuildID)
			if build != nil && snapshot.ComparisonImagePath != nil {
				// Create or update baseline
				h.baselineRepo.Upsert(r.Context(), repository.CreateBaselineParams{
					ProjectID:        build.ProjectID,
					Name:             snapshot.Name,
					Branch:           build.Branch,
					ImagePath:        *snapshot.ComparisonImagePath,
					Width:            snapshot.Width,
					Height:           snapshot.Height,
					Browser:          snapshot.Browser,
					Viewport:         snapshot.Viewport,
					SourceSnapshotID: &snapshot.ID,
				})
			}
		}
	}

	// Update build stats
	snapshot, _ := h.repo.GetByID(r.Context(), id)
	if snapshot != nil {
		h.buildRepo.UpdateStats(r.Context(), snapshot.BuildID)
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// BatchReview updates review status for multiple snapshots
func (h *SnapshotHandlers) BatchReview(w http.ResponseWriter, r *http.Request) {
	var req models.BatchReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.SnapshotIDs) == 0 || req.ReviewedBy == "" {
		respondError(w, http.StatusBadRequest, "snapshot_ids and reviewed_by are required")
		return
	}

	if err := h.repo.BatchUpdateReviewStatus(r.Context(), req); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to batch update review status")
		return
	}

	// If approved, update baselines for each snapshot
	if req.ReviewStatus == models.ReviewStatusApproved {
		for _, snapshotID := range req.SnapshotIDs {
			snapshot, err := h.repo.GetByID(r.Context(), snapshotID)
			if err != nil || snapshot == nil || snapshot.ComparisonImagePath == nil {
				continue
			}

			build, err := h.buildRepo.GetByID(r.Context(), snapshot.BuildID)
			if err != nil || build == nil {
				continue
			}

			h.baselineRepo.Upsert(r.Context(), repository.CreateBaselineParams{
				ProjectID:        build.ProjectID,
				Name:             snapshot.Name,
				Branch:           build.Branch,
				ImagePath:        *snapshot.ComparisonImagePath,
				Width:            snapshot.Width,
				Height:           snapshot.Height,
				Browser:          snapshot.Browser,
				Viewport:         snapshot.Viewport,
				SourceSnapshotID: &snapshot.ID,
			})

			// Update build stats
			h.buildRepo.UpdateStats(r.Context(), snapshot.BuildID)
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{"updated": len(req.SnapshotIDs)})
}

// GetImage serves a snapshot image
func (h *SnapshotHandlers) GetImage(w http.ResponseWriter, r *http.Request) {
	imageType := chi.URLParam(r, "imageType") // base, comparison, or diff
	idStr := chi.URLParam(r, "snapshotID")

	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid snapshot ID")
		return
	}

	snapshot, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Snapshot not found")
		return
	}

	var imagePath *string
	switch imageType {
	case "base":
		imagePath = snapshot.BaseImagePath
	case "comparison":
		imagePath = snapshot.ComparisonImagePath
	case "diff":
		imagePath = snapshot.DiffImagePath
	default:
		respondError(w, http.StatusBadRequest, "Invalid image type")
		return
	}

	if imagePath == nil {
		respondError(w, http.StatusNotFound, "Image not found")
		return
	}

	file, err := h.storage.GetFile(*imagePath)
	if err != nil {
		respondError(w, http.StatusNotFound, "Image file not found")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	io.Copy(w, file)
}

// Delete deletes a snapshot
func (h *SnapshotHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "snapshotID")
	id, err := parseUUID(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid snapshot ID")
		return
	}

	snapshot, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "Snapshot not found")
		return
	}

	// Delete images
	if snapshot.ComparisonImagePath != nil {
		h.storage.DeleteFile(*snapshot.ComparisonImagePath)
	}
	if snapshot.DiffImagePath != nil {
		h.storage.DeleteFile(*snapshot.DiffImagePath)
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete snapshot")
		return
	}

	// Update build stats
	h.buildRepo.UpdateStats(r.Context(), snapshot.BuildID)

	respondJSON(w, http.StatusNoContent, nil)
}
