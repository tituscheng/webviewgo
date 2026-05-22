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
| `SetMinSize()` | ✅ | ❌ | ❌ | ❌ |
| `SetMaxSize()` | ✅ | ❌ | ❌ | ❌ |
| `SetFullscreen()` | ✅ | ❌ | ❌ | ❌ |
| `SetAlwaysOnTop()` | ✅ | ❌ | ❌ | ❌ |
| Frameless window | ⚠️ (style only) | ❌ | ❌ | ➖ |
| Transparent window | ❌ | ❌ | ❌ | ➖ |
| Center window | ✅ (always) | ❌ | ❌ | ➖ |

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
| `Bind()` — JS→Go call | ✅ | ✅ | ❌ | ➖ |
| `Bind()` — Go→JS response | ✅ | ✅ | ❌ | ➖ |
| Promise wrapper | ✅ | ✅ | ❌ | ➖ |
| TypeScript generation | ❌ | ❌ | ❌ | ➖ |

## Custom Protocols

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `RegisterScheme()` | ✅ | ❌ | ❌ | ✅ |
| `FSHandler` adapter | ❌ | ❌ | ❌ | ❌ |
| `HTTPHandler` adapter | ❌ | ❌ | ❌ | ❌ |
| Response delivery | ✅ | ❌ | ❌ | ➖ |
| Request body forwarding | ✅ | ➖ | ➖ | ➖ |

## Cookies & Sessions

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| SQLite cookie store | ✅ | ✅ | ✅ | ✅ |
| `http.CookieJar` impl | ✅ | ✅ | ✅ | ✅ |
| Session isolation | ✅ | ✅ | ✅ | ✅ |
| Native cookie sync | ✅ | ❌ | ❌ | ➖ |
| `SameSite` sync | ❌ | ❌ | ❌ | ➖ |

## Dialogs

| Feature | macOS | Linux | Windows | Headless |
|---------|:-----:|:-----:|:-------:|:--------:|
| `OpenDialog()` | ✅ | ❌ | ✅ | ❌ |
| `SaveDialog()` | ✅ | ❌ | ✅ | ❌ |
| `MessageDialog()` | ✅ | ❌ | ✅ | ❌ |
| File filters | ❌ | ❌ | ❌ | ➖ |
| Multiple file selection | ✅ | ❌ | ❌ | ➖ |

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
| Custom UserAgent | ⚠️ (JS hack) | ⚠️ (JS hack) | ⚠️ (JS hack) | ➖ |
| Proxy support | ❌ | ❌ | ❌ | ➖ |

---

## Platform Readiness Summary

| Platform | Status | Usable For |
|----------|--------|------------|
| **macOS** | 🟢 Production-ready | Navigation, JS interop, dialogs, cookies, clipboard, notifications |
| **Linux** | 🔴 Broken | Basic navigation works; JS interop completely broken due to wrong API + thread safety |
| **Windows** | 🟡 Partially Ready | WebView2 init, navigation, JS interop, dialogs, clipboard all work. RegisterScheme missing. Needs Windows testing. |
| **Headless** | 🟢 CI-ready | All non-UI operations work; useful for unit testing |
