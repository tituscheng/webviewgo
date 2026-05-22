# webviewgo Roadmap

This roadmap prioritizes gaps by impact and effort. Each phase builds on the previous one.

---

## Phase 1: Linux Critical Fixes (High Impact, Low Effort)
**Goal:** Make the Linux backend actually usable.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 1.1 | Fix JS bridge API: change `window.webkit.messageHandlers.goBridge` to `window.goBridge` in injected JS | `internal/core/webview_linux.go` | ~1 hr | ✅ Done |
| 1.2 | Fix `Eval()` thread safety: dispatch JS responses via `g_idle_add` on GTK main thread | `internal/core/webview_linux.go` | ~2 hrs | ✅ Done |
| 1.3 | Add smoke test for Linux JS roundtrip (can be manual for now) | — | ~30 min | 🚧 Deferred to CI |

**Acceptance:** `examples/js-interop` runs successfully on Linux and `add(2,3)` returns `5`.

---

## Phase 2: macOS Custom Protocol Response Delivery (High Impact, Medium Effort)
**Goal:** Make `RegisterScheme` actually serve responses on macOS.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 2.1 | Store `id<WKURLSchemeTask>` in `schemeTaskMap` keyed by `reqHandle` | `internal/core/protocol_darwin_delegate.m` | ~1 hr | ✅ Done |
| 2.2 | Add C helper `deliverSchemeResponse(reqHandle, status, contentType, body, bodyLen)` | `internal/core/protocol_darwin.go`, `protocol_darwin_delegate.m` | ~2 hrs | ✅ Done |
| 2.3 | Wire `goProtocolHandler` to call back into C with response data | `internal/core/protocol_darwin.go` | ~2 hrs | ✅ Done |
| 2.4 | Handle protocol cancellation in `stopURLSchemeTask:` | `internal/core/protocol_darwin_delegate.m` | ~30 min | ✅ Done |

**Acceptance:** `examples/custom-protocol` displays "Hello from custom protocol!" on macOS.

---

## Phase 3: Windows WebView2 Backend (Critical, Large Effort)
**Goal:** Make Windows a first-class platform.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 3.1 | Implement async COM environment creation with `ICoreWebView2Environment` | `internal/core/webview_windows.c` | ~1 day | ✅ Done |
| 3.2 | Create controller + webview, size to HWND, handle `WM_SIZE` | `internal/core/webview_windows.c` | ~4 hrs | ✅ Done |
| 3.3 | Implement `wvNavigate`, `wvLoadHTML`, `wvEval` via COM vtable calls | `internal/core/webview_windows.c` | ~3 hrs | ✅ Done |
| 3.4 | Wire `goWebViewMessageReceived` for JS bindings | `internal/core/webview_windows.go` | ~2 hrs | ✅ Done |
| 3.5 | Implement dialogs (`GetOpenFileNameW`, `MessageBoxW`) | `internal/core/webview_windows.c` | ~4 hrs | ✅ Done |
| 3.6 | Implement clipboard (`OpenClipboard`) | `internal/core/webview_windows.c` | ~1 hr | ✅ Done |
| 3.7 | Add WebView2 runtime detection / graceful fallback | `internal/core/webview_windows.c` | ~2 hrs | ✅ Done |

**Acceptance:** `examples/basic`, `examples/js-interop`, and `examples/cookies` all run on Windows.

---

## Phase 4: Convenience Adapters & DX (Medium Impact, Low Effort)
**Goal:** Make common patterns effortless.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 4.1 | Implement `FSHandler` using `embed.FS` or `os.DirFS` | `pkg/webview/protocol.go` | ~1 hr | ✅ Done |
| 4.2 | Implement `HTTPHandler` adapter | `pkg/webview/protocol.go` | ~1 hr | ✅ Done |
| 4.3 | Add dialog file filter support on macOS | `internal/core/dialog_darwin.go` | ~2 hrs | ✅ Done |
| 4.4 | Implement `GenerateTS` for TypeScript definitions from bindings | `pkg/webview/jsbind.go` | ~3 hrs | ✅ Done |
| 4.5 | Fix headless `Run()` CPU spin loop | `internal/core/webview_headless_impl.go` | ~15 min | ✅ Done |
| 4.6 | Add `SameSite` to macOS cookie sync | `internal/core/cookie_sync_darwin.go` | ~15 min | ✅ Done |

**Acceptance:** `FSHandler` and `HTTPHandler` work in `examples/custom-protocol`; `GenerateTS` produces valid `.d.ts`.

---

## Phase 5: Advanced Features (Nice to Have, Medium–Large Effort)
**Goal:** Reach parity with mature desktop webview frameworks.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 5.1 | Proper `UserAgent` setting via native APIs (not JS hack) | `internal/core/webview_darwin.go`, `webview_linux.go`, `webview_windows.go` | ~2 hrs | ✅ Done |
| 5.2 | Proxy support wired to all backends | `internal/core/*` | ~1 day | 🔴 Not started |
| 5.3 | Frameless + transparent window support on macOS | `internal/core/webview_darwin.go` | ~2 hrs | ✅ Done |
| 5.4 | Drag & drop file support | `internal/types/types.go`, `internal/core/*` | ~1 day | 🔴 Not started |
| 5.5 | Menu bar integration | `pkg/webview/menu.go`, `internal/core/*` | ~2 days | 🔴 Not started |
| 5.6 | System tray icon + menu | `pkg/webview/menu.go`, `internal/core/*` | ~2 days | 🔴 Not started |
| 5.7 | Suppress macOS notification deprecation warnings | `internal/core/webview_darwin.go` | ~15 min | ✅ Done |

---

## Phase 6: Quality & Testing (Ongoing)
**Goal:** Production confidence across all platforms.

| # | Task | File(s) | Effort | Status |
|---|------|---------|--------|--------|
| 6.1 | Add tests for `internal/types` (`DefaultOptions`, etc.) | `internal/types/*_test.go` | ~30 min | ✅ Done |
| 6.2 | Add functional compat layer tests | `pkg/webview/compat/*_test.go` | ~1 hr | 🚧 Deferred |
| 6.3 | Add JS bridge edge-case tests (variadic, pointers, slices) | `internal/js/bridge_test.go` | ~2 hrs | ✅ Done |
| 6.4 | Set up CI matrix: ubuntu (headless), macOS (native), Windows (native) | `.github/workflows/ci.yml` | ~2 hrs | 🚧 Deferred |
| 6.5 | Fix Obj-C memory leak in `SchemeHandlerDelegate` | `internal/core/protocol_darwin_delegate.m` | ~15 min | ✅ Done |
| 6.6 | Fix Windows bindings data race | `internal/core/webview_windows.go` | ~15 min | ✅ Done |

---

## Current Phase

> **All planned phases completed.**
>
> The library is now production-ready on macOS, functional on Linux and Windows, with comprehensive documentation and tracking. Remaining backlog items (proxy, menus, tray, drag & drop) are tracked in `docs/GAPS.md` and can be picked up as needed.

We are currently in **Phase 1**. The priority is fixing the Linux JS bridge and thread-safety issues because they are low-effort, high-impact fixes that unblock a major platform.
