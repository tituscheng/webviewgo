# Migration Guide: webview/webview_go → webviewgo

## Compatibility Layer

For a quick drop-in replacement, use the compat package:

```go
import webview "github.com/tituscheng/webviewgo/pkg/webview/compat"

func main() {
    w := webview.New(false)
    defer webview.Destroy(w)
    w.SetTitle("Minimal webview example")
    w.SetSize(800, 600, webview.HintNone)
    w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")
    webview.Run(w)
}
```

## API Changes

### Construction

**Old:**
```go
w := webview.New(false)
```

**New:**
```go
wv, err := webview.New(webview.Options{
    Devtools: false,
    Title:    "My App",
    Width:    1024,
    Height:   768,
})
```

### Lifecycle

**Old:**
```go
w.Run()
w.Terminate()
w.Destroy()
```

**New:**
```go
wv.Run()
wv.Terminate()
wv.Destroy() // returns error; use defer
```

### Bindings

**Old:**
```go
w.Bind("add", func(a, b int) int { return a + b })
```

**New:**
```go
wv.Bind("add", func(a, b int) (int, error) { return a + b, nil })
```

Return values are wrapped in JS Promises. Errors become promise rejections.

### Cookies

Previously unavailable. Now first-class:

```go
cm := wv.CookieManager()
cm.SetCookie(webview.Cookie{...})
```

### Data Directory

Previously hardcoded or unmanaged. Now explicit:

```go
wv, _ := webview.New(webview.Options{
    AppName: "MyApp",
    Profile: "default",
})
```

Or override:

```go
os.Setenv("WEBVIEW_DATA_DIR", "/tmp/webview-data")
```
