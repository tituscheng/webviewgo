//go:build integration

package core

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
)

func TestIntegration_HeadlessBindingRoundTrip(t *testing.T) {
	p, err := newHeadless(types.Options{AppName: "integration-bind"})
	if err != nil {
		t.Fatalf("newHeadless: %v", err)
	}
	hw := p.(*headlessWebView)
	defer hw.Destroy()

	if err := hw.Bind("echo", func(args []any) (any, error) {
		return "ok", nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	host := newPlatformBridgeHost(
		hw.lookupBinding,
		hw.isTerminated,
		func(script string) { _ = hw.Eval(script) },
		nil,
	)

	dispatchBridgeMessage(host, bridgeMessage{
		Bind: "echo",
		Args: json.RawMessage(`["hi"]`),
		CB:   "__go_testcb",
	})

	deadline := time.After(2 * time.Second)
	for {
		hw.mu.RLock()
		n := len(hw.evals)
		joined := strings.Join(hw.evals, "")
		hw.mu.RUnlock()
		if n > 0 && strings.Contains(joined, `window['__go_testcb'].resolve("ok")`) {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("binding response not eval'd, evals=%v", hw.evals)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestIntegration_HeadlessSchemeFromOptions(t *testing.T) {
	p, err := newHeadless(types.Options{
		AppName: "integration-scheme",
		Schemes: map[string]types.SchemeHandler{
			"app": func(req *types.Request) *types.Response {
				return &types.Response{StatusCode: http.StatusOK}
			},
		},
	})
	if err != nil {
		t.Fatalf("newHeadless: %v", err)
	}
	hw := p.(*headlessWebView)
	defer hw.Destroy()

	hw.mu.RLock()
	_, ok := hw.schemes["app"]
	hw.mu.RUnlock()
	if !ok {
		t.Fatal("expected scheme from Options.Schemes")
	}
}

func (w *headlessWebView) lookupBinding(name string) bridgeBindings {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return bridgeBindings{w.rawBindings[name], w.bindings[name]}
}

func (w *headlessWebView) isTerminated() bool {
	return w.terminated.Load()
}
