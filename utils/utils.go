//go:build linux

package utils

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func ParsePath(_ context.Context, name string) (remote, local string) {
	// Remote format: user@host:/remote/repo:/local/repo
	remotePattern := `^([^@]+)@([^:]+):([^:]+):([^:]+)$`
	remoteRegex := regexp.MustCompile(remotePattern)

	if matches := remoteRegex.FindStringSubmatch(name); matches != nil {
		_path := strings.Split(name, ":")
		return _path[0] + ":" + _path[1], _path[2]
	}

	return "", name
}
