package webview

import (
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"

	"github.com/tituscheng/webviewgo/internal/types"
)

// SchemeHandler handles requests for a custom URL scheme.
type SchemeHandler = types.SchemeHandler

// Request represents a custom scheme request.
type Request = types.Request

// Response is the handler's reply.
type Response = types.Response

// FSHandler returns a SchemeHandler that serves files from an fs.FS.
// The prefix is stripped from the request path before lookup.
// fsys can be an embed.FS, os.DirFS, or any fs.FS implementation.
func FSHandler(fsys any, prefix string) SchemeHandler {
	var filesystem fs.FS
	switch v := fsys.(type) {
	case fs.FS:
		filesystem = v
	default:
		return func(req *Request) *Response {
			return &Response{StatusCode: http.StatusInternalServerError}
		}
	}

	return func(req *Request) *Response {
		u, err := url.Parse(req.URL)
		if err != nil {
			return &Response{StatusCode: http.StatusBadRequest}
		}

		p := strings.TrimPrefix(u.Path, prefix)
		p = strings.TrimPrefix(p, "/")
		if p == "" {
			p = "index.html"
		}

		f, err := filesystem.Open(p)
		if err != nil {
			return &Response{StatusCode: http.StatusNotFound}
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return &Response{StatusCode: http.StatusInternalServerError}
		}
		if info.IsDir() {
			// Try index.html inside the directory
			indexPath := path.Join(p, "index.html")
			idx, err := filesystem.Open(indexPath)
			if err != nil {
				return &Response{StatusCode: http.StatusForbidden}
			}
			defer idx.Close()
			f = idx
			p = indexPath
		}

		body, err := io.ReadAll(f)
		if err != nil {
			return &Response{StatusCode: http.StatusInternalServerError}
		}

		ct := mime.TypeByExtension(path.Ext(p))
		if ct == "" {
			ct = "application/octet-stream"
		}

		return &Response{
			StatusCode: http.StatusOK,
			Headers:    http.Header{"Content-Type": {ct}},
			Body:       io.NopCloser(strings.NewReader(string(body))),
		}
	}
}

// HTTPHandler returns a SchemeHandler that adapts a standard http.Handler.
func HTTPHandler(h http.Handler) SchemeHandler {
	return func(req *Request) *Response {
		u, err := url.Parse(req.URL)
		if err != nil {
			return &Response{StatusCode: http.StatusBadRequest}
		}

		httpReq := httptest.NewRequest(req.Method, u.String(), nil)
		if req.Body != nil {
			httpReq.Body = io.NopCloser(strings.NewReader(string(req.Body)))
			httpReq.ContentLength = int64(len(req.Body))
		}
		for k, v := range req.Headers {
			httpReq.Header[k] = v
		}

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httpReq)

		var body []byte
		if rr.Body != nil {
			body = rr.Body.Bytes()
		}

		return &Response{
			StatusCode: rr.Code,
			Headers:    rr.Header(),
			Body:       io.NopCloser(strings.NewReader(string(body))),
		}
	}
}

// NewResponse creates a Response with the given status and body.
func NewResponse(status int, body []byte) *Response {
	return &Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(string(body))),
	}
}

// MustParseURL is a helper that panics if the URL is invalid.
// Useful for constructing custom scheme URLs at init time.
func MustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(fmt.Sprintf("webview: invalid URL %q: %v", raw, err))
	}
	return u
}
