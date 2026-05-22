package cookie

import (
	"context"
	"testing"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_SetAndGet(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	c := types.Cookie{
		Name:   "session",
		Value:  "abc123",
		Domain: "example.com",
		Path:   "/",
	}
	if err := s.SetCookie(ctx, c); err != nil {
		t.Fatalf("set cookie: %v", err)
	}

	cookies, err := s.GetCookies(ctx, "https://example.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Value != "abc123" {
		t.Fatalf("expected value abc123, got %s", cookies[0].Value)
	}
}

func TestStore_Get_NoMatch(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	cookies, err := s.GetCookies(ctx, "https://other.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies, got %d", len(cookies))
	}
}

func TestStore_SessionIsolation(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	c1 := types.Cookie{SessionID: "sess-a", Name: "a", Value: "1", Domain: "x.com", Path: "/"}
	c2 := types.Cookie{SessionID: "sess-b", Name: "b", Value: "2", Domain: "x.com", Path: "/"}
	c3 := types.Cookie{SessionID: "", Name: "c", Value: "3", Domain: "x.com", Path: "/"}

	for _, c := range []types.Cookie{c1, c2, c3} {
		if err := s.SetCookie(ctx, c); err != nil {
			t.Fatalf("set cookie: %v", err)
		}
	}

	// Query with sess-a should return a + persistent c
	cookies, err := s.GetCookies(ctx, "https://x.com/", "sess-a")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
}

func TestStore_DeleteCookie(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	c := types.Cookie{Name: "del", Value: "v", Domain: "del.com", Path: "/"}
	if err := s.SetCookie(ctx, c); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	if err := s.DeleteCookie(ctx, "del", "del.com", "/"); err != nil {
		t.Fatalf("delete cookie: %v", err)
	}
	cookies, err := s.GetCookies(ctx, "https://del.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies after delete, got %d", len(cookies))
	}
}

func TestStore_Clear(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if err := s.SetCookie(ctx, types.Cookie{Name: "p", Value: "1", Domain: "c.com", Path: "/"}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	if err := s.SetCookie(ctx, types.Cookie{SessionID: "s", Name: "s", Value: "2", Domain: "c.com", Path: "/"}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	if err := s.Clear(ctx); err != nil {
		t.Fatalf("clear: %v", err)
	}

	cookies, err := s.GetCookies(ctx, "https://c.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 persistent cookies after clear, got %d", len(cookies))
	}

	// Session cookies should remain
	cookies, err = s.GetCookies(ctx, "https://c.com/", "s")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 session cookie after clear, got %d", len(cookies))
	}
}

func TestStore_ClearSession(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if err := s.SetCookie(ctx, types.Cookie{SessionID: "s1", Name: "a", Value: "1", Domain: "cs.com", Path: "/"}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	if err := s.SetCookie(ctx, types.Cookie{SessionID: "s2", Name: "b", Value: "2", Domain: "cs.com", Path: "/"}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	if err := s.ClearSession(ctx, "s1"); err != nil {
		t.Fatalf("clear session: %v", err)
	}

	cookies, err := s.GetCookies(ctx, "https://cs.com/", "s1")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 cookies for cleared session, got %d", len(cookies))
	}

	cookies, err = s.GetCookies(ctx, "https://cs.com/", "s2")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie for remaining session, got %d", len(cookies))
	}
}

func TestStore_Expired(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	c := types.Cookie{
		Name:    "exp",
		Value:   "v",
		Domain:  "ex.com",
		Path:    "/",
		Expires: time.Now().Add(-time.Hour),
	}
	if err := s.SetCookie(ctx, c); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	cookies, err := s.GetCookies(ctx, "https://ex.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 0 {
		t.Fatalf("expected 0 expired cookies, got %d", len(cookies))
	}
}

func TestStore_Update(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	c := types.Cookie{Name: "u", Value: "1", Domain: "up.com", Path: "/"}
	if err := s.SetCookie(ctx, c); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	c.Value = "2"
	if err := s.SetCookie(ctx, c); err != nil {
		t.Fatalf("update cookie: %v", err)
	}
	cookies, err := s.GetCookies(ctx, "https://up.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(cookies) != 1 || cookies[0].Value != "2" {
		t.Fatalf("expected updated value 2, got %+v", cookies)
	}
}
