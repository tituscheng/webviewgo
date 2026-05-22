#import <WebKit/WebKit.h>
#import <Foundation/Foundation.h>

extern void goProtocolHandler(uintptr_t handle, char *scheme, char *url, char *method,
                              char *headers, void *body, int bodyLen, uintptr_t reqHandle);

@interface SchemeHandlerDelegate : NSObject <WKURLSchemeHandler>
@property uintptr_t handle;
@property(retain) NSString *scheme;
@end

static NSMutableDictionary *schemeTaskMap = nil;
static dispatch_queue_t schemeTaskQueue = nil;

// serializeHeaders renders an HTTP header dictionary as "Key: Value\n" lines
// for transport across the cgo boundary.
static NSString *serializeHeaders(NSDictionary<NSString *, NSString *> *fields) {
    NSMutableString *out = [NSMutableString string];
    [fields enumerateKeysAndObjectsUsingBlock:^(NSString *k, NSString *v, BOOL *stop) {
        [out appendFormat:@"%@: %@\n", k, v];
    }];
    return out;
}

// parseHeaders turns "Key: Value\n" lines back into a dictionary.
static NSMutableDictionary *parseHeaders(const char *headers) {
    NSMutableDictionary *out = [NSMutableDictionary dictionary];
    if (!headers) return out;
    NSString *blob = [NSString stringWithUTF8String:headers];
    for (NSString *line in [blob componentsSeparatedByString:@"\n"]) {
        NSRange sep = [line rangeOfString:@": "];
        if (sep.location != NSNotFound) {
            NSString *key = [line substringToIndex:sep.location];
            NSString *val = [line substringFromIndex:sep.location + 2];
            if (key.length > 0) out[key] = val;
        }
    }
    return out;
}

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
    NSString *headers = serializeHeaders(req.allHTTPHeaderFields);
    NSData *bodyData = req.HTTPBody ?: [NSData data];

    static uintptr_t reqSeq = 1;
    uintptr_t reqHandle = reqSeq++;

    dispatch_sync(schemeTaskQueue, ^{
        [schemeTaskMap setObject:urlSchemeTask forKey:@(reqHandle)];
    });

    goProtocolHandler(self.handle, (char *)[self.scheme UTF8String], (char *)[url UTF8String],
                      (char *)[method UTF8String], (char *)[headers UTF8String],
                      (void *)[bodyData bytes], (int)[bodyData length], reqHandle);
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

// deliverSchemeResponse completes a custom scheme request with the given
// response. It looks up the task by reqHandle, sends the HTTP response, body,
// and finish signal, then removes the task from the map.
//
// WKURLSchemeTask methods must be invoked on the main thread, so the actual
// delivery is dispatched there. The body and headers are copied into
// Objective-C objects on the calling thread, so the caller may free the C
// buffers as soon as this function returns. A task may be stopped concurrently
// (removed from the map) between lookup and delivery; the @try/@catch guards
// against the exception WebKit raises when delivering to a stopped task.
void deliverSchemeResponse(uintptr_t reqHandle, int statusCode, char *headers,
                           void *body, int bodyLen) {
    if (!schemeTaskQueue) return;

    NSMutableDictionary *headerFields = parseHeaders(headers);
    if (!headerFields[@"Content-Type"]) {
        headerFields[@"Content-Type"] = @"application/octet-stream";
    }
    NSData *data = (body && bodyLen > 0) ? [NSData dataWithBytes:body length:bodyLen] : nil;
    NSNumber *key = @(reqHandle);
    int sc = statusCode;

    dispatch_async(dispatch_get_main_queue(), ^{
        __block id<WKURLSchemeTask> task = nil;
        dispatch_sync(schemeTaskQueue, ^{
            task = [schemeTaskMap objectForKey:key];
            if (task) {
                [schemeTaskMap removeObjectForKey:key];
            }
        });
        if (!task) {
            return;
        }
        @try {
            NSHTTPURLResponse *response = [[NSHTTPURLResponse alloc] initWithURL:task.request.URL
                                                                      statusCode:sc
                                                                     HTTPVersion:@"HTTP/1.1"
                                                                    headerFields:headerFields];
            [task didReceiveResponse:response];
            [response release];
            if (data) {
                [task didReceiveData:data];
            }
            [task didFinish];
        } @catch (NSException *ex) {
            // Task was stopped before/while delivering; nothing more to do.
        }
    });
}
