package webview

import (
	"context"
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
	platform        core.Platform
	cm              types.CookieManager
	logger          *slog.Logger
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	userAgent       string
	uaInjected      bool
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
	if w.userAgent == "" || w.uaInjected {
		return
	}
	w.uaInjected = true
	script := fmt.Sprintf(`navigator.__defineGetter__('userAgent', function(){ return %q; });`, w.userAgent)
	if err := w.platform.Eval(script); err != nil {
		w.logger.Warn("user-agent fallback injection failed", "error", err)
	}
}
func (w *webview) Reload()                              { w.platform.Reload() }
func (w *webview) Back()                                { w.platform.Back() }
func (w *webview) Forward()                             { w.platform.Forward() }
func (w *webview) Eval(script string) error             { return w.platform.Eval(script) }

func (w *webview) Bind(name string, fn any) error {
	bridge, err := js.Wrap(fn)
	if err != nil {
		return fmt.Errorf("webview: Bind %q: %w", name, err)
	}
	return w.platform.Bind(name, bridge)
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
var _ http.CookieJar = (func() http.CookieJar {
	m, _ := cookie.NewManager(":memory:")
	if m != nil {
		defer m.Close()
		return m.AsJar()
	}
	return nil
})()
