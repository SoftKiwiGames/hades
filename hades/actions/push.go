package actions

import (
	"context"
	"fmt"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type PushAction struct {
	Registry string
	Artifact string
	Name     string
	Tag      string
}

func NewPushAction(action *schema.ActionPush) Action {
	return &PushAction{
		Registry: action.Registry,
		Artifact: action.Artifact,
		Name:     action.Name,
		Tag:      action.Tag,
	}
}

func (a *PushAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	// Expand environment variables in fields
	registry, err := expandEnv(a.Registry, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand registry: %w", err)
	}

	artifact, err := expandEnv(a.Artifact, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand artifact: %w", err)
	}

	name, err := expandEnv(a.Name, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand name: %w", err)
	}

	tag, err := expandEnv(a.Tag, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand tag: %w", err)
	}

	// Get the registry
	reg, err := runtime.RegistryMgr.GetRegistry(registry)
	if err != nil {
		return fmt.Errorf("failed to get registry: %w", err)
	}

	// Get the artifact from artifact manager
	artifactData, err := runtime.ArtifactMgr.Get(artifact)
	if err != nil {
		return fmt.Errorf("failed to get artifact %s: %w", artifact, err)
	}
	defer artifactData.Close()

	// Push to registry
	if err := reg.Push(ctx, name, tag, artifactData); err != nil {
		return fmt.Errorf("failed to push to registry: %w", err)
	}

	return nil
}

func (a *PushAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	// Expand for dry-run display
	registry, _ := expandEnv(a.Registry, runtime.Env)
	artifact, _ := expandEnv(a.Artifact, runtime.Env)
	name, _ := expandEnv(a.Name, runtime.Env)
	tag, _ := expandEnv(a.Tag, runtime.Env)
	return fmt.Sprintf("push: artifact=%s to registry=%s as %s:%s", artifact, registry, name, tag)
}
