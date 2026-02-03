package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/crzytrane/diffit/internal/repository"
	"github.com/crzytrane/diffit/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handlers contains all HTTP handlers and their dependencies
type Handlers struct {
	Projects  *ProjectHandlers
	Builds    *BuildHandlers
	Snapshots *SnapshotHandlers
	Baselines *BaselineHandlers
	storage   *storage.Storage
}

// New creates a new Handlers instance with all dependencies
func New(pool *pgxpool.Pool, storage *storage.Storage) *Handlers {
	projectRepo := repository.NewProjectRepository(pool)
	buildRepo := repository.NewBuildRepository(pool)
	snapshotRepo := repository.NewSnapshotRepository(pool)
	baselineRepo := repository.NewBaselineRepository(pool)

	return &Handlers{
		Projects:  NewProjectHandlers(projectRepo, storage),
		Builds:    NewBuildHandlers(buildRepo, projectRepo, storage),
		Snapshots: NewSnapshotHandlers(snapshotRepo, buildRepo, baselineRepo, storage),
		Baselines: NewBaselineHandlers(baselineRepo, projectRepo, snapshotRepo, storage),
		storage:   storage,
	}
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parsePagination(r *http.Request) models.PaginationParams {
	page := 1
	perPage := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil {
			perPage = parsed
		}
	}

	return models.NewPaginationParams(page, perPage)
}

func paginatedResponse(data interface{}, total int, pagination models.PaginationParams) models.PaginatedResponse {
	totalPages := total / pagination.PerPage
	if total%pagination.PerPage > 0 {
		totalPages++
	}

	return models.PaginatedResponse{
		Data:       data,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
