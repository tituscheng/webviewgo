//go:build darwin

package core

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import <stdlib.h>
#import <string.h>
#import "webview_darwin_delegate.h"
#import "darwin_compat.h"

static NSApplication *sharedApp(void) {
    return [NSApplication sharedApplication];
}

static void setActivationPolicyAccessory(void) {
    [[NSApplication sharedApplication] setActivationPolicy:NSApplicationActivationPolicyAccessory];
}

static void setActivationPolicyRegular(void) {
    [[NSApplication sharedApplication] setActivationPolicy:NSApplicationActivationPolicyRegular];
}

static void activateApp(void) {
    [[NSApplication sharedApplication] activateIgnoringOtherApps:YES];
}

static void deactivateApp(void) {
    [[NSApplication sharedApplication] deactivate];
}

static void runApp(void) {
    [NSApp run];
}

static void stopApp(void) {
    [NSApp stop:nil];
    NSEvent *event = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                        location:NSMakePoint(0, 0)
                                   modifierFlags:0
                                       timestamp:0
                                    windowNumber:0
                                         context:nil
                                         subtype:0
                                           data1:0
                                           data2:0];
    [NSApp postEvent:event atStart:YES];
}

// Window / WebView helpers use void* to avoid cgo struct/objc class mismatch.
static void *createWindow(int width, int height, int styleMask, const char *title) {
    NSRect frame = NSMakeRect(0, 0, width, height);
    NSWindow *window = [[NSWindow alloc] initWithContentRect:frame
                                                   styleMask:styleMask
                                                     backing:NSBackingStoreBuffered
                                                       defer:NO];
    [window setTitle:[NSString stringWithUTF8String:title]];
    [window center];
    return window;
}

static void *createWebView(void *windowPtr, uintptr_t handle, int enableDevTools) {
    NSWindow *window = (NSWindow *)windowPtr;
    WKWebViewConfiguration *config = [[WKWebViewConfiguration alloc] init];
    if (enableDevTools) {
        [config.preferences setValue:@YES forKey:@"developerExtrasEnabled"];
    }

    WKUserContentController *controller = config.userContentController;
    WebViewDelegate *delegate = [[WebViewDelegate alloc] init];
    delegate.handle = handle;
    [controller addScriptMessageHandler:delegate name:@"goBridge"];

    NSRect bounds = [[window contentView] bounds];
    WKWebView *webView = [[WKWebView alloc] initWithFrame:bounds configuration:config];
    webView.navigationDelegate = delegate;
    webView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    [[window contentView] addSubview:webView];

    [window setDelegate:delegate];
    return webView;
}

static void webViewNavigate(void *webViewPtr, const char *url) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    NSURL *nsurl = [NSURL URLWithString:[NSString stringWithUTF8String:url]];
    NSURLRequest *req = [NSURLRequest requestWithURL:nsurl];
    [webView loadRequest:req];
}

static void webViewLoadHTML(void *webViewPtr, const char *html, const char *baseURL) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    NSString *nshtml = [NSString stringWithUTF8String:html];
    NSURL *nsbase = [NSURL URLWithString:[NSString stringWithUTF8String:baseURL]];
    [webView loadHTMLString:nshtml baseURL:nsbase];
}

static void webViewEval(void *webViewPtr, const char *script, uintptr_t handle) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    NSString *nsscript = [NSString stringWithUTF8String:script];
    [webView evaluateJavaScript:nsscript completionHandler:^(id result, NSError *error) {
        // Async; handled via promise wrapper on Go side.
    }];
}

static void windowSetTitle(void *windowPtr, const char *title) {
    NSWindow *window = (NSWindow *)windowPtr;
    [window setTitle:[NSString stringWithUTF8String:title]];
}

static void windowSetSize(void *windowPtr, int width, int height) {
    NSWindow *window = (NSWindow *)windowPtr;
    NSRect frame = [window frame];
    frame.size.width = width;
    frame.size.height = height;
    [window setFrame:frame display:YES];
}

static void windowShow(void *windowPtr) {
    NSWindow *window = (NSWindow *)windowPtr;
    [window makeKeyAndOrderFront:nil];
}

static void windowHide(void *windowPtr) {
    NSWindow *window = (NSWindow *)windowPtr;
    [window orderOut:nil];
}

static void windowSetFullscreen(void *windowPtr, int fullscreen) {
    NSWindow *window = (NSWindow *)windowPtr;
    if (fullscreen) {
        if (([window styleMask] & NSWindowStyleMaskFullScreen) == 0) {
            [window toggleFullScreen:nil];
        }
    } else {
        if (([window styleMask] & NSWindowStyleMaskFullScreen) != 0) {
            [window toggleFullScreen:nil];
        }
    }
}

static void windowSetAlwaysOnTop(void *windowPtr, int onTop) {
    NSWindow *window = (NSWindow *)windowPtr;
    if (onTop) {
        [window setLevel:NSFloatingWindowLevel];
    } else {
        [window setLevel:NSNormalWindowLevel];
    }
}

static void windowSetMinSize(void *windowPtr, int width, int height) {
    NSWindow *window = (NSWindow *)windowPtr;
    [window setMinSize:NSMakeSize(width, height)];
}

static void windowSetMaxSize(void *windowPtr, int width, int height) {
    NSWindow *window = (NSWindow *)windowPtr;
    [window setMaxSize:NSMakeSize(width, height)];
}

static void webViewReload(void *webViewPtr) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    [webView reload];
}

static void webViewGoBack(void *webViewPtr) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    if ([webView canGoBack]) [webView goBack];
}

static void webViewGoForward(void *webViewPtr) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    if ([webView canGoForward]) [webView goForward];
}

static void setUserAgent(void *webViewPtr, const char *ua) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    webView.customUserAgent = [NSString stringWithUTF8String:ua];
}

static void makeTransparent(void *windowPtr, void *webViewPtr) {
    (void)webViewPtr;
    NSWindow *window = (NSWindow *)windowPtr;
    [window setOpaque:NO];
    [window setBackgroundColor:[NSColor clearColor]];
}

static char *clipboardReadText(void) {
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    NSString *text = [pb stringForType:NSPasteboardTypeString];
    if (!text) return NULL;
    return strdup([text UTF8String]);
}

static void clipboardWriteText(const char *text) {
    NSPasteboard *pb = [NSPasteboard generalPasteboard];
    [pb clearContents];
    [pb setString:[NSString stringWithUTF8String:text] forType:NSPasteboardTypeString];
}

static void showNotification(const char *title, const char *body) {
    SUPPRESS_DEPRECATED_DECLARATIONS
    NSUserNotification *note = [[NSUserNotification alloc] init];
    note.title = [NSString stringWithUTF8String:title];
    note.informativeText = [NSString stringWithUTF8String:body];
    [[NSUserNotificationCenter defaultUserNotificationCenter] deliverNotification:note];
    RESTORE_DEPRECATED_DECLARATIONS
}

*/
import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

// darwinWebView is the macOS WKWebView backend.
type darwinWebView struct {
	handle     uintptr
	window     unsafe.Pointer
	webView    unsafe.Pointer
	logger     *slog.Logger
	bindings   map[string]func([]any) (any, error)
	schemes    map[string]types.SchemeHandler
	mu         sync.RWMutex
	done       chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	terminated bool
	pending    sync.WaitGroup
}

// init pins the main goroutine to the main OS thread. On macOS, AppKit
// requires that NSApplication's run loop and all NSWindow/WKWebView calls
// happen on the process's main thread (thread 0). Package init functions run
// on the main goroutine before main(), while it is still on the main thread,
// so this is the only reliable place to lock it. Callers must invoke New and
// Run from the main goroutine.
func init() {
	runtime.LockOSThread()
}

func newNative(opts types.Options) (Platform, error) {
	C.setActivationPolicyRegular()
	C.activateApp()

	title := C.CString(opts.Title)
	defer C.free(unsafe.Pointer(title))

	styleMask := C.int(C.NSWindowStyleMaskTitled | C.NSWindowStyleMaskClosable |
		C.NSWindowStyleMaskMiniaturizable | C.NSWindowStyleMaskResizable)
	if opts.Frameless {
		styleMask = C.NSWindowStyleMaskBorderless | C.NSWindowStyleMaskResizable
	}

	window := C.createWindow(C.int(opts.Width), C.int(opts.Height), styleMask, title)
	if window == nil {
		return nil, fmt.Errorf("core: failed to create window")
	}

	if opts.Transparent {
		C.makeTransparent(window, nil)
	}

	wv := &darwinWebView{
		logger:   logOpts(opts),
		bindings: make(map[string]func([]any) (any, error)),
		schemes:  make(map[string]types.SchemeHandler),
		done:     make(chan struct{}),
	}
	wv.ctx, wv.cancel = context.WithCancel(context.Background())
	wv.handle = nextHandle(wv)
	wv.window = unsafe.Pointer(window)

	webView := C.createWebView(window, C.uintptr_t(wv.handle), boolInt(opts.Devtools))
	if webView == nil {
		return nil, fmt.Errorf("core: failed to create webview")
	}
	wv.webView = unsafe.Pointer(webView)

	if opts.Transparent {
		C.makeTransparent(window, webView)
	}

	if opts.UserAgent != "" {
		ua := C.CString(opts.UserAgent)
		C.setUserAgent(wv.webView, ua)
		C.free(unsafe.Pointer(ua))
	}

	C.windowShow(window)

	return wv, nil
}

func (w *darwinWebView) Run() error {
	defer close(w.done)
	C.runApp()
	return nil
}

func (w *darwinWebView) Terminate() {
	w.mu.Lock()
	if w.terminated {
		w.mu.Unlock()
		return
	}
	w.terminated = true
	w.mu.Unlock()
	w.cancel()
	C.stopApp()
}

func (w *darwinWebView) Destroy() error {
	w.Terminate()
	select {
	case <-w.done:
	case <-w.ctx.Done():
	}
	// Drain any in-flight scheme tasks with a timeout to avoid hanging.
	done := make(chan struct{})
	go func() {
		w.pending.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		w.logger.Warn("destroy timed out waiting for pending scheme tasks")
	}
	releaseHandle(w.handle)
	C.deactivateApp()
	return nil
}

func (w *darwinWebView) SetUserAgent(ua string) {
	cs := C.CString(ua)
	defer C.free(unsafe.Pointer(cs))
	C.setUserAgent(w.webView, cs)
}

func (w *darwinWebView) SetTitle(title string) {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	C.windowSetTitle(w.window, cs)
}

func (w *darwinWebView) SetSize(width, height int, hint types.Hint) {
	switch hint {
	case types.HintMin:
		C.windowSetMinSize(w.window, C.int(width), C.int(height))
	case types.HintMax:
		C.windowSetMaxSize(w.window, C.int(width), C.int(height))
	default:
		C.windowSetSize(w.window, C.int(width), C.int(height))
	}
}

func (w *darwinWebView) SetMinSize(width, height int) {
	C.windowSetMinSize(w.window, C.int(width), C.int(height))
}

func (w *darwinWebView) SetMaxSize(width, height int) {
	C.windowSetMaxSize(w.window, C.int(width), C.int(height))
}

func (w *darwinWebView) SetFullscreen(fullscreen bool) {
	C.windowSetFullscreen(w.window, boolInt(fullscreen))
}

func (w *darwinWebView) SetAlwaysOnTop(alwaysOnTop bool) {
	C.windowSetAlwaysOnTop(w.window, boolInt(alwaysOnTop))
}

func (w *darwinWebView) Show() {
	C.windowShow(w.window)
}

func (w *darwinWebView) Hide() {
	C.windowHide(w.window)
}

func (w *darwinWebView) Navigate(url string) error {
	cs := C.CString(url)
	defer C.free(unsafe.Pointer(cs))
	C.webViewNavigate(w.webView, cs)
	return nil
}

func (w *darwinWebView) LoadHTML(html, baseURL string) error {
	chs := C.CString(html)
	defer C.free(unsafe.Pointer(chs))
	cbase := C.CString(baseURL)
	defer C.free(unsafe.Pointer(cbase))
	C.webViewLoadHTML(w.webView, chs, cbase)
	return nil
}

func (w *darwinWebView) Reload() {
	C.webViewReload(w.webView)
}

func (w *darwinWebView) Back() {
	C.webViewGoBack(w.webView)
}

func (w *darwinWebView) Forward() {
	C.webViewGoForward(w.webView)
}

func (w *darwinWebView) Eval(script string) error {
	cs := C.CString(script)
	defer C.free(unsafe.Pointer(cs))
	C.webViewEval(w.webView, cs, C.uintptr_t(w.handle))
	return nil
}

func (w *darwinWebView) Bind(name string, fn func(args []any) (any, error)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bindings[name] = fn

	script := fmt.Sprintf(`
	window.%s = function(...args) {
		return new Promise((resolve, reject) => {
			const id = '__go_' + Math.random().toString(36).slice(2);
			window[id] = { resolve, reject };
			window.webkit.messageHandlers.goBridge.postMessage({
				bind: %q,
				args: args,
				cb: id
			});
		});
	};
`, name, name)
	_ = w.Eval(script)
	return nil
}

func (w *darwinWebView) ClipboardReadText() (string, error) {
	cs := C.clipboardReadText()
	if cs == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(cs))
	return C.GoString(cs), nil
}

func (w *darwinWebView) ClipboardWriteText(text string) error {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.clipboardWriteText(cs)
	return nil
}

func (w *darwinWebView) Notify(title, body string) error {
	ct := C.CString(title)
	defer C.free(unsafe.Pointer(ct))
	cb := C.CString(body)
	defer C.free(unsafe.Pointer(cb))
	C.showNotification(ct, cb)
	return nil
}

//export goWebViewMessageReceived
func goWebViewMessageReceived(handle C.uintptr_t, name *C.char, body *C.char) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	dw := wv.(*darwinWebView)
	n := C.GoString(name)
	b := C.GoString(body)

	if n == "goBridge" {
		var msg struct {
			Bind string `json:"bind"`
			Args []any  `json:"args"`
			CB   string `json:"cb"`
		}
		if err := json.Unmarshal([]byte(b), &msg); err != nil {
			dw.logger.Error("failed to unmarshal bridge message", "error", err)
			return
		}
		dw.mu.RLock()
		fn, ok := dw.bindings[msg.Bind]
		dw.mu.RUnlock()
		if !ok {
			dw.logger.Warn("unknown binding", "name", msg.Bind)
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					dw.logger.Error("binding callback panic", "name", msg.Bind, "recover", r)
				}
			}()
			res, err := fn(msg.Args)
			var script string
			if err != nil {
				es, _ := json.Marshal(err.Error())
				script = fmt.Sprintf("window['%s'].reject(new Error(%s)); delete window['%s'];", msg.CB, es, msg.CB)
			} else {
				rs, _ := json.Marshal(res)
				script = fmt.Sprintf("window['%s'].resolve(%s); delete window['%s'];", msg.CB, rs, msg.CB)
			}
			dw.mu.RLock()
			term := dw.terminated
			dw.mu.RUnlock()
			if term {
				return
			}
			_ = dw.Eval(script)
		}()
	}
}

//export goWebViewWindowWillClose
func goWebViewWindowWillClose(handle C.uintptr_t) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	wv.Terminate()
}

//export goWebViewNavigationFinished
func goWebViewNavigationFinished(handle C.uintptr_t, url *C.char) {
	// Hook for future use (e.g., protocol injection)
}

func boolInt(b bool) C.int {
	if b {
		return 1
	}
	return 0
}
