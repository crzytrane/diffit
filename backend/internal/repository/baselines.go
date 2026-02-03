package repository

import (
	"context"
	"fmt"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BaselineRepository struct {
	pool *pgxpool.Pool
}

func NewBaselineRepository(pool *pgxpool.Pool) *BaselineRepository {
	return &BaselineRepository{pool: pool}
}

type CreateBaselineParams struct {
	ProjectID        uuid.UUID
	Name             string
	Branch           string
	ImagePath        string
	Width            *int
	Height           *int
	Browser          *string
	Viewport         *string
	SourceSnapshotID *uuid.UUID
}

func (r *BaselineRepository) Create(ctx context.Context, params CreateBaselineParams) (*models.Baseline, error) {
	var baseline models.Baseline
	err := r.pool.QueryRow(ctx, `
		INSERT INTO baselines (project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
	`, params.ProjectID, params.Name, params.Branch, params.ImagePath, params.Width, params.Height,
		params.Browser, params.Viewport, params.SourceSnapshotID).Scan(
		&baseline.ID,
		&baseline.ProjectID,
		&baseline.Name,
		&baseline.Branch,
		&baseline.ImagePath,
		&baseline.Width,
		&baseline.Height,
		&baseline.Browser,
		&baseline.Viewport,
		&baseline.SourceSnapshotID,
		&baseline.CreatedAt,
		&baseline.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create baseline: %w", err)
	}

	return &baseline, nil
}

func (r *BaselineRepository) Upsert(ctx context.Context, params CreateBaselineParams) (*models.Baseline, error) {
	var baseline models.Baseline
	err := r.pool.QueryRow(ctx, `
		INSERT INTO baselines (project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (project_id, name, branch, browser, viewport)
		DO UPDATE SET
			image_path = EXCLUDED.image_path,
			width = EXCLUDED.width,
			height = EXCLUDED.height,
			source_snapshot_id = EXCLUDED.source_snapshot_id,
			updated_at = NOW()
		RETURNING id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
	`, params.ProjectID, params.Name, params.Branch, params.ImagePath, params.Width, params.Height,
		params.Browser, params.Viewport, params.SourceSnapshotID).Scan(
		&baseline.ID,
		&baseline.ProjectID,
		&baseline.Name,
		&baseline.Branch,
		&baseline.ImagePath,
		&baseline.Width,
		&baseline.Height,
		&baseline.Browser,
		&baseline.Viewport,
		&baseline.SourceSnapshotID,
		&baseline.CreatedAt,
		&baseline.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert baseline: %w", err)
	}

	return &baseline, nil
}

func (r *BaselineRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Baseline, error) {
	var baseline models.Baseline
	err := r.pool.QueryRow(ctx, `
		SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
		FROM baselines WHERE id = $1
	`, id).Scan(
		&baseline.ID,
		&baseline.ProjectID,
		&baseline.Name,
		&baseline.Branch,
		&baseline.ImagePath,
		&baseline.Width,
		&baseline.Height,
		&baseline.Browser,
		&baseline.Viewport,
		&baseline.SourceSnapshotID,
		&baseline.CreatedAt,
		&baseline.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get baseline: %w", err)
	}

	return &baseline, nil
}

func (r *BaselineRepository) FindByKey(ctx context.Context, projectID uuid.UUID, name, branch string, browser, viewport *string) (*models.Baseline, error) {
	var baseline models.Baseline

	// Handle nullable browser and viewport in the query
	var err error
	if browser == nil && viewport == nil {
		err = r.pool.QueryRow(ctx, `
			SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
			FROM baselines
			WHERE project_id = $1 AND name = $2 AND branch = $3 AND browser IS NULL AND viewport IS NULL
		`, projectID, name, branch).Scan(
			&baseline.ID, &baseline.ProjectID, &baseline.Name, &baseline.Branch, &baseline.ImagePath,
			&baseline.Width, &baseline.Height, &baseline.Browser, &baseline.Viewport,
			&baseline.SourceSnapshotID, &baseline.CreatedAt, &baseline.UpdatedAt,
		)
	} else if browser != nil && viewport == nil {
		err = r.pool.QueryRow(ctx, `
			SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
			FROM baselines
			WHERE project_id = $1 AND name = $2 AND branch = $3 AND browser = $4 AND viewport IS NULL
		`, projectID, name, branch, *browser).Scan(
			&baseline.ID, &baseline.ProjectID, &baseline.Name, &baseline.Branch, &baseline.ImagePath,
			&baseline.Width, &baseline.Height, &baseline.Browser, &baseline.Viewport,
			&baseline.SourceSnapshotID, &baseline.CreatedAt, &baseline.UpdatedAt,
		)
	} else if browser == nil && viewport != nil {
		err = r.pool.QueryRow(ctx, `
			SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
			FROM baselines
			WHERE project_id = $1 AND name = $2 AND branch = $3 AND browser IS NULL AND viewport = $4
		`, projectID, name, branch, *viewport).Scan(
			&baseline.ID, &baseline.ProjectID, &baseline.Name, &baseline.Branch, &baseline.ImagePath,
			&baseline.Width, &baseline.Height, &baseline.Browser, &baseline.Viewport,
			&baseline.SourceSnapshotID, &baseline.CreatedAt, &baseline.UpdatedAt,
		)
	} else {
		err = r.pool.QueryRow(ctx, `
			SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
			FROM baselines
			WHERE project_id = $1 AND name = $2 AND branch = $3 AND browser = $4 AND viewport = $5
		`, projectID, name, branch, *browser, *viewport).Scan(
			&baseline.ID, &baseline.ProjectID, &baseline.Name, &baseline.Branch, &baseline.ImagePath,
			&baseline.Width, &baseline.Height, &baseline.Browser, &baseline.Viewport,
			&baseline.SourceSnapshotID, &baseline.CreatedAt, &baseline.UpdatedAt,
		)
	}

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find baseline by key: %w", err)
	}

	return &baseline, nil
}

func (r *BaselineRepository) ListByProject(ctx context.Context, projectID uuid.UUID, pagination models.PaginationParams) ([]models.Baseline, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM baselines WHERE project_id = $1`, projectID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count baselines: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
		FROM baselines
		WHERE project_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`, projectID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list baselines: %w", err)
	}
	defer rows.Close()

	var baselines []models.Baseline
	for rows.Next() {
		var baseline models.Baseline
		if err := rows.Scan(
			&baseline.ID,
			&baseline.ProjectID,
			&baseline.Name,
			&baseline.Branch,
			&baseline.ImagePath,
			&baseline.Width,
			&baseline.Height,
			&baseline.Browser,
			&baseline.Viewport,
			&baseline.SourceSnapshotID,
			&baseline.CreatedAt,
			&baseline.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan baseline: %w", err)
		}
		baselines = append(baselines, baseline)
	}

	return baselines, total, nil
}

func (r *BaselineRepository) ListByProjectAndBranch(ctx context.Context, projectID uuid.UUID, branch string, pagination models.PaginationParams) ([]models.Baseline, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM baselines WHERE project_id = $1 AND branch = $2`, projectID, branch).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count baselines: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, name, branch, image_path, width, height, browser, viewport, source_snapshot_id, created_at, updated_at
		FROM baselines
		WHERE project_id = $1 AND branch = $2
		ORDER BY name ASC
		LIMIT $3 OFFSET $4
	`, projectID, branch, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list baselines by branch: %w", err)
	}
	defer rows.Close()

	var baselines []models.Baseline
	for rows.Next() {
		var baseline models.Baseline
		if err := rows.Scan(
			&baseline.ID,
			&baseline.ProjectID,
			&baseline.Name,
			&baseline.Branch,
			&baseline.ImagePath,
			&baseline.Width,
			&baseline.Height,
			&baseline.Browser,
			&baseline.Viewport,
			&baseline.SourceSnapshotID,
			&baseline.CreatedAt,
			&baseline.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan baseline: %w", err)
		}
		baselines = append(baselines, baseline)
	}

	return baselines, total, nil
}

func (r *BaselineRepository) UpdateImagePath(ctx context.Context, id uuid.UUID, imagePath string, width, height *int) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE baselines
		SET image_path = $2, width = COALESCE($3, width), height = COALESCE($4, height)
		WHERE id = $1
	`, id, imagePath, width, height)
	if err != nil {
		return fmt.Errorf("failed to update baseline image path: %w", err)
	}
	return nil
}

func (r *BaselineRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM baselines WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete baseline: %w", err)
	}
	return nil
}

func (r *BaselineRepository) DeleteByProject(ctx context.Context, projectID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM baselines WHERE project_id = $1`, projectID)
	if err != nil {
		return fmt.Errorf("failed to delete baselines by project: %w", err)
	}
	return nil
}
