package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"studydesk/internal/timer"
)

func TestTimerStateRepositoryPersistsSnapshots(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "timers.db")
	db, err := OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})
	repo := NewTimerStateRepository(db)

	if _, found, err := repo.LoadTimerState(ctx, timer.ModePomodoro); err != nil || found {
		t.Fatalf("LoadTimerState() for missing value = found %t, error %v; want not found", found, err)
	}
	want := timer.Snapshot{
		Mode: timer.ModePomodoro, Phase: timer.PhaseFocus, Status: timer.StatusRunning,
		Duration: 25 * time.Minute, Elapsed: 3 * time.Minute,
		StartedAtUnixNano: time.Date(2026, 1, 1, 0, 3, 0, 0, time.UTC).UnixNano(),
	}
	if err := repo.SaveTimerState(ctx, want); err != nil {
		t.Fatalf("SaveTimerState() returned an error: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close SQLite database before reopening: %v", err)
	}
	db, err = OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("reopen SQLite database: %v", err)
	}
	repo = NewTimerStateRepository(db)
	got, found, err := repo.LoadTimerState(ctx, timer.ModePomodoro)
	if err != nil {
		t.Fatalf("LoadTimerState() returned an error: %v", err)
	}
	if !found || got != want {
		t.Errorf("LoadTimerState() = %+v, found %t; want %+v", got, found, want)
	}
}

func TestTimerStateRepositoryRejectsCorruptAndUnsupportedState(t *testing.T) {
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
	repo := NewTimerStateRepository(db)

	if _, _, err := repo.LoadTimerState(ctx, timer.Mode("invalid")); err == nil {
		t.Fatal("LoadTimerState() accepted an unsupported mode")
	}
	if err := repo.SaveTimerState(ctx, timer.Snapshot{Mode: timer.Mode("invalid")}); err == nil {
		t.Fatal("SaveTimerState() accepted an unsupported mode")
	}
	if _, err := db.ExecContext(ctx,
		"INSERT INTO settings (key, value) VALUES (?, ?)",
		timerStateSettingPrefix+string(timer.ModePomodoro), "{invalid",
	); err != nil {
		t.Fatalf("insert corrupt timer state: %v", err)
	}
	if _, _, err := repo.LoadTimerState(ctx, timer.ModePomodoro); err == nil {
		t.Fatal("LoadTimerState() accepted corrupt persisted data")
	}
}
