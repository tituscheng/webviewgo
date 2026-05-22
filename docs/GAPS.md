# Gap Tracker

Comprehensive inventory of all known gaps, stubs, TODOs, and technical debt in the webviewgo library.

## Legend

- 🔴 **Critical** — Blocks real usage on that platform
- 🟠 **High** — Major feature missing or broken
- 🟡 **Medium** — Test coverage or polish gap
- 🟢 **Low** — Minor code quality issue
- ✅ **Fixed** — Resolved and verified

---

## Platform Backends

### macOS (darwin)

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| D01 | 🟠 | Custom protocol response not delivered to webview | `internal/core/protocol_darwin.go`, `protocol_darwin_delegate.m` | ✅ Fixed | Response now delivered via `deliverSchemeResponse` using `g_idle_add` pattern; task stored in `schemeTaskMap` keyed by `reqHandle` |
| D02 | 🟢 | `NSUserNotification` deprecated since macOS 11.0 | `internal/core/webview_darwin.go` | ✅ Fixed | Deprecation warnings suppressed with `#pragma clang diagnostic ignored` |
| D03 | 🟢 | `SameSite` not synced to `WKHTTPCookieStore` | `internal/core/cookie_sync_darwin.go` | ✅ Fixed | `NSHTTPCookieSameSitePolicy` set from `SameSite` enum |
| D04 | 🟢 | `SchemeHandlerDelegate` memory leak (non-ARC `retain`) | `internal/core/protocol_darwin_delegate.m` | ✅ Fixed | Added `dealloc` with `[_scheme release]` |
| D05 | 🟠 | `UserAgent` set via JS hack instead of native API | `pkg/webview/webview.go` | ✅ Fixed | Native `customUserAgent` on macOS, `webkit_settings_set_user_agent` on Linux. Windows still uses JS fallback. |
| D06 | 🟠 | Frameless/Transparent options incomplete | `internal/core/webview_darwin.go` | ✅ Fixed | `setOpaque:NO` + `setBackgroundColor:[NSColor clearColor]` on window. WebView transparency requires page CSS. |
| D07 | 🟢 | `Center` option ignored; window always centers | `internal/core/webview_darwin.go` | ✅ Fixed | `createWindow` centers only when `opts.Center` is true |
| D09 | 🔴 | Main goroutine not pinned to main thread before `main()` | `internal/core/webview_darwin.go` | ✅ Fixed | `runtime.LockOSThread()` moved from `newNative` to `init()`. AppKit requires thread 0 for `NSApp run`. |
| D10 | 🔴 | Cookie sync deadlocks main thread before `Run()` | `internal/core/cookie_sync_darwin.go` | ✅ Fixed | `waitForCookieStore()` pumps `NSRunLoop` when on main thread; `dispatch_semaphore_wait` alone blocks the completion handler from running. |
| D08 | 🟢 | Dialog file filters ignored | `internal/core/dialog_darwin.go` | ✅ Fixed | `setAllowedFileTypes` wired for open/save dialogs on macOS |
| D11 | 🔴 | Binding callback goroutines panic into void + use-after-free on Destroy | `internal/core/webview_darwin.go`, `webview_linux.go`, `webview_windows.go` | ✅ Fixed | Added `recover()` in goroutine; check `terminated` before `Eval` / `evalOnMainThread` |
| D12 | 🟠 | Scheme response `Body` readers never closed | `internal/core/protocol_darwin.go` | ✅ Fixed | Close `resp.Body` via `io.Closer` after `io.ReadAll` |
| D13 | 🔴 | Scheme task goroutines not coordinated with webview Destroy | `internal/core/protocol_darwin.go`, `webview_darwin.go` | ✅ Fixed | Check `terminated` before handler + delivery; `sync.WaitGroup` in `Destroy` with 5s timeout |
| D14 | 🔴 | Custom schemes never wired to WKWebViewConfiguration | `internal/core/webview_darwin.go`, `protocol_darwin.go` | ✅ Fixed | Schemes from `Options.Schemes` registered in `createWebView`; late `RegisterScheme` returns actionable error |
| D15 | 🔴 | Bridge callback id (`cb`) not validated — JS injection | `internal/core/webview.go`, platform handlers | ✅ Fixed | Shared `dispatchBridgeMessage` validates `__go_<alphanum>` ids; unknown bindings reject the promise |

### Linux (WebKitGTK)

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| L01 | 🔴 | JS bridge uses WKWebView API instead of WebKitGTK | `internal/core/webview_linux.go` | ✅ Fixed | Injected JS now uses `window.goBridge.postMessage` |
| L02 | 🔴 | `Eval()` called from goroutine (GTK thread safety) | `internal/core/webview_linux.go` | ✅ Fixed | Responses now queued via `g_idle_add` to run on GTK main thread |
| L03 | 🟠 | Custom protocols not implemented | `internal/core/protocol_linux.go` | ✅ Fixed | `webkit_web_context_register_uri_scheme` with async Go handler + response delivery |
| L04 | 🟠 | Dialogs not implemented | `internal/core/dialog_linux.go` | ✅ Fixed | GTK file chooser + message dialog |
| L05 | 🟠 | `SetMinSize`, `SetMaxSize`, `SetFullscreen`, `SetAlwaysOnTop` no-ops | `internal/core/webview_linux.go` | ✅ Fixed | GTK geometry hints, fullscreen, and keep-above |
| L06 | 🟢 | Notifications no-op | `internal/core/webview_linux.go` | 🔴 Open | `Notify` returns `nil` with no action |

### Windows (WebView2)

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| W01 | 🔴 | WebView2 COM object never created | `internal/core/webview_windows.c` | ✅ Fixed | Full async COM init with `ICoreWebView2Environment` + `ICoreWebView2Controller` + `ICoreWebView2` |
| W02 | 🔴 | `Navigate`, `LoadHTML`, `Eval` are no-ops | `internal/core/webview_windows.c` | ✅ Fixed | Implemented via COM vtable offsets with UTF-8→UTF-16 conversion |
| W03 | 🔴 | JS messages discarded without dispatch | `internal/core/webview_windows.go` | ✅ Fixed | `goWebViewMessageReceived` parses JSON, dispatches to bindings, injects promise resolution via `Eval` |
| W04 | 🟠 | Dialogs not implemented | `internal/core/webview_windows.go` | ✅ Fixed | `GetOpenFileNameW`, `GetSaveFileNameW`, `MessageBoxW` |
| W05 | 🟠 | Clipboard no-op | `internal/core/webview_windows.go` | ✅ Fixed | `OpenClipboard` / `GetClipboardData` / `SetClipboardData` with UTF-16 |
| W06 | 🟠 | `Reload`, `Back`, `Forward` no-ops | `internal/core/webview_windows.go` | ✅ Fixed | COM vtable calls to `Reload`, `GoBack`, `GoForward` |
| W07 | 🟠 | `SetMinSize`, `SetMaxSize`, `SetFullscreen`, `SetAlwaysOnTop` no-ops | `internal/core/webview_windows.go` | ✅ Fixed | `SetWindowPos`, `ShowWindow`, `HWND_TOPMOST` |
| W08 | 🟢 | Bindings data race | `internal/core/webview_windows.go` | ✅ Fixed | `goWebViewMessageReceived` uses `RLock` before reading `bindings` |

### Headless (All Platforms)

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| H01 | 🟢 | `Run()` is a 100% CPU spin loop | `internal/core/webview_headless_impl.go` | ✅ Fixed | `time.Sleep(100ms)` inside the loop |
| H02 | 🟡 | Dialogs return errors (expected, but could be configurable) | `internal/core/webview_headless_impl.go` | 🔴 Open | By design for CI, but limits headless testing of dialog code paths |

---

## Public API Gaps

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| A01 | 🟠 | `FSHandler` returns `StatusNotImplemented` | `pkg/webview/protocol.go` | ✅ Fixed | Serves any `fs.FS` with automatic `index.html`, mime-type detection |
| A02 | 🟠 | `HTTPHandler` returns `StatusNotImplemented` | `pkg/webview/protocol.go` | ✅ Fixed | Adapts `http.Handler` via `httptest.ResponseRecorder` |
| A03 | 🟠 | `GenerateTS` not implemented | `pkg/webview/jsbind.go` | ✅ Fixed | Reflection-based TS generation with struct field support |
| A04 | 🟠 | Menu / Tray types exist but unintegrated | `pkg/webview/menu.go` | 🔴 Open | No `WebView` methods, no backend wiring |
| A05 | 🟠 | Drag & Drop option exists but unwired | `internal/types/types.go` | 🔴 Open | `OnDrop` in `Options` never read |
| A06 | 🟠 | Proxy option exists but unwired | `internal/types/types.go` | 🔴 Open | `Proxy` in `Options` never passed to backends |
| A07 | 🟢 | Compat layer `New()` panics on error | `pkg/webview/compat/compat.go` | 🔴 Open | Matches old API but poor practice |
| A08 | 🔴 | `Eval()` signature promises `(any, error)` but always returns `(nil, nil)` on all platforms | `pkg/webview/webview.go`, `internal/core/webview*.go` | ✅ Fixed | Changed signature to `Eval(script string) error` to match actual behavior and original webview_go contract |

---

## Test Coverage Gaps

| # | Severity | Description | File(s) | Status | Notes |
|---|----------|-------------|---------|--------|-------|
| T01 | 🟡 | `internal/core` at 13.8% | `internal/core/*` | 🟡 Partial | CGO hard to test; headless tests exist but don't cover native paths |
| T02 | 🟡 | `internal/types` at 0.0% | `internal/types/*` | ✅ Fixed | Added tests for constants, structs, DefaultOptions |
| T03 | 🟡 | `pkg/webview/compat` at 0.0% | `pkg/webview/compat/*` | 🟡 Partial | Only compile-time alias check |
| T04 | 🟡 | JS bridge missing edge-case tests | `internal/js/bridge_test.go` | ✅ Fixed | Added tests for slice args, map args, pointer returns, error wrapping, multi-return validation |
| T05 | 🟡 | No cross-platform integration tests | `internal/core/bridge_integration_test.go` | ✅ Fixed | `//go:build integration` tests for headless bind round-trip and `Options.Schemes` |
| T06 | 🟡 | Dialog success paths untested | `pkg/webview/webview_test.go` | 🔴 Open | Only error paths tested (headless returns errors) |

---

## Recently Fixed

| # | Description | File(s) | Fixed In | PR/Commit |
|---|-------------|---------|----------|-----------|
| D09 | Main goroutine pinning (`LockOSThread` in `init()`) | `internal/core/webview_darwin.go`, `webview_linux.go`, `webview_windows.go` | v0.1 | Prevents AppKit/GTK COM init from happening on a goroutine-scheduled thread instead of thread 0. |
| D10 | Cookie sync main-thread deadlock (`waitForCookieStore`) | `internal/core/cookie_sync_darwin.go` | v0.1 | `WKHTTPCookieStore` completion handlers run on main thread; `dispatch_semaphore_wait` without run-loop pumping deadlocks if called before `[NSApp run]`. |
| A08 | `Eval()` API contract fix: `(any, error)` → `error` | `pkg/webview/webview.go`, `internal/core/*` | v0.2 | All backends discarded eval results; signature falsely promised return values. Reverted to `error`-only to match original library contract. |
| D11 | Binding callback panic recovery + post-Destroy guard | `internal/core/webview_darwin.go`, `webview_linux.go`, `webview_windows.go` | v0.2 | Goroutines could panic silently or call `Eval` on freed native objects after `Destroy`. Added `recover()` and `terminated` check. |
| D12 | Scheme response Body close leak | `internal/core/protocol_darwin.go` | v0.2 | `FSHandler` and `HTTPHandler` returned live readers that were never closed after `io.ReadAll`. |
| D13 | Scheme task lifetime coordination with Destroy | `internal/core/protocol_darwin.go`, `webview_darwin.go` | v0.2 | Pending scheme goroutines could outlive the WKWebView. Added `terminated` guards and `sync.WaitGroup` drain in `Destroy`. |
| D14 | macOS scheme registration wired at webview creation | `internal/core/webview_darwin.go`, `internal/types/types.go` | v0.3 | `Options.Schemes` pre-registers handlers on `WKWebViewConfiguration`; late `RegisterScheme` returns clear error. |
| D15 | Bridge callback id validation + unknown binding reject | `internal/core/bridge.go` | v0.3 | Prevents JS injection via crafted `cb` field; rejects unknown bindings instead of hanging promises. |
| L03–L05 | Linux custom schemes, dialogs, window chrome | `internal/core/protocol_linux.go`, `dialog_linux.go`, `webview_linux.go` | v0.3 | WebKitGTK URI schemes, GTK dialogs, geometry/fullscreen/keep-above. |
