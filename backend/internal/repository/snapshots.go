package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotRepository struct {
	pool *pgxpool.Pool
}

func NewSnapshotRepository(pool *pgxpool.Pool) *SnapshotRepository {
	return &SnapshotRepository{pool: pool}
}

func (r *SnapshotRepository) Create(ctx context.Context, req models.CreateSnapshotRequest) (*models.Snapshot, error) {
	var snapshot models.Snapshot
	err := r.pool.QueryRow(ctx, `
		INSERT INTO snapshots (build_id, name, width, height, browser, viewport, status, review_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, build_id, baseline_id, name, width, height, browser, viewport,
		          base_image_path, comparison_image_path, diff_image_path, diff_percentage,
		          status, review_status, reviewed_by, reviewed_at, created_at, updated_at
	`, req.BuildID, req.Name, req.Width, req.Height, req.Browser, req.Viewport,
		models.SnapshotStatusPending, models.ReviewStatusUnreviewed).Scan(
		&snapshot.ID,
		&snapshot.BuildID,
		&snapshot.BaselineID,
		&snapshot.Name,
		&snapshot.Width,
		&snapshot.Height,
		&snapshot.Browser,
		&snapshot.Viewport,
		&snapshot.BaseImagePath,
		&snapshot.ComparisonImagePath,
		&snapshot.DiffImagePath,
		&snapshot.DiffPercentage,
		&snapshot.Status,
		&snapshot.ReviewStatus,
		&snapshot.ReviewedBy,
		&snapshot.ReviewedAt,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return &snapshot, nil
}

func (r *SnapshotRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Snapshot, error) {
	var snapshot models.Snapshot
	err := r.pool.QueryRow(ctx, `
		SELECT id, build_id, baseline_id, name, width, height, browser, viewport,
		       base_image_path, comparison_image_path, diff_image_path, diff_percentage,
		       status, review_status, reviewed_by, reviewed_at, created_at, updated_at
		FROM snapshots WHERE id = $1
	`, id).Scan(
		&snapshot.ID,
		&snapshot.BuildID,
		&snapshot.BaselineID,
		&snapshot.Name,
		&snapshot.Width,
		&snapshot.Height,
		&snapshot.Browser,
		&snapshot.Viewport,
		&snapshot.BaseImagePath,
		&snapshot.ComparisonImagePath,
		&snapshot.DiffImagePath,
		&snapshot.DiffPercentage,
		&snapshot.Status,
		&snapshot.ReviewStatus,
		&snapshot.ReviewedBy,
		&snapshot.ReviewedAt,
		&snapshot.CreatedAt,
		&snapshot.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return &snapshot, nil
}

func (r *SnapshotRepository) ListByBuild(ctx context.Context, buildID uuid.UUID, pagination models.PaginationParams) ([]models.Snapshot, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM snapshots WHERE build_id = $1`, buildID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count snapshots: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, build_id, baseline_id, name, width, height, browser, viewport,
		       base_image_path, comparison_image_path, diff_image_path, diff_percentage,
		       status, review_status, reviewed_by, reviewed_at, created_at, updated_at
		FROM snapshots
		WHERE build_id = $1
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`, buildID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []models.Snapshot
	for rows.Next() {
		var snapshot models.Snapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.BuildID,
			&snapshot.BaselineID,
			&snapshot.Name,
			&snapshot.Width,
			&snapshot.Height,
			&snapshot.Browser,
			&snapshot.Viewport,
			&snapshot.BaseImagePath,
			&snapshot.ComparisonImagePath,
			&snapshot.DiffImagePath,
			&snapshot.DiffPercentage,
			&snapshot.Status,
			&snapshot.ReviewStatus,
			&snapshot.ReviewedBy,
			&snapshot.ReviewedAt,
			&snapshot.CreatedAt,
			&snapshot.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, total, nil
}

func (r *SnapshotRepository) ListByBuildWithFilter(ctx context.Context, buildID uuid.UUID, reviewStatus *models.ReviewStatus, pagination models.PaginationParams) ([]models.Snapshot, int, error) {
	var total int
	var err error

	if reviewStatus != nil {
		err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM snapshots WHERE build_id = $1 AND review_status = $2`, buildID, *reviewStatus).Scan(&total)
	} else {
		err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM snapshots WHERE build_id = $1`, buildID).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count snapshots: %w", err)
	}

	query := `
		SELECT id, build_id, baseline_id, name, width, height, browser, viewport,
		       base_image_path, comparison_image_path, diff_image_path, diff_percentage,
		       status, review_status, reviewed_by, reviewed_at, created_at, updated_at
		FROM snapshots
		WHERE build_id = $1
	`
	args := []interface{}{buildID}

	if reviewStatus != nil {
		query += ` AND review_status = $2 ORDER BY name ASC LIMIT $3 OFFSET $4`
		args = append(args, *reviewStatus, pagination.Limit(), pagination.Offset())
	} else {
		query += ` ORDER BY name ASC LIMIT $2 OFFSET $3`
		args = append(args, pagination.Limit(), pagination.Offset())
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []models.Snapshot
	for rows.Next() {
		var snapshot models.Snapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.BuildID,
			&snapshot.BaselineID,
			&snapshot.Name,
			&snapshot.Width,
			&snapshot.Height,
			&snapshot.Browser,
			&snapshot.Viewport,
			&snapshot.BaseImagePath,
			&snapshot.ComparisonImagePath,
			&snapshot.DiffImagePath,
			&snapshot.DiffPercentage,
			&snapshot.Status,
			&snapshot.ReviewStatus,
			&snapshot.ReviewedBy,
			&snapshot.ReviewedAt,
			&snapshot.CreatedAt,
			&snapshot.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, total, nil
}

func (r *SnapshotRepository) UpdateImagePaths(ctx context.Context, id uuid.UUID, baseImagePath, comparisonImagePath, diffImagePath *string, diffPercentage *float64) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE snapshots
		SET base_image_path = COALESCE($2, base_image_path),
		    comparison_image_path = COALESCE($3, comparison_image_path),
		    diff_image_path = COALESCE($4, diff_image_path),
		    diff_percentage = COALESCE($5, diff_percentage)
		WHERE id = $1
	`, id, baseImagePath, comparisonImagePath, diffImagePath, diffPercentage)
	if err != nil {
		return fmt.Errorf("failed to update snapshot image paths: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.SnapshotStatus) error {
	_, err := r.pool.Exec(ctx, `UPDATE snapshots SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("failed to update snapshot status: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) UpdateReviewStatus(ctx context.Context, id uuid.UUID, req models.ReviewSnapshotRequest) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE snapshots
		SET review_status = $2, reviewed_by = $3, reviewed_at = $4
		WHERE id = $1
	`, id, req.ReviewStatus, req.ReviewedBy, now)
	if err != nil {
		return fmt.Errorf("failed to update review status: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) BatchUpdateReviewStatus(ctx context.Context, req models.BatchReviewRequest) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE snapshots
		SET review_status = $2, reviewed_by = $3, reviewed_at = $4
		WHERE id = ANY($1)
	`, req.SnapshotIDs, req.ReviewStatus, req.ReviewedBy, now)
	if err != nil {
		return fmt.Errorf("failed to batch update review status: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) SetBaseline(ctx context.Context, id uuid.UUID, baselineID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE snapshots SET baseline_id = $2 WHERE id = $1`, id, baselineID)
	if err != nil {
		return fmt.Errorf("failed to set baseline: %w", err)
	}
	return nil
}

func (r *SnapshotRepository) GetChangedSnapshots(ctx context.Context, buildID uuid.UUID) ([]models.Snapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, build_id, baseline_id, name, width, height, browser, viewport,
		       base_image_path, comparison_image_path, diff_image_path, diff_percentage,
		       status, review_status, reviewed_by, reviewed_at, created_at, updated_at
		FROM snapshots
		WHERE build_id = $1 AND (diff_percentage > 0 OR baseline_id IS NULL)
		ORDER BY name ASC
	`, buildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []models.Snapshot
	for rows.Next() {
		var snapshot models.Snapshot
		if err := rows.Scan(
			&snapshot.ID,
			&snapshot.BuildID,
			&snapshot.BaselineID,
			&snapshot.Name,
			&snapshot.Width,
			&snapshot.Height,
			&snapshot.Browser,
			&snapshot.Viewport,
			&snapshot.BaseImagePath,
			&snapshot.ComparisonImagePath,
			&snapshot.DiffImagePath,
			&snapshot.DiffPercentage,
			&snapshot.Status,
			&snapshot.ReviewStatus,
			&snapshot.ReviewedBy,
			&snapshot.ReviewedAt,
			&snapshot.CreatedAt,
			&snapshot.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func (r *SnapshotRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM snapshots WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}
	return nil
}
