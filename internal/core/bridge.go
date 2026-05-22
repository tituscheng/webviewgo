package core

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// bridgeMessage is the JSON payload posted by injected binding shims.
type bridgeMessage struct {
	Bind string          `json:"bind"`
	Args json.RawMessage `json:"args"`
	CB   string          `json:"cb"`
}

// bridgeBindings holds the registered handlers for a binding name.
type bridgeBindings struct {
	raw    func(json.RawMessage) (json.RawMessage, error)
	normal func([]any) (any, error)
}

// bridgeHost is implemented by each platform backend to route bridge responses.
type bridgeHost interface {
	lookupBinding(name string) bridgeBindings
	isTerminated() bool
	enqueueScript(script string)
	logger() *slog.Logger
}

// isValidBridgeCallbackID reports whether cb matches the id format generated
// by injected binding shims (__go_<alphanum>). Callback ids are interpolated
// into settle scripts, so untrusted values must be rejected.
func isValidBridgeCallbackID(cb string) bool {
	const prefix = "__go_"
	if !strings.HasPrefix(cb, prefix) {
		return false
	}
	suffix := cb[len(prefix):]
	if suffix == "" {
		return false
	}
	for _, r := range suffix {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}

// bindRejectScript builds a promise rejection for cb. cb must already be
// validated with isValidBridgeCallbackID.
func bindRejectScript(cb, errMsg string) string {
	es, _ := json.Marshal(errMsg)
	return fmt.Sprintf("window['%s'].reject(new Error(%s)); delete window['%s'];", cb, es, cb)
}

// dispatchBridgeMessage handles a goBridge postMessage from page JavaScript.
func dispatchBridgeMessage(host bridgeHost, msg bridgeMessage) {
	log := host.logger()

	if !isValidBridgeCallbackID(msg.CB) {
		log.Warn("invalid bridge callback id", "cb", msg.CB)
		return
	}

	bindings := host.lookupBinding(msg.Bind)
	if bindings.raw == nil && bindings.normal == nil {
		log.Warn("unknown binding", "name", msg.Bind)
		host.enqueueScript(bindRejectScript(msg.CB, fmt.Sprintf("unknown binding: %s", msg.Bind)))
		return
	}

	bindName := msg.Bind
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("binding callback panic", "name", bindName, "recover", r)
			}
		}()

		if host.isTerminated() {
			return
		}

		script := bindResponseScript(msg.CB, msg.Args, bindings.raw, bindings.normal)

		if host.isTerminated() {
			return
		}
		host.enqueueScript(script)
	}()
}

// parseBridgeMessage unmarshals a goBridge body and dispatches it. Returns
// false when the payload could not be parsed.
func parseBridgeMessage(host bridgeHost, body string) bool {
	var msg bridgeMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		host.logger().Error("failed to unmarshal bridge message", "error", err)
		return false
	}
	dispatchBridgeMessage(host, msg)
	return true
}

type platformBridgeHost struct {
	lookup   func(name string) bridgeBindings
	term     func() bool
	enqueue  func(string)
	log      *slog.Logger
}

func (h platformBridgeHost) lookupBinding(name string) bridgeBindings { return h.lookup(name) }
func (h platformBridgeHost) isTerminated() bool                       { return h.term() }
func (h platformBridgeHost) enqueueScript(script string)              { h.enqueue(script) }
func (h platformBridgeHost) logger() *slog.Logger                     { return h.log }

func newPlatformBridgeHost(
	lookup func(name string) bridgeBindings,
	term func() bool,
	enqueue func(string),
	log *slog.Logger,
) bridgeHost {
	if log == nil {
		log = slog.Default()
	}
	return platformBridgeHost{lookup: lookup, term: term, enqueue: enqueue, log: log}
}
