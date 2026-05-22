#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>

extern void goProtocolHandler(uintptr_t handle, char *scheme, char *url, char *method,
                              void *body, int bodyLen, uintptr_t reqHandle);

@interface SchemeHandlerDelegate : NSObject <WKURLSchemeHandler>
@property uintptr_t handle;
@property(retain) NSString *scheme;
@end

static NSMutableDictionary *schemeTaskMap = nil;
static dispatch_queue_t schemeTaskQueue = nil;

@implementation SchemeHandlerDelegate

- (void)dealloc {
    [_scheme release];
    [super dealloc];
}

- (void)webView:(WKWebView *)webView startURLSchemeTask:(id<WKURLSchemeTask>)urlSchemeTask {
    if (!schemeTaskMap) {
        schemeTaskMap = [NSMutableDictionary dictionary];
    }
    if (!schemeTaskQueue) {
        schemeTaskQueue = dispatch_queue_create("webviewgo.schemeTask", DISPATCH_QUEUE_SERIAL);
    }

    NSURLRequest *req = urlSchemeTask.request;
    NSString *url = req.URL.absoluteString ?: @"";
    NSString *method = req.HTTPMethod ?: @"GET";
    NSData *bodyData = req.HTTPBody ?: [NSData data];

    static uintptr_t reqSeq = 1;
    uintptr_t reqHandle = reqSeq++;

    dispatch_sync(schemeTaskQueue, ^{
        [schemeTaskMap setObject:urlSchemeTask forKey:@(reqHandle)];
    });

    goProtocolHandler(self.handle, (char *)[self.scheme UTF8String], (char *)[url UTF8String],
                      (char *)[method UTF8String], (void *)[bodyData bytes], (int)[bodyData length],
                      reqHandle);
}

- (void)webView:(WKWebView *)webView stopURLSchemeTask:(id<WKURLSchemeTask>)urlSchemeTask {
    if (!schemeTaskQueue) return;
    dispatch_sync(schemeTaskQueue, ^{
        NSArray *keys = [schemeTaskMap allKeysForObject:urlSchemeTask];
        for (NSNumber *key in keys) {
            [schemeTaskMap removeObjectForKey:key];
        }
    });
}

@end

// deliverSchemeResponse completes a custom scheme request with the given response.
// It looks up the task by reqHandle, sends the HTTP response, body, and finish signal,
// then removes the task from the map.
void deliverSchemeResponse(uintptr_t reqHandle, int statusCode, char *contentType,
                           void *body, int bodyLen) {
    if (!schemeTaskQueue) return;

    __block id<WKURLSchemeTask> task = nil;
    dispatch_sync(schemeTaskQueue, ^{
        task = [schemeTaskMap objectForKey:@(reqHandle)];
        if (task) {
            [schemeTaskMap removeObjectForKey:@(reqHandle)];
        }
    });

    if (!task) {
        return;
    }

    NSString *ct = contentType ? [NSString stringWithUTF8String:contentType] : @"application/octet-stream";
    NSDictionary *headers = @{@"Content-Type": ct};

    NSHTTPURLResponse *response = [[NSHTTPURLResponse alloc] initWithURL:task.request.URL
                                                              statusCode:statusCode
                                                             HTTPVersion:@"HTTP/1.1"
                                                            headerFields:headers];
    [task didReceiveResponse:response];
    [response release];

    if (body && bodyLen > 0) {
        [task didReceiveData:[NSData dataWithBytes:body length:bodyLen]];
    }

    [task didFinish];
}
