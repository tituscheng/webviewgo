package cookie

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tituscheng/webviewgo/internal/types"
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
		INSERT INTO cookies (session_id, name, value, domain, path, expires, secure, http_only, same_site)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(session_id, name, domain, path) DO UPDATE SET
			value = excluded.value,
			expires = excluded.expires,
			secure = excluded.secure,
			http_only = excluded.http_only,
			same_site = excluded.same_site,
			updated_at = strftime('%s', 'now')
	`, c.SessionID, c.Name, c.Value, c.Domain, c.Path, expires, boolInt(c.Secure), boolInt(c.HTTPOnly), int(c.SameSite))
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

	// Simple domain matching: exact or suffix with leading dot.
	domain := u.Hostname()
	rows, err := s.db.QueryContext(ctx, `
		SELECT session_id, name, value, domain, path, expires, secure, http_only, same_site
		FROM cookies
		WHERE (domain = ? OR domain = ? OR substr(?, -length(domain)-1) = '.' || domain)
		  AND (session_id = ? OR session_id = '')
		ORDER BY length(path) DESC
	`, domain, "."+domain, domain, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cookie: query cookies: %w", err)
	}
	defer rows.Close()

	var cookies []types.Cookie
	for rows.Next() {
		var c types.Cookie
		var expires sql.NullInt64
		if err := rows.Scan(&c.SessionID, &c.Name, &c.Value, &c.Domain, &c.Path, &expires, &c.Secure, &c.HTTPOnly, &c.SameSite); err != nil {
			return nil, fmt.Errorf("cookie: scan cookie: %w", err)
		}
		if expires.Valid {
			c.Expires = time.Unix(expires.Int64, 0).UTC()
		}
		if pathMatch(u.Path, c.Path) && !isExpired(c) {
			cookies = append(cookies, c)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cookie: iterate cookies: %w", err)
	}
	return cookies, nil
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
		SELECT session_id, name, value, domain, path, expires, secure, http_only, same_site
		FROM cookies
		WHERE session_id = ? OR ? = ''
	`, sessionID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("cookie: query all: %w", err)
	}
	defer rows.Close()

	var cookies []types.Cookie
	for rows.Next() {
		var c types.Cookie
		var expires sql.NullInt64
		if err := rows.Scan(&c.SessionID, &c.Name, &c.Value, &c.Domain, &c.Path, &expires, &c.Secure, &c.HTTPOnly, &c.SameSite); err != nil {
			return nil, fmt.Errorf("cookie: scan all: %w", err)
		}
		if expires.Valid {
			c.Expires = time.Unix(expires.Int64, 0).UTC()
		}
		cookies = append(cookies, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cookie: iterate all: %w", err)
	}
	return cookies, nil
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
	// sqlite.Error with code 2067 = SQLITE_CONSTRAINT_UNIQUE
	// Use error string fallback for compatibility.
	if fmt.Sprintf("%T", err) == "*sqlite.Error" {
		// We can't import the specific type easily across versions,
		// so rely on string matching.
		return strings.Contains(err.Error(), "UNIQUE constraint failed")
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
