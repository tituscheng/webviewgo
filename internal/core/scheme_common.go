package core

import (
	"net/http"
	"strings"
)

const maxSchemeBodySize = 100 << 20 // 100 MiB

// headerBlob renders an http.Header as "Key: Value\n" lines for the cgo call.
func headerBlob(h http.Header) string {
	var b strings.Builder
	for k, vals := range h {
		for _, v := range vals {
			b.WriteString(k)
			b.WriteString(": ")
			b.WriteString(v)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// parseHeaderBlob parses "Key: Value\n" lines back into an http.Header.
func parseHeaderBlob(s string) http.Header {
	h := http.Header{}
	for _, line := range strings.Split(s, "\n") {
		if i := strings.Index(line, ": "); i >= 0 {
			h.Add(line[:i], line[i+2:])
		}
	}
	return h
}
