package cookie

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Jar implements http.CookieJar backed by a Store.
type Jar struct {
	store     *Store
	sessionID string
	mu        sync.RWMutex
}

// NewJar creates a new Jar for the given store and optional session.
func NewJar(store *Store, sessionID string) *Jar {
	return &Jar{store: store, sessionID: sessionID}
}

// SetCookies implements http.CookieJar.
func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	ctx := context.Background()
	sid := j.SessionID()
	host := canonicalHost(u.Hostname())
	for _, hc := range cookies {
		c := fromHTTP(hc, sid)
		// A cookie with no Domain attribute is host-only: it may only be sent
		// back to the exact host that set it.
		c.HostOnly = hc.Domain == ""
		c.Domain = canonicalHost(effectiveDomain(host, hc.Domain))
		if c.Path == "" {
			c.Path = "/"
		}
		_ = j.store.SetCookie(ctx, c)
	}
}

// Cookies implements http.CookieJar.
func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	ctx := context.Background()
	items, err := j.store.GetCookies(ctx, u.String(), j.SessionID())
	if err != nil {
		return nil
	}
	var out []*http.Cookie
	for _, c := range items {
		out = append(out, toHTTP(c))
	}
	return out
}

// SessionID returns the session identifier for this jar.
func (j *Jar) SessionID() string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.sessionID
}

// SetSessionID changes the active session.
func (j *Jar) SetSessionID(id string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.sessionID = id
}

func fromHTTP(c *http.Cookie, sessionID string) types.Cookie {
	wc := types.Cookie{
		SessionID: sessionID,
		Name:      c.Name,
		Value:     c.Value,
		Domain:    c.Domain,
		Path:      c.Path,
		Secure:    c.Secure,
		HTTPOnly:  c.HttpOnly,
		Raw:       c.Raw,
	}
	if !c.Expires.IsZero() {
		wc.Expires = c.Expires
	}
	switch c.SameSite {
	case http.SameSiteLaxMode:
		wc.SameSite = types.SameSiteLax
	case http.SameSiteStrictMode:
		wc.SameSite = types.SameSiteStrict
	case http.SameSiteNoneMode:
		wc.SameSite = types.SameSiteNone
	}
	return wc
}

func toHTTP(c types.Cookie) *http.Cookie {
	hc := &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Domain:   c.Domain,
		Path:     c.Path,
		Expires:  c.Expires,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
		Raw:      c.Raw,
	}
	switch c.SameSite {
	case types.SameSiteLax:
		hc.SameSite = http.SameSiteLaxMode
	case types.SameSiteStrict:
		hc.SameSite = http.SameSiteStrictMode
	case types.SameSiteNone:
		hc.SameSite = http.SameSiteNoneMode
	}
	return hc
}

func effectiveDomain(hostname, cookieDomain string) string {
	if cookieDomain != "" {
		return cookieDomain
	}
	return hostname
}

// canonicalHost normalises a host or cookie domain for comparison: lower-cased
// with any trailing root dot and any leading "." (a Domain attribute prefix)
// removed.
func canonicalHost(host string) string {
	host = strings.TrimSuffix(host, ".")
	host = strings.TrimPrefix(host, ".")
	return strings.ToLower(host)
}
