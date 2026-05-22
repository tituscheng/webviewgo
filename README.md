# webviewgo

A production-grade, lightweight desktop webview library for Go. Inspired by
`webview/webview_go` but rewritten from scratch with modern architecture,
first-class persistence, rich JS interop, and native UI helpers.

## Features

- **SQLite-backed Cookie & Session Manager** — persistent, session-isolated,
  with optional `http.CookieJar` implementation and native store sync.
- **Rich JS ↔ Go Interop** — automatic JSON marshaling for structs, slices,
  maps; async Promise support.
- **Custom Protocols** — register custom URL schemes with `embed.FS` or
  `http.Handler` adapters.
- **Native UI Helpers** — dialogs, clipboard, notifications, frameless windows.
- **Profiles & Data Isolation** — platform-native data directories, configurable
  per-profile.
- **Headless Mode** — for CI and integration testing.
- **Cross-Platform** — macOS (WKWebView), Windows (WebView2), Linux (WebKitGTK).

## Requirements

| Platform | Requirements |
|----------|-------------|
| macOS | Xcode Command Line Tools (Cocoa, WebKit frameworks) |
| Windows | WebView2 Runtime (auto-installed on modern Windows) |
| Linux | `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` |

## Installation

```bash
go get github.com/tituscheng/webviewgo
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/tituscheng/webviewgo/pkg/webview"
)

func main() {
    wv, err := webview.New(webview.Options{
        Title:     "Hello",
        Width:     1024,
        Height:    768,
        Resizable: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer wv.Destroy()

    wv.Navigate("https://go.dev")
    wv.Run()
}
```

> `AppName` is auto-derived from the executable name when omitted.

## Examples

```bash
go run examples/basic/main.go          # Minimal webview
go run examples/cookies/main.go        # Cookie storage & session isolation
go run examples/js-interop/main.go     # Go ↔ JS binding
go run examples/custom-protocol/main.go # Custom URL scheme handler
```

## Data Directory

By default, the library resolves data directories using platform-native conventions:

- **Windows**: `%LOCALAPPDATA%\<AppName>\profiles\<Profile>\`
- **macOS**: `~/Library/Application Support/<AppName>/profiles/<Profile>/`
- **Linux**: `~/.local/share/<AppName>/profiles/<Profile>/`

Override with `WEBVIEW_DATA_DIR` or `Options.DataDir`.

## Cookies & Sessions

```go
wv, _ := webview.New(webview.Options{AppName: "myapp"})
cm := wv.CookieManager()

cm.SetCookie(webview.Cookie{
    Name:    "token",
    Value:   "abc",
    Domain:  "example.com",
    Path:    "/",
    Expires: time.Now().Add(24 * time.Hour),
})
```

## JS Interop

```go
wv.Bind("add", func(a, b int) (int, error) {
    return a + b, nil
})
```

JavaScript receives a Promise:

```js
const result = await window.add(2, 3);
```

## Headless Testing

```go
wv, _ := webview.New(webview.Options{Headless: true})
```

## Platform Matrix

See [`docs/PLATFORM_MATRIX.md`](docs/PLATFORM_MATRIX.md) for the full feature-by-platform breakdown.

## Migration from webview/webview_go

See `pkg/webview/compat/` for a drop-in replacement API, or read `MIGRATION.md`.

## License

MIT
