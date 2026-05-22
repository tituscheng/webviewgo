package core

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// countStmts returns how many wrapped settle statements appear across all
// delivered batches.
func countStmts(batches []string) int {
	n := 0
	for _, b := range batches {
		n += strings.Count(b, "try{")
	}
	return n
}

func TestResponsePump_DeliversAllWrapped(t *testing.T) {
	var mu sync.Mutex
	var got []string
	deliver := func(s string) {
		mu.Lock()
		got = append(got, s)
		mu.Unlock()
	}

	p := newResponsePump(deliver)
	defer p.shutdown()

	const n = 50
	for i := 0; i < n; i++ {
		p.enqueue("call();")
	}

	deadline := time.After(2 * time.Second)
	for {
		mu.Lock()
		total := countStmts(got)
		sample := strings.Join(got, "")
		mu.Unlock()
		if total == n {
			if !strings.Contains(sample, "try{call();}catch(e){}") {
				t.Fatalf("delivered scripts not wrapped in try/catch: %q", sample)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("expected %d statements delivered, got %d", n, total)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestResponsePump_DropsAfterShutdown(t *testing.T) {
	var mu sync.Mutex
	delivered := 0
	p := newResponsePump(func(string) {
		mu.Lock()
		delivered++
		mu.Unlock()
	})
	p.shutdown()
	p.shutdown() // idempotent, must not panic

	for i := 0; i < 10; i++ {
		p.enqueue("call();")
	}
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if delivered != 0 {
		t.Fatalf("expected no delivery after shutdown, got %d", delivered)
	}
}

func TestBindResponseScript_Raw(t *testing.T) {
	raw := func(args json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`{"ok":true}`), nil
	}
	got := bindResponseScript("cb1", json.RawMessage(`[]`), raw, nil)
	if !strings.Contains(got, `window['cb1'].resolve({"ok":true});`) {
		t.Fatalf("raw resolve not inserted verbatim: %q", got)
	}

	empty := func(args json.RawMessage) (json.RawMessage, error) { return nil, nil }
	got = bindResponseScript("cb2", nil, empty, nil)
	if !strings.Contains(got, "resolve(undefined)") {
		t.Fatalf("empty raw result should resolve undefined: %q", got)
	}

	fail := func(args json.RawMessage) (json.RawMessage, error) { return nil, errors.New("boom") }
	got = bindResponseScript("cb3", nil, fail, nil)
	if !strings.Contains(got, `reject(new Error("boom"))`) {
		t.Fatalf("raw error should reject: %q", got)
	}
}

// TestResponsePump_Batches proves the amortization: when many responses arrive
// while a delivery is in flight, they coalesce into far fewer deliver calls
// than statements. The first deliver blocks until a burst is enqueued, so the
// follow-up delivery is guaranteed to carry a batch.
func TestResponsePump_Batches(t *testing.T) {
	release := make(chan struct{})
	var mu sync.Mutex
	var batches []string
	first := true

	p := newResponsePump(func(s string) {
		mu.Lock()
		batches = append(batches, s)
		blockFirst := first
		first = false
		mu.Unlock()
		if blockFirst {
			<-release // hold the pump so the burst accumulates in the buffer
		}
	})
	defer p.shutdown()

	p.enqueue("a();") // triggers the first (blocking) deliver
	// Give the pump a moment to enter the blocking deliver, then burst.
	time.Sleep(20 * time.Millisecond)
	const burst = 20
	for i := 0; i < burst; i++ {
		p.enqueue("b();")
	}
	close(release)

	deadline := time.After(2 * time.Second)
	for {
		mu.Lock()
		total := countStmts(batches)
		count := len(batches)
		mu.Unlock()
		if total == burst+1 {
			if count > burst {
				t.Fatalf("expected batching (deliver calls %d <= statements %d)", count, burst+1)
			}
			return
		}
		select {
		case <-deadline:
			t.Fatalf("only %d/%d statements delivered", total, burst+1)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestBindResponseScript_Normal(t *testing.T) {
	normal := func(args []any) (any, error) {
		return len(args), nil
	}
	got := bindResponseScript("cb", json.RawMessage(`[1,2,3]`), nil, normal)
	if !strings.Contains(got, "window['cb'].resolve(3);") {
		t.Fatalf("normal resolve wrong: %q", got)
	}

	fail := func(args []any) (any, error) { return nil, errors.New("nope") }
	got = bindResponseScript("cb", json.RawMessage(`[]`), nil, fail)
	if !strings.Contains(got, `reject(new Error("nope"))`) {
		t.Fatalf("normal error should reject: %q", got)
	}
}

func BenchmarkResponsePumpEnqueue(b *testing.B) {
	p := newResponsePump(func(string) {})
	defer p.shutdown()
	script := "window['cb'].resolve(1); delete window['cb'];"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.enqueue(script)
	}
}
