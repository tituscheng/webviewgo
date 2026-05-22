//go:build darwin

package core

/*
#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>
#import "protocol_darwin_delegate.h"

// Forward declaration for the response delivery function defined in protocol_darwin_delegate.m
extern void deliverSchemeResponse(uintptr_t reqHandle, int statusCode, char *headers,
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
	"strings"
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

const maxSchemeBodySize = 100 << 20 // 100 MiB

// deliverText completes a scheme task with a plain-text status response.
func deliverText(reqHandle C.uintptr_t, status int, msg string) {
	headers := C.CString("Content-Type: text/plain")
	body := C.CString(msg)
	C.deliverSchemeResponse(reqHandle, C.int(status), headers, unsafe.Pointer(body), C.int(len(msg)))
	C.free(unsafe.Pointer(headers))
	C.free(unsafe.Pointer(body))
}

// headerBlob renders an http.Header as "Key: Value\n" lines for the cgo call.
func headerBlob(h http.Header) string {
	var b strings.Builder
	for k, vals := range h {
		for _, v := range vals {
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// parseHeaderBlob parses "Key: Value\n" lines back into an http.Header.
func parseHeaderBlob(s string) http.Header {
	h := http.Header{}
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, ": "); i >= 0 {
			h.Add(line[:i], line[i+2:])
		}
	}
	return h
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

		dw.mu.RLock()
		term = dw.terminated
		dw.mu.RUnlock()
		if term {
			return
		}

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
