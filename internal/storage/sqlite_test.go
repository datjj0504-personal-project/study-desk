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
	if _, err := db.ExecContext(ctx, `
		INSERT INTO study_sessions (kind, started_at, ended_at, duration_seconds)
		VALUES ('manual', 'start', 'end', -1)
	`); err == nil {
		t.Fatal("study_sessions accepted a negative duration")
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO notes (title, subject_id, created_at, updated_at)
		VALUES ('Orphan note', 999, 'created', 'updated')
	`); err == nil {
		t.Fatal("notes accepted a reference to a missing subject")
	}

	if _, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES ('Math', 'created')"); err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES ('Math', 'created')"); err == nil {
		t.Fatal("subjects accepted a duplicate name")
	}

	result, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES ('Science', 'created')")
	if err != nil {
		t.Fatalf("insert second subject: %v", err)
	}
	subjectID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get second subject ID: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO notes (title, subject_id, created_at, updated_at)
		VALUES ('Related note', ?, 'created', 'updated')
	`, subjectID); err != nil {
		t.Fatalf("insert note with subject: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM subjects WHERE id = ?", subjectID); err != nil {
		t.Fatalf("delete referenced subject: %v", err)
	}
	var noteSubjectID *int64
	if err := db.QueryRowContext(ctx, "SELECT subject_id FROM notes WHERE title = 'Related note'").Scan(&noteSubjectID); err != nil {
		t.Fatalf("read note subject after deleting subject: %v", err)
	}
	if noteSubjectID != nil {
		t.Errorf("note subject after deleting subject = %d, want NULL", *noteSubjectID)
	}
}

func TestOpenDatabasePersistsDataAfterReopen(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "studydesk.db")

	db, err := OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := db.Close(); err != nil {
				t.Errorf("close initial SQLite database: %v", err)
			}
		}
	})

	result, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES (?, ?)", "Physics", "2026-10-24T00:00:00Z")
	if err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	subjectID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get subject ID: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO notes (title, body, subject_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "Motion", "Velocity and acceleration", subjectID, "created", "updated"); err != nil {
		t.Fatalf("insert note: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO study_sessions (subject_id, kind, started_at, ended_at, duration_seconds, note)
		VALUES (?, 'pomodoro', 'start', 'end', 1500, 'Reviewed motion')
	`, subjectID); err != nil {
		t.Fatalf("insert study session: %v", err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES ('theme', 'dark')"); err != nil {
		t.Fatalf("insert setting: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close initial SQLite database: %v", err)
	}
	closed = true

	reopenedDB, err := OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	t.Cleanup(func() {
		if err := reopenedDB.Close(); err != nil {
			t.Errorf("close reopened SQLite database: %v", err)
		}
	})

	var gotName string
	if err := reopenedDB.QueryRowContext(ctx, "SELECT name FROM subjects WHERE id = ?", subjectID).Scan(&gotName); err != nil {
		t.Fatalf("read persisted subject: %v", err)
	}
	if gotName != "Physics" {
		t.Errorf("persisted subject name = %q, want %q", gotName, "Physics")
	}

	var gotBody, gotTitle string
	if err := reopenedDB.QueryRowContext(ctx, "SELECT title, body FROM notes WHERE subject_id = ?", subjectID).Scan(&gotTitle, &gotBody); err != nil {
		t.Fatalf("read persisted note: %v", err)
	}
	if gotTitle != "Motion" || gotBody != "Velocity and acceleration" {
		t.Errorf("persisted note = (%q, %q), want (%q, %q)", gotTitle, gotBody, "Motion", "Velocity and acceleration")
	}

	var gotDuration int
	var gotSessionNote string
	if err := reopenedDB.QueryRowContext(ctx, "SELECT duration_seconds, note FROM study_sessions WHERE subject_id = ?", subjectID).Scan(&gotDuration, &gotSessionNote); err != nil {
		t.Fatalf("read persisted study session: %v", err)
	}
	if gotDuration != 1500 || gotSessionNote != "Reviewed motion" {
		t.Errorf("persisted study session = (%d, %q), want (%d, %q)", gotDuration, gotSessionNote, 1500, "Reviewed motion")
	}

	var gotSetting string
	if err := reopenedDB.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = 'theme'").Scan(&gotSetting); err != nil {
		t.Fatalf("read persisted setting: %v", err)
	}
	if gotSetting != "dark" {
		t.Errorf("persisted setting = %q, want %q", gotSetting, "dark")
	}
}

func TestApplyMigrationRollsBackOnFailure(t *testing.T) {
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

	err = applyMigration(ctx, db, 99, "0099_failure.sql", `
		CREATE TABLE should_rollback (id INTEGER PRIMARY KEY);
		INSERT INTO missing_table (id) VALUES (1);
	`)
	if err == nil {
		t.Fatal("applyMigration() succeeded despite an invalid statement")
	}

	var tableExists bool
	if err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'should_rollback')",
	).Scan(&tableExists); err != nil {
		t.Fatalf("check rolled-back table: %v", err)
	}
	if tableExists {
		t.Fatal("failed migration left its table behind")
	}

	var migrationRecorded bool
	if err := db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 99)",
	).Scan(&migrationRecorded); err != nil {
		t.Fatalf("check failed migration record: %v", err)
	}
	if migrationRecorded {
		t.Fatal("failed migration was recorded as applied")
	}
}
