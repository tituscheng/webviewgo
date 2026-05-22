//go:build linux

package core

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <stdlib.h>

static char **openDialog(GtkWindow *parent, int allowFiles, int allowDirs, int allowMultiple,
                         const char *title, const char *directory, int *outCount) {
    GtkWidget *dialog = gtk_file_chooser_dialog_new(
        title ? title : "Open File",
        parent,
        allowDirs && !allowFiles ? GTK_FILE_CHOOSER_ACTION_SELECT_FOLDER : GTK_FILE_CHOOSER_ACTION_OPEN,
        "_Cancel", GTK_RESPONSE_CANCEL,
        "_Open", GTK_RESPONSE_ACCEPT,
        NULL);

    if (directory) {
        gtk_file_chooser_set_current_folder(GTK_FILE_CHOOSER(dialog), directory);
    }
    gtk_file_chooser_set_select_multiple(GTK_FILE_CHOOSER(dialog), allowMultiple);

    if (gtk_dialog_run(GTK_DIALOG(dialog)) != GTK_RESPONSE_ACCEPT) {
        gtk_widget_destroy(dialog);
        *outCount = 0;
        return NULL;
    }

    GSList *files = gtk_file_chooser_get_filenames(GTK_FILE_CHOOSER(dialog));
    *outCount = g_slist_length(files);
    char **paths = (char **)malloc(sizeof(char *) * (*outCount));
    int i = 0;
    for (GSList *l = files; l != NULL; l = l->next, i++) {
        paths[i] = (char *)l->data;
    }
    g_slist_free(files);
    gtk_widget_destroy(dialog);
    return paths;
}

static char *saveDialog(GtkWindow *parent, const char *title, const char *directory,
                        const char *defaultFile) {
    GtkWidget *dialog = gtk_file_chooser_dialog_new(
        title ? title : "Save File",
        parent,
        GTK_FILE_CHOOSER_ACTION_SAVE,
        "_Cancel", GTK_RESPONSE_CANCEL,
        "_Save", GTK_RESPONSE_ACCEPT,
        NULL);

    gtk_file_chooser_set_do_overwrite_confirmation(GTK_FILE_CHOOSER(dialog), TRUE);
    if (directory) {
        gtk_file_chooser_set_current_folder(GTK_FILE_CHOOSER(dialog), directory);
    }
    if (defaultFile) {
        gtk_file_chooser_set_current_name(GTK_FILE_CHOOSER(dialog), defaultFile);
    }

    if (gtk_dialog_run(GTK_DIALOG(dialog)) != GTK_RESPONSE_ACCEPT) {
        gtk_widget_destroy(dialog);
        return NULL;
    }

    char *path = gtk_file_chooser_get_filename(GTK_FILE_CHOOSER(dialog));
    gtk_widget_destroy(dialog);
    return path;
}

static int messageDialog(GtkWindow *parent, const char *title, const char *message,
                         int level, int buttons) {
    GtkMessageType msgType = GTK_MESSAGE_INFO;
    switch (level) {
        case 1: msgType = GTK_MESSAGE_WARNING; break;
        case 2: msgType = GTK_MESSAGE_ERROR; break;
        case 3: msgType = GTK_MESSAGE_QUESTION; break;
        default: msgType = GTK_MESSAGE_INFO; break;
    }

    GtkButtonsType btnType = GTK_BUTTONS_OK;
    switch (buttons) {
        case 1: btnType = GTK_BUTTONS_OK_CANCEL; break;
        case 2: btnType = GTK_BUTTONS_YES_NO; break;
        case 3: btnType = GTK_BUTTONS_YES_NO; break; // GTK has no native YesNoCancel
        case 4: btnType = GTK_BUTTONS_OK_CANCEL; break;
        default: btnType = GTK_BUTTONS_OK; break;
    }

    GtkWidget *dialog = gtk_message_dialog_new(
        parent,
        GTK_DIALOG_MODAL,
        msgType,
        btnType,
        "%s",
        message ? message : "");

    if (title) {
        gtk_window_set_title(GTK_WINDOW(dialog), title);
    }

    int response = gtk_dialog_run(GTK_DIALOG(dialog));
    gtk_widget_destroy(dialog);

    switch (buttons) {
        case 1: return (response == GTK_RESPONSE_OK) ? 1 : 0;
        case 2: return (response == GTK_RESPONSE_YES) ? 2 : 3;
        case 3: return (response == GTK_RESPONSE_YES) ? 2 : 3;
        case 4: return (response == GTK_RESPONSE_OK) ? 5 : 0;
        default: return (response == GTK_RESPONSE_OK) ? 1 : 0;
    }
}

static void freeCStringArray(char **arr, int count) {
    for (int i = 0; i < count; i++) {
        g_free(arr[i]);
    }
    free(arr);
}
*/
import "C"
import (
	"unsafe"

	"github.com/tituscheng/webviewgo/internal/types"
)

func (w *linuxWebView) OpenDialog(opts types.OpenDialogOptions) ([]string, error) {
	title := cStrOrNull(opts.Title)
	dir := cStrOrNull(opts.Directory)
	defer freeCStr(title)
	defer freeCStr(dir)

	var count C.int
	paths := C.openDialog(
		(*C.GtkWindow)(w.window),
		boolInt(opts.AllowFiles), boolInt(opts.AllowDirs), boolInt(opts.AllowMultiple),
		title, dir, &count,
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

func (w *linuxWebView) SaveDialog(opts types.SaveDialogOptions) (string, error) {
	title := cStrOrNull(opts.Title)
	dir := cStrOrNull(opts.Directory)
	defFile := cStrOrNull(opts.DefaultFile)
	defer freeCStr(title)
	defer freeCStr(dir)
	defer freeCStr(defFile)

	path := C.saveDialog((*C.GtkWindow)(w.window), title, dir, defFile)
	if path == nil {
		return "", nil
	}
	defer C.g_free(C.gpointer(path))
	return C.GoString(path), nil
}

func (w *linuxWebView) MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error) {
	title := cStrOrNull(opts.Title)
	msg := cStrOrNull(opts.Message)
	defer freeCStr(title)
	defer freeCStr(msg)

	res := C.messageDialog((*C.GtkWindow)(w.window), title, msg, C.int(opts.Level), C.int(opts.Buttons))
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
