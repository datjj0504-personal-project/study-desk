package settings

import (
	"errors"
	"testing"
	"time"
)

func TestDefaultTimerSettings(t *testing.T) {
	settings := DefaultTimerSettings()
	if err := settings.Validate(); err != nil {
		t.Fatalf("DefaultTimerSettings() is invalid: %v", err)
	}
	durations := settings.Durations()
	if durations.Focus != 25*time.Minute || durations.ShortBreak != 5*time.Minute ||
		durations.LongBreak != 15*time.Minute || durations.LongBreakEvery != 4 {
		t.Errorf("Durations() = %+v, want default Pomodoro durations", durations)
	}
}

func TestTimerSettingsValidate(t *testing.T) {
	tests := []struct {
		name     string
		settings TimerSettings
		wantErr  bool
	}{
		{name: "valid", settings: DefaultTimerSettings()},
		{name: "minimum values", settings: TimerSettings{FocusMinutes: 1, ShortBreakMinutes: 1, LongBreakMinutes: 1, LongBreakEvery: 1}},
		{name: "maximum values", settings: TimerSettings{FocusMinutes: 180, ShortBreakMinutes: 60, LongBreakMinutes: 60, LongBreakEvery: 12}},
		{name: "zero focus", settings: TimerSettings{FocusMinutes: 0, ShortBreakMinutes: 5, LongBreakMinutes: 15, LongBreakEvery: 4}, wantErr: true},
		{name: "excessive duration", settings: TimerSettings{FocusMinutes: 181, ShortBreakMinutes: 5, LongBreakMinutes: 15, LongBreakEvery: 4}, wantErr: true},
		{name: "invalid interval", settings: TimerSettings{FocusMinutes: 25, ShortBreakMinutes: 5, LongBreakMinutes: 15, LongBreakEvery: 13}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.settings.Validate()
			if errors.Is(err, ErrInvalidSettings) != test.wantErr {
				t.Errorf("Validate() error = %v, want invalid=%t", err, test.wantErr)
			}
		})
	}
}
