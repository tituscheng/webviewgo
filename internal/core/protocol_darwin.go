//go:build darwin

package core

/*
#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>
#import "protocol_darwin_delegate.h"

// Forward declaration for the response delivery function defined in protocol_darwin_delegate.m
extern void deliverSchemeResponse(uintptr_t reqHandle, int statusCode, char *contentType,
                                  void *body, int bodyLen);

static void registerScheme(void *configPtr, const char *scheme, uintptr_t handle) {
    WKWebViewConfiguration *config = (WKWebViewConfiguration *)configPtr;
    SchemeHandlerDelegate *delegate = [[SchemeHandlerDelegate alloc] init];
    delegate.handle = handle;
    delegate.scheme = [NSString stringWithUTF8String:scheme];
    [config setURLSchemeHandler:delegate forURLScheme:[NSString stringWithUTF8String:scheme]];
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

func (w *darwinWebView) RegisterScheme(scheme string, handler types.SchemeHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.schemes[scheme] = handler

	// Note: WKURLSchemeHandler must be registered before the WKWebView is created.
	// Since our webview is already created, this only works if we re-create the webview
	// or if the scheme was pre-registered. For now, return an error if the webview exists.
	if w.webView != nil {
		// TODO: support dynamic scheme registration by recreating webview or using a proxy.
		return fmt.Errorf("core: schemes must be registered before navigation on darwin")
	}
	return nil
}

const maxSchemeResponseBody = 100 << 20 // 100 MiB

//export goProtocolHandler
func goProtocolHandler(handle C.uintptr_t, scheme *C.char, url *C.char, method *C.char,
	body unsafe.Pointer, bodyLen C.int, reqHandle C.uintptr_t) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	dw := wv.(*darwinWebView)
	s := C.GoString(scheme)
	u := C.GoString(url)
	m := C.GoString(method)

	var reqBody []byte
	if body != nil && bodyLen > 0 {
		if int(bodyLen) > maxSchemeResponseBody {
			// Body too large — send 413 and clean up.
			cct := C.CString("text/plain")
			msg := C.CString("Request body too large")
			C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(http.StatusRequestEntityTooLarge), cct, unsafe.Pointer(msg), C.int(len("Request body too large")))
			C.free(unsafe.Pointer(cct))
			C.free(unsafe.Pointer(msg))
			return
		}
		reqBody = C.GoBytes(body, bodyLen)
	}

	dw.mu.RLock()
	handler, ok := dw.schemes[s]
	dw.mu.RUnlock()
	if !ok {
		// No handler registered — send 404 so the task is cleaned up.
		cct := C.CString("text/plain")
		msg := C.CString("Not Found")
		C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(http.StatusNotFound), cct, unsafe.Pointer(msg), C.int(len("Not Found")))
		C.free(unsafe.Pointer(cct))
		C.free(unsafe.Pointer(msg))
		return
	}

	dw.pending.Add(1)
	go func() {
		defer dw.pending.Done()
		defer func() {
			if r := recover(); r != nil {
				// Ensure the scheme task is always completed, even on panic.
				cct := C.CString("text/plain")
				msg := C.CString("Internal Server Error")
				C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(http.StatusInternalServerError), cct, unsafe.Pointer(msg), C.int(len("Internal Server Error")))
				C.free(unsafe.Pointer(cct))
				C.free(unsafe.Pointer(msg))
			}
		}()

		dw.mu.RLock()
		term := dw.terminated
		dw.mu.RUnlock()
		if term {
			cct := C.CString("text/plain")
			msg := C.CString("Service Unavailable")
			C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(http.StatusServiceUnavailable), cct, unsafe.Pointer(msg), C.int(len("Service Unavailable")))
			C.free(unsafe.Pointer(cct))
			C.free(unsafe.Pointer(msg))
			return
		}

		resp := handler(&types.Request{
			Method: m,
			URL:    u,
			Body:   reqBody,
		})
		if resp == nil {
			resp = &types.Response{StatusCode: http.StatusNotFound}
		}

		var respBody []byte
		if resp.Body != nil {
			var err error
			respBody, err = io.ReadAll(io.LimitReader(resp.Body, maxSchemeResponseBody))
			if c, ok := resp.Body.(io.Closer); ok {
				c.Close()
			}
			if err != nil {
				resp.StatusCode = http.StatusInternalServerError
				respBody = nil
			}
		}

		dw.mu.RLock()
		term = dw.terminated
		dw.mu.RUnlock()
		if term {
			return
		}

		ct := resp.Headers.Get("Content-Type")
		if ct == "" {
			ct = "application/octet-stream"
		}

		var cct *C.char
		if ct != "" {
			cct = C.CString(ct)
		}

		var cbody unsafe.Pointer
		if len(respBody) > 0 {
			cbody = C.CBytes(respBody)
		}

		C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(resp.StatusCode), cct, cbody, C.int(len(respBody)))

		if cct != nil {
			C.free(unsafe.Pointer(cct))
		}
		if cbody != nil {
			C.free(cbody)
		}
	}()
}
