package types

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// Options configures a new WebView instance.
type Options struct {
	DataDir     string
	AppName     string
	Profile     string
	Title       string
	Width       int
	Height      int
	Resizable   bool
	Frameless   bool
	Transparent bool
	Center      bool
	UserAgent   string
	// Proxy configures an upstream HTTP proxy for the webview's network stack.
	//
	// Not yet implemented by any backend; setting it currently has no effect.
	Proxy    *url.URL
	Devtools bool
	Headless bool
	Logger   *slog.Logger
	// OnDrop is invoked when files are dropped onto the window.
	//
	// Not yet implemented by any backend; the callback is currently never
	// invoked.
	OnDrop func(files []string)
}

// Cookie represents an HTTP cookie with session isolation.
type Cookie struct {
	SessionID string
	Name      string
	Value     string
	Domain    string
	Path      string
	Expires   time.Time
	Secure    bool
	HTTPOnly  bool
	// HostOnly marks a cookie that may only be sent to the exact host that set
	// it (i.e. the cookie was set without an explicit Domain attribute). When
	// false the cookie is a domain cookie and also matches subdomains.
	HostOnly bool
	SameSite SameSite
	Raw      string
}

// SameSite describes the SameSite attribute.
type SameSite int

const (
	SameSiteNone SameSite = iota
	SameSiteLax
	SameSiteStrict
)

// CookieManager controls cookie storage and synchronization.
type CookieManager interface {
	SetCookie(c Cookie) error
	GetCookies(url string, sessionID string) ([]Cookie, error)
	DeleteCookie(name, domain, path string) error
	Clear() error
	ClearSession(sessionID string) error
	SaveSession(sessionID string) error
	LoadSession(sessionID string) error
	AsJar() http.CookieJar
	Close() error
}

// SchemeHandler handles requests for a custom URL scheme.
type SchemeHandler func(req *Request) *Response

// Request represents a custom scheme request.
type Request struct {
	Method  string
	URL     string
	Headers http.Header
	Body    []byte
}

// Response is the handler's reply.
type Response struct {
	StatusCode    int
	Headers       http.Header
	Body          io.Reader
	ContentLength int64
}

// Hint controls how window size is interpreted.
type Hint int

const (
	HintNone Hint = iota
	HintMin
	HintMax
	HintFixed
)

// DialogResult is the outcome of a message dialog.
type DialogResult int

const (
	DialogCancel DialogResult = iota
	DialogOK
	DialogYes
	DialogNo
	DialogAbort
	DialogRetry
	DialogIgnore
)

// DialogLevel controls the message dialog icon.
type DialogLevel int

const (
	DialogInfo DialogLevel = iota
	DialogWarning
	DialogError
	DialogQuestion
)

// DialogButtons controls which buttons appear.
type DialogButtons int

const (
	DialogButtonsOK DialogButtons = iota
	DialogButtonsOKCancel
	DialogButtonsYesNo
	DialogButtonsYesNoCancel
	DialogButtonsRetryCancel
	DialogButtonsAbortRetryIgnore
)

// OpenDialogOptions configures a file-open dialog.
type OpenDialogOptions struct {
	Title         string
	Directory     string
	DefaultFile   string
	Filters       []FileFilter
	AllowFiles    bool
	AllowDirs     bool
	AllowMultiple bool
}

// SaveDialogOptions configures a file-save dialog.
type SaveDialogOptions struct {
	Title       string
	Directory   string
	DefaultFile string
	Filters     []FileFilter
}

// MessageDialogOptions configures a message dialog.
type MessageDialogOptions struct {
	Title   string
	Message string
	Level   DialogLevel
	Buttons DialogButtons
}

// FileFilter describes a file type filter for dialogs.
type FileFilter struct {
	DisplayName string
	Pattern     string
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Title:     "WebView",
		Width:     1024,
		Height:    768,
		Resizable: true,
		Center:    true,
		Profile:   "default",
		Logger:    slog.Default(),
	}
}
