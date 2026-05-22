//go:build darwin

package profile

import (
	"os"
	"path/filepath"
)

func defaultDataDir(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, "Library", "Application Support", appName)
}
