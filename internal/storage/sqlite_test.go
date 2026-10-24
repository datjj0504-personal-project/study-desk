package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpenSQLite(t *testing.T) {
	db, err := OpenSQLite(":memory:")
	if err != nil {
		t.Fatalf("OpenSQLite() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("ping SQLite database: %v", err)
	}

	var version string
	if err := db.QueryRowContext(context.Background(), "SELECT sqlite_version()").Scan(&version); err != nil {
		t.Fatalf("query SQLite version: %v", err)
	}
	if version == "" {
		t.Fatal("SQLite returned an empty version")
	}
}

func TestOpenDatabaseAppliesMigrations(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "studydesk.db")

	db, err := OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})

	for _, table := range []string{"subjects", "notes", "study_sessions", "settings"} {
		var exists bool
		if err := db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?)",
			table,
		).Scan(&exists); err != nil {
			t.Fatalf("check for table %q: %v", table, err)
		}
		if !exists {
			t.Errorf("migration did not create table %q", table)
		}
	}

	var version int
	if err := db.QueryRowContext(ctx, "SELECT version FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("read applied migration version: %v", err)
	}
	if version != 1 {
		t.Errorf("applied migration version = %d, want 1", version)
	}

	var foreignKeys int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign key setting: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign keys = %d, want 1", foreignKeys)
	}

	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("ApplyMigrations() on initialized database returned an error: %v", err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("count applied migrations: %v", err)
	}
	if version != 1 {
		t.Errorf("migration count after applying again = %d, want 1", version)
	}
}

func TestOpenDatabaseEnforcesSchemaConstraints(t *testing.T) {
	ctx := context.Background()
	db, err := OpenDatabase(ctx, ":memory:")
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})

	if _, err := db.ExecContext(ctx, `
		INSERT INTO study_sessions (kind, started_at, ended_at, duration_seconds)
		VALUES ('invalid', 'start', 'end', 0)
	`); err == nil {
		t.Fatal("study_sessions accepted an invalid kind")
	}
}
