package actions

import (
	"context"
	"fmt"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type RunAction struct {
	Command string
}

func NewRunAction(action *schema.ActionRun) Action {
	return &RunAction{
		Command: string(*action),
	}
}

func (a *RunAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	// Create SSH session
	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	// Expand environment variables in the command
	cmd := ExpandEnvVars(a.Command, runtime.Env)

	// Execute command - use runtime's stdout/stderr to ensure output goes to logs
	if err := sess.Run(ctx, cmd, runtime.Stdout, runtime.Stderr); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func (a *RunAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	// Expand environment variables for dry-run display
	cmd := ExpandEnvVars(a.Command, runtime.Env)
	return fmt.Sprintf("run: %s", cmd)
}
