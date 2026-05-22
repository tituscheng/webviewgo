CREATE TABLE IF NOT EXISTS cookies (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id  TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL,
    value       TEXT NOT NULL,
    domain      TEXT NOT NULL,
    path        TEXT NOT NULL DEFAULT '/',
    expires     INTEGER, -- Unix timestamp, NULL for session cookies
    secure      INTEGER NOT NULL DEFAULT 0,
    http_only   INTEGER NOT NULL DEFAULT 0,
    host_only   INTEGER NOT NULL DEFAULT 0,
    same_site   INTEGER NOT NULL DEFAULT 0,
    created_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at  INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    UNIQUE(session_id, name, domain, path)
);

CREATE INDEX IF NOT EXISTS idx_cookies_session ON cookies(session_id);
CREATE INDEX IF NOT EXISTS idx_cookies_domain ON cookies(domain);
