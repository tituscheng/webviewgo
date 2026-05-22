package cookie

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/tituscheng/webviewgo/internal/types"
)

// Manager coordinates the SQLite store, optional native sync, and the Jar.
type Manager struct {
	store     *Store
	sessionID string
	mu        sync.RWMutex
	onSync    func([]types.Cookie) error // optional native sync callback
}

// NewManager opens a manager backed by the store at the given path.
func NewManager(dbPath string) (*Manager, error) {
	store, err := Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &Manager{store: store}, nil
}

// SetCookie persists a cookie.
func (m *Manager) SetCookie(c types.Cookie) error {
	ctx := context.Background()
	if err := m.store.SetCookie(ctx, c); err != nil {
		return err
	}
	return m.flush(ctx)
}

// GetCookies returns matching cookies for a URL and optional session.
func (m *Manager) GetCookies(url string, sessionID string) ([]types.Cookie, error) {
	ctx := context.Background()
	return m.store.GetCookies(ctx, url, sessionID)
}

// DeleteCookie removes a specific cookie.
func (m *Manager) DeleteCookie(name, domain, path string) error {
	ctx := context.Background()
	if err := m.store.DeleteCookie(ctx, name, domain, path); err != nil {
		return err
	}
	return m.flush(ctx)
}

// Clear removes all persistent cookies.
func (m *Manager) Clear() error {
	ctx := context.Background()
	if err := m.store.Clear(ctx); err != nil {
		return err
	}
	return m.flush(ctx)
}

// ClearSession removes all cookies for a session.
func (m *Manager) ClearSession(sessionID string) error {
	ctx := context.Background()
	if err := m.store.ClearSession(ctx, sessionID); err != nil {
		return err
	}
	return m.flush(ctx)
}

// SaveSession is a no-op for the SQLite manager (already persisted).
func (m *Manager) SaveSession(sessionID string) error {
	return nil
}

// LoadSession makes sessionID the active session and pushes its cookies
// (plus shared persistent cookies) to the native store. Subsequent writes
// flush against this session.
func (m *Manager) LoadSession(sessionID string) error {
	ctx := context.Background()
	m.mu.Lock()
	m.sessionID = sessionID
	onSync := m.onSync
	m.mu.Unlock()

	cookies, err := m.store.SyncSet(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("cookie: load session: %w", err)
	}
	if onSync != nil {
		return onSync(cookies)
	}
	return nil
}

// AsJar returns an http.CookieJar backed by this manager.
func (m *Manager) AsJar() http.CookieJar {
	m.mu.RLock()
	sid := m.sessionID
	m.mu.RUnlock()
	return NewJar(m.store, sid)
}

// Close releases resources.
func (m *Manager) Close() error {
	return m.store.Close()
}

// SetSyncCallback registers a callback to push cookies to the native store.
func (m *Manager) SetSyncCallback(fn func([]types.Cookie) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onSync = fn
}

// flush calls the sync callback if registered. It pushes only the active
// session's cookies plus shared persistent cookies, never other sessions',
// so session isolation is preserved in the native cookie store.
func (m *Manager) flush(ctx context.Context) error {
	m.mu.RLock()
	onSync := m.onSync
	sid := m.sessionID
	m.mu.RUnlock()
	if onSync == nil {
		return nil
	}
	cookies, err := m.store.SyncSet(ctx, sid)
	if err != nil {
		return fmt.Errorf("cookie: flush: %w", err)
	}
	return onSync(cookies)
}
