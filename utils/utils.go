//go:build linux

package utils

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const (
	PermDir  = 0755
	PermFile = 0644
)

func ExpandTilde(name string) string {
	if !strings.HasPrefix(name, "~") {
		return name
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, name[1:])
}

func ParsePath(_ context.Context, name string) (user, host, dir string) {
	// Format: user@host:/remote/repo
	remotePattern := `^([^@]+)@([^:]+):([^:]+)$`
	remoteRegex := regexp.MustCompile(remotePattern)

	if matches := remoteRegex.FindStringSubmatch(name); matches != nil {
		dir = strings.Split(name, ":")[1]
		buf := strings.Split(strings.Split(name, ":")[0], "@")
		return buf[0], buf[1], dir
	}

	return "", "", name
}

func WriteTable(_ context.Context, data [][]string) error {
	table := tablewriter.NewWriter(os.Stdout)

	table.Header(data[0])
	_ = table.Bulk(data[1:])
	_ = table.Render()

	return nil
}

// OverlayClean provides a user-friendly way to clean directories in overlayfs workspaces
// Use this instead of 'rm -rf' when working inside git workspaces to avoid "Directory not empty" errors
func OverlayClean(targetPath string) error {
	if targetPath == "" {
		return nil
	}

	// First attempt: Standard removal
	if err := os.RemoveAll(targetPath); err == nil {
		return nil
	}

	// Second attempt: Fix permissions and try again
	_ = exec.Command("find", targetPath, "-type", "d", "-exec", "chmod", "755", "{}", "+").Run()
	_ = exec.Command("find", targetPath, "-type", "f", "-exec", "chmod", "644", "{}", "+").Run()
	if err := os.RemoveAll(targetPath); err == nil {
		return nil
	}

	// Third attempt: Use find with -delete (handles overlayfs better than rm -rf)
	if err := exec.Command("find", targetPath, "-depth", "-delete").Run(); err == nil {
		return nil
	}

	// Fourth attempt: Remove files first, then directories
	_ = exec.Command("find", targetPath, "-type", "f", "-delete").Run()
	_ = exec.Command("find", targetPath, "-type", "d", "-empty", "-delete").Run()

	// Final check - try to remove the target directory if it still exists
	_ = os.Remove(targetPath)

	return nil
}
