# Database Design

The initial schema is implemented by the embedded `0001_init.sql` migration in
[migrations/](../../migrations/README.md). Future schema changes must be added
as new, versioned migrations rather than editing an applied migration.

SQLite, one local file. Timestamps are stored as UTC ISO-8601 text; IDs are integer primary keys.

```sql
CREATE TABLE subjects (
  id         INTEGER PRIMARY KEY,
  name       TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL
);

CREATE TABLE notes (
  id         INTEGER PRIMARY KEY,
  title      TEXT NOT NULL,
  body       TEXT NOT NULL DEFAULT '',
  subject_id INTEGER REFERENCES subjects(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX idx_notes_updated_at ON notes(updated_at);

CREATE TABLE study_sessions (
  id               INTEGER PRIMARY KEY,
  subject_id       INTEGER REFERENCES subjects(id) ON DELETE SET NULL,
  kind             TEXT NOT NULL CHECK (kind IN ('pomodoro','stopwatch','manual')),
  started_at       TEXT NOT NULL,
  ended_at         TEXT NOT NULL,
  duration_seconds INTEGER NOT NULL CHECK (duration_seconds >= 0),
  note             TEXT NOT NULL DEFAULT ''
);
CREATE INDEX idx_sessions_started_at ON study_sessions(started_at);

CREATE TABLE settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
```

The application uses the pure-Go [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) driver to avoid requiring a C toolchain for Windows builds; see [ADR 0002](../../ai/decisions/0002-use-pure-go-sqlite-driver.md). Full-text search (FTS5) for notes remains an open design question.
