// custom-protocol demonstrates a custom URL scheme handler.
//
//go:build ignore

package main

import (
	"log"
	"log/slog"
	"strings"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

func main() {
	wv, err := webview.New(webview.Options{
		Title:     "Custom Protocol Example",
		Width:     1024,
		Height:    768,
		Resizable: true,
		Devtools:  true,
		Logger:    slog.Default(),
		AppName:   "custom-protocol-example",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer wv.Destroy()

	if err := wv.RegisterScheme("app", func(req *webview.Request) *webview.Response {
		return &webview.Response{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"text/html"}},
			Body:       strings.NewReader("<h1>Hello from custom protocol!</h1>"),
		}
	}); err != nil {
		log.Printf("RegisterScheme: %v (expected on some platforms)", err)
	}

	if err := wv.Navigate("app://index.html"); err != nil {
		log.Fatal(err)
	}

	if err := wv.Run(); err != nil {
		log.Fatal(err)
	}
}

