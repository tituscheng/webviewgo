package core

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Platform is the minimal surface a platform backend must implement.
type Platform interface {
	Run() error
	Terminate()
	Destroy() error

	SetTitle(title string)
	SetSize(width, height int, hint types.Hint)
	SetMinSize(width, height int)
	SetMaxSize(width, height int)
	SetFullscreen(fullscreen bool)
	SetAlwaysOnTop(alwaysOnTop bool)
	Show()
	Hide()

	Navigate(url string) error
	LoadHTML(html, baseURL string) error
	Reload()
	Back()
	Forward()

	Eval(script string) error
	Bind(name string, fn func(args []any) (any, error)) error

	// BindRaw is the high-performance path. See pkg/webview.WebView.BindRaw for docs.
	BindRaw(name string, fn func(args json.RawMessage) (json.RawMessage, error)) error

	RegisterScheme(scheme string, handler types.SchemeHandler) error

	OpenDialog(opts types.OpenDialogOptions) ([]string, error)
	SaveDialog(opts types.SaveDialogOptions) (string, error)
	MessageDialog(opts types.MessageDialogOptions) (types.DialogResult, error)

	ClipboardReadText() (string, error)
	ClipboardWriteText(text string) error
	Notify(title, body string) error
}

var (
	platforms sync.Map // map[uintptr]Platform
	handleSeq atomic.Uintptr
)

func nextHandle(p Platform) uintptr {
	h := handleSeq.Add(1)
	platforms.Store(h, p)
	return h
}

func releaseHandle(h uintptr) {
	platforms.Delete(h)
}

func getPlatform(h uintptr) (Platform, bool) {
	v, ok := platforms.Load(h)
	if !ok {
		return nil, false
	}
	return v.(Platform), true
}

// New creates the best available platform backend.
func New(opts types.Options) (Platform, error) {
	if opts.Headless {
		return newHeadless(opts)
	}
	return newNative(opts)
}

func logOpts(opts types.Options) *slog.Logger {
	if opts.Logger != nil {
		return opts.Logger
	}
	return slog.Default()
}

// bindResponseScript runs the matched binding and returns the JavaScript that
// settles the promise identified by cb. Exactly one of rawFn / normalFn is
// non-nil. The logic is platform-independent; each backend supplies the
// lookup, goroutine, and main-thread delivery around it.
func bindResponseScript(cb string, args json.RawMessage,
	rawFn func(json.RawMessage) (json.RawMessage, error),
	normalFn func([]any) (any, error)) string {

	reject := func(err error) string {
		es, _ := json.Marshal(err.Error())
		return fmt.Sprintf("window['%s'].reject(new Error(%s)); delete window['%s'];", cb, es, cb)
	}

	if rawFn != nil {
		// Raw path: caller owns (de)serialization. The returned bytes are
		// inserted directly as the resolved value, so they must be a valid JS
		// expression (any valid JSON is).
		res, err := rawFn(args)
		if err != nil {
			return reject(err)
		}
		if len(res) == 0 {
			return fmt.Sprintf("window['%s'].resolve(undefined); delete window['%s'];", cb, cb)
		}
		return fmt.Sprintf("window['%s'].resolve(%s); delete window['%s'];", cb, res, cb)
	}

	var a []any
	_ = json.Unmarshal(args, &a) // best effort; nil flows into the binding as no args
	res, err := normalFn(a)
	if err != nil {
		return reject(err)
	}
	rs, _ := json.Marshal(res)
	return fmt.Sprintf("window['%s'].resolve(%s); delete window['%s'];", cb, rs, cb)
}

// responsePump batches resolve/reject scripts into fewer main-thread script
// evaluations. Native script evaluation has non-trivial per-call cost (cgo plus
// cross-thread dispatch), so coalescing high-frequency binding responses
// meaningfully reduces overhead. Each statement is wrapped in try/catch before
// batching so that one failing settle (e.g. a callback whose page already
// navigated away) cannot abort the rest of the batch.
type responsePump struct {
	ch       chan string
	stop     chan struct{}
	deliver  func(string) // sends a script to the UI/main thread
	stopped  atomic.Bool
	stopOnce sync.Once
}

const (
	responsePumpBuffer   = 512
	responsePumpMaxBatch = 32
	responsePumpMaxBytes = 64 * 1024
)

// newResponsePump starts a pump that hands batches to deliver. deliver must be
// safe to call from a background goroutine (i.e. it dispatches to the UI
// thread).
func newResponsePump(deliver func(string)) *responsePump {
	p := &responsePump{
		ch:      make(chan string, responsePumpBuffer),
		stop:    make(chan struct{}),
		deliver: deliver,
	}
	go p.run()
	return p
}

// enqueue schedules a settle script for batched delivery. If the buffer is full
// it delivers directly to preserve correctness; if the pump is stopped it drops
// the script (the webview is going away).
func (p *responsePump) enqueue(script string) {
	if p.stopped.Load() {
		return
	}
	select {
	case p.ch <- script:
	case <-p.stop:
	default:
		p.deliver(wrapStmt(script))
	}
}

// shutdown stops the pump. Safe to call multiple times and concurrently.
func (p *responsePump) shutdown() {
	p.stopOnce.Do(func() {
		p.stopped.Store(true)
		close(p.stop)
	})
}

func (p *responsePump) run() {
	for {
		select {
		case <-p.stop:
			return
		case first := <-p.ch:
			batch := wrapStmt(first)
		drain:
			for i := 1; i < responsePumpMaxBatch && len(batch) < responsePumpMaxBytes; i++ {
				select {
				case more := <-p.ch:
					batch += "\n" + wrapStmt(more)
				default:
					break drain
				}
			}
			p.deliver(batch)
		}
	}
}

// wrapStmt isolates a settle statement so a throw in one batched script does
// not prevent the others from running.
func wrapStmt(s string) string {
	return "try{" + s + "}catch(e){}"
}
