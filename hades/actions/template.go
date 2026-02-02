package actions

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type TemplateAction struct {
	Src string
	Dst string
}

func NewTemplateAction(action *schema.ActionTemplate) Action {
	return &TemplateAction{
		Src: action.Src,
		Dst: action.Dst,
	}
}

func (a *TemplateAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	// Read template file
	tmplData, err := os.ReadFile(a.Src)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", a.Src, err)
	}

	// Parse template
	tmpl, err := template.New(a.Src).Parse(string(tmplData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Build template context
	data := map[string]interface{}{
		"Env":    runtime.Env,
		"Host":   runtime.Host.Name,
		"Target": runtime.Target,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create SSH session
	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	// Copy rendered template to remote host
	if err := sess.CopyFile(ctx, &buf, a.Dst, 0644); err != nil {
		return fmt.Errorf("failed to copy rendered template to %s: %w", a.Dst, err)
	}

	return nil
}

func (a *TemplateAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	return fmt.Sprintf("template: %s -> %s", a.Src, a.Dst)
}
