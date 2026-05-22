#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#include <stdlib.h>
#include <string.h>

// --- Minimal COM definitions for WebView2 ---

typedef long HRESULT;
#define S_OK ((HRESULT)0L)
#define E_NOINTERFACE ((HRESULT)0x80004002L)
#define E_FAIL ((HRESULT)0x80004005L)

#define STDMETHODCALLTYPE __stdcall

typedef struct IUnknown IUnknown;
typedef struct ICoreWebView2Environment ICoreWebView2Environment;
typedef struct ICoreWebView2Controller ICoreWebView2Controller;
typedef struct ICoreWebView2 ICoreWebView2;
typedef struct ICoreWebView2WebMessageReceivedEventArgs ICoreWebView2WebMessageReceivedEventArgs;
typedef struct EventRegistrationToken { LONGLONG value; } EventRegistrationToken;

typedef struct {
    HRESULT (STDMETHODCALLTYPE *QueryInterface)(IUnknown *self, REFIID riid, void **ppv);
    ULONG (STDMETHODCALLTYPE *AddRef)(IUnknown *self);
    ULONG (STDMETHODCALLTYPE *Release)(IUnknown *self);
} IUnknownVtbl;

struct IUnknown {
    const IUnknownVtbl *lpVtbl;
};

// --- WebView2 method indices (ICoreWebView2) ---
// IUnknown=0-2, then:
//  3=get_Settings, 4=get_Source, 5=Navigate, 6=NavigateToString,
//  7-26=event add/remove pairs,
// 27=AddScriptToExecuteOnDocumentCreated, 28=RemoveScriptToExecuteOnDocumentCreated,
// 29=ExecuteScript, 30=CapturePreview, 31=Reload,
// 32=PostWebMessageAsString, 33=PostWebMessageAsJSON,
// 34=add_WebMessageReceived, 35=remove_WebMessageReceived,
// ... 40=GoBack, 41=GoForward
#define WV2_NAVIGATE 5
#define WV2_NAVIGATE_TO_STRING 6
#define WV2_ADD_SCRIPT_ON_DOCUMENT_CREATED 27
#define WV2_EXECUTE_SCRIPT 29
#define WV2_RELOAD 31
#define WV2_POST_WEB_MESSAGE_AS_JSON 33
#define WV2_ADD_WEB_MESSAGE_RECEIVED 34
#define WV2_ADD_PERMISSION_REQUESTED 23
#define WV2_GO_BACK 40
#define WV2_GO_FORWARD 41

// Custom window message used to marshal JS evaluation onto the UI thread.
// WebView2 (like all WebView2 calls) must be used on the thread that created
// the message loop; bind callbacks complete on a background goroutine.
#define WV_WM_EVAL (WM_APP + 1)

static HRESULT STDMETHODCALLTYPE wv2_navigate(ICoreWebView2 *wv, LPCWSTR url) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, LPCWSTR);
    return ((Fn)((void ***)wv)[0][WV2_NAVIGATE])(wv, url);
}

static HRESULT STDMETHODCALLTYPE wv2_navigate_to_string(ICoreWebView2 *wv, LPCWSTR html) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, LPCWSTR);
    return ((Fn)((void ***)wv)[0][WV2_NAVIGATE_TO_STRING])(wv, html);
}

static HRESULT STDMETHODCALLTYPE wv2_execute_script(ICoreWebView2 *wv, LPCWSTR script) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, LPCWSTR);
    return ((Fn)((void ***)wv)[0][WV2_EXECUTE_SCRIPT])(wv, script);
}

static HRESULT STDMETHODCALLTYPE wv2_add_script_on_document_created(ICoreWebView2 *wv, LPCWSTR script, IUnknown *handler) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, LPCWSTR, IUnknown *);
    return ((Fn)((void ***)wv)[0][WV2_ADD_SCRIPT_ON_DOCUMENT_CREATED])(wv, script, handler);
}

static HRESULT STDMETHODCALLTYPE wv2_reload(ICoreWebView2 *wv) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *);
    return ((Fn)((void ***)wv)[0][WV2_RELOAD])(wv);
}

static HRESULT STDMETHODCALLTYPE wv2_go_back(ICoreWebView2 *wv) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *);
    return ((Fn)((void ***)wv)[0][WV2_GO_BACK])(wv);
}

static HRESULT STDMETHODCALLTYPE wv2_go_forward(ICoreWebView2 *wv) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *);
    return ((Fn)((void ***)wv)[0][WV2_GO_FORWARD])(wv);
}

static HRESULT STDMETHODCALLTYPE wv2_post_web_message_as_json(ICoreWebView2 *wv, LPCWSTR json) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, LPCWSTR);
    return ((Fn)((void ***)wv)[0][WV2_POST_WEB_MESSAGE_AS_JSON])(wv, json);
}

static HRESULT STDMETHODCALLTYPE wv2_add_web_message_received(ICoreWebView2 *wv, IUnknown *handler, EventRegistrationToken *token) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2 *, IUnknown *, EventRegistrationToken *);
    return ((Fn)((void ***)wv)[0][WV2_ADD_WEB_MESSAGE_RECEIVED])(wv, handler, token);
}

// --- ICoreWebView2Controller helpers ---
// IUnknown=0-2, 3=get_IsVisible, 4=put_IsVisible, 5=get_Bounds, 6=put_Bounds,
// ... 25=get_CoreWebView2
#define WVC_PUT_BOUNDS 6
#define WVC_GET_CORE_WEBVIEW2 25

static HRESULT STDMETHODCALLTYPE wvc_put_bounds(ICoreWebView2Controller *c, RECT bounds) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2Controller *, RECT);
    return ((Fn)((void ***)c)[0][WVC_PUT_BOUNDS])(c, bounds);
}

static HRESULT STDMETHODCALLTYPE wvc_get_core_webview2(ICoreWebView2Controller *c, ICoreWebView2 **wv) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2Controller *, ICoreWebView2 **);
    return ((Fn)((void ***)c)[0][WVC_GET_CORE_WEBVIEW2])(c, wv);
}

// --- ICoreWebView2Environment helpers ---
// IUnknown=0-2, 3=CreateCoreWebView2Controller, 4=CreateWebResourceResponse
#define WVE_CREATE_CONTROLLER 3

static HRESULT STDMETHODCALLTYPE wve_create_controller(ICoreWebView2Environment *env, HWND parent, IUnknown *handler) {
    typedef HRESULT (STDMETHODCALLTYPE *Fn)(ICoreWebView2Environment *, HWND, IUnknown *);
    return ((Fn)((void ***)env)[0][WVE_CREATE_CONTROLLER])(env, parent, handler);
}

// --- WebView2Loader function pointer ---
typedef HRESULT (STDMETHODCALLTYPE *CreateWebView2Fn)(
    const wchar_t *browserExecutableFolder,
    const wchar_t *userDataFolder,
    void *environmentOptions,
    IUnknown *environmentCreatedHandler);

// --- Globals ---
static HWND g_hwnd = NULL;
static HMODULE g_webview2 = NULL;
static ICoreWebView2Controller *g_controller = NULL;
static ICoreWebView2 *g_webview = NULL;
static int g_running = 0;
static uintptr_t g_handle = 0;
static HANDLE g_initEvent = NULL;

// --- Go callbacks ---
extern void goWebViewMessageReceived(uintptr_t handle, char *name, char *body);
extern void goWebViewWindowWillClose(uintptr_t handle);

// --- COM callback helpers ---

typedef HRESULT (STDMETHODCALLTYPE *Invoke2Fn)(IUnknown *self, HRESULT errorCode, void *result);

typedef struct {
    IUnknownVtbl base;
    Invoke2Fn Invoke;
} ICompletedHandlerVtbl2;

typedef struct {
    const ICompletedHandlerVtbl2 *lpVtbl;
    int refCount;
} ICompletedHandler2;

static HRESULT STDMETHODCALLTYPE handler2_QueryInterface(IUnknown *self, REFIID riid, void **ppv) {
    *ppv = self;
    return S_OK;
}

static ULONG STDMETHODCALLTYPE handler2_AddRef(IUnknown *self) {
    ICompletedHandler2 *h = (ICompletedHandler2 *)self;
    return (ULONG)++h->refCount;
}

static ULONG STDMETHODCALLTYPE handler2_Release(IUnknown *self) {
    ICompletedHandler2 *h = (ICompletedHandler2 *)self;
    return (ULONG)--h->refCount;
}

// --- Environment created handler ---

static HRESULT STDMETHODCALLTYPE envHandler_Invoke(IUnknown *self, HRESULT errorCode, ICoreWebView2Environment *env) {
    if (FAILED(errorCode) || !env) {
        SetEvent(g_initEvent);
        return E_FAIL;
    }

    static ICompletedHandlerVtbl2 controllerVtbl = {
        { handler2_QueryInterface, handler2_AddRef, handler2_Release },
        NULL // set below
    };
    static ICompletedHandler2 controllerHandler = { &controllerVtbl, 1 };

    // Controller handler inline to avoid forward-declaration issues
    HRESULT STDMETHODCALLTYPE controllerHandler_Invoke(IUnknown *s, HRESULT errorCode, ICoreWebView2Controller *controller) {
        (void)s;
        if (FAILED(errorCode) || !controller) {
            SetEvent(g_initEvent);
            return E_FAIL;
        }
        g_controller = controller;

        ICoreWebView2 *wv = NULL;
        if (SUCCEEDED(wvc_get_core_webview2(controller, &wv)) && wv) {
            g_webview = wv;

            // Set bounds to fill the window
            RECT bounds;
            GetClientRect(g_hwnd, &bounds);
            wvc_put_bounds(controller, bounds);

            // Add web message received handler for JS bindings
            static ICompletedHandlerVtbl2 msgVtbl = {
                { handler2_QueryInterface, handler2_AddRef, handler2_Release },
                NULL
            };
            static ICompletedHandler2 msgHandler = { &msgVtbl, 1 };

            HRESULT STDMETHODCALLTYPE msgHandler_Invoke(IUnknown *s, ICoreWebView2 *sender, ICoreWebView2WebMessageReceivedEventArgs *args) {
                (void)s; (void)sender;
                if (!args) return S_OK;

                // Extract the web message (JSON string) from args
                // ICoreWebView2WebMessageReceivedEventArgs has get_WebMessageAsJson at index 3
                typedef HRESULT (STDMETHODCALLTYPE *GetJsonFn)(ICoreWebView2WebMessageReceivedEventArgs *, LPWSTR *);
                LPWSTR json = NULL;
                HRESULT hr = ((GetJsonFn)((void ***)args)[0][3])(args, &json);
                if (SUCCEEDED(hr) && json) {
                    // Convert UTF-16 to UTF-8
                    int len = WideCharToMultiByte(CP_UTF8, 0, json, -1, NULL, 0, NULL, NULL);
                    if (len > 0) {
                        char *cjson = (char *)malloc(len);
                        if (cjson) {
                            WideCharToMultiByte(CP_UTF8, 0, json, -1, cjson, len, NULL, NULL);
                            goWebViewMessageReceived(g_handle, "goBridge", cjson);
                            free(cjson);
                        }
                    }
                    // Free the returned string using CoTaskMemFree (index 2 = Release on args)
                    // Actually the string is allocated by the COM object; we need to free it.
                    // WebView2 strings are allocated with CoTaskMemAlloc.
                    CoTaskMemFree(json);
                }
                return S_OK;
            }
            msgVtbl.Invoke = msgHandler_Invoke;

            EventRegistrationToken token;
            wv2_add_web_message_received(wv, (IUnknown *)&msgHandler, &token);

            // Add permission requested handler to allow JS clipboard access
            static ICompletedHandlerVtbl2 permVtbl = {
                { handler2_QueryInterface, handler2_AddRef, handler2_Release },
                NULL
            };
            static ICompletedHandler2 permHandler = { &permVtbl, 1 };

            HRESULT STDMETHODCALLTYPE permHandler_Invoke(IUnknown *s, ICoreWebView2 *sender, ICoreWebView2PermissionRequestedEventArgs *args) {
                (void)s; (void)sender;
                if (!args) return S_OK;
                // ICoreWebView2PermissionRequestedEventArgs:
                // 3=get_PermissionKind, 6=put_State
                typedef HRESULT (STDMETHODCALLTYPE *GetKindFn)(ICoreWebView2PermissionRequestedEventArgs *, int *);
                typedef HRESULT (STDMETHODCALLTYPE *PutStateFn)(ICoreWebView2PermissionRequestedEventArgs *, int);
                int kind = 0;
                HRESULT hr = ((GetKindFn)((void ***)args)[0][3])(args, &kind);
                if (SUCCEEDED(hr) && kind == 6) { // COREWEBVIEW2_PERMISSION_KIND_CLIPBOARD_READ
                    ((PutStateFn)((void ***)args)[0][6])(args, 1); // COREWEBVIEW2_PERMISSION_STATE_ALLOW
                }
                return S_OK;
            }
            permVtbl.Invoke = permHandler_Invoke;

            EventRegistrationToken permToken;
            typedef HRESULT (STDMETHODCALLTYPE *AddPermFn)(ICoreWebView2 *, IUnknown *, EventRegistrationToken *);
            ((AddPermFn)((void ***)wv)[0][WV2_ADD_PERMISSION_REQUESTED])(wv, (IUnknown *)&permHandler, &permToken);
        }

        SetEvent(g_initEvent);
        return S_OK;
    }
    controllerVtbl.Invoke = controllerHandler_Invoke;

    wve_create_controller(env, g_hwnd, (IUnknown *)&controllerHandler);
    return S_OK;
}

// --- Win32 window proc ---

static LRESULT CALLBACK wndProc(HWND hwnd, UINT msg, WPARAM wParam, LPARAM lParam) {
    switch (msg) {
    case WV_WM_EVAL: {
        // Runs on the UI thread; lParam is a heap-allocated UTF-16 script that
        // we own and must free.
        wchar_t *script = (wchar_t *)lParam;
        if (script) {
            if (g_webview) {
                wv2_execute_script(g_webview, script);
            }
            free(script);
        }
        return 0;
    }
    case WM_SIZE:
        if (g_controller && g_hwnd) {
            RECT bounds;
            GetClientRect(g_hwnd, &bounds);
            wvc_put_bounds(g_controller, bounds);
        }
        return 0;
    case WM_CLOSE:
        if (g_handle) {
            goWebViewWindowWillClose(g_handle);
        }
        DestroyWindow(hwnd);
        return 0;
    case WM_DESTROY:
        PostQuitMessage(0);
        return 0;
    }
    return DefWindowProc(hwnd, msg, wParam, lParam);
}

// --- Window creation ---

HWND wvCreateWindow(int width, int height, const char *title) {
    HINSTANCE hInstance = GetModuleHandle(NULL);
    WNDCLASSEX wc = {0};
    wc.cbSize = sizeof(WNDCLASSEX);
    wc.lpfnWndProc = wndProc;
    wc.hInstance = hInstance;
    wc.lpszClassName = "webviewgoWindow";
    RegisterClassEx(&wc);

    HWND hwnd = CreateWindowEx(0, "webviewgoWindow", title,
        WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT,
        width, height, NULL, NULL, hInstance, NULL);
    return hwnd;
}

// --- WebView2 initialization ---

int wvInitWebView2(HWND hwnd, uintptr_t handle, int width, int height) {
    (void)width; (void)height;
    g_hwnd = hwnd;
    g_handle = handle;

    g_webview2 = LoadLibrary("WebView2Loader.dll");
    if (!g_webview2) {
        return -1;
    }

    CreateWebView2Fn createEnv = (CreateWebView2Fn)GetProcAddress(g_webview2, "CreateCoreWebView2EnvironmentWithOptions");
    if (!createEnv) {
        FreeLibrary(g_webview2);
        g_webview2 = NULL;
        return -1;
    }

    g_initEvent = CreateEvent(NULL, TRUE, FALSE, NULL);
    if (!g_initEvent) {
        FreeLibrary(g_webview2);
        g_webview2 = NULL;
        return -1;
    }

    static ICompletedHandlerVtbl2 envVtbl = {
        { handler2_QueryInterface, handler2_AddRef, handler2_Release },
        envHandler_Invoke
    };
    static ICompletedHandler2 envHandler = { &envVtbl, 1 };

    HRESULT hr = createEnv(NULL, NULL, NULL, (IUnknown *)&envHandler);
    if (FAILED(hr)) {
        CloseHandle(g_initEvent);
        g_initEvent = NULL;
        FreeLibrary(g_webview2);
        g_webview2 = NULL;
        return -1;
    }

    // Pump messages until the webview is initialized
    while (WaitForSingleObject(g_initEvent, 0) != WAIT_OBJECT_0) {
        MSG msg;
        if (PeekMessage(&msg, NULL, 0, 0, PM_REMOVE)) {
            TranslateMessage(&msg);
            DispatchMessage(&msg);
        } else {
            Sleep(1);
        }
    }

    CloseHandle(g_initEvent);
    g_initEvent = NULL;

    if (!g_webview) {
        // Initialization failed inside the callback
        if (g_controller) {
            IUnknown *ctrl = (IUnknown *)g_controller;
            ctrl->lpVtbl->Release(ctrl);
            g_controller = NULL;
        }
        FreeLibrary(g_webview2);
        g_webview2 = NULL;
        return -1;
    }

    return 0;
}

// --- Cleanup ---

void wvDestroyWebView2() {
    if (g_webview) {
        IUnknown *wv = (IUnknown *)g_webview;
        wv->lpVtbl->Release(wv);
        g_webview = NULL;
    }
    if (g_controller) {
        IUnknown *ctrl = (IUnknown *)g_controller;
        ctrl->lpVtbl->Release(ctrl);
        g_controller = NULL;
    }
    if (g_webview2) {
        FreeLibrary(g_webview2);
        g_webview2 = NULL;
    }
}

// --- Navigation ---

static wchar_t *utf8_to_utf16(const char *src) {
    if (!src) return NULL;
    int len = MultiByteToWideChar(CP_UTF8, 0, src, -1, NULL, 0);
    if (len <= 0) return NULL;
    wchar_t *dst = (wchar_t *)malloc(len * sizeof(wchar_t));
    if (dst) {
        MultiByteToWideChar(CP_UTF8, 0, src, -1, dst, len);
    }
    return dst;
}

void wvNavigate(const char *url) {
    if (!g_webview) return;
    wchar_t *wurl = utf8_to_utf16(url);
    if (wurl) {
        wv2_navigate(g_webview, wurl);
        free(wurl);
    }
}

void wvLoadHTML(const char *html) {
    if (!g_webview) return;
    wchar_t *whtml = utf8_to_utf16(html);
    if (whtml) {
        wv2_navigate_to_string(g_webview, whtml);
        free(whtml);
    }
}

void wvEval(const char *script) {
    if (!g_webview) return;
    wchar_t *wscript = utf8_to_utf16(script);
    if (wscript) {
        wv2_execute_script(g_webview, wscript);
        free(wscript);
    }
}

// wvEvalAsync marshals JS evaluation onto the UI thread. Safe to call from any
// thread (PostMessageW is thread-safe). The UI thread frees the script.
void wvEvalAsync(const char *script) {
    if (!g_hwnd) return;
    wchar_t *wscript = utf8_to_utf16(script);
    if (!wscript) return;
    if (!PostMessageW(g_hwnd, WV_WM_EVAL, 0, (LPARAM)wscript)) {
        free(wscript); // post failed; avoid leaking the buffer
    }
}

// No-op completion handler for AddScriptToExecuteOnDocumentCreated. Its Invoke
// receives (HRESULT errorCode, LPCWSTR id); the id pointer maps onto the
// generic void* slot.
static HRESULT STDMETHODCALLTYPE addScriptHandler_Invoke(IUnknown *self, HRESULT errorCode, void *id) {
    (void)self; (void)errorCode; (void)id;
    return S_OK;
}

// wvAddUserScript registers a script that runs at document creation on every
// navigation, so JS-to-Go bindings persist across page loads. Must run on the
// UI thread (called during Bind, before the message loop spins on a worker).
void wvAddUserScript(const char *script) {
    if (!g_webview) return;
    wchar_t *wscript = utf8_to_utf16(script);
    if (!wscript) return;

    static ICompletedHandlerVtbl2 vtbl = {
        { handler2_QueryInterface, handler2_AddRef, handler2_Release },
        addScriptHandler_Invoke
    };
    static ICompletedHandler2 handler = { &vtbl, 1 };

    wv2_add_script_on_document_created(g_webview, wscript, (IUnknown *)&handler);
    free(wscript);
}

void wvReload() {
    if (!g_webview) return;
    wv2_reload(g_webview);
}

void wvGoBack() {
    if (!g_webview) return;
    wv2_go_back(g_webview);
}

void wvGoForward() {
    if (!g_webview) return;
    wv2_go_forward(g_webview);
}

// --- Window controls ---

void wvSetTitle(HWND hwnd, const char *title) {
    SetWindowText(hwnd, title);
}

void wvSetSize(HWND hwnd, int width, int height) {
    SetWindowPos(hwnd, NULL, 0, 0, width, height, SWP_NOMOVE | SWP_NOZORDER);
    if (g_controller && g_hwnd) {
        RECT bounds;
        GetClientRect(g_hwnd, &bounds);
        wvc_put_bounds(g_controller, bounds);
    }
}

void wvShow(HWND hwnd) {
    ShowWindow(hwnd, SW_SHOW);
    UpdateWindow(hwnd);
}

void wvHide(HWND hwnd) {
    ShowWindow(hwnd, SW_HIDE);
}

// --- Message loop ---

void wvRunMsgLoop() {
    g_running = 1;
    MSG msg;
    while (g_running && GetMessage(&msg, NULL, 0, 0)) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }
}

void wvTerminateMsgLoop() {
    g_running = 0;
    PostQuitMessage(0);
}

// --- Dialog helpers ---

#include <commdlg.h>

char **openDialogW(int allowFiles, int allowDirs, int allowMultiple,
                   const wchar_t *title, const wchar_t *directory,
                   int *outCount) {
    (void)allowDirs; // Windows file dialog doesn't support directory picking in the same dialog
    OPENFILENAMEW ofn = {0};
    ofn.lStructSize = sizeof(OPENFILENAMEW);

    wchar_t fileBuffer[65536] = {0};
    ofn.lpstrFile = fileBuffer;
    ofn.nMaxFile = 65536;
    ofn.lpstrTitle = title;
    if (directory) {
        ofn.lpstrInitialDir = directory;
    }
    ofn.Flags = OFN_EXPLORER | OFN_NOCHANGEDIR;
    if (allowMultiple) {
        ofn.Flags |= OFN_ALLOWMULTISELECT;
    }

    BOOL ok = GetOpenFileNameW(&ofn);
    if (!ok) {
        *outCount = 0;
        return NULL;
    }

    // Count selected files
    int count = 1;
    if (allowMultiple) {
        wchar_t *p = fileBuffer + wcslen(fileBuffer) + 1;
        if (*p) {
            count = 0;
            while (*p) {
                count++;
                p += wcslen(p) + 1;
            }
        }
    }

    char **result = (char **)malloc(sizeof(char *) * count);
    if (!result) {
        *outCount = 0;
        return NULL;
    }

    if (allowMultiple && count > 1) {
        wchar_t dir[MAX_PATH];
        wcscpy(dir, fileBuffer);
        int dirLen = wcslen(dir);
        if (dir[dirLen - 1] != L'\\' && dir[dirLen - 1] != L'/') {
            wcscat(dir, L"\\");
        }

        wchar_t *p = fileBuffer + wcslen(fileBuffer) + 1;
        for (int i = 0; i < count; i++) {
            wchar_t fullPath[MAX_PATH];
            wcscpy(fullPath, dir);
            wcscat(fullPath, p);
            int len = WideCharToMultiByte(CP_UTF8, 0, fullPath, -1, NULL, 0, NULL, NULL);
            result[i] = (char *)malloc(len);
            WideCharToMultiByte(CP_UTF8, 0, fullPath, -1, result[i], len, NULL, NULL);
            p += wcslen(p) + 1;
        }
    } else {
        int len = WideCharToMultiByte(CP_UTF8, 0, fileBuffer, -1, NULL, 0, NULL, NULL);
        result[0] = (char *)malloc(len);
        WideCharToMultiByte(CP_UTF8, 0, fileBuffer, -1, result[0], len, NULL, NULL);
    }

    *outCount = count;
    return result;
}

wchar_t *saveDialogW(const wchar_t *title, const wchar_t *directory,
                     const wchar_t *defaultFile) {
    OPENFILENAMEW ofn = {0};
    ofn.lStructSize = sizeof(OPENFILENAMEW);

    wchar_t fileBuffer[MAX_PATH] = {0};
    if (defaultFile) {
        wcsncpy(fileBuffer, defaultFile, MAX_PATH - 1);
    }
    ofn.lpstrFile = fileBuffer;
    ofn.nMaxFile = MAX_PATH;
    ofn.lpstrTitle = title;
    if (directory) {
        ofn.lpstrInitialDir = directory;
    }
    ofn.Flags = OFN_EXPLORER | OFN_NOCHANGEDIR | OFN_OVERWRITEPROMPT;

    BOOL ok = GetSaveFileNameW(&ofn);
    if (!ok) {
        return NULL;
    }
    return wcsdup(fileBuffer);
}

int messageDialogW(const wchar_t *title, const wchar_t *message,
                   int level, int buttons) {
    UINT uType = 0;
    switch (level) {
        case 1: uType |= MB_ICONWARNING; break;
        case 2: uType |= MB_ICONERROR; break;
        default: uType |= MB_ICONINFORMATION; break;
    }
    switch (buttons) {
        case 1: uType |= MB_OKCANCEL; break;
        case 2: uType |= MB_YESNO; break;
        case 3: uType |= MB_YESNOCANCEL; break;
        case 4: uType |= MB_RETRYCANCEL; break;
        case 5: uType |= MB_ABORTRETRYIGNORE; break;
        default: uType |= MB_OK; break;
    }

    int result = MessageBoxW(NULL, message, title, uType);
    switch (buttons) {
        case 1: return (result == IDOK) ? 1 : 0;
        case 2: return (result == IDYES) ? 2 : 3;
        case 3:
            if (result == IDYES) return 2;
            if (result == IDNO) return 3;
            return 0;
        case 4: return (result == IDRETRY) ? 5 : 0;
        case 5:
            if (result == IDABORT) return 4;
            if (result == IDRETRY) return 5;
            return 6;
        default: return (result == IDOK) ? 1 : 0;
    }
}

void freeWStringArray(char **arr, int count) {
    for (int i = 0; i < count; i++) {
        free(arr[i]);
    }
    free(arr);
}

// --- Clipboard helpers ---

char *clipboardReadTextW() {
    if (!OpenClipboard(NULL)) {
        return NULL;
    }
    HANDLE hData = GetClipboardData(CF_UNICODETEXT);
    if (!hData) {
        CloseClipboard();
        return NULL;
    }
    wchar_t *wtext = (wchar_t *)GlobalLock(hData);
    if (!wtext) {
        CloseClipboard();
        return NULL;
    }
    int len = WideCharToMultiByte(CP_UTF8, 0, wtext, -1, NULL, 0, NULL, NULL);
    char *text = NULL;
    if (len > 0) {
        text = (char *)malloc(len);
        if (text) {
            WideCharToMultiByte(CP_UTF8, 0, wtext, -1, text, len, NULL, NULL);
        }
    }
    GlobalUnlock(hData);
    CloseClipboard();
    return text;
}

void clipboardWriteTextW(const wchar_t *text) {
    if (!text) return;
    if (!OpenClipboard(NULL)) return;
    EmptyClipboard();

    int len = (int)(wcslen(text) + 1) * sizeof(wchar_t);
    HGLOBAL hMem = GlobalAlloc(GMEM_MOVEABLE, len);
    if (hMem) {
        wchar_t *dest = (wchar_t *)GlobalLock(hMem);
        if (dest) {
            memcpy(dest, text, len);
            GlobalUnlock(hMem);
            SetClipboardData(CF_UNICODETEXT, hMem);
        }
    }
    CloseClipboard();
}
