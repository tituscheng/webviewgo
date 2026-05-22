//go:build windows

package core

/*
#cgo LDFLAGS: -lcomctl32 -lole32 -loleaut32 -luuid -lgdi32

#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <stdlib.h>

// C helpers implemented in webview_windows.c
extern int wvInitWebView2(HWND hwnd, uintptr_t handle, int width, int height);
extern void wvDestroyWebView2();
extern void wvNavigate(const char *url);
extern void wvLoadHTML(const char *html);
extern void wvEval(const char *script);
extern void wvEvalAsync(const char *script);
extern void wvAddUserScript(const char *script);
extern void wvReload();
extern void wvGoBack();
extern void wvGoForward();
extern void wvSetTitle(HWND hwnd, const char *title);
extern void wvSetSize(HWND hwnd, int width, int height);
extern void wvShow(HWND hwnd);
extern void wvHide(HWND hwnd);
extern void wvRunMsgLoop();
extern void wvTerminateMsgLoop();
extern HWND wvCreateWindow(int width, int height, const char *title);

// Dialog & clipboard helpers
extern char **openDialogW(int allowFiles, int allowDirs, int allowMultiple,
                          const wchar_t *title, const wchar_t *directory,
                          int *outCount);
extern wchar_t *saveDialogW(const wchar_t *title, const wchar_t *directory,
                            const wchar_t *defaultFile);
extern int messageDialogW(const wchar_t *title, const wchar_t *message,
                          int level, int buttons);
extern void freeWStringArray(char **arr, int count);
extern char *clipboardReadTextW();
extern void clipboardWriteTextW(const wchar_t *text);

// Go callbacks
extern void goWebViewMessageReceived(uintptr_t handle, char *name, char *body);
extern void goWebViewWindowWillClose(uintptr_t handle);
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

// The WebView2 backend keeps the controller/webview in C global state, so only
// one instance can exist per process. This guard rejects a second New() rather
// than silently clobbering the first.
var windowsInstanceActive atomic.Bool

// windowsWebView is the Windows WebView2 backend.
type windowsWebView struct {
	handle      uintptr
	hwnd        unsafe.Pointer
	logger      *slog.Logger
	bindings    map[string]func([]any) (any, error)
	rawBindings map[string]func(json.RawMessage) (json.RawMessage, error)
	schemes     map[string]types.SchemeHandler
	mu          sync.RWMutex
	done        chan struct{}
	terminated  bool
	pump        *responsePump // batches Bind/BindRaw responses onto the UI thread
}

// init pins the main goroutine to the main OS thread. Windows COM/WebView2
// initialization must happen on the thread that created the message loop.
// Package init functions run on the main goroutine before main(), while it is
// still on the main thread, so this is the only reliable place to lock it.
func init() {
	runtime.LockOSThread()
}

func newNative(opts types.Options) (Platform, error) {
	if !windowsInstanceActive.CompareAndSwap(false, true) {
		return nil, fmt.Errorf("core: only one webview instance is supported per process on windows")
	}

	title := C.CString(opts.Title)
	defer C.free(unsafe.Pointer(title))

	hwnd := C.wvCreateWindow(C.int(opts.Width), C.int(opts.Height), title)
	if hwnd == nil {
		windowsInstanceActive.Store(false)
		return nil, fmt.Errorf("core: failed to create window")
	}

	wv := &windowsWebView{
		logger:      logOpts(opts),
		bindings:    make(map[string]func([]any) (any, error)),
		rawBindings: make(map[string]func(json.RawMessage) (json.RawMessage, error)),
		schemes:     make(map[string]types.SchemeHandler),
		done:        make(chan struct{}),
	}
	wv.handle = nextHandle(wv)
	wv.hwnd = unsafe.Pointer(hwnd)
	wv.pump = newResponsePump(wv.evalAsync)

	if C.wvInitWebView2(hwnd, C.uintptr_t(wv.handle), C.int(opts.Width), C.int(opts.Height)) != 0 {
		releaseHandle(wv.handle)
		windowsInstanceActive.Store(false)
		return nil, fmt.Errorf("core: failed to initialize WebView2. Ensure the WebView2 Runtime is installed.")
	}

	return wv, nil
}

func (w *windowsWebView) Run() error {
	defer close(w.done)
	C.wvRunMsgLoop()
	return nil
}

func (w *windowsWebView) Terminate() {
	w.mu.Lock()
	if w.terminated {
		w.mu.Unlock()
		return
	}
	w.terminated = true
	w.mu.Unlock()
	if w.pump != nil {
		w.pump.shutdown()
	}
	C.wvTerminateMsgLoop()
}

// isTerminated reports whether the webview has been terminated.
func (w *windowsWebView) isTerminated() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.terminated
}

func (w *windowsWebView) Destroy() error {
	w.Terminate()
	select {
	case <-w.done:
	case <-time.After(5 * time.Second):
		w.logger.Warn("destroy timed out waiting for message loop exit")
	}
	C.wvDestroyWebView2()
	releaseHandle(w.handle)
	windowsInstanceActive.Store(false)
	return nil
}

func (w *windowsWebView) SetTitle(title string) {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	C.wvSetTitle(C.HWND(w.hwnd), cs)
}

func (w *windowsWebView) SetSize(width, height int, hint types.Hint) {
	C.wvSetSize(C.HWND(w.hwnd), C.int(width), C.int(height))
}

func (w *windowsWebView) SetMinSize(width, height int) {
	// Track min size via MINMAXINFO in window proc would be needed for full support.
	// For now, set the window size as a best-effort approximation.
	C.wvSetSize(C.HWND(w.hwnd), C.int(width), C.int(height))
}

func (w *windowsWebView) SetMaxSize(width, height int) {
	// Same limitation as SetMinSize.
}

func (w *windowsWebView) SetFullscreen(fullscreen bool) {
	if fullscreen {
		C.ShowWindow(C.HWND(w.hwnd), C.SW_MAXIMIZE)
	} else {
		C.ShowWindow(C.HWND(w.hwnd), C.SW_RESTORE)
	}
}

func (w *windowsWebView) SetAlwaysOnTop(alwaysOnTop bool) {
	var flags C.UINT
	if alwaysOnTop {
		flags = C.HWND_TOPMOST
	} else {
		flags = C.HWND_NOTOPMOST
	}
	C.SetWindowPos(C.HWND(w.hwnd), C.HWND(flags), 0, 0, 0, 0,
		C.SWP_NOMOVE|C.SWP_NOSIZE|C.SWP_SHOWWINDOW)
}

func (w *windowsWebView) Show() { C.wvShow(C.HWND(w.hwnd)) }
func (w *windowsWebView) Hide() { C.wvHide(C.HWND(w.hwnd)) }

func (w *windowsWebView) Navigate(url string) error {
	cs := C.CString(url)
	defer C.free(unsafe.Pointer(cs))
	C.wvNavigate(cs)
	return nil
}

func (w *windowsWebView) LoadHTML(html, baseURL string) error {
	cs := C.CString(html)
	defer C.free(unsafe.Pointer(cs))
	C.wvLoadHTML(cs)
	return nil
}

func (w *windowsWebView) Reload()  { C.wvReload() }
func (w *windowsWebView) Back()    { C.wvGoBack() }
func (w *windowsWebView) Forward() { C.wvGoForward() }

func (w *windowsWebView) Eval(script string) error {
	cs := C.CString(script)
	defer C.free(unsafe.Pointer(cs))
	C.wvEval(cs)
	return nil
}

// evalAsync runs script on the UI thread; safe to call from a goroutine.
func (w *windowsWebView) evalAsync(script string) {
	cs := C.CString(script)
	defer C.free(unsafe.Pointer(cs))
	C.wvEvalAsync(cs)
}

func (w *windowsWebView) Bind(name string, fn func(args []any) (any, error)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.rawBindings[name]; ok {
		return fmt.Errorf("webview: Bind %q: already bound as raw binding", name)
	}
	w.bindings[name] = fn
	w.installBindingLocked(name)
	return nil
}

func (w *windowsWebView) BindRaw(name string, fn func(json.RawMessage) (json.RawMessage, error)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.bindings[name]; ok {
		return fmt.Errorf("webview: BindRaw %q: already bound as normal binding", name)
	}
	if _, ok := w.rawBindings[name]; ok {
		return fmt.Errorf("webview: BindRaw %q: already bound", name)
	}
	w.rawBindings[name] = fn
	w.installBindingLocked(name)
	return nil
}

// installBindingLocked injects the JS shim for a binding as a document-created
// script (so it survives navigation) and evaluates it once for the current
// document. Responses are delivered by evaluating a resolve/reject expression
// directly, so no inbound message listener is needed. The caller must hold w.mu.
func (w *windowsWebView) installBindingLocked(name string) {
	script := fmt.Sprintf(`
window.%s = function(...args) {
	return new Promise((resolve, reject) => {
		const id = '__go_' + Math.random().toString(36).slice(2);
		window[id] = { resolve, reject };
		window.chrome.webview.postMessage({bind: %q, args: args, cb: id});
	});
};
`, name, name)
	cs := C.CString(script)
	C.wvAddUserScript(cs)
	C.free(unsafe.Pointer(cs))
	_ = w.Eval(script)
}

func (w *windowsWebView) RegisterScheme(scheme string, handler types.SchemeHandler) error {
	return fmt.Errorf("core: RegisterScheme not yet implemented on windows")
}

func (w *windowsWebView) OpenDialog(opts types.OpenDialogOptions) ([]string, error) {
	var ctitle, cdir *C.wchar_t
	if opts.Title != "" {
		ctitle = utf8ToWide(opts.Title)
		defer C.free(unsafe.Pointer(ctitle))
	}
	if opts.Directory != "" {
		cdir = utf8ToWide(opts.Directory)
		defer C.free(unsafe.Pointer(cdir))
	}

	var count C.int
	paths := C.openDialogW(
		boolInt(opts.AllowFiles), boolInt(opts.AllowDirs), boolInt(opts.AllowMultiple),
		ctitle, cdir, &count,
	)
	if paths == nil {
		return nil, nil
	}
	defer C.freeWStringArray(paths, count)

	var out []string
	slice := unsafe.Slice((**C.char)(unsafe.Pointer(paths)), int(count))
	for _, p := range slice {
		out = append(out, C.GoString(p))
	}
	return out, nil
}

func (w *windowsWebView) SaveDialog(opts types.SaveDialogOptions) (string, error) {
	var ctitle, cdir, cfile *C.wchar_t
	if opts.Title != "" {
		ctitle = utf8ToWide(opts.Title)
		defer C.free(unsafe.Pointer(ctitle))
	}
	if opts.Directory != "" {
		cdir = utf8ToWide(opts.Directory)
		defer C.free(unsafe.Pointer(cdir))
	}
	if opts.DefaultFile != "" {
		cfile = utf8ToWide(opts.DefaultFile)
		defer C.free(unsafe.Pointer(cfile))
	}

	path := C.saveDialogW(ctitle, cdir, cfile)
	if path == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(path))

	// Convert wide char result to Go string
	len := C.WideCharToMultiByte(C.CP_UTF8, 0, path, -1, nil, 0, nil, nil)
	if len <= 0 {
		return "", nil
	}
	buf := make([]byte, len)
	C.WideCharToMultiByte(C.CP_UTF8, 0, path, -1, (*C.char)(unsafe.Pointer(&buf[0])), C.int(len), nil, nil)
	return string(buf[:len-1]), nil // exclude null terminator
}

func (w *windowsWebView) MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error) {
	var ctitle, cmsg *C.wchar_t
	if opts.Title != "" {
		ctitle = utf8ToWide(opts.Title)
		defer C.free(unsafe.Pointer(ctitle))
	}
	if opts.Message != "" {
		cmsg = utf8ToWide(opts.Message)
		defer C.free(unsafe.Pointer(cmsg))
	}

	res := C.messageDialogW(ctitle, cmsg, C.int(opts.Level), C.int(opts.Buttons))
	return types.DialogResult(res), nil
}

func (w *windowsWebView) ClipboardReadText() (string, error) {
	cs := C.clipboardReadTextW()
	if cs == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(cs))
	return C.GoString(cs), nil
}

func (w *windowsWebView) ClipboardWriteText(text string) error {
	wcs := utf8ToWide(text)
	if wcs == nil {
		return nil
	}
	defer C.free(unsafe.Pointer(wcs))
	C.clipboardWriteTextW(wcs)
	return nil
}

func (w *windowsWebView) Notify(title, body string) error {
	return nil
}

//export goWebViewMessageReceived
func goWebViewMessageReceived(handle C.uintptr_t, name *C.char, body *C.char) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	ww := wv.(*windowsWebView)
	n := C.GoString(name)
	b := C.GoString(body)

	if n == "goBridge" {
		parseBridgeMessage(newPlatformBridgeHost(
			ww.lookupBinding,
			ww.isTerminated,
			ww.pump.enqueue,
			ww.logger,
		), b)
	}
}

func (w *windowsWebView) lookupBinding(name string) bridgeBindings {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return bridgeBindings{w.rawBindings[name], w.bindings[name]}
}

//export goWebViewWindowWillClose
func goWebViewWindowWillClose(handle C.uintptr_t) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	wv.Terminate()
}

func utf8ToWide(s string) *C.wchar_t {
	if s == "" {
		return nil
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	len := C.MultiByteToWideChar(C.CP_UTF8, 0, cs, -1, nil, 0)
	if len <= 0 {
		return nil
	}
	wcs := (*C.wchar_t)(C.malloc(C.size_t(len) * C.size_t(unsafe.Sizeof(C.wchar_t(0)))))
	if wcs == nil {
		return nil
	}
	C.MultiByteToWideChar(C.CP_UTF8, 0, cs, -1, wcs, len)
	return wcs
}
