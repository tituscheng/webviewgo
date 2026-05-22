package compat

import (
	"testing"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

func TestCompatAliases(t *testing.T) {
	// Verify that type aliases and constants are correctly exported.
	var _ webview.WebView = (WebView)(nil)
	var _ Hint = HintNone
	_ = webview.DialogCancel
	_ = webview.DialogButtonsOK
}
