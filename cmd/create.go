//go:build linux

package cmd

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

var (
	createName string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create workspace for git repo",
	Args:  cobra.RangeArgs(1, 1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config := GetConfig()
		repo := args[0]
		name := createName
		if name == "" {
			name = fmt.Sprintf("%s-%s", path.Base(repo), generateHash(repo))
		}
		if err := runCreate(ctx, config, repo, name); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.PersistentFlags().StringVarP(&createName, "name", "n", "", "workspace name")
}

func generateHash(name string) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 7)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}

	return string(result)
}

func runCreate(ctx context.Context, cfg *config.Config, repo, name string) error {
	repoPath := utils.ExpandTilde(repo)
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

	user, host, _ := utils.ParsePath(ctx, repoPath)

	if user != "" && host != "" {
		var mounted bool
		var err error
		for _, port := range cfg.Sshfs.Ports {
			if err = MountSshfs(ctx, repoPath, sshfsPath, strings.Join(cfg.Sshfs.Options, ","), port); err == nil {
				mounted = true
				break
			} else {
				_ = UnmountSshfs(ctx, sshfsPath)
			}
		}
		if !mounted {
			return err
		}
		repoPath = sshfsPath
	}

	if err := MountOverlay(ctx, repoPath, overlayPath); err != nil {
		_ = UnmountSshfs(ctx, sshfsPath)
		return err
	}

	return nil
}

func MountSshfs(_ context.Context, repo, mount, options string, port int) error {
	if repo == "" || mount == "" {
		return errors.New("repo and mount are required\n")
	}

	if err := os.MkdirAll(mount, utils.PermDir); err != nil {
		return errors.Wrap(err, "failed to make directory\n")
	}

	cmd := exec.Command("sshfs",
		repo,
		path.Clean(mount),
		"-o", options,
		"-o", fmt.Sprintf("port=%d", port),
	)

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "failed to mount sshfs\n")
	}

	fmt.Printf("successfully mounted sshfs at %s\n", path.Clean(mount))

	return nil
}

func MountOverlay(_ context.Context, repo, mount string) error {
	if repo == "" || mount == "" {
		return errors.New("repo and mount are required\n")
	}

	mountDir := path.Dir(path.Clean(mount))
	mountName := path.Base(path.Clean(mount))

	upperPath := path.Join(mountDir, "upper-"+mountName)
	workPath := path.Join(mountDir, "work-"+mountName)

	dirs := []string{mount, upperPath, workPath}

	for _, item := range dirs {
		if err := os.MkdirAll(item, utils.PermDir); err != nil {
			return errors.Wrap(err, "failed to make directory\n")
		}
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=off", path.Clean(repo), upperPath, workPath)

	if err := syscall.Mount("overlay", mount, "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("successfully mounted overlay at %s\n", mount)

	return nil
}
