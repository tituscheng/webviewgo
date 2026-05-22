// Package compat provides a familiar API for users migrating from
// github.com/webview/webview_go. It wraps the modern pkg/webview API
// with the legacy function names.
package compat

import (
	"fmt"
	"runtime"

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
func New(debug bool) WebView {
	runtime.LockOSThread()
	wv, err := webview.New(webview.Options{
		Devtools: debug,
		AppName:  "webviewgo-compat",
	})
	if err != nil {
		panic(fmt.Sprintf("compat: failed to create webview: %v", err))
	}
	return wv
}

// Destroy releases the webview.
func Destroy(w WebView) {
	_ = w.Destroy()
}

// Run starts the event loop.
func Run(w WebView) {
	_ = w.Run()
}

// Terminate signals the event loop to stop.
func Terminate(w WebView) {
	w.Terminate()
}

// Navigate loads a URL.
func Navigate(w WebView, url string) {
	_ = w.Navigate(url)
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

// Eval evaluates JavaScript in the webview.
func Eval(w WebView, script string) {
	_, _ = w.Eval(script)
}

// Init injects JavaScript before page load (best-effort via eval).
func Init(w WebView, js string) {
	_, _ = w.Eval(js)
}
