//go:build windows

package profile

import (
	"os"
	"path/filepath"
)

func defaultDataDir(appName string) string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.Getenv("APPDATA")
	}
	if localAppData == "" {
		localAppData = "."
	}
	return filepath.Join(localAppData, appName)
}
