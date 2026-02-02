package registry

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type filesystemRegistry struct {
	basePath string
}

func NewFilesystemRegistry(basePath string) (Registry, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create registry directory: %w", err)
	}

	return &filesystemRegistry{
		basePath: basePath,
	}, nil
}

func (r *filesystemRegistry) Push(ctx context.Context, name, tag string, data io.Reader) error {
	// Create directory structure: <basePath>/<name>/
	nameDir := filepath.Join(r.basePath, name)
	if err := os.MkdirAll(nameDir, 0755); err != nil {
		return fmt.Errorf("failed to create name directory: %w", err)
	}

	// Write to file: <basePath>/<name>/<tag>
	filePath := filepath.Join(nameDir, tag)

	// Check if already exists (immutable)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("artifact %s:%s already exists (registries are immutable)", name, tag)
	}

	// Write to temporary file first
	tmpPath := filePath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tmpPath) // Clean up if we fail

	// Copy data
	if _, err := io.Copy(file, data); err != nil {
		file.Close()
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("failed to rename to final path: %w", err)
	}

	return nil
}

func (r *filesystemRegistry) Pull(ctx context.Context, name, tag string) (io.ReadCloser, error) {
	filePath := filepath.Join(r.basePath, name, tag)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("artifact %s:%s not found in registry", name, tag)
		}
		return nil, fmt.Errorf("failed to open artifact: %w", err)
	}

	return file, nil
}

func (r *filesystemRegistry) Exists(ctx context.Context, name, tag string) (bool, error) {
	filePath := filepath.Join(r.basePath, name, tag)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
