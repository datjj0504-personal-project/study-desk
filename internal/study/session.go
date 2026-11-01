package study

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSession  = errors.New("invalid study session")
	ErrSessionNotFound = errors.New("study session not found")
)

type SessionKind string

const (
	SessionKindPomodoro  SessionKind = "pomodoro"
	SessionKindStopwatch SessionKind = "stopwatch"
	SessionKindManual    SessionKind = "manual"
)

// Session records a completed or manually entered study period.
type Session struct {
	ID              int64       `json:"id"`
	SubjectID       *int64      `json:"subjectId"`
	Kind            SessionKind `json:"kind"`
	StartedAt       time.Time   `json:"startedAt"`
	EndedAt         time.Time   `json:"endedAt"`
	DurationSeconds int64       `json:"durationSeconds"`
	Note            string      `json:"note"`
}

func (s Session) Validate() error {
	if s.ID < 0 || (s.SubjectID != nil && *s.SubjectID < 1) {
		return fmt.Errorf("%w: IDs must be positive when provided", ErrInvalidSession)
	}
	switch s.Kind {
	case SessionKindPomodoro, SessionKindStopwatch, SessionKindManual:
	default:
		return fmt.Errorf("%w: unsupported kind %q", ErrInvalidSession, s.Kind)
	}
	if s.StartedAt.IsZero() || s.EndedAt.IsZero() || s.EndedAt.Before(s.StartedAt) {
		return fmt.Errorf("%w: start and end times are required and end cannot precede start", ErrInvalidSession)
	}
	if s.DurationSeconds < 0 {
		return fmt.Errorf("%w: duration cannot be negative", ErrInvalidSession)
	}
	return nil
}
