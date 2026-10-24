package study

import (
	"errors"
	"testing"
	"time"
)

func validSession() Session {
	start := time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC)
	return Session{
		Kind: SessionKindPomodoro, StartedAt: start, EndedAt: start.Add(25 * time.Minute),
		DurationSeconds: 25 * 60, Note: "Review chapter",
	}
}

func TestSessionValidate(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*Session)
		wantErr bool
	}{
		{name: "valid Pomodoro"},
		{name: "valid stopwatch", change: func(session *Session) { session.Kind = SessionKindStopwatch }},
		{name: "valid manual session", change: func(session *Session) { session.Kind = SessionKindManual }},
		{name: "zero duration", change: func(session *Session) { session.DurationSeconds = 0 }},
		{name: "invalid kind", change: func(session *Session) { session.Kind = "break" }, wantErr: true},
		{name: "missing start", change: func(session *Session) { session.StartedAt = time.Time{} }, wantErr: true},
		{name: "missing end", change: func(session *Session) { session.EndedAt = time.Time{} }, wantErr: true},
		{name: "end before start", change: func(session *Session) { session.EndedAt = session.StartedAt.Add(-time.Second) }, wantErr: true},
		{name: "negative duration", change: func(session *Session) { session.DurationSeconds = -1 }, wantErr: true},
		{name: "negative ID", change: func(session *Session) { session.ID = -1 }, wantErr: true},
		{name: "invalid subject ID", change: func(session *Session) { id := int64(0); session.SubjectID = &id }, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session := validSession()
			if test.change != nil {
				test.change(&session)
			}
			err := session.Validate()
			if errors.Is(err, ErrInvalidSession) != test.wantErr {
				t.Errorf("Validate() error = %v, want invalid=%t", err, test.wantErr)
			}
		})
	}
}
