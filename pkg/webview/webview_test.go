package webview

import (
	"context"
	"testing"
	"time"
)

func TestNew_WithHeadless(t *testing.T) {
	wv, err := New(Options{
		Title:    "test",
		Width:    800,
		Height:   600,
		Headless: true,
		AppName:  "webviewgo-test",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	if wv == nil {
		t.Fatal("expected non-nil webview")
	}
}

func TestNewWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := NewWithContext(ctx, Options{
		Headless: true,
		AppName:  "webviewgo-test-ctx",
	})
	// Cancellation during construction is acceptable
	if err != nil {
		t.Logf("NewWithContext after cancel: %v", err)
	}
}

func TestWebView_HeadlessLifecycle(t *testing.T) {
	wv, err := New(Options{
		Title:    "lifecycle",
		Headless: true,
		AppName:  "webviewgo-test-lifecycle",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	wv.SetTitle("updated")
	wv.SetSize(100, 100, HintNone)
	wv.SetMinSize(50, 50)
	wv.SetMaxSize(200, 200)
	wv.SetFullscreen(false)
	wv.SetAlwaysOnTop(false)
	wv.Show()
	wv.Hide()

	if err := wv.Navigate("https://example.com"); err != nil {
		t.Fatalf("Navigate: %v", err)
	}
	if err := wv.LoadHTML("<h1>hi</h1>", "https://example.com"); err != nil {
		t.Fatalf("LoadHTML: %v", err)
	}
	wv.Reload()
	wv.Back()
	wv.Forward()

	if err := wv.Eval("1+1"); err != nil {
		t.Fatalf("Eval: %v", err)
	}

	if err := wv.Destroy(); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
}

func TestWebView_CookieManager(t *testing.T) {
	wv, err := New(Options{
		Headless: true,
		AppName:  "webviewgo-test-cookies",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	cm := wv.CookieManager()
	if cm == nil {
		t.Fatal("expected non-nil cookie manager")
	}

	if err := cm.SetCookie(Cookie{
		Name:    "x",
		Value:   "1",
		Domain:  "test.com",
		Path:    "/",
		Expires: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("SetCookie: %v", err)
	}

	cookies, err := cm.GetCookies("https://test.com/", "")
	if err != nil {
		t.Fatalf("GetCookies: %v", err)
	}
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
}

func TestWebView_Bind(t *testing.T) {
	wv, err := New(Options{
		Headless: true,
		AppName:  "webviewgo-test-bind",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	if err := wv.Bind("add", func(a, b int) (int, error) {
		return a + b, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	if err := wv.Bind("greet", func(name string) string {
		return "hello " + name
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
}

func TestWebView_ClipboardAndNotify(t *testing.T) {
	wv, err := New(Options{
		Headless: true,
		AppName:  "webviewgo-test-clipboard",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	if err := wv.ClipboardWriteText("hello"); err != nil {
		t.Fatalf("ClipboardWriteText: %v", err)
	}
	_, _ = wv.ClipboardReadText()
	_ = wv.Notify("t", "b")
}

func TestWebView_Dialogs(t *testing.T) {
	wv, err := New(Options{
		Headless: true,
		AppName:  "webviewgo-test-dialogs",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	_, err = wv.OpenDialog(OpenDialogOptions{Title: "Open"})
	if err == nil {
		t.Fatal("expected error for OpenDialog in headless")
	}

	_, err = wv.SaveDialog(SaveDialogOptions{Title: "Save"})
	if err == nil {
		t.Fatal("expected error for SaveDialog in headless")
	}

	_, err = wv.MessageDialog(MessageDialogOptions{Title: "Msg", Message: "hello"})
	if err == nil {
		t.Fatal("expected error for MessageDialog in headless")
	}
}

func TestBind_RejectsInvalidName(t *testing.T) {
	wv, err := New(Options{Headless: true, AppName: "webviewgo-test-bind"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer wv.Destroy()

	valid := []string{"foo", "_bar", "$baz", "fn123"}
	for _, name := range valid {
		if err := wv.Bind(name, func() error { return nil }); err != nil {
			t.Errorf("Bind(%q) unexpectedly rejected: %v", name, err)
		}
	}

	invalid := []string{"", "1foo", "a.b", "x;evil()", "win dow", "a-b"}
	for _, name := range invalid {
		if err := wv.Bind(name, func() error { return nil }); err == nil {
			t.Errorf("Bind(%q) should have been rejected", name)
		}
	}
}
