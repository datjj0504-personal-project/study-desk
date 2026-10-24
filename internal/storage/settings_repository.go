package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"studydesk/internal/settings"
)

const (
	settingFocusMinutes         = "timer.focus_minutes"
	settingShortBreakMinutes    = "timer.short_break_minutes"
	settingLongBreakMinutes     = "timer.long_break_minutes"
	settingLongBreakEvery       = "timer.long_break_every"
	settingNotificationsEnabled = "timer.notifications_enabled"
)

type SettingsRepository struct {
	db *sql.DB
}

var _ settings.Repository = (*SettingsRepository)(nil)

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) LoadTimerSettings(ctx context.Context) (settings.TimerSettings, error) {
	result := settings.DefaultTimerSettings()
	values := map[string]*int{
		settingFocusMinutes:      &result.FocusMinutes,
		settingShortBreakMinutes: &result.ShortBreakMinutes,
		settingLongBreakMinutes:  &result.LongBreakMinutes,
		settingLongBreakEvery:    &result.LongBreakEvery,
	}
	for key, target := range values {
		value, found, err := r.loadValue(ctx, key)
		if err != nil {
			return settings.TimerSettings{}, err
		}
		if found {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				return settings.TimerSettings{}, fmt.Errorf("parse setting %q: %w", key, err)
			}
			*target = parsed
		}
	}
	notificationValue, found, err := r.loadValue(ctx, settingNotificationsEnabled)
	if err != nil {
		return settings.TimerSettings{}, err
	}
	if found {
		result.NotificationsEnabled, err = strconv.ParseBool(notificationValue)
		if err != nil {
			return settings.TimerSettings{}, fmt.Errorf("parse setting %q: %w", settingNotificationsEnabled, err)
		}
	}
	if err := result.Validate(); err != nil {
		return settings.TimerSettings{}, fmt.Errorf("load timer settings: %w", err)
	}
	return result, nil
}

func (r *SettingsRepository) SaveTimerSettings(ctx context.Context, value settings.TimerSettings) error {
	if err := value.Validate(); err != nil {
		return err
	}
	values := map[string]string{
		settingFocusMinutes:         strconv.Itoa(value.FocusMinutes),
		settingShortBreakMinutes:    strconv.Itoa(value.ShortBreakMinutes),
		settingLongBreakMinutes:     strconv.Itoa(value.LongBreakMinutes),
		settingLongBreakEvery:       strconv.Itoa(value.LongBreakEvery),
		settingNotificationsEnabled: strconv.FormatBool(value.NotificationsEnabled),
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save timer settings: %w", err)
	}
	for key, settingValue := range values {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO settings (key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, settingValue); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("save timer setting %q: %w (rollback failed: %v)", key, err, rollbackErr)
			}
			return fmt.Errorf("save timer setting %q: %w", key, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit timer settings: %w", err)
	}
	return nil
}

func (r *SettingsRepository) loadValue(ctx context.Context, key string) (string, bool, error) {
	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load setting %q: %w", key, err)
	}
	return value, true, nil
}
