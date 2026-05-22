package types

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Title != "WebView" {
		t.Errorf("expected default title 'WebView', got %q", opts.Title)
	}
	if opts.Width != 1024 {
		t.Errorf("expected default width 1024, got %d", opts.Width)
	}
	if opts.Height != 768 {
		t.Errorf("expected default height 768, got %d", opts.Height)
	}
	if !opts.Resizable {
		t.Error("expected Resizable to be true by default")
	}
	if opts.Profile != "default" {
		t.Errorf("expected default profile 'default', got %q", opts.Profile)
	}
	if opts.Logger == nil {
		t.Error("expected default logger to be non-nil")
	}
}

func TestDialogResultConstants(t *testing.T) {
	if DialogCancel != 0 {
		t.Errorf("DialogCancel expected 0, got %d", DialogCancel)
	}
	if DialogOK != 1 {
		t.Errorf("DialogOK expected 1, got %d", DialogOK)
	}
	if DialogYes != 2 {
		t.Errorf("DialogYes expected 2, got %d", DialogYes)
	}
	if DialogNo != 3 {
		t.Errorf("DialogNo expected 3, got %d", DialogNo)
	}
}

func TestDialogLevelConstants(t *testing.T) {
	if DialogInfo != 0 {
		t.Errorf("DialogInfo expected 0, got %d", DialogInfo)
	}
	if DialogWarning != 1 {
		t.Errorf("DialogWarning expected 1, got %d", DialogWarning)
	}
	if DialogError != 2 {
		t.Errorf("DialogError expected 2, got %d", DialogError)
	}
}

func TestDialogButtonsConstants(t *testing.T) {
	if DialogButtonsOK != 0 {
		t.Errorf("DialogButtonsOK expected 0, got %d", DialogButtonsOK)
	}
	if DialogButtonsOKCancel != 1 {
		t.Errorf("DialogButtonsOKCancel expected 1, got %d", DialogButtonsOKCancel)
	}
	if DialogButtonsYesNo != 2 {
		t.Errorf("DialogButtonsYesNo expected 2, got %d", DialogButtonsYesNo)
	}
}

func TestHintConstants(t *testing.T) {
	if HintNone != 0 {
		t.Errorf("HintNone expected 0, got %d", HintNone)
	}
	if HintMin != 1 {
		t.Errorf("HintMin expected 1, got %d", HintMin)
	}
	if HintMax != 2 {
		t.Errorf("HintMax expected 2, got %d", HintMax)
	}
	if HintFixed != 3 {
		t.Errorf("HintFixed expected 3, got %d", HintFixed)
	}
}

func TestSameSiteConstants(t *testing.T) {
	if SameSiteNone != 0 {
		t.Errorf("SameSiteNone expected 0, got %d", SameSiteNone)
	}
	if SameSiteLax != 1 {
		t.Errorf("SameSiteLax expected 1, got %d", SameSiteLax)
	}
	if SameSiteStrict != 2 {
		t.Errorf("SameSiteStrict expected 2, got %d", SameSiteStrict)
	}
}

func TestCookieStruct(t *testing.T) {
	c := Cookie{
		SessionID: "sess-1",
		Name:      "token",
		Value:     "abc",
		Domain:    "example.com",
		Path:      "/",
		Secure:    true,
		HTTPOnly:  true,
		SameSite:  SameSiteLax,
		Raw:       "token=abc",
	}
	if c.Name != "token" {
		t.Errorf("expected name 'token', got %q", c.Name)
	}
	if c.SameSite != SameSiteLax {
		t.Errorf("expected SameSiteLax, got %d", c.SameSite)
	}
}

func TestRequestResponseTypes(t *testing.T) {
	req := Request{
		Method: "GET",
		URL:    "app://index.html",
		Body:   []byte("test"),
	}
	if req.Method != "GET" {
		t.Errorf("expected method GET, got %q", req.Method)
	}

	resp := Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/html"}},
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.Headers.Get("Content-Type") != "text/html" {
		t.Errorf("expected Content-Type text/html, got %q", resp.Headers.Get("Content-Type"))
	}
}
