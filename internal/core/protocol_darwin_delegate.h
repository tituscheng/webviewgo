#ifndef PROTOCOL_DARWIN_DELEGATE_H
#define PROTOCOL_DARWIN_DELEGATE_H

#import <WebKit/WebKit.h>

extern void goProtocolHandler(uintptr_t handle, char *scheme, char *url, char *method,
                              void *body, int bodyLen, uintptr_t reqHandle);

@interface SchemeHandlerDelegate : NSObject <WKURLSchemeHandler>
@property uintptr_t handle;
@property(retain) NSString *scheme;
@end

#endif // PROTOCOL_DARWIN_DELEGATE_H
