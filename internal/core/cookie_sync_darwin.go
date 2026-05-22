//go:build darwin

package core

/*
#import <WebKit/WebKit.h>
#import <dispatch/dispatch.h>

// waitForCookieStore blocks until sem is signaled. WKHTTPCookieStore delivers
// its completion handlers on the main thread, so if we are called on the main
// thread (e.g. SetCookie before Run starts the app run loop) we must pump the
// run loop while waiting — otherwise the handler can never run and we
// deadlock. Off the main thread, a plain wait is correct: the main run loop
// (driven by [NSApp run]) delivers the handler.
static void waitForCookieStore(dispatch_semaphore_t sem) {
    if ([NSThread isMainThread]) {
        while (dispatch_semaphore_wait(sem, DISPATCH_TIME_NOW) != 0) {
            [[NSRunLoop currentRunLoop] runMode:NSDefaultRunLoopMode
                                     beforeDate:[NSDate dateWithTimeIntervalSinceNow:0.01]];
        }
    } else {
        dispatch_semaphore_wait(sem, DISPATCH_TIME_FOREVER);
    }
}

static void syncCookie(void *webViewPtr, const char *name, const char *value,
                       const char *domain, const char *path, double expires,
                       int secure, int httpOnly, int sameSite) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    WKHTTPCookieStore *store = webView.configuration.websiteDataStore.httpCookieStore;

    NSMutableDictionary *props = [NSMutableDictionary dictionary];
    props[NSHTTPCookieName] = [NSString stringWithUTF8String:name];
    props[NSHTTPCookieValue] = [NSString stringWithUTF8String:value];
    props[NSHTTPCookieDomain] = [NSString stringWithUTF8String:domain];
    props[NSHTTPCookiePath] = [NSString stringWithUTF8String:path];
    if (expires > 0) {
        props[NSHTTPCookieExpires] = [NSDate dateWithTimeIntervalSince1970:expires];
    }
    if (secure) {
        props[NSHTTPCookieSecure] = @"TRUE";
    }
    if (httpOnly) {
        // NSHTTPCookie doesn't have a direct HTTPOnly property in the dictionary,
        // but WKHTTPCookieStore handles it via the properties dict on modern iOS/macOS.
        // The key is NSHTTPCookieVersion = @"1" for HTTPOnly support in some contexts.
        // For WKWebView, HTTPOnly is typically handled by the cookie store itself.
    }
    if (sameSite == 1) {
        props[NSHTTPCookieSameSitePolicy] = @"Lax";
    } else if (sameSite == 2) {
        props[NSHTTPCookieSameSitePolicy] = @"Strict";
    } else {
        props[NSHTTPCookieSameSitePolicy] = @"None";
    }

    NSHTTPCookie *cookie = [NSHTTPCookie cookieWithProperties:props];
    if (!cookie) return;

    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    [store setCookie:cookie completionHandler:^{
        dispatch_semaphore_signal(sem);
    }];
    waitForCookieStore(sem);
    dispatch_release(sem);
}

static void clearAllCookies(void *webViewPtr) {
    WKWebView *webView = (WKWebView *)webViewPtr;
    WKHTTPCookieStore *store = webView.configuration.websiteDataStore.httpCookieStore;

    // Fetch first, then delete at the top level. Deleting inside the
    // getAllCookies handler would require re-entrant run-loop pumping (see
    // waitForCookieStore), so we hoist the snapshot out instead.
    __block NSArray<NSHTTPCookie *> *snapshot = nil;
    dispatch_semaphore_t sem = dispatch_semaphore_create(0);
    [store getAllCookies:^(NSArray<NSHTTPCookie *> *cookies) {
        snapshot = [cookies copy];
        dispatch_semaphore_signal(sem);
    }];
    waitForCookieStore(sem);
    dispatch_release(sem);

    for (NSHTTPCookie *c in snapshot) {
        dispatch_semaphore_t inner = dispatch_semaphore_create(0);
        [store deleteCookie:c completionHandler:^{
            dispatch_semaphore_signal(inner);
        }];
        waitForCookieStore(inner);
        dispatch_release(inner);
    }
    [snapshot release];
}
*/
import "C"
import (
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

func (w *darwinWebView) SyncCookiesToNative(cookies []types.Cookie) error {
	C.clearAllCookies(w.webView)
	for _, c := range cookies {
		name := C.CString(c.Name)
		value := C.CString(c.Value)
		domain := C.CString(c.Domain)
		path := C.CString(c.Path)
		var expires C.double
		if !c.Expires.IsZero() {
			expires = C.double(c.Expires.Unix())
		}
		C.syncCookie(w.webView, name, value, domain, path, expires, boolInt(c.Secure), boolInt(c.HTTPOnly), C.int(c.SameSite))
		C.free(unsafe.Pointer(name))
		C.free(unsafe.Pointer(value))
		C.free(unsafe.Pointer(domain))
		C.free(unsafe.Pointer(path))
	}
	return nil
}
