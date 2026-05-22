package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew_ExplicitDir(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "dev", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit dataDir is used as-is; no /profiles/<name> appended.
	if p.Dir != tmp {
		t.Fatalf("unexpected dir: %s, want %s", p.Dir, tmp)
	}
	if _, err := os.Stat(p.Dir); err != nil {
		t.Fatalf("directory not created: %v", err)
	}
}

func TestNew_EnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("WEBVIEW_DATA_DIR", tmp)
	p, err := New("testapp", "prod", "/ignored")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// WEBVIEW_DATA_DIR is treated as explicit; used as-is.
	if p.Dir != tmp {
		t.Fatalf("unexpected dir: %s, want %s", p.Dir, tmp)
	}
}

func TestNew_DefaultProfile(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit dataDir is used as-is even with empty profile name.
	if p.Dir != tmp {
		t.Fatalf("unexpected dir: %s, want %s", p.Dir, tmp)
	}
}

func TestNew_MissingAppNameAndDir(t *testing.T) {
	_, err := New("", "default", "")
	if err == nil {
		t.Fatal("expected error when both AppName and DataDir are empty")
	}
}

func TestProfile_CookieDB(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "default", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(p.Dir, "cookies.db")
	if got := p.CookieDB(); got != want {
		t.Fatalf("CookieDB() = %s, want %s", got, want)
	}
}

func TestProfile_CacheDir(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "default", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(p.Dir, "cache")
	if got := p.CacheDir(); got != want {
		t.Fatalf("CacheDir() = %s, want %s", got, want)
	}
}
