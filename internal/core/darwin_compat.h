#ifndef DARWIN_COMPAT_H
#define DARWIN_COMPAT_H

// ============================================================================
// Deprecation Suppression Macros
// ============================================================================
// Apple deprecates APIs that remain functionally necessary for libraries
// (e.g. setAllowedFileTypes: supports custom extensions without Info.plist;
// NSUserNotification avoids async UserNotifications.framework complexity).
// These macros centralize the suppression so the intent is self-documenting.
//
// Usage:
//   SUPPRESS_DEPRECATED_DECLARATIONS
//   [someObject deprecatedMethod];
//   RESTORE_DEPRECATED_DECLARATIONS
// ============================================================================

#define SUPPRESS_DEPRECATED_DECLARATIONS        \
    _Pragma("clang diagnostic push")            \
    _Pragma("clang diagnostic ignored \"-Wdeprecated-declarations\"")

#define RESTORE_DEPRECATED_DECLARATIONS         \
    _Pragma("clang diagnostic pop")

#endif // DARWIN_COMPAT_H
