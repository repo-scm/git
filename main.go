//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	directoryPerm = 0755
)

var (
	BuildTime string
	CommitID  string
)

var (
	mountPath      string
	repositoryPath string
	unmountPath    string
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
	rootCmd.PersistentFlags().StringVarP(&repositoryPath, "repository", "r", "", "repository path")

	rootCmd.MarkFlagsOneRequired("mount", "unmount")
	rootCmd.MarkFlagsMutuallyExclusive("mount", "unmount")

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	if unmountPath != "" {
		if err := unmountFs(ctx); err != nil {
			return errors.Wrap(err, "failed to unmount fs\n")
		}
		return nil
	}

	if err := mountFs(ctx); err != nil {
		return errors.Wrap(err, "failed to mount fs\n")
	}

	return nil
}

func mountFs(_ context.Context) error {
	if repositoryPath == "" {
		return errors.New("repository path is required\n")
	}

	repositoryDir := path.Dir(path.Clean(repositoryPath))

	mountDir := path.Dir(path.Clean(mountPath))
	mountName := path.Base(path.Clean(mountPath))

	upperPath := path.Join(repositoryDir, "cow-"+mountName)
	workPath := path.Join(mountDir, "work-"+mountName)

	dirs := []string{mountPath, upperPath, workPath}

	for _, item := range dirs {
		if err := os.MkdirAll(item, directoryPerm); err != nil {
			return errors.Wrap(err, "failed to make directory\n")
		}
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=off", repositoryPath, upperPath, workPath)

	if err := syscall.Mount("overlay", mountPath, "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("\nSuccessfully mounted overlay fs at %s\n", upperPath)

	return nil
}

func unmountFs(_ context.Context) error {
	repositoryDir := path.Dir(path.Clean(repositoryPath))

	unmountDir := path.Dir(path.Clean(unmountPath))
	unmountName := path.Base(path.Clean(unmountPath))

	upperPath := path.Join(repositoryDir, "cow-"+unmountName)
	workPath := path.Join(unmountDir, "work-"+unmountName)

	if err := syscall.Unmount(unmountPath, 0); err != nil {
		return errors.Wrap(err, "failed to unmount overlay fs\n")
	}

	if err := os.RemoveAll(unmountPath); err != nil {
		return errors.Wrap(err, "failed to remove directory\n")
	}

	if err := os.RemoveAll(workPath); err != nil {
		return errors.Wrap(err, "failed to remove directory\n")
	}

	if err := os.RemoveAll(upperPath); err != nil {
		return errors.Wrap(err, "failed to remove directory\n")
	}

	fmt.Printf("\nSuccessfully unmounted overlay fs\n")

	return nil
}
