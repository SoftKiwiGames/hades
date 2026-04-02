package actions

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// Expand env vars in src and dst
	src := ExpandEnvVars(a.Src, runtime.Env)
	dst := ExpandEnvVars(a.Dst, runtime.Env)

	// Read template file
	resolvedSrc := runtime.ResolvePath(src)
	tmplData, err := os.ReadFile(resolvedSrc)
	if err != nil {
		return fmt.Errorf("failed to read template file %s: %w", resolvedSrc, err)
	}

	// Parse template
	tmplDir := filepath.Dir(resolvedSrc)
	tmpl, err := template.New(src).Funcs(template.FuncMap{
		"readFile": func(path string) (string, error) {
			if !filepath.IsAbs(path) {
				path = filepath.Join(tmplDir, path)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("readFile: %w", err)
			}
			return string(data), nil
		},
	}).Parse(string(tmplData))
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

	// Write rendered template to intermediate file for inspection
	// Structure: logs/<runID>/rendered/<hostName>/<templatePath>
	renderedPath := filepath.Join("logs", runtime.RunID, "rendered", runtime.Host.Name, src)
	if err := os.MkdirAll(filepath.Dir(renderedPath), 0755); err != nil {
		return fmt.Errorf("failed to create rendered directory: %w", err)
	}
	if err := os.WriteFile(renderedPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write rendered template: %w", err)
	}

	// Create SSH session
	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	// Copy rendered template to remote host
	if err := sess.CopyFile(ctx, &buf, dst, 0644); err != nil {
		return fmt.Errorf("failed to copy rendered template to %s: %w", dst, err)
	}

	return nil
}

func (a *TemplateAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	src := ExpandEnvVars(a.Src, runtime.Env)
	dst := ExpandEnvVars(a.Dst, runtime.Env)
	return fmt.Sprintf("template: %s -> %s", src, dst)
}
