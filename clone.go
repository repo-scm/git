//go:build linux

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	branchName = "master"

	cloneDepth    = 1
	cloneDuration = 100 * time.Millisecond

	overlaySourceDir = "source"
	overlayLowerDir  = "lower"
	overlayUpperDir  = "upper"
	overlayWorkDir   = "work"
	overlayIndex     = "off"
	overlayMergedDir = "merged"

	dirPerm = 0755
)

var (
	BuildTime string
	CommitID  string
)

var (
	configFile string
	repoUrl    string
	repoBranch string
	destDir    string
	unmountDir string
)

type Config struct {
	Clone Clone `yaml:"clone"`
}

type Clone struct {
	Depth        int  `yaml:"depth"`
	SingleBranch bool `yaml:"single_branch"`
}

var rootCmd = &cobra.Command{
	Use:     "clone",
	Short:   "git clone with copy-on-write",
	Version: BuildTime + "-" + CommitID,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		if err := validArgs(ctx); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		cfg, err := loadConfig(ctx)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := run(ctx, &cfg); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

// nolint:gochecknoinits
func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&destDir, "dest-dir", "d", "", "dest dir")
	rootCmd.PersistentFlags().StringVarP(&repoUrl, "repo-url", "r", "", "repo url")
	rootCmd.PersistentFlags().StringVarP(&repoBranch, "repo-branch", "b", "", "repo branch")
	rootCmd.PersistentFlags().StringVarP(&unmountDir, "unmount-dir", "u", "", "unmount dir")

	_ = rootCmd.MarkFlagRequired("config-file")

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func validArgs(_ context.Context) error {
	if unmountDir != "" {
		if repoUrl != "" || destDir != "" {
			return errors.New("please use either --unmount-dir or --repo-url/--dest-dir\n")
		}
	} else {
		if repoUrl == "" {
			return errors.New("please use --repo-url\n")
		}
		if repoBranch == "" {
			repoBranch = branchName
		}
		if destDir == "" {
			url := strings.TrimSuffix(repoUrl, ".git")
			destDir = path.Base(url)
		}
	}

	return nil
}

func loadConfig(_ context.Context) (Config, error) {
	var cfg Config

	if configFile == "" {
		return cfg, errors.New("invalid config file\n")
	}

	if _, err := os.Stat(configFile); err != nil {
		return cfg, errors.New("config file not found\n")
	}

	buf, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, errors.Wrap(err, "failed to read file\n")
	}

	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return cfg, errors.Wrap(err, "failed to unmarshal file\n")
	}

	return cfg, nil
}

func run(ctx context.Context, cfg *Config) error {
	if unmountDir != "" {
		if err := unmountFs(ctx, cfg); err != nil {
			return errors.Wrap(err, "failed to unmount fs\n")
		}
		return nil
	}

	if err := mountFs(ctx, cfg); err != nil {
		return errors.Wrap(err, "failed to mount fs\n")
	}

	if err := cloneRepo(ctx, cfg); err != nil {
		return errors.Wrap(err, "failed to clone repo\n")
	}

	return nil
}

func mountFs(_ context.Context, _ *Config) error {
	sourceDir := path.Join(destDir, overlaySourceDir)
	lowerDir := path.Join(destDir, overlayLowerDir)
	upperDir := path.Join(destDir, overlayUpperDir)
	workDir := path.Join(destDir, overlayWorkDir)
	index := overlayIndex
	mergedDir := path.Join(destDir, overlayMergedDir)

	dirs := []string{sourceDir, lowerDir, upperDir, workDir, mergedDir}

	for _, item := range dirs {
		if err := os.MkdirAll(item, dirPerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create dir%s\n", item))
		}
	}

	flags := syscall.MS_BIND | syscall.MS_REC
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=%s", lowerDir, upperDir, workDir, index)

	if err := syscall.Mount(sourceDir, mergedDir, "overlay", uintptr(flags), opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("\nSuccessfully mounted overlay fs at %s\n", mergedDir)

	return nil
}

func unmountFs(_ context.Context, _ *Config) error {
	mergedDir := path.Join(unmountDir, overlayMergedDir)

	if err := syscall.Unmount(mergedDir, 0); err != nil {
		return errors.Wrap(err, "failed to unmount overlay fs\n")
	}

	if err := os.RemoveAll(unmountDir); err != nil {
		return errors.Wrap(err, "failed to remove dir\n")
	}

	fmt.Printf("\nSuccessfully unmounted overlay fs\n")

	return nil
}

func cloneRepo(ctx context.Context, cfg *Config) error {
	var cmd *exec.Cmd
	var depth int

	if cfg.Clone.Depth > cloneDepth {
		depth = cfg.Clone.Depth
	} else {
		depth = cloneDepth
	}

	mergedDir := path.Join(destDir, overlayMergedDir)

	if cfg.Clone.SingleBranch {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--single-branch", "--branch", repoBranch, "--", repoUrl, mergedDir)
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--no-single-branch", "--branch", repoBranch, "--", repoUrl, mergedDir)
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start cmd\n")
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	spinner := []string{"-", "\\", "|", "/"}
	i := 0

	for {
		select {
		case err := <-done:
			if err != nil {
				return errors.Wrap(err, "failed to wait cmd\n")
			} else {
				fmt.Printf("\nSuccessfully cloned repo %s\n", repoUrl)
				return nil
			}
		case <-time.After(cloneDuration):
			fmt.Printf("\rCloning repository... %s", spinner[i])
			i = (i + 1) % len(spinner)
		}
	}
}
