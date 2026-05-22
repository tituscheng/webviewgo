package core

import (
	"testing"

	"github.com/tituscheng/webviewgo/internal/types"
)

func TestHeadlessLifecycle(t *testing.T) {
	w, err := newHeadless(types.Options{
		Title:  "test",
		Width:  100,
		Height: 100,
	})
	if err != nil {
		t.Fatalf("newHeadless: %v", err)
	}

	w.SetTitle("updated")
	w.SetSize(200, 200, types.HintNone)
	w.SetMinSize(50, 50)
	w.SetMaxSize(300, 300)
	w.SetFullscreen(true)
	w.SetAlwaysOnTop(true)
	w.Show()
	w.Hide()

	if err := w.Navigate("https://example.com"); err != nil {
		t.Fatalf("Navigate: %v", err)
	}
	if err := w.LoadHTML("<h1>hi</h1>", "https://example.com"); err != nil {
		t.Fatalf("LoadHTML: %v", err)
	}
	w.Reload()
	w.Back()
	w.Forward()

	if err := w.Eval("1"); err != nil {
		t.Fatalf("Eval: %v", err)
	}

	if err := w.Bind("fn", func(args []any) (any, error) { return nil, nil }); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	if err := w.RegisterScheme("app", func(req *types.Request) *types.Response {
		return nil
	}); err != nil {
		t.Fatalf("RegisterScheme: %v", err)
	}

	w.Terminate()
	if err := w.Destroy(); err != nil {
		t.Fatalf("Destroy: %v", err)
	}
}

func TestHeadlessDialogErrors(t *testing.T) {
	w, err := newHeadless(types.Options{})
	if err != nil {
		t.Fatalf("newHeadless: %v", err)
	}

	_, err = w.OpenDialog(types.OpenDialogOptions{})
	if err == nil {
		t.Fatal("expected error for OpenDialog")
	}

	_, err = w.SaveDialog(types.SaveDialogOptions{})
	if err == nil {
		t.Fatal("expected error for SaveDialog")
	}

	_, err = w.MessageDialog(types.MessageDialogOptions{})
	if err == nil {
		t.Fatal("expected error for MessageDialog")
	}
}
