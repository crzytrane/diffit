package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BuildRepository struct {
	pool *pgxpool.Pool
}

func NewBuildRepository(pool *pgxpool.Pool) *BuildRepository {
	return &BuildRepository{pool: pool}
}

func (r *BuildRepository) Create(ctx context.Context, req models.CreateBuildRequest) (*models.Build, error) {
	var build models.Build
	err := r.pool.QueryRow(ctx, `
		INSERT INTO builds (project_id, branch, commit_sha, commit_message, pull_request_number, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, project_id, build_number, branch, commit_sha, commit_message,
		          pull_request_number, status, total_snapshots, changed_snapshots,
		          approved_snapshots, created_at, updated_at, finished_at
	`, req.ProjectID, req.Branch, req.CommitSHA, req.CommitMessage, req.PullRequestNumber, models.BuildStatusPending).Scan(
		&build.ID,
		&build.ProjectID,
		&build.BuildNumber,
		&build.Branch,
		&build.CommitSHA,
		&build.CommitMessage,
		&build.PullRequestNumber,
		&build.Status,
		&build.TotalSnapshots,
		&build.ChangedSnapshots,
		&build.ApprovedSnapshots,
		&build.CreatedAt,
		&build.UpdatedAt,
		&build.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create build: %w", err)
	}

	return &build, nil
}

func (r *BuildRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Build, error) {
	var build models.Build
	err := r.pool.QueryRow(ctx, `
		SELECT id, project_id, build_number, branch, commit_sha, commit_message,
		       pull_request_number, status, total_snapshots, changed_snapshots,
		       approved_snapshots, created_at, updated_at, finished_at
		FROM builds WHERE id = $1
	`, id).Scan(
		&build.ID,
		&build.ProjectID,
		&build.BuildNumber,
		&build.Branch,
		&build.CommitSHA,
		&build.CommitMessage,
		&build.PullRequestNumber,
		&build.Status,
		&build.TotalSnapshots,
		&build.ChangedSnapshots,
		&build.ApprovedSnapshots,
		&build.CreatedAt,
		&build.UpdatedAt,
		&build.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get build: %w", err)
	}

	return &build, nil
}

func (r *BuildRepository) ListByProject(ctx context.Context, projectID uuid.UUID, pagination models.PaginationParams) ([]models.Build, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM builds WHERE project_id = $1`, projectID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count builds: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, build_number, branch, commit_sha, commit_message,
		       pull_request_number, status, total_snapshots, changed_snapshots,
		       approved_snapshots, created_at, updated_at, finished_at
		FROM builds
		WHERE project_id = $1
		ORDER BY build_number DESC
		LIMIT $2 OFFSET $3
	`, projectID, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list builds: %w", err)
	}
	defer rows.Close()

	var builds []models.Build
	for rows.Next() {
		var build models.Build
		if err := rows.Scan(
			&build.ID,
			&build.ProjectID,
			&build.BuildNumber,
			&build.Branch,
			&build.CommitSHA,
			&build.CommitMessage,
			&build.PullRequestNumber,
			&build.Status,
			&build.TotalSnapshots,
			&build.ChangedSnapshots,
			&build.ApprovedSnapshots,
			&build.CreatedAt,
			&build.UpdatedAt,
			&build.FinishedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan build: %w", err)
		}
		builds = append(builds, build)
	}

	return builds, total, nil
}

func (r *BuildRepository) ListByBranch(ctx context.Context, projectID uuid.UUID, branch string, pagination models.PaginationParams) ([]models.Build, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM builds WHERE project_id = $1 AND branch = $2`, projectID, branch).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count builds: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, project_id, build_number, branch, commit_sha, commit_message,
		       pull_request_number, status, total_snapshots, changed_snapshots,
		       approved_snapshots, created_at, updated_at, finished_at
		FROM builds
		WHERE project_id = $1 AND branch = $2
		ORDER BY build_number DESC
		LIMIT $3 OFFSET $4
	`, projectID, branch, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list builds by branch: %w", err)
	}
	defer rows.Close()

	var builds []models.Build
	for rows.Next() {
		var build models.Build
		if err := rows.Scan(
			&build.ID,
			&build.ProjectID,
			&build.BuildNumber,
			&build.Branch,
			&build.CommitSHA,
			&build.CommitMessage,
			&build.PullRequestNumber,
			&build.Status,
			&build.TotalSnapshots,
			&build.ChangedSnapshots,
			&build.ApprovedSnapshots,
			&build.CreatedAt,
			&build.UpdatedAt,
			&build.FinishedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan build: %w", err)
		}
		builds = append(builds, build)
	}

	return builds, total, nil
}

func (r *BuildRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.BuildStatus) error {
	var finishedAt *time.Time
	if status == models.BuildStatusCompleted || status == models.BuildStatusFailed {
		now := time.Now()
		finishedAt = &now
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE builds SET status = $2, finished_at = $3 WHERE id = $1
	`, id, status, finishedAt)
	if err != nil {
		return fmt.Errorf("failed to update build status: %w", err)
	}
	return nil
}

func (r *BuildRepository) UpdateStats(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE builds SET
			total_snapshots = (SELECT COUNT(*) FROM snapshots WHERE build_id = $1),
			changed_snapshots = (SELECT COUNT(*) FROM snapshots WHERE build_id = $1 AND diff_percentage > 0),
			approved_snapshots = (SELECT COUNT(*) FROM snapshots WHERE build_id = $1 AND review_status = 'approved')
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("failed to update build stats: %w", err)
	}
	return nil
}

func (r *BuildRepository) GetLatestByBranch(ctx context.Context, projectID uuid.UUID, branch string) (*models.Build, error) {
	var build models.Build
	err := r.pool.QueryRow(ctx, `
		SELECT id, project_id, build_number, branch, commit_sha, commit_message,
		       pull_request_number, status, total_snapshots, changed_snapshots,
		       approved_snapshots, created_at, updated_at, finished_at
		FROM builds
		WHERE project_id = $1 AND branch = $2
		ORDER BY build_number DESC
		LIMIT 1
	`, projectID, branch).Scan(
		&build.ID,
		&build.ProjectID,
		&build.BuildNumber,
		&build.Branch,
		&build.CommitSHA,
		&build.CommitMessage,
		&build.PullRequestNumber,
		&build.Status,
		&build.TotalSnapshots,
		&build.ChangedSnapshots,
		&build.ApprovedSnapshots,
		&build.CreatedAt,
		&build.UpdatedAt,
		&build.FinishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest build: %w", err)
	}

	return &build, nil
}

func (r *BuildRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM builds WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete build: %w", err)
	}
	return nil
}
