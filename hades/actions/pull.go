package actions

import (
	"context"
	"fmt"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type PullAction struct {
	Registry string
	Name     string
	Tag      string
	To       string
}

func NewPullAction(action *schema.ActionPull) Action {
	return &PullAction{
		Registry: action.Registry,
		Name:     action.Name,
		Tag:      action.Tag,
		To:       action.To,
	}
}

func (a *PullAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	// Expand environment variables in fields
	registry, err := expandEnv(a.Registry, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand registry: %w", err)
	}

	name, err := expandEnv(a.Name, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand name: %w", err)
	}

	tag, err := expandEnv(a.Tag, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand tag: %w", err)
	}

	to, err := expandEnv(a.To, runtime.Env)
	if err != nil {
		return fmt.Errorf("failed to expand to: %w", err)
	}

	// Get the registry
	reg, err := runtime.RegistryMgr.GetRegistry(registry)
	if err != nil {
		return fmt.Errorf("failed to get registry: %w", err)
	}

	// Pull from registry
	artifact, err := reg.Pull(ctx, name, tag)
	if err != nil {
		return fmt.Errorf("failed to pull from registry: %w", err)
	}
	defer artifact.Close()

	// Create SSH session
	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	// Copy to remote host
	if err := sess.CopyFile(ctx, artifact, to, 0644); err != nil {
		return fmt.Errorf("failed to copy to host: %w", err)
	}

	return nil
}

func (a *PullAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	// Expand for dry-run display
	registry, _ := expandEnv(a.Registry, runtime.Env)
	name, _ := expandEnv(a.Name, runtime.Env)
	tag, _ := expandEnv(a.Tag, runtime.Env)
	to, _ := expandEnv(a.To, runtime.Env)
	return fmt.Sprintf("pull: %s:%s from registry=%s to %s", name, tag, registry, to)
}
