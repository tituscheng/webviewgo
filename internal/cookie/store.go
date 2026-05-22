package cookie

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
	"modernc.org/sqlite"
	_ "modernc.org/sqlite" // driver
)

//go:embed schema/*.sql
var schemaFS embed.FS

// Store is a SQLite-backed cookie store with session isolation.
type Store struct {
	db   *sql.DB
	mu   sync.RWMutex
	path string
}

// Open opens or creates the cookie store at the given path.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, fmt.Errorf("cookie: open database: %w", err)
	}
	db.SetMaxOpenConns(1) // WAL mode recommended with single writer
	db.SetConnMaxLifetime(0)

	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	entries, err := schemaFS.ReadDir("schema")
	if err != nil {
		return fmt.Errorf("cookie: read schema dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		data, err := schemaFS.ReadFile("schema/" + entry.Name())
		if err != nil {
			return fmt.Errorf("cookie: read schema %s: %w", entry.Name(), err)
		}
		if _, err := s.db.Exec(string(data)); err != nil {
			// Idempotent ADD COLUMN migrations re-run on every Open; ignore the
			// duplicate-column error they raise once the column already exists.
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("cookie: apply schema %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// SetCookie inserts or replaces a cookie.
func (s *Store) SetCookie(ctx context.Context, c types.Cookie) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expires any
	if !c.Expires.IsZero() {
		expires = c.Expires.Unix()
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cookies (session_id, name, value, domain, path, expires, secure, http_only, host_only, same_site)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id, name, domain, path) DO UPDATE SET
			value = excluded.value,
			expires = excluded.expires,
			secure = excluded.secure,
			http_only = excluded.http_only,
			host_only = excluded.host_only,
			same_site = excluded.same_site,
			updated_at = strftime('%s', 'now')
	`, c.SessionID, c.Name, c.Value, c.Domain, c.Path, expires, boolInt(c.Secure), boolInt(c.HTTPOnly), boolInt(c.HostOnly), int(c.SameSite))
	if err != nil {
		return fmt.Errorf("cookie: set cookie: %w", err)
	}
	return nil
}

// GetCookies returns cookies matching the URL and optional session.
func (s *Store) GetCookies(ctx context.Context, rawURL, sessionID string) ([]types.Cookie, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("cookie: parse url: %w", err)
	}

	host := canonicalHost(u.Hostname())
	secureChannel := u.Scheme == "https" || u.Scheme == "wss"

	// Candidate selection is scoped to the session in SQL; the finer-grained
	// domain/path/secure/expiry rules are applied in Go for clarity and to
	// honour host-only scoping correctly.
	rows, err := s.db.QueryContext(ctx, `
		SELECT session_id, name, value, domain, path, expires, secure, http_only, host_only, same_site
		FROM cookies
		WHERE session_id = ? OR session_id = ''
		ORDER BY length(path) DESC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cookie: query cookies: %w", err)
	}
	defer rows.Close()

	var cookies []types.Cookie
	for rows.Next() {
		var c types.Cookie
		var expires sql.NullInt64
		if err := rows.Scan(&c.SessionID, &c.Name, &c.Value, &c.Domain, &c.Path, &expires, &c.Secure, &c.HTTPOnly, &c.HostOnly, &c.SameSite); err != nil {
			return nil, fmt.Errorf("cookie: scan cookie: %w", err)
		}
		if expires.Valid {
			c.Expires = time.Unix(expires.Int64, 0).UTC()
		}
		if isExpired(c) {
			continue
		}
		if c.Secure && !secureChannel {
			continue // Secure cookies are only sent over secure channels.
		}
		if !domainMatch(host, c.Domain, c.HostOnly) {
			continue
		}
		if !pathMatch(u.Path, c.Path) {
			continue
		}
		cookies = append(cookies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cookie: iterate cookies: %w", err)
	}
	return cookies, nil
}

// SyncSet returns the cookies that should be mirrored to a native cookie store
// for the given session: the session's own cookies plus shared persistent
// cookies (session_id = ”). Unlike All it never returns other sessions'
// cookies, preserving session isolation when syncing to the platform webview.
func (s *Store) SyncSet(ctx context.Context, sessionID string) ([]types.Cookie, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT session_id, name, value, domain, path, expires, secure, http_only, host_only, same_site
		FROM cookies
		WHERE session_id = ? OR session_id = ''
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cookie: query sync set: %w", err)
	}
	defer rows.Close()
	return scanCookies(rows)
}

// scanCookies materialises a result set of the full cookie column list.
func scanCookies(rows *sql.Rows) ([]types.Cookie, error) {
	var cookies []types.Cookie
	for rows.Next() {
		var c types.Cookie
		var expires sql.NullInt64
		if err := rows.Scan(&c.SessionID, &c.Name, &c.Value, &c.Domain, &c.Path, &expires, &c.Secure, &c.HTTPOnly, &c.HostOnly, &c.SameSite); err != nil {
			return nil, fmt.Errorf("cookie: scan cookie: %w", err)
		}
		if expires.Valid {
			c.Expires = time.Unix(expires.Int64, 0).UTC()
		}
		cookies = append(cookies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cookie: iterate cookies: %w", err)
	}
	return cookies, nil
}

// domainMatch reports whether a cookie scoped to cookieDomain may be sent to
// host. Host-only cookies require an exact host match; domain cookies also
// match subdomains (RFC 6265 §5.1.3).
//
// A cookie with an empty domain never matches: it cannot be safely scoped, so
// the conservative choice is to send it to no one. The jar always populates the
// domain from the request host, so this only guards malformed cookies inserted
// directly through the store.
//
// host and cookieDomain are compared as canonical hostnames. For IP-literal
// hosts (IPv4 or bracket-stripped IPv6) only the exact-match branch can ever
// fire, since the suffix rule requires a dotted parent domain.
func domainMatch(host, cookieDomain string, hostOnly bool) bool {
	cookieDomain = canonicalHost(cookieDomain)
	if cookieDomain == "" {
		return false
	}
	if host == cookieDomain {
		return true
	}
	if hostOnly {
		return false
	}
	return strings.HasSuffix(host, "."+cookieDomain)
}

// DeleteCookie removes a specific cookie.
func (s *Store) DeleteCookie(ctx context.Context, name, domain, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `
		DELETE FROM cookies WHERE name = ? AND domain = ? AND path = ?
	`, name, domain, path)
	if err != nil {
		return fmt.Errorf("cookie: delete cookie: %w", err)
	}
	return nil
}

// Clear removes all persistent cookies (session_id = ”).
func (s *Store) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM cookies WHERE session_id = ''`)
	if err != nil {
		return fmt.Errorf("cookie: clear: %w", err)
	}
	return nil
}

// ClearSession removes all cookies for a session.
func (s *Store) ClearSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(ctx, `DELETE FROM cookies WHERE session_id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("cookie: clear session: %w", err)
	}
	return nil
}

// All returns every cookie in the store (for sync/debug).
func (s *Store) All(ctx context.Context, sessionID string) ([]types.Cookie, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(ctx, `
		SELECT session_id, name, value, domain, path, expires, secure, http_only, host_only, same_site
		FROM cookies
		WHERE session_id = ? OR ? = ''
	`, sessionID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cookie: query all: %w", err)
	}
	defer rows.Close()
	return scanCookies(rows)
}

// Close closes the underlying database.
func (s *Store) Close() error {
	return s.db.Close()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func pathMatch(reqPath, cookiePath string) bool {
	if cookiePath == "" {
		cookiePath = "/"
	}
	if !strings.HasPrefix(reqPath, cookiePath) {
		return false
	}
	if len(reqPath) == len(cookiePath) {
		return true
	}
	if strings.HasSuffix(cookiePath, "/") {
		return true
	}
	return reqPath[len(cookiePath)] == '/'
}

func isExpired(c types.Cookie) bool {
	if c.Expires.IsZero() {
		return false
	}
	return time.Now().After(c.Expires)
}

// IsConstraintError reports whether err is a SQLite unique-constraint violation.
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code() == 2067 // SQLITE_CONSTRAINT_UNIQUE
	}
	// Fallback for wrapped errors or other drivers.
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
