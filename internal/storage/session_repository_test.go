package storage

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"studydesk/internal/study"
)

func TestSessionRepositoryPersistsAndListsSessions(t *testing.T) {
	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "sessions.db")
	db, err := OpenDatabase(ctx, databasePath)
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := db.Close(); err != nil {
				t.Errorf("close SQLite database: %v", err)
			}
		}
	})
	repo := NewSessionRepository(db)
	startedAt := time.Date(2026, 10, 24, 8, 0, 0, 123456789, time.FixedZone("UTC+2", 2*60*60))
	endedAt := startedAt.Add(25 * time.Minute)

	subjectResult, err := db.ExecContext(ctx,
		"INSERT INTO subjects (name, created_at) VALUES (?, ?)",
		"Physics", "2026-10-24T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	subjectID, err := subjectResult.LastInsertId()
	if err != nil {
		t.Fatalf("get subject ID: %v", err)
	}
	first, err := repo.CreateSession(ctx, study.Session{
		SubjectID: &subjectID, Kind: study.SessionKindPomodoro,
		StartedAt: startedAt, EndedAt: endedAt, DurationSeconds: 25 * 60,
		Note: "Reviewed motion",
	})
	if err != nil {
		t.Fatalf("CreateSession() returned an error: %v", err)
	}
	if first.ID < 1 || first.StartedAt.Location() != time.UTC || first.EndedAt.Location() != time.UTC {
		t.Fatalf("CreateSession() did not return generated ID and normalized UTC times: %+v", first)
	}
	if first.StartedAt.UnixNano() != startedAt.UnixNano() || first.EndedAt.UnixNano() != endedAt.UnixNano() {
		t.Errorf("CreateSession() changed timestamp instants: %+v", first)
	}
	secondStart := startedAt.Add(time.Hour)
	second, err := repo.CreateSession(ctx, study.Session{
		Kind: study.SessionKindManual, StartedAt: secondStart,
		EndedAt: secondStart.Add(10 * time.Minute), DurationSeconds: 600,
	})
	if err != nil {
		t.Fatalf("CreateSession() second session returned an error: %v", err)
	}

	got, err := repo.GetSession(ctx, first.ID)
	if err != nil {
		t.Fatalf("GetSession() returned an error: %v", err)
	}
	if got.ID != first.ID || got.SubjectID == nil || *got.SubjectID != subjectID ||
		got.Kind != study.SessionKindPomodoro || got.DurationSeconds != 1500 ||
		got.Note != "Reviewed motion" || !got.StartedAt.Equal(startedAt) || !got.EndedAt.Equal(endedAt) {
		t.Errorf("GetSession() = %+v, want persisted Pomodoro session", got)
	}

	listed, err := repo.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions() returned an error: %v", err)
	}
	if len(listed) != 2 || listed[0].ID != second.ID || listed[1].ID != first.ID {
		t.Errorf("ListSessions() = %+v, want newest start time first", listed)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close database before reopening: %v", err)
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
	reopenedRepo := NewSessionRepository(reopenedDB)
	persisted, err := reopenedRepo.GetSession(ctx, first.ID)
	if err != nil {
		t.Fatalf("GetSession() after reopen returned an error: %v", err)
	}
	if persisted.ID != first.ID || persisted.DurationSeconds != first.DurationSeconds || persisted.Note != first.Note {
		t.Errorf("persisted session after reopen = %+v, want %+v", persisted, first)
	}
}

func TestSessionRepositoryRejectsInvalidAndMissingSessions(t *testing.T) {
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
	repo := NewSessionRepository(db)

	valid := study.Session{
		Kind:            study.SessionKindStopwatch,
		StartedAt:       time.Date(2026, 10, 24, 8, 0, 0, 0, time.UTC),
		EndedAt:         time.Date(2026, 10, 24, 8, 1, 0, 0, time.UTC),
		DurationSeconds: 60,
	}
	invalid := valid
	invalid.Kind = "unknown"
	if _, err := repo.CreateSession(ctx, invalid); !errors.Is(err, study.ErrInvalidSession) {
		t.Errorf("CreateSession() invalid error = %v, want %v", err, study.ErrInvalidSession)
	}
	withID := valid
	withID.ID = 10
	if _, err := repo.CreateSession(ctx, withID); !errors.Is(err, study.ErrInvalidSession) {
		t.Errorf("CreateSession() with caller-supplied ID error = %v, want %v", err, study.ErrInvalidSession)
	}
	subjectID := int64(999)
	invalidSubject := valid
	invalidSubject.SubjectID = &subjectID
	if _, err := repo.CreateSession(ctx, invalidSubject); err == nil {
		t.Fatal("CreateSession() accepted a nonexistent subject")
	}
	if _, err := repo.GetSession(ctx, 1); !errors.Is(err, study.ErrSessionNotFound) {
		t.Errorf("GetSession() missing error = %v, want %v", err, study.ErrSessionNotFound)
	}
	if _, err := repo.GetSession(ctx, 0); !errors.Is(err, study.ErrSessionNotFound) {
		t.Errorf("GetSession() invalid ID error = %v, want %v", err, study.ErrSessionNotFound)
	}
	listed, err := repo.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions() empty returned an error: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("ListSessions() empty = %+v, want empty result", listed)
	}
}

func TestSessionRepositoryDeduplicatesCompletedSessions(t *testing.T) {
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
	repo := NewSessionRepository(db)
	start := time.Date(2026, 10, 24, 10, 0, 0, 0, time.UTC)
	session := study.Session{
		Kind:      study.SessionKindPomodoro,
		StartedAt: start, EndedAt: start.Add(25 * time.Minute),
		DurationSeconds: 1500,
	}
	first, err := repo.CreateCompletedSession(ctx, session, "pomodoro:unique-session")
	if err != nil {
		t.Fatalf("CreateCompletedSession() returned an error: %v", err)
	}
	if _, err := repo.CreateCompletedSession(ctx, session, ""); !errors.Is(err, study.ErrInvalidSession) {
		t.Errorf("CreateCompletedSession() without key error = %v, want %v", err, study.ErrInvalidSession)
	}
	duplicate := session
	duplicate.Note = "retry"
	got, err := repo.CreateCompletedSession(ctx, duplicate, "pomodoro:unique-session")
	if err != nil {
		t.Fatalf("CreateCompletedSession() retry returned an error: %v", err)
	}
	if got.ID != first.ID || got.Note != first.Note {
		t.Errorf("duplicate completion = %+v, want original session %+v", got, first)
	}
	listed, err := repo.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions() returned an error: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("ListSessions() returned %d sessions, want one idempotent completion", len(listed))
	}
}

func TestSessionRepositoryAggregatesStudyStatistics(t *testing.T) {
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
	repo := NewSessionRepository(db)
	physicsResult, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES (?, ?)", "Physics", "2026-10-24T00:00:00Z")
	if err != nil {
		t.Fatalf("insert Physics subject: %v", err)
	}
	physicsID, err := physicsResult.LastInsertId()
	if err != nil {
		t.Fatalf("get Physics subject ID: %v", err)
	}
	biologyResult, err := db.ExecContext(ctx, "INSERT INTO subjects (name, created_at) VALUES (?, ?)", "Biology", "2026-10-24T00:00:00Z")
	if err != nil {
		t.Fatalf("insert Biology subject: %v", err)
	}
	biologyID, err := biologyResult.LastInsertId()
	if err != nil {
		t.Fatalf("get Biology subject ID: %v", err)
	}

	window := study.StatisticsWindow{
		TodayStart: time.Date(2026, time.October, 21, 0, 0, 0, 0, time.UTC),
		TodayEnd:   time.Date(2026, time.October, 21, 12, 0, 0, 0, time.UTC),
		WeekStart:  time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC),
		WeekEnd:    time.Date(2026, time.October, 21, 12, 0, 0, 0, time.UTC),
	}
	empty, err := repo.GetStatistics(ctx, window)
	if err != nil {
		t.Fatalf("GetStatistics() on an empty repository returned an error: %v", err)
	}
	if empty.TodaySeconds != 0 || empty.WeekSeconds != 0 || empty.AllTimeSeconds != 0 || empty.BySubject == nil || len(empty.BySubject) != 0 {
		t.Errorf("empty statistics = %+v, want zero totals and an empty subject list", empty)
	}

	sessions := []study.Session{
		{SubjectID: &physicsID, Kind: study.SessionKindPomodoro, StartedAt: window.TodayStart.Add(time.Hour), EndedAt: window.TodayStart.Add(2 * time.Hour), DurationSeconds: 100},
		{Kind: study.SessionKindStopwatch, StartedAt: window.WeekStart.Add(time.Hour), EndedAt: window.WeekStart.Add(2 * time.Hour), DurationSeconds: 200},
		{SubjectID: &physicsID, Kind: study.SessionKindManual, StartedAt: window.WeekStart.Add(-time.Hour), EndedAt: window.WeekStart, DurationSeconds: 300},
		{SubjectID: &biologyID, Kind: study.SessionKindManual, StartedAt: window.TodayEnd, EndedAt: window.TodayEnd.Add(time.Hour), DurationSeconds: 400},
	}
	for _, session := range sessions {
		if _, err := repo.CreateSession(ctx, session); err != nil {
			t.Fatalf("CreateSession() returned an error: %v", err)
		}
	}

	got, err := repo.GetStatistics(ctx, window)
	if err != nil {
		t.Fatalf("GetStatistics() returned an error: %v", err)
	}
	if got.TodaySeconds != 100 || got.WeekSeconds != 300 || got.AllTimeSeconds != 1000 {
		t.Errorf("statistics totals = today %d, week %d, all-time %d; want 100, 300, 1000",
			got.TodaySeconds, got.WeekSeconds, got.AllTimeSeconds)
	}
	if len(got.BySubject) != 3 {
		t.Fatalf("GetStatistics() returned %d subject rows, want 3: %+v", len(got.BySubject), got.BySubject)
	}
	durations := make(map[string]int64, len(got.BySubject))
	for _, subject := range got.BySubject {
		durations[subject.SubjectName] = subject.DurationSeconds
	}
	if durations["Physics"] != 400 || durations["Biology"] != 400 || durations["Unassigned"] != 200 {
		t.Errorf("subject durations = %v, want Physics 400, Biology 400, Unassigned 200", durations)
	}
}

func TestSessionRepositorySubjectDeletionSetsNull(t *testing.T) {
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
	repo := NewSessionRepository(db)
	result, err := db.ExecContext(ctx,
		"INSERT INTO subjects (name, created_at) VALUES (?, ?)",
		"Chemistry", "2026-10-24T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert subject: %v", err)
	}
	subjectID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("get subject ID: %v", err)
	}
	start := time.Date(2026, 10, 24, 9, 0, 0, 0, time.UTC)
	session, err := repo.CreateSession(ctx, study.Session{
		SubjectID: &subjectID, Kind: study.SessionKindManual,
		StartedAt: start, EndedAt: start.Add(time.Minute), DurationSeconds: 60,
	})
	if err != nil {
		t.Fatalf("CreateSession() returned an error: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM subjects WHERE id = ?", subjectID); err != nil {
		t.Fatalf("delete subject: %v", err)
	}
	got, err := repo.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSession() after subject deletion returned an error: %v", err)
	}
	if got.SubjectID != nil {
		t.Errorf("subject ID after deletion = %v, want nil", *got.SubjectID)
	}
}
