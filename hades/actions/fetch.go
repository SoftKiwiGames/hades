package actions

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SoftKiwiGames/hades/hades/schema"
	"github.com/SoftKiwiGames/hades/hades/types"
)

type FetchAction struct {
	Src string
	Dst string
}

func NewFetchAction(action *schema.ActionFetch) Action {
	return &FetchAction{
		Src: action.Src,
		Dst: action.Dst,
	}
}

func (a *FetchAction) Execute(ctx context.Context, runtime *types.Runtime) error {
	src := ExpandEnvVars(a.Src, runtime.Env)
	dst := runtime.ResolvePath(ExpandEnvVars(a.Dst, runtime.Env))

	sess, err := runtime.SSHClient.Connect(ctx, runtime.Host)
	if err != nil {
		return fmt.Errorf("failed to connect to host: %w", err)
	}
	defer sess.Close()

	remoteChecksum, _, err := getRemoteChecksum(ctx, sess, src)
	if err != nil {
		return fmt.Errorf("failed to get remote checksum: %w", err)
	}

	if remoteChecksum != "" {
		if f, err := os.Open(dst); err == nil {
			localChecksum, err := calculateChecksum(f)
			f.Close()
			if err == nil && localChecksum == remoteChecksum {
				fmt.Fprintf(runtime.Stdout, "Skipping %s (already up to date)\n", dst)
				return nil
			}
		}
	}

	reader, err := sess.ReadFile(ctx, src)
	if err != nil {
		return err
	}
	defer reader.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", dst, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return fmt.Errorf("failed to write local file: %w", err)
	}

	fmt.Fprintf(runtime.Stdout, "Fetched %s:%s to %s\n", runtime.Host.Name, src, dst)
	return nil
}

func (a *FetchAction) DryRun(ctx context.Context, runtime *types.Runtime) string {
	src := ExpandEnvVars(a.Src, runtime.Env)
	dst := runtime.ResolvePath(ExpandEnvVars(a.Dst, runtime.Env))
	return fmt.Sprintf("fetch: %s:%s to %s", runtime.Host.Name, src, dst)
}
