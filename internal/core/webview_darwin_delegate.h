#ifndef WEBVIEW_DARWIN_DELEGATE_H
#define WEBVIEW_DARWIN_DELEGATE_H

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

// Go callbacks (implemented in Go) — char* not const char* to match cgo export.
extern void goWebViewMessageReceived(uintptr_t handle, char* name, char* body);
extern void goWebViewWindowWillClose(uintptr_t handle);
extern void goWebViewNavigationFinished(uintptr_t handle, char* url);

@interface WebViewDelegate : NSObject <WKScriptMessageHandler, WKNavigationDelegate, NSWindowDelegate>
@property uintptr_t handle;
@end

#endif // WEBVIEW_DARWIN_DELEGATE_H
