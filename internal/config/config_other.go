//go:build !windows

package config

import (
	"os"
	"path/filepath"
)

// DocumentsDir returns $HOME/Documents on non-Windows platforms.
func DocumentsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents"), nil
}
