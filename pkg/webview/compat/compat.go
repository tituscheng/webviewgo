// Package compat provides a familiar API for users migrating from
// github.com/webview/webview_go. It wraps the modern pkg/webview API
// with the legacy function names.
package compat

import (
	"fmt"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

// WebView is an alias for compatibility.
type WebView = webview.WebView

// Hint is an alias for compatibility.
type Hint = webview.Hint

const (
	HintNone  = webview.HintNone
	HintMin   = webview.HintMin
	HintMax   = webview.HintMax
	HintFixed = webview.HintFixed
)

// New creates a webview in compatibility mode.
// Must be called from the main goroutine (same as the underlying webview
// package — the core backend locks to the main OS thread in init()).
func New(debug bool) WebView {
	wv, err := webview.New(webview.Options{
		Devtools: debug,
		AppName:  "webviewgo-compat",
	})
	if err != nil {
		panic(fmt.Sprintf("compat: failed to create webview: %v", err))
	}
	return wv
}

// Destroy releases the webview. Panics on error (matches legacy API).
func Destroy(w WebView) {
	if err := w.Destroy(); err != nil {
		panic(fmt.Sprintf("compat: Destroy failed: %v", err))
	}
}

// Run starts the event loop. Panics on error (matches legacy API).
func Run(w WebView) {
	if err := w.Run(); err != nil {
		panic(fmt.Sprintf("compat: Run failed: %v", err))
	}
}

// Terminate signals the event loop to stop.
func Terminate(w WebView) {
	w.Terminate()
}

// Navigate loads a URL. Panics on error (matches legacy API).
func Navigate(w WebView, url string) {
	if err := w.Navigate(url); err != nil {
		panic(fmt.Sprintf("compat: Navigate failed: %v", err))
	}
}

// SetTitle sets the window title.
func SetTitle(w WebView, title string) {
	w.SetTitle(title)
}

// SetSize sets the window size.
func SetSize(w WebView, width, height int, hint Hint) {
	w.SetSize(width, height, hint)
}

// Bind registers a Go function callable from JS.
func Bind(w WebView, name string, fn any) error {
	return w.Bind(name, fn)
}

// Eval evaluates JavaScript in the webview. Panics on error.
func Eval(w WebView, script string) {
	if _, err := w.Eval(script); err != nil {
		panic(fmt.Sprintf("compat: Eval failed: %v", err))
	}
}

// Init injects JavaScript before page load (best-effort via eval). Panics on error.
func Init(w WebView, js string) {
	if _, err := w.Eval(js); err != nil {
		panic(fmt.Sprintf("compat: Init failed: %v", err))
	}
}
