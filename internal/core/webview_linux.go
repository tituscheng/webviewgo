//go:build linux

package core

/*
#cgo pkg-config: gtk+-3.0 webkit2gtk-4.0
#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <stdlib.h>

// Go callbacks
extern void goWebViewMessageReceived(uintptr_t handle, char *name, char *body);
extern void goWebViewWindowWillClose(uintptr_t handle);

static void onScriptMessage(WebKitUserContentManager *manager,
                            WebKitJavascriptResult *result,
                            gpointer user_data) {
    JSCValue *value = webkit_javascript_result_get_js_value(result);
    char *json = jsc_value_to_json(value, 0);
    goWebViewMessageReceived((uintptr_t)user_data, "goBridge", json ? json : "");
    g_free(json);
}

static void onDestroy(GtkWidget *widget, gpointer user_data) {
    goWebViewWindowWillClose((uintptr_t)user_data);
}

static WebKitWebView *createWebView(uintptr_t handle, int enableDevTools, const char *userAgent) {
    WebKitUserContentManager *cm = webkit_user_content_manager_new();
    g_signal_connect(cm, "script-message-received::goBridge", G_CALLBACK(onScriptMessage), (gpointer)handle);
    webkit_user_content_manager_register_script_message_handler(cm, "goBridge");

    WebKitSettings *settings = webkit_settings_new();
    if (enableDevTools) {
        webkit_settings_set_enable_developer_extras(settings, TRUE);
    }
    if (userAgent && userAgent[0]) {
        webkit_settings_set_user_agent(settings, userAgent);
    }

    WebKitWebView *webView = WEBKIT_WEB_VIEW(g_object_new(WEBKIT_TYPE_WEB_VIEW,
        "settings", settings,
        "user-content-manager", cm,
        NULL));
    g_object_unref(settings);
    return webView;
}

static GtkWindow *createWindow(int width, int height, const char *title) {
    gtk_init(NULL, NULL);
    GtkWindow *window = GTK_WINDOW(gtk_window_new(GTK_WINDOW_TOPLEVEL));
    gtk_window_set_default_size(window, width, height);
    gtk_window_set_title(window, title);
    return window;
}

static void webViewNavigate(WebKitWebView *webView, const char *url) {
    webkit_web_view_load_uri(webView, url);
}

static void webViewLoadHTML(WebKitWebView *webView, const char *html, const char *baseURL) {
    webkit_web_view_load_html(webView, html, baseURL);
}

static void webViewEval(WebKitWebView *webView, const char *script) {
    webkit_web_view_run_javascript(webView, script, NULL, NULL, NULL);
}

static void webViewReload(WebKitWebView *webView) {
    webkit_web_view_reload(webView);
}

static void webViewGoBack(WebKitWebView *webView) {
    webkit_web_view_go_back(webView);
}

static void webViewGoForward(WebKitWebView *webView) {
    webkit_web_view_go_forward(webView);
}

static void windowSetTitle(GtkWindow *window, const char *title) {
    gtk_window_set_title(window, title);
}

static void windowSetSize(GtkWindow *window, int width, int height) {
    gtk_window_resize(window, width, height);
}

static void windowShow(GtkWindow *window) {
    gtk_widget_show_all(GTK_WIDGET(window));
}

static void windowHide(GtkWindow *window) {
    gtk_widget_hide(GTK_WIDGET(window));
}

static void runGtk() {
    gtk_main();
}

static void stopGtk() {
    gtk_main_quit();
}

static char *clipboardReadText() {
    GtkClipboard *cb = gtk_clipboard_get(GDK_SELECTION_CLIPBOARD);
    return gtk_clipboard_wait_for_text(cb);
}

static void clipboardWriteText(const char *text) {
    GtkClipboard *cb = gtk_clipboard_get(GDK_SELECTION_CLIPBOARD);
    gtk_clipboard_set_text(cb, text, -1);
}

// evalOnMainThread queues a JavaScript evaluation on the GTK main thread.
// WebKitGTK functions must be called from the thread running gtk_main().
typedef struct {
    WebKitWebView *webView;
    char *script;
} JsEvalData;

static gboolean runJsIdle(gpointer user_data) {
    JsEvalData *data = (JsEvalData*)user_data;
    webkit_web_view_run_javascript(data->webView, data->script, NULL, NULL, NULL);
    g_free(data->script);
    g_free(data);
    return G_SOURCE_REMOVE;
}

static void evalOnMainThread(WebKitWebView *webView, const char *script) {
    JsEvalData *data = g_new(JsEvalData, 1);
    data->webView = webView;
    data->script = g_strdup(script);
    g_idle_add(runJsIdle, data);
}

// addUserScript installs a script that runs at document start on every page
// load (current and future navigations), so JS-to-Go bindings survive
// navigation instead of being defined once via a transient eval.
static void addUserScript(WebKitWebView *webView, const char *source) {
    WebKitUserContentManager *cm = webkit_web_view_get_user_content_manager(webView);
    WebKitUserScript *script = webkit_user_script_new(source,
        WEBKIT_USER_CONTENT_INJECT_ALL_FRAMES,
        WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, NULL, NULL);
    webkit_user_content_manager_add_script(cm, script);
    webkit_user_script_unref(script);
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
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

// linuxWebView is the Linux WebKitGTK backend.
type linuxWebView struct {
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
}

// init pins the main goroutine to the main OS thread. WebKitGTK/GTK requires
// that its main loop and widget calls happen on the thread that initialized
// GTK. Package init functions run on the main goroutine before main(), while
// it is still on the main thread, so this is the only reliable place to lock
// it. Callers must invoke New and Run from the main goroutine.
func init() {
	runtime.LockOSThread()
}

func newNative(opts types.Options) (Platform, error) {
	title := C.CString(opts.Title)
	defer C.free(unsafe.Pointer(title))

	window := C.createWindow(C.int(opts.Width), C.int(opts.Height), title)
	if window == nil {
		return nil, fmt.Errorf("core: failed to create window")
	}

	wv := &linuxWebView{
		logger:   logOpts(opts),
		bindings: make(map[string]func([]any) (any, error)),
		schemes:  make(map[string]types.SchemeHandler),
		done:     make(chan struct{}),
	}
	wv.ctx, wv.cancel = context.WithCancel(context.Background())
	wv.handle = nextHandle(wv)
	wv.window = unsafe.Pointer(window)

	ua := C.CString(opts.UserAgent)
	defer C.free(unsafe.Pointer(ua))
	webView := C.createWebView(C.uintptr_t(wv.handle), boolInt(opts.Devtools), ua)
	if webView == nil {
		return nil, fmt.Errorf("core: failed to create webview")
	}
	wv.webView = unsafe.Pointer(webView)

	gtkWindow := (*C.GtkWindow)(window)
	C.gtk_container_add((*C.GtkContainer)(unsafe.Pointer(gtkWindow)), (*C.GtkWidget)(webView))

	destroySignal := C.CString("destroy")
	C.g_signal_connect((*C.GObject)(unsafe.Pointer(gtkWindow)), destroySignal,
		C.GCallback(C.onDestroy), C.gpointer(wv.handle))
	C.free(unsafe.Pointer(destroySignal))

	return wv, nil
}

func (w *linuxWebView) Run() error {
	defer close(w.done)
	C.runGtk()
	return nil
}

func (w *linuxWebView) Terminate() {
	w.mu.Lock()
	if w.terminated {
		w.mu.Unlock()
		return
	}
	w.terminated = true
	w.mu.Unlock()
	w.cancel()
	C.stopGtk()
}

func (w *linuxWebView) Destroy() error {
	w.Terminate()
	select {
	case <-w.done:
	case <-w.ctx.Done():
	}
	releaseHandle(w.handle)
	return nil
}

func (w *linuxWebView) SetUserAgent(ua string) {
	settings := C.webkit_web_view_get_settings((*C.WebKitWebView)(w.webView))
	if settings != nil {
		cs := C.CString(ua)
		defer C.free(unsafe.Pointer(cs))
		C.webkit_settings_set_user_agent(settings, cs)
	}
}

func (w *linuxWebView) SetTitle(title string) {
	cs := C.CString(title)
	defer C.free(unsafe.Pointer(cs))
	C.windowSetTitle((*C.GtkWindow)(w.window), cs)
}

func (w *linuxWebView) SetSize(width, height int, hint types.Hint) {
	C.windowSetSize((*C.GtkWindow)(w.window), C.int(width), C.int(height))
}

func (w *linuxWebView) SetMinSize(width, height int) {}
func (w *linuxWebView) SetMaxSize(width, height int) {}
func (w *linuxWebView) SetFullscreen(bool)           {}
func (w *linuxWebView) SetAlwaysOnTop(bool)          {}
func (w *linuxWebView) Show()                        { C.windowShow((*C.GtkWindow)(w.window)) }
func (w *linuxWebView) Hide()                        { C.windowHide((*C.GtkWindow)(w.window)) }

func (w *linuxWebView) Navigate(url string) error {
	cs := C.CString(url)
	defer C.free(unsafe.Pointer(cs))
	C.webViewNavigate((*C.WebKitWebView)(w.webView), cs)
	return nil
}

func (w *linuxWebView) LoadHTML(html, baseURL string) error {
	chs := C.CString(html)
	defer C.free(unsafe.Pointer(chs))
	cbase := C.CString(baseURL)
	defer C.free(unsafe.Pointer(cbase))
	C.webViewLoadHTML((*C.WebKitWebView)(w.webView), chs, cbase)
	return nil
}

func (w *linuxWebView) Reload()  { C.webViewReload((*C.WebKitWebView)(w.webView)) }
func (w *linuxWebView) Back()    { C.webViewGoBack((*C.WebKitWebView)(w.webView)) }
func (w *linuxWebView) Forward() { C.webViewGoForward((*C.WebKitWebView)(w.webView)) }

func (w *linuxWebView) Eval(script string) error {
	cs := C.CString(script)
	defer C.free(unsafe.Pointer(cs))
	C.webViewEval((*C.WebKitWebView)(w.webView), cs)
	return nil
}

func (w *linuxWebView) Bind(name string, fn func(args []any) (any, error)) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bindings[name] = fn
	script := fmt.Sprintf(`
	window.%s = function(...args) {
		return new Promise((resolve, reject) => {
			const id = '__go_' + Math.random().toString(36).slice(2);
			window[id] = { resolve, reject };
			window.goBridge.postMessage({
				bind: %q,
				args: args,
				cb: id
			});
		});
	};
`, name, name)
	// Install as a document-start user script so the binding persists across
	// navigations, and eval once so it is available on the current document.
	cs := C.CString(script)
	C.addUserScript((*C.WebKitWebView)(w.webView), cs)
	C.free(unsafe.Pointer(cs))
	_ = w.Eval(script)
	return nil
}

func (w *linuxWebView) RegisterScheme(scheme string, handler types.SchemeHandler) error {
	return fmt.Errorf("core: RegisterScheme not yet implemented on linux")
}

func (w *linuxWebView) OpenDialog(opts types.OpenDialogOptions) ([]string, error) {
	return nil, fmt.Errorf("core: OpenDialog not yet implemented on linux")
}

func (w *linuxWebView) SaveDialog(opts types.SaveDialogOptions) (string, error) {
	return "", fmt.Errorf("core: SaveDialog not yet implemented on linux")
}

func (w *linuxWebView) MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error) {
	return types.DialogCancel, fmt.Errorf("core: MessageDialog not yet implemented on linux")
}

func (w *linuxWebView) ClipboardReadText() (string, error) {
	cs := C.clipboardReadText()
	if cs == nil {
		return "", nil
	}
	defer C.g_free(C.gpointer(cs))
	return C.GoString(cs), nil
}

func (w *linuxWebView) ClipboardWriteText(text string) error {
	cs := C.CString(text)
	defer C.free(unsafe.Pointer(cs))
	C.clipboardWriteText(cs)
	return nil
}

func (w *linuxWebView) Notify(title, body string) error {
	return nil
}

//export goWebViewMessageReceived
func goWebViewMessageReceived(handle C.uintptr_t, name *C.char, body *C.char) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	lw := wv.(*linuxWebView)
	n := C.GoString(name)
	b := C.GoString(body)

	if n == "goBridge" {
		var msg struct {
			Bind string `json:"bind"`
			Args []any  `json:"args"`
			CB   string `json:"cb"`
		}
		if err := json.Unmarshal([]byte(b), &msg); err != nil {
			lw.logger.Error("failed to unmarshal bridge message", "error", err)
			return
		}
		lw.mu.RLock()
		fn, ok := lw.bindings[msg.Bind]
		lw.mu.RUnlock()
		if !ok {
			lw.logger.Warn("unknown binding", "name", msg.Bind)
			return
		}
		go func() {
			defer func() {
				if r := recover(); r != nil {
					lw.logger.Error("binding callback panic", "name", msg.Bind, "recover", r)
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
			lw.mu.RLock()
			term := lw.terminated
			lw.mu.RUnlock()
			if term {
				return
			}
			cs := C.CString(script)
			C.evalOnMainThread((*C.WebKitWebView)(lw.webView), cs)
			C.free(unsafe.Pointer(cs))
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
