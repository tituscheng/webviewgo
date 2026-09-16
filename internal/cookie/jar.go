package cookie

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Jar implements http.CookieJar backed by a Store.
type Jar struct {
	store     *Store
	sessionID string
	mu        sync.RWMutex
	flushFn   func() error
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
	changed := false
	for _, hc := range cookies {
		if hc == nil {
			continue
		}
		path := hc.Path
		if path == "" {
			path = defaultPath(u.Path)
		}
		domain, hostOnly, ok := cookieDomainForHost(host, hc.Domain)
		if !ok {
			continue
		}
		if hc.MaxAge < 0 {
			_ = j.store.DeleteCookie(ctx, sid, hc.Name, domain, path)
			changed = true
			continue
		}
		c := fromHTTP(hc, sid)
		c.HostOnly = hostOnly
		c.Domain = domain
		c.Path = path
		_ = j.store.SetCookie(ctx, c)
		changed = true
	}
	if changed {
		j.flush()
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
	switch {
	case c.MaxAge > 0:
		wc.Expires = time.Now().Add(time.Duration(c.MaxAge) * time.Second)
	case !c.Expires.IsZero():
		wc.Expires = c.Expires
	}
	switch c.SameSite {
	case http.SameSiteLaxMode:
		wc.SameSite = types.SameSiteLax
	case http.SameSiteStrictMode:
		wc.SameSite = types.SameSiteStrict
	case http.SameSiteNoneMode:
		wc.SameSite = types.SameSiteNone
	default:
		wc.SameSite = types.SameSiteDefault
	}
	return wc
}

func toHTTP(c types.Cookie) *http.Cookie {
	hc := &http.Cookie{
		Name:     c.Name,
		Value:    c.Value,
		Path:     c.Path,
		Expires:  c.Expires,
		Secure:   c.Secure,
		HttpOnly: c.HTTPOnly,
		Raw:      c.Raw,
	}
	if !c.HostOnly {
		hc.Domain = c.Domain
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

func (j *Jar) flush() {
	if j.flushFn == nil {
		return
	}
	_ = j.flushFn()
}

// cookieDomainForHost applies RFC 6265 §5.3 Domain-attribute checks. ok is
// false when the cookie must be ignored.
func cookieDomainForHost(host, cookieDomain string) (domain string, hostOnly, ok bool) {
	if cookieDomain == "" {
		return host, true, host != ""
	}
	domain = canonicalHost(cookieDomain)
	if domain == "" {
		return "", false, false
	}
	if net.ParseIP(host) != nil {
		// IP hosts only accept an exact-match Domain, otherwise host-only.
		if domain != host {
			return "", false, false
		}
		return host, true, true
	}
	if !domainMatch(host, domain, false) {
		return "", false, false
	}
	// Heuristic public-suffix guard without an extra dependency: a Domain
	// with no dot (e.g. "com") would otherwise match every host in that TLD.
	if domain != host && !strings.Contains(domain, ".") {
		return "", false, false
	}
	return domain, false, true
}

// defaultPath is RFC 6265 §5.1.4: the directory of the request URI, not "/".
func defaultPath(reqPath string) string {
	if reqPath == "" || reqPath[0] != '/' {
		return "/"
	}
	i := strings.LastIndex(reqPath, "/")
	if i <= 0 {
		return "/"
	}
	return reqPath[:i]
}

// canonicalHost normalises a host or cookie domain for comparison: lower-cased
// with any trailing root dot and any leading "." (a Domain attribute prefix)
// removed. IP-literal hosts (IPv4, or IPv6 already stripped of brackets by
// url.Hostname) pass through unchanged and only ever compare by exact match.
func canonicalHost(host string) string {
	host = strings.TrimSuffix(host, ".")
	host = strings.TrimPrefix(host, ".")
	return strings.ToLower(host)
}
