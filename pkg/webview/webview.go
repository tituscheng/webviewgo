package webview

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/tituscheng/webviewgo/internal/cookie"
	"github.com/tituscheng/webviewgo/internal/core"
	"github.com/tituscheng/webviewgo/internal/js"
	"github.com/tituscheng/webviewgo/internal/profile"
	"github.com/tituscheng/webviewgo/internal/types"
)

// WebView is the primary interface for controlling a native webview window.
type WebView interface {
	Run() error
	Terminate()
	Destroy() error

	SetTitle(title string)
	SetSize(width, height int, hint Hint)
	SetMinSize(width, height int)
	SetMaxSize(width, height int)
	SetFullscreen(fullscreen bool)
	SetAlwaysOnTop(alwaysOnTop bool)
	Show()
	Hide()

	Navigate(url string) error
	LoadHTML(html, baseURL string) error
	Reload()
	Back()
	Forward()

	Eval(script string) error
	Bind(name string, fn any) error

	// BindRaw is the high-performance variant of Bind. It bypasses reflection
	// and the standard JSON (de)serialization used by Bind. The handler receives
	// the raw JSON-encoded argument array (as produced by the browser's
	// postMessage) and must return a JSON-encoded value (as a string that will
	// be inserted directly into the resolve expression).
	//
	// Use BindRaw for hot paths where you control serialization (e.g. with
	// jsoniter, msgpack, protobuf, or a binary format encoded inside a JSON string).
	BindRaw(name string, fn func(args json.RawMessage) (json.RawMessage, error)) error

	CookieManager() CookieManager
	RegisterScheme(scheme string, handler SchemeHandler) error

	OpenDialog(opts OpenDialogOptions) ([]string, error)
	SaveDialog(opts SaveDialogOptions) (string, error)
	MessageDialog(opts MessageDialogOptions) (DialogResult, error)

	ClipboardReadText() (string, error)
	ClipboardWriteText(text string) error
	Notify(title, body string) error
}

// webview is the concrete implementation of WebView.
type webview struct {
	platform   core.Platform
	cm         types.CookieManager
	logger     *slog.Logger
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	userAgent  string
	uaInjected bool
}

// New creates a WebView with the given options.
func New(opts Options) (WebView, error) {
	return NewWithContext(context.Background(), opts)
}

// NewWithContext creates a WebView bound to a context.
func NewWithContext(ctx context.Context, opts Options) (WebView, error) {
	if opts.Profile == "" {
		opts.Profile = "default"
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.AppName == "" && opts.DataDir == "" {
		if exe := os.Args[0]; exe != "" {
			opts.AppName = filepath.Base(exe)
		} else {
			opts.AppName = "webviewgo"
		}
	}

	p, err := core.New(opts)
	if err != nil {
		return nil, fmt.Errorf("webview: create platform backend: %w", err)
	}

	prof, err := profile.New(opts.AppName, opts.Profile, opts.DataDir)
	if err != nil {
		return nil, fmt.Errorf("webview: resolve profile: %w", err)
	}

	cm, err := cookie.NewManager(prof.CookieDB())
	if err != nil {
		return nil, fmt.Errorf("webview: open cookie manager: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	wv := &webview{
		platform: p,
		cm:       cm,
		logger:   opts.Logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	if syncer, ok := p.(interface{ SyncCookiesToNative([]types.Cookie) error }); ok {
		cm.SetSyncCallback(syncer.SyncCookiesToNative)
	}

	if opts.UserAgent != "" {
		if setter, ok := p.(interface{ SetUserAgent(string) }); ok {
			setter.SetUserAgent(opts.UserAgent)
		} else {
			// Fallback for platforms without native UA support (e.g. Windows).
			// Injected lazily on first Navigate/LoadHTML to avoid racing
			// with document creation.
			wv.userAgent = opts.UserAgent
		}
	}

	go func() {
		<-ctx.Done()
		wv.Terminate()
	}()

	return wv, nil
}

func (w *webview) Run() error {
	return w.platform.Run()
}

func (w *webview) Terminate() {
	w.cancel()
	w.platform.Terminate()
}

func (w *webview) Destroy() error {
	w.cancel()
	if err := w.cm.Close(); err != nil {
		w.logger.Warn("cookie manager close failed", "error", err)
	}
	return w.platform.Destroy()
}

func (w *webview) SetTitle(title string)                { w.platform.SetTitle(title) }
func (w *webview) SetSize(width, height int, hint Hint) { w.platform.SetSize(width, height, hint) }
func (w *webview) SetMinSize(width, height int)         { w.platform.SetMinSize(width, height) }
func (w *webview) SetMaxSize(width, height int)         { w.platform.SetMaxSize(width, height) }
func (w *webview) SetFullscreen(fullscreen bool)        { w.platform.SetFullscreen(fullscreen) }
func (w *webview) SetAlwaysOnTop(alwaysOnTop bool)      { w.platform.SetAlwaysOnTop(alwaysOnTop) }
func (w *webview) Show()                                { w.platform.Show() }
func (w *webview) Hide()                                { w.platform.Hide() }
func (w *webview) Navigate(url string) error {
	w.injectUserAgent()
	return w.platform.Navigate(url)
}

func (w *webview) LoadHTML(html, baseURL string) error {
	w.injectUserAgent()
	return w.platform.LoadHTML(html, baseURL)
}

func (w *webview) injectUserAgent() {
	w.mu.Lock()
	if w.userAgent == "" || w.uaInjected {
		w.mu.Unlock()
		return
	}
	w.uaInjected = true
	ua := w.userAgent
	w.mu.Unlock()

	script := fmt.Sprintf(`navigator.__defineGetter__('userAgent', function(){ return %q; });`, ua)
	if err := w.platform.Eval(script); err != nil {
		w.logger.Warn("user-agent fallback injection failed", "error", err)
	}
}
func (w *webview) Reload()                  { w.platform.Reload() }
func (w *webview) Back()                    { w.platform.Back() }
func (w *webview) Forward()                 { w.platform.Forward() }
func (w *webview) Eval(script string) error { return w.platform.Eval(script) }

func (w *webview) Bind(name string, fn any) error {
	if !isValidJSIdentifier(name) {
		return fmt.Errorf("webview: Bind %q: name must be a valid JavaScript identifier", name)
	}
	bridge, err := js.Wrap(fn)
	if err != nil {
		return fmt.Errorf("webview: Bind %q: %w", name, err)
	}
	return w.platform.Bind(name, bridge)
}

func (w *webview) BindRaw(name string, fn func(args json.RawMessage) (json.RawMessage, error)) error {
	if !isValidJSIdentifier(name) {
		return fmt.Errorf("webview: BindRaw %q: name must be a valid JavaScript identifier", name)
	}
	return w.platform.BindRaw(name, fn)
}

// isValidJSIdentifier reports whether name is a safe top-level identifier to
// interpolate into the injected bridge script. This prevents script injection
// through the binding name (which is otherwise written verbatim into JS).
func isValidJSIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '_', r == '$':
			// always allowed
		case (r >= '0' && r <= '9') && i > 0:
			// digits allowed, but not as the first character
		default:
			return false
		}
	}
	return true
}

func (w *webview) CookieManager() CookieManager {
	return w.cm
}

func (w *webview) RegisterScheme(scheme string, handler SchemeHandler) error {
	return w.platform.RegisterScheme(scheme, handler)
}

func (w *webview) OpenDialog(opts OpenDialogOptions) ([]string, error) {
	return w.platform.OpenDialog(opts)
}

func (w *webview) SaveDialog(opts SaveDialogOptions) (string, error) {
	return w.platform.SaveDialog(opts)
}

func (w *webview) MessageDialog(opts MessageDialogOptions) (DialogResult, error) {
	return w.platform.MessageDialog(opts)
}

func (w *webview) ClipboardReadText() (string, error)   { return w.platform.ClipboardReadText() }
func (w *webview) ClipboardWriteText(text string) error { return w.platform.ClipboardWriteText(text) }
func (w *webview) Notify(title, body string) error      { return w.platform.Notify(title, body) }

var _ WebView = (*webview)(nil)

// Compile-time interface checks for internal wiring.
var _ CookieManager = (*cookie.Manager)(nil)
var _ http.CookieJar = (*cookie.Jar)(nil)
