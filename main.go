//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
	sshfsPort      int
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
	rootCmd.PersistentFlags().IntVarP(&sshfsPort, "port", "p", 22, "sshfs port")

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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	if unmountPath != "" {
		if err := mount.UnmountOverlay(ctx, expandTilde(localRepo), expandTilde(unmountPath)); err != nil {
			fmt.Print(err.Error())
		}
		if remoteRepo != "" {
			if err := mount.UnmountSshfs(ctx, expandTilde(localRepo)); err != nil {
				fmt.Print(err.Error())
			}
		}
		return nil
	}

	if remoteRepo != "" {
		if err := mount.MountSshfs(ctx, remoteRepo, expandTilde(localRepo), sshfsPort); err != nil {
			_ = mount.UnmountSshfs(ctx, expandTilde(localRepo))
			return err
		}
	}

	if err := mount.MountOverlay(ctx, expandTilde(localRepo), expandTilde(mountPath)); err != nil {
		_ = mount.UnmountSshfs(ctx, expandTilde(localRepo))
		return err
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
