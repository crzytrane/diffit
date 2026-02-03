package models

import (
	"time"

	"github.com/google/uuid"
)

// BuildStatus represents the status of a build
type BuildStatus string

const (
	BuildStatusPending    BuildStatus = "pending"
	BuildStatusProcessing BuildStatus = "processing"
	BuildStatusCompleted  BuildStatus = "completed"
	BuildStatusFailed     BuildStatus = "failed"
)

// SnapshotStatus represents the processing status of a snapshot
type SnapshotStatus string

const (
	SnapshotStatusPending    SnapshotStatus = "pending"
	SnapshotStatusProcessing SnapshotStatus = "processing"
	SnapshotStatusCompleted  SnapshotStatus = "completed"
	SnapshotStatusFailed     SnapshotStatus = "failed"
)

// ReviewStatus represents the review status of a snapshot
type ReviewStatus string

const (
	ReviewStatusUnreviewed ReviewStatus = "unreviewed"
	ReviewStatusApproved   ReviewStatus = "approved"
	ReviewStatusRejected   ReviewStatus = "rejected"
)

// Project represents a visual testing project
type Project struct {
	ID            uuid.UUID  `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	RepositoryURL *string    `json:"repository_url,omitempty"`
	DefaultBranch string     `json:"default_branch"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Build represents a collection of snapshots from a single CI run
type Build struct {
	ID                uuid.UUID   `json:"id"`
	ProjectID         uuid.UUID   `json:"project_id"`
	BuildNumber       int         `json:"build_number"`
	Branch            string      `json:"branch"`
	CommitSHA         *string     `json:"commit_sha,omitempty"`
	CommitMessage     *string     `json:"commit_message,omitempty"`
	PullRequestNumber *int        `json:"pull_request_number,omitempty"`
	Status            BuildStatus `json:"status"`
	TotalSnapshots    int         `json:"total_snapshots"`
	ChangedSnapshots  int         `json:"changed_snapshots"`
	ApprovedSnapshots int         `json:"approved_snapshots"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	FinishedAt        *time.Time  `json:"finished_at,omitempty"`
}

// Snapshot represents a single screenshot comparison
type Snapshot struct {
	ID                  uuid.UUID      `json:"id"`
	BuildID             uuid.UUID      `json:"build_id"`
	BaselineID          *uuid.UUID     `json:"baseline_id,omitempty"`
	Name                string         `json:"name"`
	Width               *int           `json:"width,omitempty"`
	Height              *int           `json:"height,omitempty"`
	Browser             *string        `json:"browser,omitempty"`
	Viewport            *string        `json:"viewport,omitempty"`
	BaseImagePath       *string        `json:"base_image_path,omitempty"`
	ComparisonImagePath *string        `json:"comparison_image_path,omitempty"`
	DiffImagePath       *string        `json:"diff_image_path,omitempty"`
	DiffPercentage      *float64       `json:"diff_percentage,omitempty"`
	Status              SnapshotStatus `json:"status"`
	ReviewStatus        ReviewStatus   `json:"review_status"`
	ReviewedBy          *string        `json:"reviewed_by,omitempty"`
	ReviewedAt          *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// Baseline represents an approved baseline image for comparison
type Baseline struct {
	ID               uuid.UUID  `json:"id"`
	ProjectID        uuid.UUID  `json:"project_id"`
	Name             string     `json:"name"`
	Branch           string     `json:"branch"`
	ImagePath        string     `json:"image_path"`
	Width            *int       `json:"width,omitempty"`
	Height           *int       `json:"height,omitempty"`
	Browser          *string    `json:"browser,omitempty"`
	Viewport         *string    `json:"viewport,omitempty"`
	SourceSnapshotID *uuid.UUID `json:"source_snapshot_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Request/Response types for API

type CreateProjectRequest struct {
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	RepositoryURL *string `json:"repository_url,omitempty"`
	DefaultBranch *string `json:"default_branch,omitempty"`
}

type UpdateProjectRequest struct {
	Name          *string `json:"name,omitempty"`
	RepositoryURL *string `json:"repository_url,omitempty"`
	DefaultBranch *string `json:"default_branch,omitempty"`
}

type CreateBuildRequest struct {
	ProjectID         uuid.UUID `json:"project_id"`
	Branch            string    `json:"branch"`
	CommitSHA         *string   `json:"commit_sha,omitempty"`
	CommitMessage     *string   `json:"commit_message,omitempty"`
	PullRequestNumber *int      `json:"pull_request_number,omitempty"`
}

type CreateSnapshotRequest struct {
	BuildID  uuid.UUID `json:"build_id"`
	Name     string    `json:"name"`
	Width    *int      `json:"width,omitempty"`
	Height   *int      `json:"height,omitempty"`
	Browser  *string   `json:"browser,omitempty"`
	Viewport *string   `json:"viewport,omitempty"`
}

type ReviewSnapshotRequest struct {
	ReviewStatus ReviewStatus `json:"review_status"`
	ReviewedBy   string       `json:"reviewed_by"`
}

type BatchReviewRequest struct {
	SnapshotIDs  []uuid.UUID  `json:"snapshot_ids"`
	ReviewStatus ReviewStatus `json:"review_status"`
	ReviewedBy   string       `json:"reviewed_by"`
}

// BuildWithStats includes build with aggregated stats
type BuildWithStats struct {
	Build
	NewSnapshots      int `json:"new_snapshots"`
	RemovedSnapshots  int `json:"removed_snapshots"`
	UnchangedSnapshots int `json:"unchanged_snapshots"`
}

// Pagination helpers
type PaginationParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p *PaginationParams) Limit() int {
	return p.PerPage
}

func NewPaginationParams(page, perPage int) PaginationParams {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return PaginationParams{Page: page, PerPage: perPage}
}
