//go:build darwin

package core

/*
#import <Cocoa/Cocoa.h>
#import "darwin_compat.h"

static char **openDialog(int allowFiles, int allowDirs, int allowMultiple,
                         const char *title, const char *directory,
                         const char *defaultFile,
                         char **filterExts, int filterCount,
                         int *outCount) {
    NSOpenPanel *panel = [NSOpenPanel openPanel];
    if (title) [panel setTitle:[NSString stringWithUTF8String:title]];
    if (directory) [panel setDirectoryURL:[NSURL fileURLWithPath:[NSString stringWithUTF8String:directory]]];
    if (defaultFile) [panel setNameFieldStringValue:[NSString stringWithUTF8String:defaultFile]];
    [panel setCanChooseFiles:allowFiles];
    [panel setCanChooseDirectories:allowDirs];
    [panel setAllowsMultipleSelection:allowMultiple];

    if (filterCount > 0 && filterExts) {
        NSMutableArray *types = [NSMutableArray array];
        for (int i = 0; i < filterCount; i++) {
            if (filterExts[i]) {
                NSString *ext = [NSString stringWithUTF8String:filterExts[i]];
                // Strip leading dot if present
                if ([ext hasPrefix:@"."]) {
                    ext = [ext substringFromIndex:1];
                }
                [types addObject:ext];
            }
        }
        if ([types count] > 0) {
            SUPPRESS_DEPRECATED_DECLARATIONS
            [panel setAllowedFileTypes:types];
            RESTORE_DEPRECATED_DECLARATIONS
        }
    }

    NSInteger result = [panel runModal];
    if (result != NSModalResponseOK) {
        *outCount = 0;
        return NULL;
    }

    NSArray *urls = [panel URLs];
    *outCount = (int)[urls count];
    char **paths = (char **)malloc(sizeof(char *) * (*outCount));
    for (int i = 0; i < *outCount; i++) {
        paths[i] = strdup([[[urls objectAtIndex:i] path] UTF8String]);
    }
    return paths;
}

static char *saveDialog(const char *title, const char *directory,
                        const char *defaultFile,
                        char **filterExts, int filterCount) {
    NSSavePanel *panel = [NSSavePanel savePanel];
    if (title) [panel setTitle:[NSString stringWithUTF8String:title]];
    if (directory) [panel setDirectoryURL:[NSURL fileURLWithPath:[NSString stringWithUTF8String:directory]]];
    if (defaultFile) [panel setNameFieldStringValue:[NSString stringWithUTF8String:defaultFile]];

    if (filterCount > 0 && filterExts) {
        NSMutableArray *types = [NSMutableArray array];
        for (int i = 0; i < filterCount; i++) {
            if (filterExts[i]) {
                NSString *ext = [NSString stringWithUTF8String:filterExts[i]];
                if ([ext hasPrefix:@"."]) {
                    ext = [ext substringFromIndex:1];
                }
                [types addObject:ext];
            }
        }
        if ([types count] > 0) {
            SUPPRESS_DEPRECATED_DECLARATIONS
            [panel setAllowedFileTypes:types];
            RESTORE_DEPRECATED_DECLARATIONS
        }
    }

    NSInteger result = [panel runModal];
    if (result != NSModalResponseOK) {
        return NULL;
    }
    return strdup([[[panel URL] path] UTF8String]);
}

static int messageDialog(const char *title, const char *message,
                         int level, int buttons) {
    NSAlert *alert = [[NSAlert alloc] init];
    if (title) [alert setMessageText:[NSString stringWithUTF8String:title]];
    if (message) [alert setInformativeText:[NSString stringWithUTF8String:message]];

    switch (level) {
        case 1: [alert setAlertStyle:NSAlertStyleWarning]; break;
        case 2: [alert setAlertStyle:NSAlertStyleCritical]; break;
        default: [alert setAlertStyle:NSAlertStyleInformational]; break;
    }

    switch (buttons) {
        case 1: // OKCancel
            [alert addButtonWithTitle:@"OK"];
            [alert addButtonWithTitle:@"Cancel"];
            break;
        case 2: // YesNo
            [alert addButtonWithTitle:@"Yes"];
            [alert addButtonWithTitle:@"No"];
            break;
        case 3: // YesNoCancel
            [alert addButtonWithTitle:@"Yes"];
            [alert addButtonWithTitle:@"No"];
            [alert addButtonWithTitle:@"Cancel"];
            break;
        case 4: // RetryCancel
            [alert addButtonWithTitle:@"Retry"];
            [alert addButtonWithTitle:@"Cancel"];
            break;
        case 5: // AbortRetryIgnore
            [alert addButtonWithTitle:@"Abort"];
            [alert addButtonWithTitle:@"Retry"];
            [alert addButtonWithTitle:@"Ignore"];
            break;
        default: // OK
            [alert addButtonWithTitle:@"OK"];
            break;
    }

    NSInteger result = [alert runModal];
    switch (buttons) {
        case 1: return (result == NSAlertFirstButtonReturn) ? 1 : 0; // OKCancel
        case 2: return (result == NSAlertFirstButtonReturn) ? 2 : 3; // YesNo
        case 3:
            if (result == NSAlertFirstButtonReturn) return 2; // Yes
            if (result == NSAlertSecondButtonReturn) return 3; // No
            return 0; // Cancel
        case 4: return (result == NSAlertFirstButtonReturn) ? 5 : 0; // RetryCancel
        case 5:
            if (result == NSAlertFirstButtonReturn) return 4; // Abort
            if (result == NSAlertSecondButtonReturn) return 5; // Retry
            return 6; // Ignore
        default: return (result == NSAlertFirstButtonReturn) ? 1 : 0;
    }
}

static void freeCStringArray(char **arr, int count) {
    for (int i = 0; i < count; i++) free(arr[i]);
    free(arr);
}
*/
import "C"
import (
	"strings"
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

func (w *darwinWebView) OpenDialog(opts types.OpenDialogOptions) ([]string, error) {
	title := cStrOrNull(opts.Title)
	dir := cStrOrNull(opts.Directory)
	defFile := cStrOrNull(opts.DefaultFile)
	defer freeCStr(title)
	defer freeCStr(dir)
	defer freeCStr(defFile)

	var count C.int
	filterExts, filterCount := filtersToC(opts.Filters)
	if filterCount > 0 {
		defer C.freeCStringArray(filterExts, C.int(filterCount))
	}
	paths := C.openDialog(
		boolInt(opts.AllowFiles), boolInt(opts.AllowDirs), boolInt(opts.AllowMultiple),
		title, dir, defFile, filterExts, C.int(filterCount), &count,
	)
	if paths == nil {
		return nil, nil
	}
	defer C.freeCStringArray(paths, count)

	var out []string
	slice := unsafe.Slice(paths, int(count))
	for _, p := range slice {
		out = append(out, C.GoString(p))
	}
	return out, nil
}

func (w *darwinWebView) SaveDialog(opts types.SaveDialogOptions) (string, error) {
	title := cStrOrNull(opts.Title)
	dir := cStrOrNull(opts.Directory)
	defFile := cStrOrNull(opts.DefaultFile)
	defer freeCStr(title)
	defer freeCStr(dir)
	defer freeCStr(defFile)

	filterExts, filterCount := filtersToC(opts.Filters)
	if filterCount > 0 {
		defer C.freeCStringArray(filterExts, C.int(filterCount))
	}
	path := C.saveDialog(title, dir, defFile, filterExts, C.int(filterCount))
	if path == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(path))
	return C.GoString(path), nil
}

func filtersToC(filters []types.FileFilter) (**C.char, int) {
	if len(filters) == 0 {
		return nil, 0
	}
	// Collect all extensions from all filters
	var exts []string
	for _, f := range filters {
		// Pattern may be "*.txt" or "*.txt;*.md" or just "txt"
		parts := strings.Split(f.Pattern, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			p = strings.TrimPrefix(p, "*.")
			if p != "" && p != "*" {
				exts = append(exts, p)
			}
		}
	}
	if len(exts) == 0 {
		return nil, 0
	}
	arr := (**C.char)(C.malloc(C.size_t(len(exts)) * C.size_t(unsafe.Sizeof(uintptr(0)))))
	slice := unsafe.Slice(arr, len(exts))
	for i, e := range exts {
		slice[i] = C.CString(e)
	}
	return arr, len(exts)
}

func (w *darwinWebView) MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error) {
	title := cStrOrNull(opts.Title)
	msg := cStrOrNull(opts.Message)
	defer freeCStr(title)
	defer freeCStr(msg)

	res := C.messageDialog(title, msg, C.int(opts.Level), C.int(opts.Buttons))
	return types.DialogResult(res), nil
}

func cStrOrNull(s string) *C.char {
	if s == "" {
		return nil
	}
	return C.CString(s)
}

func freeCStr(p *C.char) {
	if p != nil {
		C.free(unsafe.Pointer(p))
	}
}
