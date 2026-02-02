package artifacts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"sync"
)

type Manager interface {
	Store(name string, data io.Reader) error
	Get(name string) (io.ReadCloser, error)
	Checksum(name string) (string, error)
	List() []string
	Clear()
}

type manager struct {
	mu        sync.RWMutex
	artifacts map[string]*artifact
}

type artifact struct {
	data     []byte
	checksum string
}

func NewManager() Manager {
	return &manager{
		artifacts: make(map[string]*artifact),
	}
}

func (m *manager) Store(name string, data io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read all data
	buf, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("failed to read artifact data: %w", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(buf)
	checksum := fmt.Sprintf("%x", hash)

	m.artifacts[name] = &artifact{
		data:     buf,
		checksum: checksum,
	}

	return nil
}

func (m *manager) Get(name string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	art, ok := m.artifacts[name]
	if !ok {
		return nil, fmt.Errorf("artifact %q not found", name)
	}

	// Return a new reader each time
	return io.NopCloser(bytes.NewReader(art.data)), nil
}

func (m *manager) Checksum(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	art, ok := m.artifacts[name]
	if !ok {
		return "", fmt.Errorf("artifact %q not found", name)
	}

	return art.checksum, nil
}

func (m *manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []string
	for name := range m.artifacts {
		names = append(names, name)
	}
	return names
}

func (m *manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.artifacts = make(map[string]*artifact)
}
