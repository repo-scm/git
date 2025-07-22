//go:build linux

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/repo-scm/git/utils"
)

var (
	cleanForce bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean directories",
	Long: `Clean directories using overlayfs-aware removal methods.

Examples:
  git clean build/                    # Clean build directory
  git clean build/ temp/ cache/       # Clean multiple directories
  git clean --force /absolute/path    # Clean with absolute path (use with caution)`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, target := range args {
			if err := runClean(target); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error cleaning %s: %v\n", target, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "allow cleaning absolute paths (use with caution)")
}

func runClean(targetPath string) error {
	// Security check: prevent cleaning system directories
	if !cleanForce && filepath.IsAbs(targetPath) {
		return fmt.Errorf("absolute paths not allowed without --force flag: %s", targetPath)
	}

	// Additional safety checks for system directories
	cleanPath := filepath.Clean(targetPath)
	if isSystemPath(cleanPath) {
		return fmt.Errorf("cannot clean system directory: %s", cleanPath)
	}

	// Check if path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Printf("Path does not exist: %s\n", targetPath)
		return nil
	}

	fmt.Printf("Cleaning directory: %s\n", targetPath)

	if err := utils.OverlayClean(targetPath); err != nil {
		return fmt.Errorf("failed to clean %s: %v", targetPath, err)
	}

	fmt.Printf("Successfully cleaned: %s\n", targetPath)

	return nil
}

func isSystemPath(path string) bool {
	systemPaths := []string{
		"/", "/usr", "/etc", "/var", "/sys", "/proc", "/dev",
		"/boot", "/lib", "/lib64", "/sbin", "/bin",
	}

	cleanPath := filepath.Clean(path)
	for _, sysPath := range systemPaths {
		if cleanPath == sysPath || strings.HasPrefix(cleanPath, sysPath+"/") {
			return true
		}
	}
	return false
}
