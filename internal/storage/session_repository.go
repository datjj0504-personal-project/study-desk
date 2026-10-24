package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"studydesk/internal/study"
)

// SessionRepository stores study sessions in SQLite.
type SessionRepository struct {
	db *sql.DB
}

var _ study.SessionRepository = (*SessionRepository)(nil)

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session study.Session) (study.Session, error) {
	return r.createSession(ctx, session, "")
}

func (r *SessionRepository) CreateCompletedSession(ctx context.Context, session study.Session, completionKey string) (study.Session, error) {
	if completionKey == "" {
		return study.Session{}, fmt.Errorf("%w: completion key is required", study.ErrInvalidSession)
	}
	return r.createSession(ctx, session, completionKey)
}

func (r *SessionRepository) createSession(ctx context.Context, session study.Session, completionKey string) (study.Session, error) {
	if session.ID != 0 {
		return study.Session{}, fmt.Errorf("%w: new sessions cannot specify an ID", study.ErrInvalidSession)
	}
	if err := session.Validate(); err != nil {
		return study.Session{}, err
	}
	startedAt := session.StartedAt.UTC()
	endedAt := session.EndedAt.UTC()
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO study_sessions
			(subject_id, kind, started_at, ended_at, duration_seconds, note, completion_key)
		VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''))
		ON CONFLICT(completion_key) DO NOTHING
	`, session.SubjectID, session.Kind, startedAt.Format(sqliteTimestampLayout),
		endedAt.Format(sqliteTimestampLayout), session.DurationSeconds, session.Note, completionKey)
	if err != nil {
		return study.Session{}, fmt.Errorf("create study session: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return study.Session{}, fmt.Errorf("check created study session: %w", err)
	}
	if rowsAffected == 0 {
		persisted, err := scanSession(r.db.QueryRowContext(ctx, `
			SELECT id, subject_id, kind, started_at, ended_at, duration_seconds, note
			FROM study_sessions WHERE completion_key = ?
		`, completionKey))
		if err != nil {
			return study.Session{}, fmt.Errorf("load existing completed study session: %w", err)
		}
		return persisted, nil
	}
	session.ID, err = result.LastInsertId()
	if err != nil {
		return study.Session{}, fmt.Errorf("get created study session ID: %w", err)
	}
	session.StartedAt = startedAt
	session.EndedAt = endedAt
	return session, nil
}

func (r *SessionRepository) GetSession(ctx context.Context, id int64) (study.Session, error) {
	if id < 1 {
		return study.Session{}, fmt.Errorf("get study session: %w", study.ErrSessionNotFound)
	}
	session, err := scanSession(r.db.QueryRowContext(ctx, `
		SELECT id, subject_id, kind, started_at, ended_at, duration_seconds, note
		FROM study_sessions
		WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return study.Session{}, fmt.Errorf("study session %d: %w", id, study.ErrSessionNotFound)
	}
	if err != nil {
		return study.Session{}, fmt.Errorf("get study session %d: %w", id, err)
	}
	return session, nil
}

func (r *SessionRepository) ListSessions(ctx context.Context) ([]study.Session, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, subject_id, kind, started_at, ended_at, duration_seconds, note
		FROM study_sessions
		ORDER BY started_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list study sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]study.Session, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan study session: %w", err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate study sessions: %w", err)
	}
	return sessions, nil
}

func (r *SessionRepository) GetStatistics(ctx context.Context, window study.StatisticsWindow) (study.Statistics, error) {
	var statistics study.Statistics
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN started_at >= ? AND started_at < ? THEN duration_seconds ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN started_at >= ? AND started_at < ? THEN duration_seconds ELSE 0 END), 0),
			COALESCE(SUM(duration_seconds), 0)
		FROM study_sessions
	`, window.TodayStart.UTC().Format(sqliteTimestampLayout), window.TodayEnd.UTC().Format(sqliteTimestampLayout),
		window.WeekStart.UTC().Format(sqliteTimestampLayout), window.WeekEnd.UTC().Format(sqliteTimestampLayout)).
		Scan(&statistics.TodaySeconds, &statistics.WeekSeconds, &statistics.AllTimeSeconds)
	if err != nil {
		return study.Statistics{}, fmt.Errorf("get study duration totals: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT ss.subject_id, COALESCE(s.name, 'Unassigned'), SUM(ss.duration_seconds)
		FROM study_sessions ss
		LEFT JOIN subjects s ON s.id = ss.subject_id
		GROUP BY ss.subject_id, s.name
		ORDER BY SUM(ss.duration_seconds) DESC, COALESCE(s.name, 'Unassigned') COLLATE NOCASE
	`)
	if err != nil {
		return study.Statistics{}, fmt.Errorf("get study duration by subject: %w", err)
	}
	defer rows.Close()

	statistics.BySubject = make([]study.SubjectStatistics, 0)
	for rows.Next() {
		var subject study.SubjectStatistics
		var subjectID sql.NullInt64
		if err := rows.Scan(&subjectID, &subject.SubjectName, &subject.DurationSeconds); err != nil {
			return study.Statistics{}, fmt.Errorf("scan subject study duration: %w", err)
		}
		if subjectID.Valid {
			subject.SubjectID = &subjectID.Int64
		}
		statistics.BySubject = append(statistics.BySubject, subject)
	}
	if err := rows.Err(); err != nil {
		return study.Statistics{}, fmt.Errorf("iterate subject study durations: %w", err)
	}
	return statistics, nil
}

type sessionScanner interface {
	Scan(...any) error
}

func scanSession(scanner sessionScanner) (study.Session, error) {
	var session study.Session
	var subjectID sql.NullInt64
	var kind, startedAt, endedAt string
	if err := scanner.Scan(
		&session.ID,
		&subjectID,
		&kind,
		&startedAt,
		&endedAt,
		&session.DurationSeconds,
		&session.Note,
	); err != nil {
		return study.Session{}, err
	}
	if subjectID.Valid {
		session.SubjectID = &subjectID.Int64
	}
	session.Kind = study.SessionKind(kind)
	var err error
	session.StartedAt, err = time.Parse(time.RFC3339Nano, startedAt)
	if err != nil {
		return study.Session{}, fmt.Errorf("parse started_at: %w", err)
	}
	session.EndedAt, err = time.Parse(time.RFC3339Nano, endedAt)
	if err != nil {
		return study.Session{}, fmt.Errorf("parse ended_at: %w", err)
	}
	if err := session.Validate(); err != nil {
		return study.Session{}, fmt.Errorf("validate persisted study session: %w", err)
	}
	return session, nil
}
