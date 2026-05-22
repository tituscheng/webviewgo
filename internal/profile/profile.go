package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Profile represents a single isolated data directory.
type Profile struct {
	AppName string
	Name    string
	Dir     string
}

// New resolves and creates a profile directory.
// Resolution order:
//  1. WEBVIEW_DATA_DIR environment variable
//  2. explicit dataDir argument (if non-empty)
//  3. platform-native default based on appName + profile
//
// When an explicit dir is provided (steps 1 or 2), it is used as-is.
// Only the platform default (step 3) appends /profiles/<name>.
func New(appName, profileName, dataDir string) (*Profile, error) {
	if profileName == "" {
		profileName = "default"
	}

	dir := os.Getenv("WEBVIEW_DATA_DIR")
	explicit := true
	if dir == "" {
		dir = dataDir
	}
	if dir == "" {
		if appName == "" {
			return nil, fmt.Errorf("profile: either AppName or DataDir must be provided")
		}
		dir = defaultDataDir(appName)
		explicit = false
	}

	if !explicit {
		dir = filepath.Join(dir, "profiles", profileName)
	}

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("profile: creating directory %s: %w", dir, err)
	}

	return &Profile{
		AppName: appName,
		Name:    profileName,
		Dir:     dir,
	}, nil
}

// CookieDB returns the path to the SQLite cookie database.
func (p *Profile) CookieDB() string {
	return filepath.Join(p.Dir, "cookies.db")
}

// CacheDir returns the path to the native webview cache directory.
func (p *Profile) CacheDir() string {
	return filepath.Join(p.Dir, "cache")
}

var (
	defaultDirOnce sync.Once
	defaultDirVal  string
	defaultDirErr  error
)
