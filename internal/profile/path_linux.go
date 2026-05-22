//go:build linux

package profile

import (
	"os"
	"path/filepath"
)

func defaultDataDir(appName string) string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, appName)
}
