#import "webview_darwin_delegate.h"

@implementation WebViewDelegate

- (void)userContentController:(WKUserContentController *)userContentController
      didReceiveScriptMessage:(WKScriptMessage *)message {
    NSString *name = message.name;
    NSString *body = @"";
    if ([message.body isKindOfClass:[NSString class]]) {
        body = message.body;
    } else if (message.body) {
        NSError *err = nil;
        NSData *data = [NSJSONSerialization dataWithJSONObject:message.body options:0 error:&err];
        if (!err && data) {
            body = [[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
        }
    }
    char *cname = strdup([name UTF8String]);
    char *cbody = strdup([body UTF8String]);
    goWebViewMessageReceived(self.handle, cname, cbody);
    free(cname);
    free(cbody);
}

- (void)webView:(WKWebView *)webView didFinishNavigation:(WKNavigation *)navigation {
    NSString *url = webView.URL.absoluteString ?: @"";
    char *curl = strdup([url UTF8String]);
    goWebViewNavigationFinished(self.handle, curl);
    free(curl);
}

- (void)windowWillClose:(NSNotification *)notification {
    goWebViewWindowWillClose(self.handle);
}

@end
