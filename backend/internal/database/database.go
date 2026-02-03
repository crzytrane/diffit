package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(databaseURL string) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) Migrate(ctx context.Context) error {
	schema := `
	-- Projects table
	CREATE TABLE IF NOT EXISTS projects (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(255) NOT NULL UNIQUE,
		repository_url VARCHAR(500),
		default_branch VARCHAR(100) DEFAULT 'main',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Builds table
	CREATE TABLE IF NOT EXISTS builds (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		build_number SERIAL,
		branch VARCHAR(255) NOT NULL,
		commit_sha VARCHAR(40),
		commit_message TEXT,
		pull_request_number INTEGER,
		status VARCHAR(50) DEFAULT 'pending',
		total_snapshots INTEGER DEFAULT 0,
		changed_snapshots INTEGER DEFAULT 0,
		approved_snapshots INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		finished_at TIMESTAMP WITH TIME ZONE
	);

	-- Baselines table (created before snapshots due to FK dependency)
	CREATE TABLE IF NOT EXISTS baselines (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
		name VARCHAR(500) NOT NULL,
		branch VARCHAR(255) NOT NULL,
		image_path VARCHAR(500) NOT NULL,
		width INTEGER,
		height INTEGER,
		browser VARCHAR(50),
		viewport VARCHAR(50),
		source_snapshot_id UUID,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		UNIQUE(project_id, name, branch, browser, viewport)
	);

	-- Snapshots table
	CREATE TABLE IF NOT EXISTS snapshots (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		build_id UUID NOT NULL REFERENCES builds(id) ON DELETE CASCADE,
		baseline_id UUID REFERENCES baselines(id) ON DELETE SET NULL,
		name VARCHAR(500) NOT NULL,
		width INTEGER,
		height INTEGER,
		browser VARCHAR(50),
		viewport VARCHAR(50),
		base_image_path VARCHAR(500),
		comparison_image_path VARCHAR(500),
		diff_image_path VARCHAR(500),
		diff_percentage DECIMAL(5, 2),
		status VARCHAR(50) DEFAULT 'pending',
		review_status VARCHAR(50) DEFAULT 'unreviewed',
		reviewed_by VARCHAR(255),
		reviewed_at TIMESTAMP WITH TIME ZONE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Indexes for performance
	CREATE INDEX IF NOT EXISTS idx_builds_project_id ON builds(project_id);
	CREATE INDEX IF NOT EXISTS idx_builds_status ON builds(status);
	CREATE INDEX IF NOT EXISTS idx_builds_branch ON builds(branch);
	CREATE INDEX IF NOT EXISTS idx_snapshots_build_id ON snapshots(build_id);
	CREATE INDEX IF NOT EXISTS idx_snapshots_status ON snapshots(status);
	CREATE INDEX IF NOT EXISTS idx_snapshots_review_status ON snapshots(review_status);
	CREATE INDEX IF NOT EXISTS idx_baselines_project_id ON baselines(project_id);
	CREATE INDEX IF NOT EXISTS idx_baselines_branch ON baselines(branch);

	-- Updated_at trigger function
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	-- Apply triggers
	DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
	CREATE TRIGGER update_projects_updated_at
		BEFORE UPDATE ON projects
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_builds_updated_at ON builds;
	CREATE TRIGGER update_builds_updated_at
		BEFORE UPDATE ON builds
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_snapshots_updated_at ON snapshots;
	CREATE TRIGGER update_snapshots_updated_at
		BEFORE UPDATE ON snapshots
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_baselines_updated_at ON baselines;
	CREATE TRIGGER update_baselines_updated_at
		BEFORE UPDATE ON baselines
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := db.Pool.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
