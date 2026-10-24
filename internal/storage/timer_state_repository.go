package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"studydesk/internal/timer"
)

const timerStateSettingPrefix = "timer.state."

type TimerStateRepository struct {
	db *sql.DB
}

var _ timer.StateRepository = (*TimerStateRepository)(nil)

func NewTimerStateRepository(db *sql.DB) *TimerStateRepository {
	return &TimerStateRepository{db: db}
}

func (r *TimerStateRepository) LoadTimerState(ctx context.Context, mode timer.Mode) (timer.Snapshot, bool, error) {
	if mode != timer.ModePomodoro && mode != timer.ModeStopwatch {
		return timer.Snapshot{}, false, fmt.Errorf("load timer state: unsupported mode %q", mode)
	}
	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", timerStateSettingPrefix+string(mode)).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return timer.Snapshot{}, false, nil
	}
	if err != nil {
		return timer.Snapshot{}, false, fmt.Errorf("load %s state: %w", mode, err)
	}
	var snapshot timer.Snapshot
	if err := json.Unmarshal([]byte(value), &snapshot); err != nil {
		return timer.Snapshot{}, false, fmt.Errorf("decode %s state: %w", mode, err)
	}
	if snapshot.Mode != mode {
		return timer.Snapshot{}, false, fmt.Errorf("load %s state: stored snapshot mode is %q", mode, snapshot.Mode)
	}
	return snapshot, true, nil
}

func (r *TimerStateRepository) SaveTimerState(ctx context.Context, snapshot timer.Snapshot) error {
	if snapshot.Mode != timer.ModePomodoro && snapshot.Mode != timer.ModeStopwatch {
		return fmt.Errorf("save timer state: unsupported mode %q", snapshot.Mode)
	}
	value, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("encode %s state: %w", snapshot.Mode, err)
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, timerStateSettingPrefix+string(snapshot.Mode), string(value)); err != nil {
		return fmt.Errorf("save %s state: %w", snapshot.Mode, err)
	}
	return nil
}
