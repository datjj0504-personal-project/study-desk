package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"studydesk/migrations"
)

const (
	databaseDirectory = "StudyDesk"
	databaseFilename  = "studydesk.db"
)

// OpenSQLite opens a database using the registered pure-Go SQLite driver.
func OpenSQLite(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}
	return db, nil
}

// OpenAppDatabase opens the user's StudyDesk database and applies pending
// migrations before returning it.
func OpenAppDatabase(ctx context.Context) (*sql.DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("find user config directory: %w", err)
	}

	databaseDir := filepath.Join(configDir, databaseDirectory)
	if err := os.MkdirAll(databaseDir, 0o700); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	return OpenDatabase(ctx, filepath.Join(databaseDir, databaseFilename))
}

// OpenDatabase opens the SQLite database at dataSourceName, configures it for
// local application use, and applies all pending embedded migrations.
func OpenDatabase(ctx context.Context, dataSourceName string) (*sql.DB, error) {
	db, err := OpenSQLite(dataSourceName)
	if err != nil {
		return nil, err
	}

	// SQLite connection-local settings must be applied consistently, so keep a
	// single pooled connection for this local desktop database.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		return nil, closeWithError(db, fmt.Errorf("ping SQLite database: %w", err))
	}
	for _, pragma := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA journal_mode = WAL",
	} {
		if _, err := db.ExecContext(ctx, pragma); err != nil {
			return nil, closeWithError(db, fmt.Errorf("configure SQLite database (%s): %w", pragma, err))
		}
	}
	if err := ApplyMigrations(ctx, db); err != nil {
		return nil, closeWithError(db, err)
	}

	return db, nil
}

// ApplyMigrations applies each embedded migration not already recorded by the
// database. Each migration and its version record are committed atomically.
func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL
		)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	entries, err := migrations.Files.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		version, err := migrationVersion(entry.Name())
		if err != nil {
			return err
		}

		var applied bool
		err = db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)",
			version,
		).Scan(&applied)
		if err != nil {
			return fmt.Errorf("check migration %q: %w", entry.Name(), err)
		}
		if applied {
			continue
		}

		script, err := migrations.Files.ReadFile(entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		if err := applyMigration(ctx, db, version, entry.Name(), string(script)); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, version int, name, script string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %q: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx, script); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("apply migration %q: %w (rollback failed: %v)", name, err, rollbackErr)
		}
		return fmt.Errorf("apply migration %q: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
		version, name, time.Now().UTC().Format(time.RFC3339Nano),
	); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("record migration %q: %w (rollback failed: %v)", name, err, rollbackErr)
		}
		return fmt.Errorf("record migration %q: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %q: %w", name, err)
	}
	return nil
}

func migrationVersion(name string) (int, error) {
	prefix, _, ok := strings.Cut(name, "_")
	if !ok || len(prefix) != 4 {
		return 0, fmt.Errorf("invalid migration filename %q: expected NNNN_description.sql", name)
	}
	version, err := strconv.Atoi(prefix)
	if err != nil || version < 1 {
		return 0, fmt.Errorf("invalid migration version in filename %q", name)
	}
	return version, nil
}

func closeWithError(db *sql.DB, cause error) error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("%w (close database: %v)", cause, err)
	}
	return cause
}
