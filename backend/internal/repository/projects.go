package repository

import (
	"context"
	"fmt"

	"github.com/crzytrane/diffit/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) Create(ctx context.Context, req models.CreateProjectRequest) (*models.Project, error) {
	defaultBranch := "main"
	if req.DefaultBranch != nil {
		defaultBranch = *req.DefaultBranch
	}

	var project models.Project
	err := r.pool.QueryRow(ctx, `
		INSERT INTO projects (name, slug, repository_url, default_branch)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, slug, repository_url, default_branch, created_at, updated_at
	`, req.Name, req.Slug, req.RepositoryURL, defaultBranch).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.RepositoryURL,
		&project.DefaultBranch,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var project models.Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, repository_url, default_branch, created_at, updated_at
		FROM projects WHERE id = $1
	`, id).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.RepositoryURL,
		&project.DefaultBranch,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) GetBySlug(ctx context.Context, slug string) (*models.Project, error) {
	var project models.Project
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, slug, repository_url, default_branch, created_at, updated_at
		FROM projects WHERE slug = $1
	`, slug).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.RepositoryURL,
		&project.DefaultBranch,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project by slug: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) List(ctx context.Context, pagination models.PaginationParams) ([]models.Project, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, name, slug, repository_url, default_branch, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pagination.Limit(), pagination.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Slug,
			&project.RepositoryURL,
			&project.DefaultBranch,
			&project.CreatedAt,
			&project.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, req models.UpdateProjectRequest) (*models.Project, error) {
	// Build dynamic update query
	var project models.Project
	err := r.pool.QueryRow(ctx, `
		UPDATE projects
		SET
			name = COALESCE($2, name),
			repository_url = COALESCE($3, repository_url),
			default_branch = COALESCE($4, default_branch)
		WHERE id = $1
		RETURNING id, name, slug, repository_url, default_branch, created_at, updated_at
	`, id, req.Name, req.RepositoryURL, req.DefaultBranch).Scan(
		&project.ID,
		&project.Name,
		&project.Slug,
		&project.RepositoryURL,
		&project.DefaultBranch,
		&project.CreatedAt,
		&project.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}
