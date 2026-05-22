package core

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIsValidBridgeCallbackID(t *testing.T) {
	valid := []string{"__go_abc123", "__go_x", "__go_0z9"}
	for _, cb := range valid {
		if !isValidBridgeCallbackID(cb) {
			t.Errorf("expected valid callback id %q", cb)
		}
	}

	invalid := []string{
		"",
		"__go_",
		"cb1",
		"__go_evil();alert(1)",
		"__go_x/y",
		"__GO_abc",
	}
	for _, cb := range invalid {
		if isValidBridgeCallbackID(cb) {
			t.Errorf("expected invalid callback id %q", cb)
		}
	}
}

func TestBindRejectScript(t *testing.T) {
	got := bindRejectScript("__go_abc", "boom")
	if !strings.Contains(got, "window['__go_abc'].reject(new Error(\"boom\"))") {
		t.Fatalf("unexpected reject script: %q", got)
	}
}

func TestDispatchBridgeMessage_UnknownBindingRejects(t *testing.T) {
	var mu sync.Mutex
	var scripts []string
	host := newPlatformBridgeHost(
		func(string) bridgeBindings { return bridgeBindings{} },
		func() bool { return false },
		func(s string) {
			mu.Lock()
			scripts = append(scripts, s)
			mu.Unlock()
		},
		nil,
	)

	dispatchBridgeMessage(host, bridgeMessage{
		Bind: "missing",
		Args: json.RawMessage(`[]`),
		CB:   "__go_cb1",
	})

	mu.Lock()
	defer mu.Unlock()
	if len(scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(scripts))
	}
	if !strings.Contains(scripts[0], `unknown binding: missing`) {
		t.Fatalf("expected unknown binding reject, got %q", scripts[0])
	}
}

func TestDispatchBridgeMessage_InvalidCallbackIgnored(t *testing.T) {
	var mu sync.Mutex
	delivered := 0
	host := newPlatformBridgeHost(
		func(string) bridgeBindings {
			return bridgeBindings{normal: func([]any) (any, error) { return 1, nil }}
		},
		func() bool { return false },
		func(string) {
			mu.Lock()
			delivered++
			mu.Unlock()
		},
		nil,
	)

	dispatchBridgeMessage(host, bridgeMessage{
		Bind: "fn",
		Args: json.RawMessage(`[]`),
		CB:   "evil();",
	})

	mu.Lock()
	defer mu.Unlock()
	if delivered != 0 {
		t.Fatalf("expected no delivery for invalid cb, got %d", delivered)
	}
}

func TestDispatchBridgeMessage_SkipsWhenTerminated(t *testing.T) {
	var mu sync.Mutex
	delivered := 0
	term := true
	host := newPlatformBridgeHost(
		func(string) bridgeBindings {
			return bridgeBindings{
				normal: func([]any) (any, error) {
					return 1, nil
				},
			}
		},
		func() bool { return term },
		func(string) {
			mu.Lock()
			delivered++
			mu.Unlock()
		},
		nil,
	)

	dispatchBridgeMessage(host, bridgeMessage{
		Bind: "fn",
		Args: json.RawMessage(`[]`),
		CB:   "__go_cb1",
	})

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if delivered != 0 {
		t.Fatalf("expected no delivery when terminated, got %d", delivered)
	}
}
