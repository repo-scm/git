//go:build linux

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

var (
	deleteAll bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete workspace for git repo",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		name := args[0]
		if err := runDelete(ctx, config, name); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.PersistentFlags().BoolVarP(&deleteAll, "all-workspaces", "a", false, "delete all workspaces")
}

func runDelete(ctx context.Context, cfg *config.Config, name string) error {
	sshfsPath := path.Join(utils.ExpandTilde(cfg.Sshfs.Mount), name)
	overlayPath := path.Join(utils.ExpandTilde(cfg.Overlay.Mount), name)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	if err := UnmountOverlay(ctx, overlayPath); err != nil {
		fmt.Println(err.Error())
	}

	if err := UnmountSshfs(ctx, sshfsPath); err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func UnmountOverlay(_ context.Context, mount string) error {
	if mount == "" {
		return errors.New("mount are required\n")
	}

	mountDir := path.Dir(path.Clean(mount))
	mountName := path.Base(path.Clean(mount))

	upperPath := path.Join(mountDir, "upper-"+mountName)
	workPath := path.Join(mountDir, "work-"+mountName)

	defer func(mount, workPath, upperPath string) {
		_ = os.RemoveAll(mount)
		_ = os.RemoveAll(workPath)
		_ = os.RemoveAll(upperPath)
	}(mount, workPath, upperPath)

	if err := syscall.Unmount(mount, 0); err != nil {
		return errors.Wrap(err, "failed to unmount overlay\n")
	}

	fmt.Printf("successfully unmounted overlay\n")

	return nil
}

func UnmountSshfs(_ context.Context, mount string) error {
	if mount == "" {
		return errors.New("mount is required\n")
	}

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(mount)

	cmd := exec.Command("fusermount", "-u", path.Clean(mount))

	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Printf("successfully unmounted sshfs\n")

	return nil
}
