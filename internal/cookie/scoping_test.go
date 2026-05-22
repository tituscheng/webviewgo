package cookie

import (
	"context"
	"testing"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Host-only cookies (set without a Domain attribute) must only be returned to
// the exact host that set them, never to subdomains.
func TestStore_HostOnlyScoping(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	hostOnly := types.Cookie{Name: "ho", Value: "1", Domain: "example.com", Path: "/", HostOnly: true}
	domainWide := types.Cookie{Name: "dw", Value: "2", Domain: "example.com", Path: "/", HostOnly: false}
	for _, c := range []types.Cookie{hostOnly, domainWide} {
		if err := s.SetCookie(ctx, c); err != nil {
			t.Fatalf("set cookie: %v", err)
		}
	}

	// Exact host: both cookies match.
	exact, err := s.GetCookies(ctx, "https://example.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(exact) != 2 {
		t.Fatalf("exact host: expected 2 cookies, got %d (%+v)", len(exact), exact)
	}

	// Subdomain: only the domain-wide cookie matches.
	sub, err := s.GetCookies(ctx, "https://sub.example.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(sub) != 1 || sub[0].Name != "dw" {
		t.Fatalf("subdomain: expected only domain-wide cookie, got %+v", sub)
	}

	// Unrelated host with the cookie domain as a suffix substring must not match.
	none, err := s.GetCookies(ctx, "https://notexample.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("unrelated host: expected 0 cookies, got %+v", none)
	}
}

// Secure cookies must only be returned over a secure channel.
func TestStore_SecureScheme(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if err := s.SetCookie(ctx, types.Cookie{Name: "sec", Value: "1", Domain: "secure.com", Path: "/", Secure: true}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}

	overHTTP, err := s.GetCookies(ctx, "http://secure.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(overHTTP) != 0 {
		t.Fatalf("expected 0 secure cookies over http, got %+v", overHTTP)
	}

	overHTTPS, err := s.GetCookies(ctx, "https://secure.com/", "")
	if err != nil {
		t.Fatalf("get cookies: %v", err)
	}
	if len(overHTTPS) != 1 {
		t.Fatalf("expected 1 secure cookie over https, got %+v", overHTTPS)
	}
}

// A malformed cookie with an empty domain must never match a request, rather
// than being broadcast everywhere.
func TestDomainMatch_EmptyDomain(t *testing.T) {
	if domainMatch("example.com", "", true) {
		t.Error("host-only cookie with empty domain should not match")
	}
	if domainMatch("example.com", "", false) {
		t.Error("domain cookie with empty domain should not match")
	}
}

// A cookie set with a leading-dot Domain attribute is normalised and matches
// both the apex and subdomains.
func TestStore_LeadingDotDomain(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()

	if err := s.SetCookie(ctx, types.Cookie{Name: "d", Value: "1", Domain: ".example.com", Path: "/"}); err != nil {
		t.Fatalf("set cookie: %v", err)
	}
	for _, host := range []string{"https://example.com/", "https://www.example.com/"} {
		got, err := s.GetCookies(ctx, host, "")
		if err != nil {
			t.Fatalf("get cookies %s: %v", host, err)
		}
		if len(got) != 1 {
			t.Fatalf("%s: expected 1 cookie, got %+v", host, got)
		}
	}
}
