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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/repo-scm/git/config"
	"github.com/repo-scm/git/utils"
)

var (
	createName string

	sshfsOptions = []string{
		"allow_other",
		"cache=yes",
		"cache_timeout=115200",
		"compression=no",
		"default_permissions",
		"follow_symlinks",
		"Cipher=aes128-ctr",
		"StrictHostKeyChecking=no",
		"UserKnownHostsFile=/dev/null",
		"ConnectTimeout=10",
		"ServerAliveInterval=15",
		"ServerAliveCountMax=3",
	}
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create workspace",
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

	createCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Usage:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  %s %s <repo_name> [flags]\n\n", cmd.Root().Name(), cmd.Name())
		if cmd.HasLocalFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Flags:\n")
			cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.Name != "help" && flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		if cmd.HasInheritedFlags() {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nGlobal Flags:\n")
			cmd.InheritedFlags().VisitAll(func(flag *pflag.Flag) {
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  -%s, --%s   %s", flag.Shorthand, flag.Name, flag.Usage)
				if flag.DefValue != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStderr(), " (default %s)", flag.DefValue)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n")
			})
		}
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\nExample:\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git create /local/repo --name your_workspace\n")
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  git create user@host:/remote/repo --name your_workspace\n")
		return nil
	})
}

func generateHash(_ string) string {
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
			if err = MountSshfs(ctx, repoPath, sshfsPath, port); err == nil {
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

func getSshfsOptions() []string {
	// Check SSHFS version to determine compatible options
	cmd := exec.Command("sshfs", "--version")
	output, err := cmd.Output()
	if err != nil {
		return sshfsOptions
	}

	versionStr := string(output)
	options := sshfsOptions

	// Add newer options if SSHFS version supports them
	// Note: big_writes was deprecated in FUSE 3.x and causes issues, so we'll skip it
	// kernel_cache is supported in most versions, but let's be conservative for very old versions
	if !strings.Contains(versionStr, "SSHFS version 2.8") &&
		!strings.Contains(versionStr, "SSHFS version 2.7") &&
		!strings.Contains(versionStr, "SSHFS version 2.6") {
		options = append(options, "kernel_cache")
	}

	return options
}

func MountSshfs(ctx context.Context, repo, mount string, port int) error {
	var cmdArgs []string

	if repo == "" || mount == "" {
		return errors.New("repo and mount are required\n")
	}

	// Parse the repo to get connection details
	user, host, _ := utils.ParsePath(ctx, repo)
	if user == "" || host == "" {
		return errors.New("invalid repo format, expected user@host:/path\n")
	}

	if err := os.MkdirAll(mount, utils.PermDir); err != nil {
		return errors.Wrap(err, "failed to make directory\n")
	}

	if err := os.Chown(mount, os.Getuid(), os.Getgid()); err != nil {
		fmt.Printf("Warning: failed to set ownership of mount point %s: %v\n", mount, err)
	}

	cmdArgs = []string{repo, path.Clean(mount)}
	cmdArgs = append(cmdArgs, "-o", fmt.Sprintf("port=%d", port))

	if os.Getuid() == 0 {
		cmdArgs = append(cmdArgs, "-o", "umask=022")
	} else {
		cmdArgs = append(cmdArgs, "-o", fmt.Sprintf("uid=%d,gid=%d,umask=022", os.Getuid(), os.Getgid()))
	}

	// Get version-appropriate SSHFS options
	versionOptions := getSshfsOptions()
	for _, opt := range versionOptions {
		cmdArgs = append(cmdArgs, "-o", opt)
	}

	// Create command with context timeout
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "sshfs", cmdArgs...)
	cmd.Env = append(os.Environ(), "SSHFS_TIMEOUT=30")

	output, err := cmd.CombinedOutput()
	if err != nil {
		if removeErr := os.RemoveAll(mount); removeErr != nil {
			fmt.Printf("Warning: failed to clean up mount directory %s: %v\n", mount, removeErr)
		}
		errorMsg := string(output)
		if strings.Contains(errorMsg, "Permission denied") || strings.Contains(errorMsg, "password") {
			return errors.Wrapf(err, "failed to mount sshfs - ensure ssh key authentication is set up for user %s", os.Getenv("USER"))
		}
		return errors.Wrapf(err, "failed to mount sshfs: %s", errorMsg)
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
		if err := os.Chown(item, os.Getuid(), os.Getgid()); err != nil {
			fmt.Printf("Warning: failed to set ownership of %s: %v\n", item, err)
		}
	}

	// Test write access to upper directory before mounting
	testFile := path.Join(upperPath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), utils.PermFile); err != nil {
		return errors.Wrap(err, "failed to write test file to upper directory - check permissions\n")
	}
	_ = os.Remove(testFile)

	cmd := exec.Command("fuse-overlayfs",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", path.Clean(repo), upperPath, workPath),
		path.Clean(mount),
	)

	if err := cmd.Run(); err != nil {
		dirsToRemove := []string{mount, upperPath, workPath}
		for _, dir := range dirsToRemove {
			if dir != mountDir {
				if removeErr := os.RemoveAll(dir); removeErr != nil {
					fmt.Printf("Warning: failed to clean up directory %s: %v\n", dir, removeErr)
				}
			}
		}
		return errors.Wrap(err, "failed to mount overlay with fuse-overlayfs\n")
	}

	fmt.Printf("successfully mounted overlay at %s\n", mount)

	return nil
}
