package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// Storage handles file storage operations
type Storage struct {
	basePath string
}

// New creates a new Storage instance
func New(basePath string) (*Storage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Storage{basePath: basePath}, nil
}

// StorageType represents different storage categories
type StorageType string

const (
	StorageTypeBaseline   StorageType = "baselines"
	StorageTypeSnapshot   StorageType = "snapshots"
	StorageTypeDiff       StorageType = "diffs"
	StorageTypeComparison StorageType = "comparisons"
)

// SaveFile saves a file and returns the relative path
func (s *Storage) SaveFile(projectID uuid.UUID, storageType StorageType, filename string, reader io.Reader) (string, error) {
	// Create directory structure: basePath/projectID/storageType/
	dir := filepath.Join(s.basePath, projectID.String(), string(storageType))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(dir, uniqueFilename)

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path from base
	relativePath := filepath.Join(projectID.String(), string(storageType), uniqueFilename)
	return relativePath, nil
}

// SaveFileWithName saves a file with a specific name (for baselines)
func (s *Storage) SaveFileWithName(projectID uuid.UUID, storageType StorageType, filename string, reader io.Reader) (string, error) {
	dir := filepath.Join(s.basePath, projectID.String(), string(storageType))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	fullPath := filepath.Join(dir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	relativePath := filepath.Join(projectID.String(), string(storageType), filename)
	return relativePath, nil
}

// GetFile returns a reader for the given path
func (s *Storage) GetFile(relativePath string) (*os.File, error) {
	fullPath := filepath.Join(s.basePath, relativePath)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// GetFullPath returns the full filesystem path for a relative path
func (s *Storage) GetFullPath(relativePath string) string {
	return filepath.Join(s.basePath, relativePath)
}

// DeleteFile removes a file
func (s *Storage) DeleteFile(relativePath string) error {
	fullPath := filepath.Join(s.basePath, relativePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DeleteProjectFiles removes all files for a project
func (s *Storage) DeleteProjectFiles(projectID uuid.UUID) error {
	dir := filepath.Join(s.basePath, projectID.String())
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("failed to delete project files: %w", err)
	}
	return nil
}

// CopyFile copies a file from one path to another
func (s *Storage) CopyFile(srcRelativePath string, projectID uuid.UUID, storageType StorageType, filename string) (string, error) {
	srcFile, err := s.GetFile(srcRelativePath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	return s.SaveFile(projectID, storageType, filename, srcFile)
}

// Exists checks if a file exists
func (s *Storage) Exists(relativePath string) bool {
	fullPath := filepath.Join(s.basePath, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}
