package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/k0kubun/go-ansi"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	minArgs = 2

	cloneDepth            = 1
	mountOverlayLowerDir  = "lower"
	mountOverlayUpperDir  = "upper"
	mountOverlayWorkDir   = "work"
	mountOverlayIndex     = "off"
	mountOverlayMergedDir = "merged"

	dirPerm  = 0755
	readSize = 1024
)

var (
	configFile string
	repoUrl    string
	destPath   string
	unmountFs  bool
)

type Config struct {
	Clone Clone `yaml:"clone"`
	Mount Mount `yaml:"mount"`
}

type Clone struct {
	Depth        int  `yaml:"depth"`
	SingleBranch bool `yaml:"single_branch"`
}

type Mount struct {
	Overlay Overlay `yaml:"overlay"`
}

type Overlay struct {
	LowerDir  string `yaml:"lower_dir"`
	UpperDir  string `yaml:"upper_dir"`
	WorkDir   string `yaml:"work_dir"`
	Index     string `yaml:"index"`
	MergedDir string `yaml:"merged_dir"`
}

var rootCmd = &cobra.Command{
	Use:   "clone",
	Short: "git clone with copy-on-write",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < minArgs {
			return errors.New("invalid argument\n")
		}
		if unmountFs {
			if repoUrl != "" || destPath != "" {
				return errors.New("please use either --unmount or --repo-url --dest-path\n")
			}
		} else {
			if repoUrl == "" {
				return errors.New("please use --repo-url\n")
			}
			if destPath == "" {
				url := strings.TrimSuffix(repoUrl, ".git")
				destPath = path.Base(url)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var cfg Config
		ctx := context.Background()
		if err := viper.Unmarshal(&cfg); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := run(ctx, &cfg); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	},
}

var progressBar = progressbar.NewOptions(1000,
	progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
	progressbar.OptionEnableColorCodes(true),
	progressbar.OptionShowBytes(false),
	progressbar.OptionSetWidth(15),
	progressbar.OptionSetDescription("[cyan]Cloning repository..."),
	progressbar.OptionSetTheme(progressbar.Theme{
		Saucer:        "[green]=[reset]",
		SaucerHead:    "[green]>[reset]",
		SaucerPadding: " ",
		BarStart:      "[",
		BarEnd:        "]",
	}),
)

// nolint:gochecknoinits
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config-file", "c", "", "config file")
	rootCmd.PersistentFlags().StringVarP(&repoUrl, "repo-url", "r", "", "repo url")
	rootCmd.PersistentFlags().StringVarP(&destPath, "dest-path", "d", "", "dest path")
	rootCmd.PersistentFlags().BoolVarP(&unmountFs, "unmount-fs", "u", false, "unmount fs")

	_ = rootCmd.MarkFlagRequired("config-file")

	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig() {
	if configFile == "" {
		return
	}

	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
}

func run(ctx context.Context, cfg *Config) error {
	if unmountFs {
		if err := unmountFilesystem(ctx, cfg); err != nil {
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

func mountFs(_ context.Context, cfg *Config) error {
	lowerDir := mountOverlayLowerDir
	if cfg.Mount.Overlay.LowerDir != "" {
		lowerDir = cfg.Mount.Overlay.LowerDir
	}
	lowerDir = path.Join(destPath, lowerDir)

	upperDir := mountOverlayUpperDir
	if cfg.Mount.Overlay.UpperDir != "" {
		upperDir = cfg.Mount.Overlay.UpperDir
	}
	upperDir = path.Join(destPath, upperDir)

	workDir := mountOverlayWorkDir
	if cfg.Mount.Overlay.WorkDir != "" {
		workDir = cfg.Mount.Overlay.WorkDir
	}
	workDir = path.Join(destPath, workDir)

	index := mountOverlayIndex
	if cfg.Mount.Overlay.Index != "" {
		index = cfg.Mount.Overlay.Index
	}

	mergedDir := mountOverlayMergedDir
	if cfg.Mount.Overlay.MergedDir != "" {
		mergedDir = cfg.Mount.Overlay.MergedDir
	}
	mergedDir = path.Join(destPath, mergedDir)

	dirs := []string{lowerDir, upperDir, workDir, mergedDir}

	for _, item := range dirs {
		if err := os.MkdirAll(item, dirPerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create directory %s\n", item))
		}
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=%s", lowerDir, upperDir, workDir, index)

	if err := syscall.Mount("overlay", mergedDir, "overlay", 0, opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("\nSuccessfully mounted overlay fs at %s\n", mergedDir)

	return nil
}

func unmountFilesystem(_ context.Context, cfg *Config) error {
	_path := path.Join(destPath, cfg.Mount.Overlay.MergedDir)

	if err := syscall.Unmount(_path, 0); err != nil {
		return errors.Wrap(err, "failed to unmount overlay\n")
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

	_path := path.Join(destPath, cfg.Mount.Overlay.MergedDir)

	if cfg.Clone.SingleBranch {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--single-branch", repoUrl, _path)
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--no-single-branch", repoUrl, _path)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to get stdout pipe\n")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start cmd\n")
	}

	updateProgressBar(stdout)

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "failed to wait cmd\n")
	}

	fmt.Printf("\nSuccessfully cloned repo %s\n", repoUrl)

	return nil
}

func updateProgressBar(closer io.ReadCloser) {
	buf := make([]byte, readSize)

	for {
		n, err := closer.Read(buf)
		if n > 0 {
			_ = progressBar.Add(1)
		}
		if err != nil {
			break
		}
	}
}
