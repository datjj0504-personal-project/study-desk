package settings

import (
	"errors"
	"fmt"
	"time"

	"studydesk/internal/timer"
)

var ErrInvalidSettings = errors.New("invalid timer settings")
var ErrInvalidTheme = errors.New("invalid theme")

type Theme string

const (
	ThemeSystem Theme = "system"
	ThemeLight  Theme = "light"
	ThemeDark   Theme = "dark"
)

func DefaultTheme() Theme {
	return ThemeDark
}

func (theme Theme) Validate() error {
	switch theme {
	case ThemeSystem, ThemeLight, ThemeDark:
		return nil
	default:
		return fmt.Errorf("%w: must be system, light, or dark", ErrInvalidTheme)
	}
}

const (
	minDurationMinutes = 1
	maxFocusMinutes    = 180
	maxBreakMinutes    = 60
	maxLongBreakEvery  = 12
)

type TimerSettings struct {
	FocusMinutes         int  `json:"focusMinutes"`
	ShortBreakMinutes    int  `json:"shortBreakMinutes"`
	LongBreakMinutes     int  `json:"longBreakMinutes"`
	LongBreakEvery       int  `json:"longBreakEvery"`
	NotificationsEnabled bool `json:"notificationsEnabled"`
}

func DefaultTimerSettings() TimerSettings {
	return TimerSettings{
		FocusMinutes:         25,
		ShortBreakMinutes:    5,
		LongBreakMinutes:     15,
		LongBreakEvery:       4,
		NotificationsEnabled: false,
	}
}

func (s TimerSettings) Validate() error {
	if s.FocusMinutes < minDurationMinutes || s.FocusMinutes > maxFocusMinutes ||
		s.ShortBreakMinutes < minDurationMinutes || s.ShortBreakMinutes > maxBreakMinutes ||
		s.LongBreakMinutes < minDurationMinutes || s.LongBreakMinutes > maxBreakMinutes ||
		s.LongBreakEvery < minDurationMinutes || s.LongBreakEvery > maxLongBreakEvery {
		return fmt.Errorf("%w: focus must be 1-180 minutes, breaks 1-60 minutes, and long break interval 1-12", ErrInvalidSettings)
	}
	return nil
}

func (s TimerSettings) Durations() timer.Durations {
	return timer.Durations{
		Focus:          time.Duration(s.FocusMinutes) * time.Minute,
		ShortBreak:     time.Duration(s.ShortBreakMinutes) * time.Minute,
		LongBreak:      time.Duration(s.LongBreakMinutes) * time.Minute,
		LongBreakEvery: s.LongBreakEvery,
	}
}
