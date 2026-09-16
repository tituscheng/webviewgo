# Platform Support Matrix

Feature-by-platform capability matrix. This is the source of truth for what works where.

## Legend

- ✅ **Full** — Feature works correctly on this platform
- ⚠️ **Partial** — Feature works but has known limitations
- 🚧 **Stub** — Code exists but is non-functional or incomplete
- ❌ **Missing** — Not implemented
- ➖ **N/A** — Not applicable to this platform

---

## Core Lifecycle

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `New()` — create window/webview | ✅ | ✅ | ✅ | ✅ |
| `Run()` — event loop | ✅ | ✅ | ✅ | ✅ |
| `Terminate()` — signal stop | ✅ | ✅ | ✅ | ✅ |
| `Destroy()` — cleanup | ✅ | ✅ | ✅ | ✅ |
| `Show()` / `Hide()` | ✅ | ✅ | ✅ | ➖ |
| `SetTitle()` | ✅ | ✅ | ✅ | ✅ |
| `SetSize()` | ✅ | ✅ | ✅ | ✅ |
| `SetMinSize()` | ✅ | ✅ | ⚠️ | ➖ |
| `SetMaxSize()` | ✅ | ✅ | ⚠️ | ➖ |
| `SetFullscreen()` | ✅ | ✅ | ⚠️ | ➖ |
| `SetAlwaysOnTop()` | ✅ | ✅ | ⚠️ | ➖ |
| Frameless window | ⚠️ (style only) | ❌ | ❌ | ➖ |
| Transparent window | ⚠️ (window only) | ❌ | ❌ | ➖ |
| Center window | ✅ | ❌ | ❌ | ➖ |

## Navigation

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `Navigate(url)` | ✅ | ✅ | ✅ | ✅ (stores URL) |
| `LoadHTML(html, baseURL)` | ✅ | ✅ | ✅ | ✅ (stores HTML) |
| `Reload()` | ✅ | ✅ | ✅ | ➖ |
| `Back()` | ✅ | ✅ | ✅ | ➖ |
| `Forward()` | ✅ | ✅ | ✅ | ➖ |

## JavaScript Interop

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `Eval(script)` | ✅ | ✅ | ✅ | ✅ (stores script) |
| `Bind(name, fn)` — register | ✅ | ✅ | ✅ | ✅ |
| `BindRaw(name, fn)` | ✅ | ✅ | ✅ | ✅ |
| `Bind()` — JS→Go call | ✅ | ✅ | ✅ | ➖ |
| `Bind()` — Go→JS response | ✅ | ✅ | ✅ | ➖ |
| Promise wrapper | ✅ | ✅ | ✅ | ➖ |
| Bridge callback validation | ✅ | ✅ | ✅ | ➖ |
| TypeScript generation | ✅ | ✅ | ✅ | ➖ |

## Custom Protocols

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `Options.Schemes` (pre-register) | ✅ | ✅ | ➖ | ✅ |
| `RegisterScheme()` after `New()` | ⚠️ (error) | ✅ | ❌ | ✅ |
| `FSHandler` adapter | ✅ | ✅ | ✅ | ✅ |
| `HTTPHandler` adapter | ✅ | ✅ | ✅ | ✅ |
| Response delivery | ✅ | ✅ | ❌ | ➖ |
| Request body forwarding | ✅ (`HTTPBody` + `HTTPBodyStream`) | ✅ | ➖ | ➖ |

## Cookies & Sessions

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| SQLite cookie store | ✅ | ✅ | ✅ | ✅ |
| `http.CookieJar` impl | ✅ | ✅ | ✅ | ✅ |
| Session isolation | ✅ | ✅ | ✅ | ✅ |
| Native cookie sync | ✅ | ❌ | ❌ | ➖ |
| `SameSite` sync | ✅ | ❌ | ❌ | ➖ |
| `HostOnly` sync | ✅ | ❌ | ❌ | ➖ |

## Dialogs

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `OpenDialog()` | ✅ | ✅ | ✅ | ❌ |
| `SaveDialog()` | ✅ | ✅ | ✅ | ❌ |
| `MessageDialog()` | ✅ | ✅ | ✅ | ❌ |
| File filters | ✅ | ⚠️ | ⚠️ | ➖ |
| Multiple file selection | ✅ | ✅ | ✅ | ➖ |

## System Integration

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `ClipboardReadText()` | ✅ | ✅ | ✅ | ✅ (returns `""`) |
| `ClipboardWriteText()` | ✅ | ✅ | ✅ | ✅ (no-op) |
| `Notify()` | ✅ (deprecated API) | ❌ | ❌ | ✅ (no-op) |
| Drag & Drop | ❌ | ❌ | ❌ | ➖ |
| Menu bar | ❌ | ❌ | ❌ | ➖ |
| System tray | ❌ | ❌ | ❌ | ➖ |

## Developer Experience

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| DevTools toggle | ✅ | ✅ | ❌ | ➖ |
| Headless mode | ✅ | ✅ | ✅ | ✅ |
| Profile isolation | ✅ | ✅ | ✅ | ✅ |
| `WEBVIEW_DATA_DIR` override | ✅ | ✅ | ✅ | ✅ |
| Custom UserAgent | ✅ | ✅ | ⚠️ (JS fallback) | ➖ |
| Proxy support | ❌ | ❌ | ❌ | ➖ |
| Integration tests (`//go:build integration`) | ✅ | ✅ | ✅ | ✅ |

---

## Platform Readiness Summary

| Platform | Status | Usable For |
|----------|--------|------------|
| **macOS** | 🟢 Production-ready | Navigation, JS interop, custom schemes (via `Options.Schemes`), dialogs, cookies, clipboard |
| **Linux** | 🟡 Mostly Ready | Navigation, JS interop, custom schemes, dialogs, window chrome; notifications missing |
| **Windows** | 🟡 Mostly Ready | WebView2 init, navigation, JS interop, dialogs, clipboard; `RegisterScheme` missing |
| **Headless** | 🟢 CI-ready | All non-UI operations work; useful for unit and integration testing |
