package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_ExplicitDir(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "dev", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(p.Dir, filepath.Join(tmp, "profiles", "dev")) {
		t.Fatalf("unexpected dir: %s", p.Dir)
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
	if !strings.HasSuffix(p.Dir, filepath.Join(tmp, "profiles", "prod")) {
		t.Fatalf("unexpected dir: %s", p.Dir)
	}
}

func TestNew_DefaultProfile(t *testing.T) {
	tmp := t.TempDir()
	p, err := New("testapp", "", tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(p.Dir, filepath.Join(tmp, "profiles", "default")) {
		t.Fatalf("unexpected dir: %s", p.Dir)
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
