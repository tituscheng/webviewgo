// js-interop demonstrates Go-to-JS binding and evaluation.
//
//go:build ignore

package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

func main() {
	wv, err := webview.New(webview.Options{
		Title:     "JS Interop Example",
		Width:     1024,
		Height:    768,
		Resizable: true,
		Devtools:  true,
		Logger:    slog.Default(),
		AppName:   "js-interop-example",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer wv.Destroy()

	// Bind a Go function callable from JS.
	if err := wv.Bind("add", func(a, b int) (int, error) {
		return a + b, nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := wv.Bind("greet", func(name string) string {
		return fmt.Sprintf("Hello, %s!", name)
	}); err != nil {
		log.Fatal(err)
	}

	if err := wv.LoadHTML(`<!doctype html>
<html>
<head><title>JS Interop</title></head>
<body>
<h1>JS ↔ Go Interop</h1>
<button id="btn">Call Go add(2, 3)</button>
<p id="result"></p>
<script>
document.getElementById('btn').onclick = async () => {
	try {
		const sum = await window.add(2, 3);
		const msg = await window.greet('WebView');
		document.getElementById('result').innerText = sum + ' ' + msg;
	} catch (e) {
		document.getElementById('result').innerText = 'Error: ' + e.message;
	}
};
</script>
</body>
</html>`, "https://example.com"); err != nil {
		log.Fatal(err)
	}

	if err := wv.Run(); err != nil {
		log.Fatal(err)
	}
}
