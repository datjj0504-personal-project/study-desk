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

	theme, err := repo.LoadTheme(ctx)
	if err != nil {
		t.Fatalf("LoadTheme() returned an error: %v", err)
	}
	if theme != settings.DefaultTheme() {
		t.Fatalf("LoadTheme() = %q, want default %q", theme, settings.DefaultTheme())
	}
	if err := repo.SaveTheme(ctx, settings.ThemeLight); err != nil {
		t.Fatalf("SaveTheme() returned an error: %v", err)
	}
	theme, err = repo.LoadTheme(ctx)
	if err != nil {
		t.Fatalf("LoadTheme() after save returned an error: %v", err)
	}
	if theme != settings.ThemeLight {
		t.Errorf("persisted theme = %q, want %q", theme, settings.ThemeLight)
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
	if err := repo.SaveTheme(ctx, settings.Theme("sepia")); !errors.Is(err, settings.ErrInvalidTheme) {
		t.Errorf("SaveTheme() invalid error = %v, want %v", err, settings.ErrInvalidTheme)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES (?, ?)", settingTheme, "sepia"); err != nil {
		t.Fatalf("insert malformed theme: %v", err)
	}
	if _, err := repo.LoadTheme(ctx); !errors.Is(err, settings.ErrInvalidTheme) {
		t.Errorf("LoadTheme() corrupt error = %v, want %v", err, settings.ErrInvalidTheme)
	}
}
