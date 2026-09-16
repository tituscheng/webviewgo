package cookie

import (
	"net/http"
	"net/url"
	"testing"
)

func TestJar_SetAndGet(t *testing.T) {
	s := openTestStore(t)
	jar := NewJar(s, "")

	u, _ := url.Parse("https://example.com/path")
	jar.SetCookies(u, []*http.Cookie{
		{Name: "id", Value: "42"},
	})

	cookies := jar.Cookies(u)
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if cookies[0].Value != "42" {
		t.Fatalf("expected value 42, got %s", cookies[0].Value)
	}
}

func TestJar_SessionIsolation(t *testing.T) {
	s := openTestStore(t)
	jarA := NewJar(s, "sess-a")
	jarB := NewJar(s, "sess-b")

	u, _ := url.Parse("https://example.com/")
	jarA.SetCookies(u, []*http.Cookie{{Name: "a", Value: "1"}})
	jarB.SetCookies(u, []*http.Cookie{{Name: "b", Value: "2"}})

	if cookies := jarA.Cookies(u); len(cookies) != 1 || cookies[0].Name != "a" {
		t.Fatalf("jarA unexpected cookies: %+v", cookies)
	}
	if cookies := jarB.Cookies(u); len(cookies) != 1 || cookies[0].Name != "b" {
		t.Fatalf("jarB unexpected cookies: %+v", cookies)
	}
}

func TestJar_RejectsCrossSiteDomain(t *testing.T) {
	s := openTestStore(t)
	jar := NewJar(s, "")
	u, _ := url.Parse("https://evil.com/")
	jar.SetCookies(u, []*http.Cookie{
		{Name: "sid", Value: "stolen", Domain: "google.com", Path: "/"},
	})
	google, _ := url.Parse("https://google.com/")
	if got := jar.Cookies(google); len(got) != 0 {
		t.Fatalf("cross-site Domain must be ignored, got %+v", got)
	}
}

func TestJar_MaxAgeDelete(t *testing.T) {
	s := openTestStore(t)
	jar := NewJar(s, "")
	u, _ := url.Parse("https://example.com/")
	jar.SetCookies(u, []*http.Cookie{{Name: "a", Value: "1", Path: "/"}})
	jar.SetCookies(u, []*http.Cookie{{Name: "a", Value: "1", Path: "/", MaxAge: -1}})
	if got := jar.Cookies(u); len(got) != 0 {
		t.Fatalf("MaxAge=-1 should delete cookie, got %+v", got)
	}
}

func TestJar_DefaultPath(t *testing.T) {
	s := openTestStore(t)
	jar := NewJar(s, "")
	u, _ := url.Parse("https://example.com/admin/login")
	jar.SetCookies(u, []*http.Cookie{{Name: "a", Value: "1"}})
	root, _ := url.Parse("https://example.com/")
	if got := jar.Cookies(root); len(got) != 0 {
		t.Fatalf("default-path cookie must not be sent to /, got %+v", got)
	}
	admin, _ := url.Parse("https://example.com/admin/users")
	if got := jar.Cookies(admin); len(got) != 1 {
		t.Fatalf("expected cookie on /admin/*, got %+v", got)
	}
}

func TestJar_SetSessionID(t *testing.T) {
	s := openTestStore(t)
	jar := NewJar(s, "old")
	if jar.SessionID() != "old" {
		t.Fatalf("expected session old, got %s", jar.SessionID())
	}
	jar.SetSessionID("new")
	if jar.SessionID() != "new" {
		t.Fatalf("expected session new, got %s", jar.SessionID())
	}
}
