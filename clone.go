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
	cloneDepth = 1
	dirPerm    = 0755
	readSize   = 1024
)

var (
	configFile string
	repoUrl    string
	destDir    string
	unmountDir string
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
	Run: func(cmd *cobra.Command, args []string) {
		var cfg Config
		if err := validArgs(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
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
	rootCmd.PersistentFlags().StringVarP(&destDir, "dest-dir", "d", "", "dest dir")
	rootCmd.PersistentFlags().StringVarP(&unmountDir, "unmount-dir", "u", "", "unmount dir")

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

func validArgs() error {
	if unmountDir != "" {
		if repoUrl != "" || destDir != "" {
			return errors.New("please use either --unmount-dir or --repo-url/--dest-dir\n")
		}
	} else {
		if repoUrl == "" {
			return errors.New("please use --repo-url\n")
		}
		if destDir == "" {
			url := strings.TrimSuffix(repoUrl, ".git")
			destDir = path.Base(url)
		}
	}

	return nil
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

func mountFs(_ context.Context, cfg *Config) error {
	lowerDir := path.Join(destDir, cfg.Mount.Overlay.LowerDir)
	upperDir := path.Join(destDir, cfg.Mount.Overlay.UpperDir)
	workDir := path.Join(destDir, cfg.Mount.Overlay.WorkDir)
	index := cfg.Mount.Overlay.Index
	mergedDir := path.Join(destDir, cfg.Mount.Overlay.MergedDir)

	dirs := []string{lowerDir, upperDir, workDir, mergedDir}

	for _, item := range dirs {
		if err := os.MkdirAll(item, dirPerm); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create dir%s\n", item))
		}
	}

	flags := syscall.MS_BIND | syscall.MS_REC
	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s,index=%s", lowerDir, upperDir, workDir, index)

	if err := syscall.Mount("overlay", mergedDir, "overlay", uintptr(flags), opts); err != nil {
		return errors.Wrap(err, "failed to mount overlay\n")
	}

	fmt.Printf("\nSuccessfully mounted overlay fs at %s\n", mergedDir)

	return nil
}

func unmountFs(_ context.Context, cfg *Config) error {
	mergedDir := path.Join(unmountDir, cfg.Mount.Overlay.MergedDir)

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

	mergedDir := path.Join(destDir, cfg.Mount.Overlay.MergedDir)

	if cfg.Clone.SingleBranch {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--single-branch", "--", repoUrl, mergedDir)
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "--depth", strconv.Itoa(depth), "--no-single-branch", "--", repoUrl, mergedDir)
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
