package webview

import (
	"net/http"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Cookie represents an HTTP cookie with session isolation.
type Cookie = types.Cookie

// SameSite describes the SameSite attribute.
type SameSite = types.SameSite

const (
	SameSiteNone   = types.SameSiteNone
	SameSiteLax    = types.SameSiteLax
	SameSiteStrict = types.SameSiteStrict
)

// CookieManager controls cookie storage and synchronization.
type CookieManager = types.CookieManager

// Concrete-type compile-time interface checks live in webview.go, which can
// import the internal cookie package.

// CookieToHTTP converts a webview.Cookie to *http.Cookie.
func CookieToHTTP(c Cookie) *http.Cookie {
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
	case SameSiteLax:
		hc.SameSite = http.SameSiteLaxMode
	case SameSiteStrict:
		hc.SameSite = http.SameSiteStrictMode
	case SameSiteNone:
		hc.SameSite = http.SameSiteNoneMode
	}
	return hc
}

// HTTPToCookie converts an *http.Cookie to a webview.Cookie.
func HTTPToCookie(c *http.Cookie, sessionID string) Cookie {
	wc := Cookie{
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
		wc.SameSite = SameSiteLax
	case http.SameSiteStrictMode:
		wc.SameSite = SameSiteStrict
	case http.SameSiteNoneMode:
		wc.SameSite = SameSiteNone
	}
	return wc
}
