package artifacts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sync"
)

type Manager interface {
	Store(name string, data io.Reader) error
	Register(name string, path string)
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
	path     string // lazy-load: file read on first access
}

func NewManager() Manager {
	return &manager{
		artifacts: make(map[string]*artifact),
	}
}

func (m *manager) Register(name string, path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.artifacts[name] = &artifact{path: path}
}

func (m *manager) load(art *artifact) error {
	if art.data != nil {
		return nil
	}
	if art.path == "" {
		return fmt.Errorf("artifact has no path and no data")
	}
	f, err := os.Open(art.path)
	if err != nil {
		return fmt.Errorf("failed to open artifact at %s: %w", art.path, err)
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read artifact at %s: %w", art.path, err)
	}

	hash := sha256.Sum256(buf)
	art.data = buf
	art.checksum = fmt.Sprintf("%x", hash)
	return nil
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
	m.mu.Lock()
	defer m.mu.Unlock()

	art, ok := m.artifacts[name]
	if !ok {
		return nil, fmt.Errorf("artifact %q not found", name)
	}

	if err := m.load(art); err != nil {
		return nil, err
	}

	// Return a new reader each time
	return io.NopCloser(bytes.NewReader(art.data)), nil
}

func (m *manager) Checksum(name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	art, ok := m.artifacts[name]
	if !ok {
		return "", fmt.Errorf("artifact %q not found", name)
	}

	if err := m.load(art); err != nil {
		return "", err
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
