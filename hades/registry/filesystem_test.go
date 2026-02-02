package registry

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesystemRegistry(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "hades-registry-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create registry
	reg, err := NewFilesystemRegistry(tmpDir)
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	ctx := context.Background()
	name := "myapp"
	tag := "v1.0.0"
	content := "test artifact content"

	// Test Push
	t.Run("Push", func(t *testing.T) {
		reader := strings.NewReader(content)
		err := reg.Push(ctx, name, tag, reader)
		if err != nil {
			t.Fatalf("push failed: %v", err)
		}

		// Verify file exists
		filePath := filepath.Join(tmpDir, name, tag)
		if _, err := os.Stat(filePath); err != nil {
			t.Fatalf("artifact file not created: %v", err)
		}
	})

	// Test Exists
	t.Run("Exists", func(t *testing.T) {
		exists, err := reg.Exists(ctx, name, tag)
		if err != nil {
			t.Fatalf("exists check failed: %v", err)
		}
		if !exists {
			t.Fatal("artifact should exist")
		}

		// Check non-existent artifact
		exists, err = reg.Exists(ctx, "nonexistent", "v0.0.0")
		if err != nil {
			t.Fatalf("exists check failed: %v", err)
		}
		if exists {
			t.Fatal("non-existent artifact should not exist")
		}
	})

	// Test Pull
	t.Run("Pull", func(t *testing.T) {
		reader, err := reg.Pull(ctx, name, tag)
		if err != nil {
			t.Fatalf("pull failed: %v", err)
		}
		defer reader.Close()

		// Read content
		data, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read pulled artifact: %v", err)
		}

		if string(data) != content {
			t.Fatalf("content mismatch: got %q, want %q", string(data), content)
		}
	})

	// Test immutability (cannot overwrite)
	t.Run("Immutability", func(t *testing.T) {
		reader := strings.NewReader("different content")
		err := reg.Push(ctx, name, tag, reader)
		if err == nil {
			t.Fatal("should not be able to overwrite existing artifact")
		}
	})

	// Test Pull non-existent
	t.Run("PullNonExistent", func(t *testing.T) {
		_, err := reg.Pull(ctx, "nonexistent", "v0.0.0")
		if err == nil {
			t.Fatal("should fail when pulling non-existent artifact")
		}
	})
}
