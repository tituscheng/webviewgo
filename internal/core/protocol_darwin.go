//go:build darwin

package core

/*
#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>
#import "protocol_darwin_delegate.h"

// Forward declaration for the response delivery function defined in protocol_darwin_delegate.m
extern void deliverSchemeResponse(uintptr_t reqHandle, int statusCode, char *headers,
                                  void *body, int bodyLen);
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

	if w.webView != nil {
		return fmt.Errorf("webview: RegisterScheme %q: on macOS custom schemes must be registered via Options.Schemes before New(); WKURLSchemeHandler cannot be installed after the webview is created", scheme)
	}
	return nil
}

// schemeNamesToC converts scheme map keys to a C string array for native
// registration at webview creation time.
func schemeNamesToC(schemes map[string]types.SchemeHandler) ([]*C.char, func()) {
	if len(schemes) == 0 {
		return nil, func() {}
	}
	names := make([]*C.char, 0, len(schemes))
	for scheme := range schemes {
		names = append(names, C.CString(scheme))
	}
	return names, func() {
		for _, p := range names {
			C.free(unsafe.Pointer(p))
		}
	}
}

// deliverText completes a scheme task with a plain-text status response.
func deliverText(reqHandle C.uintptr_t, status int, msg string) {
	headers := C.CString("Content-Type: text/plain")
	body := C.CString(msg)
	C.deliverSchemeResponse(reqHandle, C.int(status), headers, unsafe.Pointer(body), C.int(len(msg)))
	C.free(unsafe.Pointer(headers))
	C.free(unsafe.Pointer(body))
}

//export goProtocolHandler
func goProtocolHandler(handle C.uintptr_t, scheme *C.char, url *C.char, method *C.char,
	headers *C.char, body unsafe.Pointer, bodyLen C.int, reqHandle C.uintptr_t) {
	wv, ok := getPlatform(uintptr(handle))
	if !ok {
		return
	}
	dw := wv.(*darwinWebView)
	s := C.GoString(scheme)
	u := C.GoString(url)
	m := C.GoString(method)
	reqHeaders := parseHeaderBlob(C.GoString(headers))

	var reqBody []byte
	if body != nil && bodyLen > 0 {
		if int(bodyLen) > maxSchemeBodySize {
			deliverText(reqHandle, http.StatusRequestEntityTooLarge, "Request body too large")
			return
		}
		reqBody = C.GoBytes(body, bodyLen)
	}

	dw.mu.RLock()
	handler, ok := dw.schemes[s]
	dw.mu.RUnlock()
	if !ok {
		// No handler registered — send 404 so the task is cleaned up.
		deliverText(reqHandle, http.StatusNotFound, "Not Found")
		return
	}

	dw.pending.Add(1)
	go func() {
		defer dw.pending.Done()
		defer func() {
			if r := recover(); r != nil {
				// Ensure the scheme task is always completed, even on panic.
				deliverText(reqHandle, http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		dw.mu.RLock()
		term := dw.terminated
		dw.mu.RUnlock()
		if term {
			deliverText(reqHandle, http.StatusServiceUnavailable, "Service Unavailable")
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

		// Always deliver the handled response so the WKURLSchemeTask is
		// completed. deliverSchemeResponse is self-guarding: if the webview was
		// torn down (task already gone, or the main run loop stopped) it no-ops,
		// and the response data is copied synchronously before dispatch.

		// Forward all response headers (not just Content-Type), defaulting the
		// content type when the handler did not set one.
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

		C.deliverSchemeResponse(C.uintptr_t(reqHandle), C.int(resp.StatusCode), cheaders, cbody, C.int(len(respBody)))

		C.free(unsafe.Pointer(cheaders))
		if cbody != nil {
			C.free(cbody)
		}
	}()
}
