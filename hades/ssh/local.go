package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// LocalClient runs commands on the local machine instead of over SSH
type LocalClient struct{}

func NewLocalClient() Client {
	return &LocalClient{}
}

func (c *LocalClient) Connect(ctx context.Context, host Host) (Session, error) {
	return &localSession{}, nil
}

func (c *LocalClient) Close() error {
	return nil
}

type localSession struct{}

func (s *localSession) Run(ctx context.Context, cmd string, stdout, stderr io.Writer) error {
	// Run command using shell
	execCmd := exec.CommandContext(ctx, "sh", "-c", cmd)
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func (s *localSession) CopyFile(ctx context.Context, content io.Reader, destPath string, mode uint32) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Use atomic write: write to temp file, then move
	tmpPath := destPath + ".tmp"

	// Create and write to temp file
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(tmpFile, content)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write content: %w", err)
	}

	// Atomic move
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to move file to final location: %w", err)
	}

	return nil
}

func (s *localSession) Close() error {
	return nil
}
