package core

import (
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
)

// headlessWebView is a no-UI backend for CI/testing.
type headlessWebView struct {
	mu         sync.RWMutex
	url        string
	html       string
	baseURL    string
	title      string
	width      int
	height     int
	terminated atomic.Bool
	bindings   map[string]func([]any) (any, error)
	schemes    map[string]types.SchemeHandler
	evals      []string
}

func newHeadless(opts types.Options) (Platform, error) {
	return &headlessWebView{
		title:    opts.Title,
		width:    opts.Width,
		height:   opts.Height,
		bindings: make(map[string]func([]any) (any, error)),
		schemes:  make(map[string]types.SchemeHandler),
	}, nil
}

func (w *headlessWebView) Run() error {
	for !w.terminated.Load() {
		// Avoid busy-waiting; 100ms is responsive enough for headless tests.
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (w *headlessWebView) Terminate() {
	w.terminated.Store(true)
}

func (w *headlessWebView) Destroy() error {
	w.Terminate()
	return nil
}

func (w *headlessWebView) SetTitle(title string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.title = title
}

func (w *headlessWebView) SetSize(width, height int, hint types.Hint) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.width = width
	w.height = height
}

func (w *headlessWebView) SetMinSize(width, height int) {}
func (w *headlessWebView) SetMaxSize(width, height int) {}
func (w *headlessWebView) SetFullscreen(bool)           {}
func (w *headlessWebView) SetAlwaysOnTop(bool)          {}
func (w *headlessWebView) Show()                        {}
func (w *headlessWebView) Hide()                        {}

func (w *headlessWebView) Navigate(urlStr string) error {
	if _, err := url.Parse(urlStr); err != nil {
		return fmt.Errorf("headless: invalid url: %w", err)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.url = urlStr
	return nil
}

func (w *headlessWebView) LoadHTML(html, baseURL string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.html = html
	w.baseURL = baseURL
	return nil
}

func (w *headlessWebView) Reload()  {}
func (w *headlessWebView) Back()    {}
func (w *headlessWebView) Forward() {}

func (w *headlessWebView) Eval(script string) (any, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evals = append(w.evals, script)
	return nil, nil
}

func (w *headlessWebView) Bind(name string, fn func(args []any) (any, error)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bindings[name] = fn
	return nil
}

func (w *headlessWebView) RegisterScheme(scheme string, handler types.SchemeHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.schemes[scheme] = handler
	return nil
}

func (w *headlessWebView) OpenDialog(opts types.OpenDialogOptions) ([]string, error) {
	return nil, fmt.Errorf("headless: OpenDialog not supported")
}

func (w *headlessWebView) SaveDialog(opts types.SaveDialogOptions) (string, error) {
	return "", fmt.Errorf("headless: SaveDialog not supported")
}

func (w *headlessWebView) MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error) {
	return types.DialogCancel, fmt.Errorf("headless: MessageDialog not supported")
}

func (w *headlessWebView) ClipboardReadText() (string, error) {
	return "", nil
}

func (w *headlessWebView) ClipboardWriteText(text string) error {
	return nil
}

func (w *headlessWebView) Notify(title, body string) error {
	return nil
}
