// basic demonstrates a minimal webview window.
//
//go:build ignore

package main

import (
	"log"
	"log/slog"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

func main() {
	wv, err := webview.New(webview.Options{
		Title:     "Basic Example",
		Width:     1024,
		Height:    768,
		Resizable: true,
		Devtools:  true,
		Logger:    slog.Default(),
		AppName:   "basic-example",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer wv.Destroy()

	if err := wv.Navigate("https://go.dev"); err != nil {
		log.Fatal(err)
	}

	if err := wv.Run(); err != nil {
		log.Fatal(err)
	}
}
