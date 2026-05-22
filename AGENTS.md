# Agent Guidance

## Project: webviewgo

A production-grade, lightweight Go webview library with SQLite-backed cookies,
rich JS interop, custom protocols, and native UI helpers.

### Build
```bash
go build ./...
make test
```

### Key Directories
- `pkg/webview/` — Public API
- `pkg/webview/compat/` — Migration layer from `webview/webview_go`
- `internal/core/` — CGO platform backends (darwin/windows/linux/headless)
- `internal/cookie/` — SQLite cookie/session manager
- `internal/profile/` — Data directory resolution
- `internal/js/` — JS-to-Go bridge reflection
- `internal/types/` — Shared type definitions (breaks import cycles)
- `examples/` — Runnable examples
- `docs/` — Roadmap, gap tracker, and platform support matrix

### Design Rules
- **Never** hardcode `~/.webviewgo`. Data directories are resolved via `Options{AppName, DataDir}` or `WEBVIEW_DATA_DIR` env var.
- `internal/types` contains all cross-package types. `pkg/webview` re-exports via type aliases.
- Platform CGO files use `//go:build` tags. `void*` is used for ObjC handles to avoid cgo struct/class mismatches.
- Cookie native sync is opt-in via the `SyncCookiesToNative` interface on platform backends.
- `runtime.LockOSThread()` must be in `init()` (not `newNative`) on darwin/linux/windows. AppKit/GTK/COM require the run loop on the process's main thread (thread 0); `init()` runs before `main()` while the goroutine is still on that thread.
- Any Cocoa API that uses completion handlers + `dispatch_semaphore_wait` can deadlock if called on the main thread before `[NSApp run]`. Use run-loop pumping (e.g. `waitForCookieStore` pattern) when on the main thread.
- Keep dependencies minimal: stdlib + `modernc.org/sqlite` only.

### Testing
- Unit tests for all `internal/` packages targeting >85% coverage.
- Table-driven tests with named subtests.
- Use `:memory:` SQLite databases in tests.
- Integration tests use `//go:build integration`.

### Documentation
- `docs/ROADMAP.md` — Phased development plan with current status
- `docs/GAPS.md` — Living inventory of all known gaps, stubs, and TODOs
- `docs/PLATFORM_MATRIX.md` — Feature-by-platform support matrix
- **Always update these docs** when fixing a gap or implementing a feature.

### CGO Best Practices
- Minimize crossings: batch operations.
- Never store Go pointers in C memory longer than the call.
- Use `uintptr` handles + `sync.Map` to reference Go objects from C callbacks.
- C callback signatures use `char*` (not `const char*`) to match cgo exports.
