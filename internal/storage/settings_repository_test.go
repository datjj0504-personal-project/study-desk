package storage

import (
	"context"
	"errors"
	"testing"

	"studydesk/internal/settings"
)

func TestSettingsRepositoryDefaultsAndPersistence(t *testing.T) {
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
	repo := NewSettingsRepository(db)

	got, err := repo.LoadTimerSettings(ctx)
	if err != nil {
		t.Fatalf("LoadTimerSettings() returned an error: %v", err)
	}
	if got != settings.DefaultTimerSettings() {
		t.Fatalf("LoadTimerSettings() = %+v, want defaults %+v", got, settings.DefaultTimerSettings())
	}

	want := settings.TimerSettings{
		FocusMinutes: 50, ShortBreakMinutes: 8, LongBreakMinutes: 20,
		LongBreakEvery: 3, NotificationsEnabled: true,
	}
	if err := repo.SaveTimerSettings(ctx, want); err != nil {
		t.Fatalf("SaveTimerSettings() returned an error: %v", err)
	}
	got, err = repo.LoadTimerSettings(ctx)
	if err != nil {
		t.Fatalf("LoadTimerSettings() after save returned an error: %v", err)
	}
	if got != want {
		t.Errorf("persisted settings = %+v, want %+v", got, want)
	}
}

func TestSettingsRepositoryRejectsInvalidAndCorruptSettings(t *testing.T) {
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
	repo := NewSettingsRepository(db)

	invalid := settings.DefaultTimerSettings()
	invalid.FocusMinutes = 0
	if err := repo.SaveTimerSettings(ctx, invalid); !errors.Is(err, settings.ErrInvalidSettings) {
		t.Errorf("SaveTimerSettings() invalid error = %v, want %v", err, settings.ErrInvalidSettings)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?)", settingFocusMinutes, "invalid"); err != nil {
		t.Fatalf("insert malformed setting: %v", err)
	}
	if _, err := repo.LoadTimerSettings(ctx); err == nil {
		t.Fatal("LoadTimerSettings() accepted malformed persisted data")
	}
}
