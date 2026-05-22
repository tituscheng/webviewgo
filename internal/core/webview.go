package core

import (
	"log/slog"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Platform is the minimal surface a platform backend must implement.
type Platform interface {
	Run() error
	Terminate()
	Destroy() error

	SetTitle(title string)
	SetSize(width, height int, hint types.Hint)
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

	Eval(script string) (any, error)
	Bind(name string, fn func(args []any) (any, error)) error

	RegisterScheme(scheme string, handler types.SchemeHandler) error

	OpenDialog(opts types.OpenDialogOptions) ([]string, error)
	SaveDialog(opts types.SaveDialogOptions) (string, error)
	MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error)

	ClipboardReadText() (string, error)
	ClipboardWriteText(text string) error
	Notify(title, body string) error
}

var (
	platforms sync.Map // map[uintptr]Platform
	handleSeq atomic.Uintptr
)

func nextHandle(p Platform) uintptr {
	h := handleSeq.Add(1)
	platforms.Store(h, p)
	return h
}

func releaseHandle(h uintptr) {
	platforms.Delete(h)
}

func getPlatform(h uintptr) (Platform, bool) {
	v, ok := platforms.Load(h)
	if !ok {
		return nil, false
	}
	return v.(Platform), true
}

// New creates the best available platform backend.
func New(opts types.Options) (Platform, error) {
	if opts.Headless {
		return newHeadless(opts)
	}
	return newNative(opts)
}

func logOpts(opts types.Options) *slog.Logger {
	if opts.Logger != nil {
		return opts.Logger
	}
	return slog.Default()
}

func resolveProxy(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
