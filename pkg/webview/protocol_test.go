package webview

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestFSHandler_OSDirFS(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	h := FSHandler(os.DirFS(dir), "")
	resp := h(&Request{Method: "GET", URL: "app://x/a.txt"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(b) != "abc" {
		t.Fatalf("got %q", b)
	}
}

func TestFSHandler_PrefixBoundary(t *testing.T) {
	mapFS := fstest.MapFS{
		"secret.txt": {Data: []byte("nope")},
		"foo.txt":    {Data: []byte("ok")},
	}
	h := FSHandler(mapFS, "/assets")
	denied := h(&Request{Method: "GET", URL: "app://x/assetsfoo/secret.txt"})
	if denied.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for prefix bypass, got %d", denied.StatusCode)
	}

	mapFS2 := fstest.MapFS{"foo.txt": {Data: []byte("ok")}}
	h2 := FSHandler(mapFS2, "/assets")
	ok := h2(&Request{Method: "GET", URL: "app://x/assets/foo.txt"})
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", ok.StatusCode)
	}
	b, err := io.ReadAll(ok.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "ok" {
		t.Fatalf("got %q", b)
	}
}

func TestHTTPHandler_DefaultStatusOK(t *testing.T) {
	h := HTTPHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	resp := h(&Request{Method: "GET", URL: "app://x/"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for handler that never WriteHeader, got %d", resp.StatusCode)
	}
}
