// cookies demonstrates persistent cookie storage and session isolation.
//
//go:build ignore

package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/tituscheng/webviewgo/pkg/webview"
)

func main() {
	wv, err := webview.New(webview.Options{
		Title:     "Cookie Example",
		Width:     1024,
		Height:    768,
		Resizable: true,
		AppName:   "webviewgo-example-cookies",
		Devtools:  true,
		Logger:    slog.Default(),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer wv.Destroy()

	cm := wv.CookieManager()

	// Set a persistent cookie (no SessionID = shared across sessions).
	if err := cm.SetCookie(webview.Cookie{
		Name:    "visit_count",
		Value:   "1",
		Domain:  "example.com",
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	}); err != nil {
		log.Fatal(err)
	}

	// Set a session-isolated cookie (tied to SessionID "user-42").
	if err := cm.SetCookie(webview.Cookie{
		SessionID: "user-42",
		Name:      "session_token",
		Value:     "abc123",
		Domain:    "example.com",
		Path:      "/",
	}); err != nil {
		log.Fatal(err)
	}

	// Bind a function so JS can query cookies from Go.
	if err := wv.Bind("getCookies", func(sessionID string) ([]cookieView, error) {
		cookies, err := cm.GetCookies("https://example.com/", sessionID)
		if err != nil {
			return nil, err
		}
		var out []cookieView
		for _, c := range cookies {
			out = append(out, cookieView{
				Name:      c.Name,
				Value:     c.Value,
				SessionID: c.SessionID,
				Expires:   c.Expires.Format(time.RFC3339),
			})
		}
		return out, nil
	}); err != nil {
		log.Fatal(err)
	}

	html := `<!doctype html>
<html>
<head><title>Cookie Example</title>
<style>
body { font-family: system-ui, sans-serif; padding: 2rem; max-width: 800px; margin: 0 auto; }
button { padding: 0.5rem 1rem; margin-right: 0.5rem; cursor: pointer; }
table { border-collapse: collapse; width: 100%; margin-top: 1rem; }
th, td { border: 1px solid #ccc; padding: 0.5rem; text-align: left; }
th { background: #f5f5f5; }
</style>
</head>
<body>
<h1>Cookie Storage Demo</h1>
<p>Click a button to query cookies from Go via the CookieManager.</p>
<button onclick="load('')">All Cookies (no session filter)</button>
<button onclick="load('user-42')">Session: user-42</button>
<button onclick="load('user-99')">Session: user-99 (none)</button>
<table id="tbl" style="display:none">
<thead><tr><th>Name</th><th>Value</th><th>SessionID</th><th>Expires</th></tr></thead>
<tbody id="body"></tbody>
</table>
<p id="empty" style="display:none;color:#666;">No cookies found.</p>
<script>
async function load(sessionID) {
	const cookies = await window.getCookies(sessionID);
	const tbl = document.getElementById('tbl');
	const body = document.getElementById('body');
	const empty = document.getElementById('empty');
	body.innerHTML = '';
	if (cookies.length === 0) {
		tbl.style.display = 'none';
		empty.style.display = 'block';
		return;
	}
	tbl.style.display = 'table';
	empty.style.display = 'none';
	for (const c of cookies) {
		const tr = document.createElement('tr');
		tr.innerHTML = '<td>' + escapeHtml(c.Name) + '</td>' +
			'<td>' + escapeHtml(c.Value) + '</td>' +
			'<td>' + escapeHtml(c.SessionID || '-') + '</td>' +
			'<td>' + escapeHtml(c.Expires || '-') + '</td>';
		body.appendChild(tr);
	}
}
function escapeHtml(s) {
	const div = document.createElement('div');
	div.textContent = s;
	return div.innerHTML;
}
</script>
</body>
</html>`

	if err := wv.LoadHTML(html, "https://example.com"); err != nil {
		log.Fatal(err)
	}

	if err := wv.Run(); err != nil {
		log.Fatal(err)
	}
}

type cookieView struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	SessionID string `json:"sessionID"`
	Expires   string `json:"expires"`
}
