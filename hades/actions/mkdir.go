package actions

import (
	"bytes"
	"context"
	"fmt"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type MkdirAction struct {
	Path string
	Mode uint32
}

func NewMkdirAction(action *schema.ActionMkdir) Action {
	return &MkdirAction{
		Path: action.Path,
		Mode: action.Mode,
	}
}

func (a *MkdirAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	// Create SSH session
	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	// Build mkdir command with mode
	cmd := fmt.Sprintf("mkdir -p %s && chmod %o %s", a.Path, a.Mode, a.Path)

	// Execute command
	var stdout, stderr bytes.Buffer
	if err := sess.Run(ctx, cmd, &stdout, &stderr); err != nil {
		return fmt.Errorf("mkdir command failed: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return nil
}

func (a *MkdirAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	return fmt.Sprintf("mkdir: %s (mode: %o)", a.Path, a.Mode)
}
