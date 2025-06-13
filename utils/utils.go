//go:build linux

package utils

import (
	"context"
	"os"
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
