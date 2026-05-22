package webview

import (
	"github.com/tituscheng/webviewgo/internal/types"
)

// Options configures a new WebView instance.
type Options = types.Options

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return types.DefaultOptions()
}

// FileFilter describes a file type filter for dialogs.
type FileFilter = types.FileFilter

// Hint controls how window size is interpreted.
type Hint = types.Hint

const (
	HintNone  = types.HintNone
	HintMin   = types.HintMin
	HintMax   = types.HintMax
	HintFixed = types.HintFixed
)

// DialogResult is the outcome of a message dialog.
type DialogResult = types.DialogResult

const (
	DialogCancel = types.DialogCancel
	DialogOK     = types.DialogOK
	DialogYes    = types.DialogYes
	DialogNo     = types.DialogNo
	DialogAbort  = types.DialogAbort
	DialogRetry  = types.DialogRetry
	DialogIgnore = types.DialogIgnore
)

// DialogLevel controls the message dialog icon.
type DialogLevel = types.DialogLevel

const (
	DialogInfo     = types.DialogInfo
	DialogWarning  = types.DialogWarning
	DialogError    = types.DialogError
	DialogQuestion = types.DialogQuestion
)

// DialogButtons controls which buttons appear.
type DialogButtons = types.DialogButtons

const (
	DialogButtonsOK               = types.DialogButtonsOK
	DialogButtonsOKCancel         = types.DialogButtonsOKCancel
	DialogButtonsYesNo            = types.DialogButtonsYesNo
	DialogButtonsYesNoCancel      = types.DialogButtonsYesNoCancel
	DialogButtonsRetryCancel      = types.DialogButtonsRetryCancel
	DialogButtonsAbortRetryIgnore = types.DialogButtonsAbortRetryIgnore
)

// OpenDialogOptions configures a file-open dialog.
type OpenDialogOptions = types.OpenDialogOptions

// SaveDialogOptions configures a file-save dialog.
type SaveDialogOptions = types.SaveDialogOptions

// MessageDialogOptions configures a message dialog.
type MessageDialogOptions = types.MessageDialogOptions
