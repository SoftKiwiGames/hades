package registry

import (
	"context"
	"fmt"
	"io"

	"github.com/SoftKiwiGames/hades/hades/schema"
)

type Registry interface {
	Push(ctx context.Context, name, tag string, data io.Reader) error
	Pull(ctx context.Context, name, tag string) (io.ReadCloser, error)
	Exists(ctx context.Context, name, tag string) (bool, error)
}

type Manager interface {
	GetRegistry(name string) (Registry, error)
}

type manager struct {
	registries map[string]Registry
}

func NewManager(configs schema.Registries) (Manager, error) {
	m := &manager{
		registries: make(map[string]Registry),
	}

	// Initialize registries from config
	for name, config := range configs {
		reg, err := createRegistry(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create registry %q: %w", name, err)
		}
		m.registries[name] = reg
	}

	return m, nil
}

func (m *manager) GetRegistry(name string) (Registry, error) {
	reg, ok := m.registries[name]
	if !ok {
		return nil, fmt.Errorf("registry %q not found", name)
	}
	return reg, nil
}

func createRegistry(config schema.RegistryConfig) (Registry, error) {
	switch config.Type {
	case "filesystem":
		if config.Path == "" {
			return nil, fmt.Errorf("filesystem registry requires 'path' field")
		}
		return NewFilesystemRegistry(config.Path)
	case "s3":
		if config.Bucket == "" {
			return nil, fmt.Errorf("s3 registry requires 'bucket' field")
		}
		return NewS3Registry(config.Bucket, config.Region, config.Endpoint)
	default:
		return nil, fmt.Errorf("unknown registry type: %s", config.Type)
	}
}
