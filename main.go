//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/repo-scm/git/mount"
)

var (
	BuildTime string
	CommitID  string
)

var (
	mountPath      string
	unmountPath    string
	repositoryPath string
)

var rootCmd = &cobra.Command{
	Use:     "git",
	Short:   "git with copy-on-write",
	Version: BuildTime + "-" + CommitID,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if err := run(ctx); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&mountPath, "mount", "m", "", "mount path")
	rootCmd.PersistentFlags().StringVarP(&unmountPath, "unmount", "u", "", "unmount path")
	rootCmd.PersistentFlags().StringVarP(&repositoryPath, "repository", "r", "", "repository path (user@host:/remote/repo:/local/repo)")

	rootCmd.MarkFlagsOneRequired("mount", "unmount")
	rootCmd.MarkFlagsMutuallyExclusive("mount", "unmount")
	_ = rootCmd.MarkFlagRequired("repository")

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	remoteRepo, localRepo := mount.ParsePath(ctx, repositoryPath)

	if unmountPath != "" {
		if err := mount.UnmountOverlay(ctx, expandTilde(localRepo), expandTilde(unmountPath)); err != nil {
			return errors.Wrap(err, "failed to unmount overlay\n")
		}
		if remoteRepo != "" {
			if err := mount.UnmountSshfs(ctx, expandTilde(localRepo)); err != nil {
				return errors.Wrap(err, "failed to unmount sshfs\n")
			}
		}
		return nil
	}

	if remoteRepo != "" {
		if err := mount.MountSshfs(ctx, remoteRepo, expandTilde(localRepo)); err != nil {
			return errors.Wrap(err, "failed to mount sshfs\n")
		}
	}

	if err := mount.MountOverlay(ctx, expandTilde(localRepo), expandTilde(mountPath)); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	return nil
}

func expandTilde(name string) string {
	if !strings.HasPrefix(name, "~") {
		return name
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, name[1:])
}
