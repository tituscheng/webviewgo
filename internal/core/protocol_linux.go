//go:build linux

package core

/*
#cgo pkg-config: webkit2gtk-4.0 libsoup-2.4
#include <webkit2/webkit2.h>
#include <libsoup/soup.h>
#include <stdlib.h>
#include <string.h>

extern void goLinuxProtocolHandler(uintptr_t handle, char *scheme, char *url, char *method,
                                   char *headers, void *body, int bodyLen, uintptr_t reqHandle);

typedef struct {
    uintptr_t handle;
    char *scheme;
} LinuxSchemeReg;

typedef struct {
    WebKitURISchemeRequest *request;
} LinuxSchemeReq;

static GHashTable *linuxSchemeReqMap = NULL;
static GMutex linuxSchemeReqMutex;

static void ensureLinuxSchemeReqMap(void) {
    if (!linuxSchemeReqMap) {
        g_mutex_init(&linuxSchemeReqMutex);
        linuxSchemeReqMap = g_hash_table_new_full(g_direct_hash, g_direct_equal, NULL, g_free);
    }
}

static void storeLinuxSchemeReq(uintptr_t reqHandle, WebKitURISchemeRequest *request) {
    ensureLinuxSchemeReqMap();
    LinuxSchemeReq *entry = g_new(LinuxSchemeReq, 1);
    entry->request = request;
    g_mutex_lock(&linuxSchemeReqMutex);
    g_hash_table_insert(linuxSchemeReqMap, (gpointer)(uintptr_t)reqHandle, entry);
    g_mutex_unlock(&linuxSchemeReqMutex);
}

static WebKitURISchemeRequest *takeLinuxSchemeReq(uintptr_t reqHandle) {
    ensureLinuxSchemeReqMap();
    g_mutex_lock(&linuxSchemeReqMutex);
    LinuxSchemeReq *entry = g_hash_table_lookup(linuxSchemeReqMap, (gpointer)(uintptr_t)reqHandle);
    if (entry) {
        g_hash_table_remove(linuxSchemeReqMap, (gpointer)(uintptr_t)reqHandle);
    }
    g_mutex_unlock(&linuxSchemeReqMutex);
    return entry ? entry->request : NULL;
}

static void appendHeader(const char *name, const char *value, gpointer user_data) {
    GString *out = (GString *)user_data;
    g_string_append_printf(out, "%s: %s\n", name, value);
}

static void linuxUriSchemeCallback(WebKitURISchemeRequest *request, gpointer user_data) {
    LinuxSchemeReg *reg = (LinuxSchemeReg *)user_data;
    GUri *uri = webkit_uri_scheme_request_get_uri(request);
    gchar *url = g_uri_to_string(uri);
    const gchar *method = webkit_uri_scheme_request_get_http_method(request);
    if (!method) {
        method = "GET";
    }

    GString *headerBlob = g_string_new("");
    SoupMessageHeaders *reqHeaders = webkit_uri_scheme_request_get_http_headers(request);
    if (reqHeaders) {
        soup_message_headers_foreach(reqHeaders, appendHeader, headerBlob);
    }

    gsize bodyLen = 0;
    GBytes *bodyBytes = NULL;
    GInputStream *bodyStream = webkit_uri_scheme_request_get_http_body(request);
    if (bodyStream) {
        GError *err = NULL;
        bodyBytes = g_input_stream_read_bytes(bodyStream, G_MAXSIZE, NULL, &err);
        if (err) {
            g_error_free(err);
            bodyBytes = NULL;
        } else if (bodyBytes) {
            bodyLen = g_bytes_get_size(bodyBytes);
        }
    }

    static uintptr_t reqSeq = 1;
    uintptr_t reqHandle = reqSeq++;
    storeLinuxSchemeReq(reqHandle, request);

    goLinuxProtocolHandler(reg->handle, reg->scheme, url, (char *)method,
                           headerBlob->str,
                           bodyBytes ? (void *)g_bytes_get_data(bodyBytes, NULL) : NULL,
                           (int)bodyLen, reqHandle);

    if (bodyBytes) {
        g_bytes_unref(bodyBytes);
    }
    g_string_free(headerBlob, TRUE);
    g_free(url);
}

static void freeLinuxSchemeReg(gpointer data) {
    LinuxSchemeReg *reg = (LinuxSchemeReg *)data;
    g_free(reg->scheme);
    g_free(reg);
}

static void registerLinuxScheme(WebKitWebView *webView, const char *scheme, uintptr_t handle) {
    WebKitWebContext *context = webkit_web_view_get_context(webView);
    LinuxSchemeReg *reg = g_new(LinuxSchemeReg, 1);
    reg->handle = handle;
    reg->scheme = g_strdup(scheme);
    webkit_web_context_register_uri_scheme(context, scheme, linuxUriSchemeCallback,
                                           reg, freeLinuxSchemeReg);
}

void deliverLinuxSchemeResponse(uintptr_t reqHandle, int statusCode, char *headers,
                                void *body, int bodyLen) {
    WebKitURISchemeRequest *request = takeLinuxSchemeReq(reqHandle);
    if (!request) {
        return;
    }

    SoupMessageHeaders *respHeaders = soup_message_headers_new(SOUP_MESSAGE_HEADERS_RESPONSE);
    soup_message_headers_set_status(respHeaders, (SoupStatus)statusCode);

    if (headers && headers[0]) {
        char *blob = g_strdup(headers);
        char *line = blob;
        while (line && *line) {
            char *nl = strchr(line, '\n');
            if (nl) {
                *nl = '\0';
            }
            char *sep = strstr(line, ": ");
            if (sep) {
                *sep = '\0';
                soup_message_headers_append(respHeaders, line, sep + 2);
            }
            if (!nl) {
                break;
            }
            line = nl + 1;
        }
        g_free(blob);
    }

    if (!soup_message_headers_get_one(respHeaders, "Content-Type")) {
        soup_message_headers_append(respHeaders, "Content-Type", "application/octet-stream");
    }

    GInputStream *stream = NULL;
    if (body && bodyLen > 0) {
        GBytes *bytes = g_bytes_new(body, (gsize)bodyLen);
        stream = g_memory_input_stream_new_from_bytes(bytes);
        g_bytes_unref(bytes);
    } else {
        stream = g_memory_input_stream_new_from_data("", 0, NULL);
    }

    webkit_uri_scheme_request_finish_with_headers(request, respHeaders, stream);
    g_object_unref(stream);
    soup_message_headers_free(respHeaders);
}
*/
import "C"
import (
	"fmt"
	"io"
	"net/http"
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

func (w *linuxWebView) RegisterScheme(scheme string, handler types.SchemeHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.schemes[scheme]; ok {
		return fmt.Errorf("webview: RegisterScheme %q: already registered", scheme)
	}
	w.schemes[scheme] = handler
	if w.webView != nil {
		cs := C.CString(scheme)
		C.registerLinuxScheme((*C.WebKitWebView)(w.webView), cs, C.uintptr_t(w.handle))
		C.free(unsafe.Pointer(cs))
	}
	return nil
}

func linuxRegisterSchemes(w *linuxWebView) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.webView == nil {
		return
	}
	for scheme := range w.schemes {
		cs := C.CString(scheme)
		C.registerLinuxScheme((*C.WebKitWebView)(w.webView), cs, C.uintptr_t(w.handle))
		C.free(unsafe.Pointer(cs))
	}
}

func linuxDeliverText(reqHandle C.uintptr_t, status int, msg string) {
	headers := C.CString("Content-Type: text/plain")
	body := C.CString(msg)
	C.deliverLinuxSchemeResponse(reqHandle, C.int(status), headers, unsafe.Pointer(body), C.int(len(msg)))
	C.free(unsafe.Pointer(headers))
	C.free(unsafe.Pointer(body))
}

//export goLinuxProtocolHandler
func goLinuxProtocolHandler(handle C.uintptr_t, scheme *C.char, url *C.char, method *C.char,
	headers *C.char, body unsafe.Pointer, bodyLen C.int, reqHandle C.uintptr_t) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	lw := wv.(*linuxWebView)
	s := C.GoString(scheme)
	u := C.GoString(url)
	m := C.GoString(method)
	reqHeaders := parseHeaderBlob(C.GoString(headers))

	var reqBody []byte
	if body != nil && bodyLen > 0 {
		if int(bodyLen) > maxSchemeBodySize {
			linuxDeliverText(reqHandle, http.StatusRequestEntityTooLarge, "Request body too large")
			return
		}
		reqBody = C.GoBytes(body, bodyLen)
	}

	lw.mu.RLock()
	handler, ok := lw.schemes[s]
	lw.mu.RUnlock()
	if !ok {
		linuxDeliverText(reqHandle, http.StatusNotFound, "Not Found")
		return
	}

	lw.pending.Add(1)
	go func() {
		defer lw.pending.Done()
		defer func() {
			if r := recover(); r != nil {
				linuxDeliverText(reqHandle, http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		if lw.isTerminated() {
			linuxDeliverText(reqHandle, http.StatusServiceUnavailable, "Service Unavailable")
			return
		}

		resp := handler(&types.Request{
			Method:  m,
			URL:     u,
			Headers: reqHeaders,
			Body:    reqBody,
		})
		if resp == nil {
			resp = &types.Response{StatusCode: http.StatusNotFound}
		}

		var respBody []byte
		if resp.Body != nil {
			var err error
			respBody, err = io.ReadAll(io.LimitReader(resp.Body, maxSchemeBodySize))
			if c, ok := resp.Body.(io.Closer); ok {
				c.Close()
			}
			if err != nil {
				resp.StatusCode = http.StatusInternalServerError
				respBody = nil
			}
		}

		hdr := resp.Headers
		if hdr == nil {
			hdr = http.Header{}
		}
		if hdr.Get("Content-Type") == "" {
			hdr = hdr.Clone()
			hdr.Set("Content-Type", "application/octet-stream")
		}
		cheaders := C.CString(headerBlob(hdr))

		var cbody unsafe.Pointer
		if len(respBody) > 0 {
			cbody = C.CBytes(respBody)
		}

		C.deliverLinuxSchemeResponse(reqHandle, C.int(resp.StatusCode), cheaders, cbody, C.int(len(respBody)))

		C.free(unsafe.Pointer(cheaders))
		if cbody != nil {
			C.free(cbody)
		}
	}()
}
