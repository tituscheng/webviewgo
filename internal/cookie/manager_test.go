package cookie

import (
	"testing"

	"github.com/tituscheng/webviewgo/internal/types"
)

func TestManager_CRUD(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	if err := m.SetCookie(types.Cookie{Name: "a", Value: "1", Domain: "x.com", Path: "/"}); err != nil {
		t.Fatalf("SetCookie: %v", err)
	}

	cookies, err := m.GetCookies("https://x.com/", "")
	if err != nil {
		t.Fatalf("GetCookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	if err := m.DeleteCookie("a", "x.com", "/"); err != nil {
		t.Fatalf("DeleteCookie: %v", err)
	}

	cookies, err = m.GetCookies("https://x.com/", "")
	if err != nil {
		t.Fatalf("GetCookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies, got %d", len(cookies))
	}
}

func TestManager_Clear(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	if err := m.SetCookie(types.Cookie{Name: "p", Value: "1", Domain: "c.com", Path: "/"}); err != nil {
		t.Fatalf("SetCookie: %v", err)
	}
	if err := m.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	cookies, err := m.GetCookies("https://c.com/", "")
	if err != nil {
		t.Fatalf("GetCookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies after clear, got %d", len(cookies))
	}
}

func TestManager_ClearSession(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	if err := m.SetCookie(types.Cookie{SessionID: "s1", Name: "a", Value: "1", Domain: "cs.com", Path: "/"}); err != nil {
		t.Fatalf("SetCookie: %v", err)
	}
	if err := m.ClearSession("s1"); err != nil {
		t.Fatalf("ClearSession: %v", err)
	}
	cookies, err := m.GetCookies("https://cs.com/", "s1")
	if err != nil {
		t.Fatalf("GetCookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 session cookies, got %d", len(cookies))
	}
}

func TestManager_LoadSession(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	if err := m.SetCookie(types.Cookie{SessionID: "s1", Name: "a", Value: "1", Domain: "ls.com", Path: "/"}); err != nil {
		t.Fatalf("SetCookie: %v", err)
	}

	var synced bool
	m.SetSyncCallback(func(cookies []types.Cookie) error {
		synced = true
		return nil
	})

	if err := m.LoadSession("s1"); err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if !synced {
		t.Fatal("expected sync callback to be called")
	}
}

func TestManager_AsJar(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	jar := m.AsJar()
	if jar == nil {
		t.Fatal("expected non-nil jar")
	}
}

func TestManager_SaveSession(t *testing.T) {
	m, err := NewManager(":memory:")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer m.Close()

	// SaveSession is a no-op for SQLite manager
	if err := m.SaveSession("s1"); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
}
