package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"studydesk/internal/notes"
)

const sqliteTimestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

// NoteRepository stores notes in SQLite.
type NoteRepository struct {
	db *sql.DB
}

var _ notes.Repository = (*NoteRepository)(nil)

// NewNoteRepository creates a SQLite-backed notes repository.
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Create validates and stores a note, assigning its ID and timestamps.
func (r *NoteRepository) Create(ctx context.Context, note notes.Note) (notes.Note, error) {
	if err := note.Validate(); err != nil {
		return notes.Note{}, err
	}
	now := time.Now().UTC()
	timestamp := now.Format(sqliteTimestampLayout)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO notes (title, body, subject_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, note.Title, note.Body, note.SubjectID, timestamp, timestamp)
	if err != nil {
		return notes.Note{}, fmt.Errorf("create note: %w", err)
	}
	note.ID, err = result.LastInsertId()
	if err != nil {
		return notes.Note{}, fmt.Errorf("get created note ID: %w", err)
	}
	note.CreatedAt = now
	note.UpdatedAt = now
	return note, nil
}

// GetByID returns the note with the given ID.
func (r *NoteRepository) GetByID(ctx context.Context, id int64) (notes.Note, error) {
	note, err := scanNote(r.db.QueryRowContext(ctx, `
		SELECT id, title, body, subject_id, created_at, updated_at
		FROM notes WHERE id = ?
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return notes.Note{}, notes.ErrNotFound
	}
	if err != nil {
		return notes.Note{}, fmt.Errorf("get note %d: %w", id, err)
	}
	return note, nil
}

// Update changes the content and optional subject of a note. Creation time is
// preserved and update time is set to the current UTC time.
func (r *NoteRepository) Update(ctx context.Context, note notes.Note) error {
	if err := note.Validate(); err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE notes
		SET title = ?, body = ?, subject_id = ?, updated_at = ?
		WHERE id = ?
	`, note.Title, note.Body, note.SubjectID, time.Now().UTC().Format(sqliteTimestampLayout), note.ID)
	if err != nil {
		return fmt.Errorf("update note %d: %w", note.ID, err)
	}
	if err := requireAffectedRow(result, note.ID); err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

// Delete removes the note with the given ID.
func (r *NoteRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete note %d: %w", id, err)
	}
	if err := requireAffectedRow(result, id); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}

// List returns notes ordered by most recently updated first.
func (r *NoteRepository) List(ctx context.Context) ([]notes.Note, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, body, subject_id, created_at, updated_at
		FROM notes
		ORDER BY updated_at DESC, id DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	result := make([]notes.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		result = append(result, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}
	return result, nil
}

// CreateSubject stores a subject and returns its generated ID.
func (r *NoteRepository) CreateSubject(ctx context.Context, subject notes.Subject) (notes.Subject, error) {
	if err := subject.Validate(); err != nil {
		return notes.Subject{}, err
	}
	subject.Name = strings.TrimSpace(subject.Name)
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO subjects (name, created_at) VALUES (?, ?)",
		subject.Name,
		time.Now().UTC().Format(sqliteTimestampLayout),
	)
	if err != nil {
		return notes.Subject{}, fmt.Errorf("create subject: %w", err)
	}
	subject.ID, err = result.LastInsertId()
	if err != nil {
		return notes.Subject{}, fmt.Errorf("get created subject ID: %w", err)
	}
	return subject, nil
}

// ListSubjects returns all subjects ordered alphabetically by name.
func (r *NoteRepository) ListSubjects(ctx context.Context) ([]notes.Subject, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM subjects ORDER BY name COLLATE NOCASE, id")
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	subjects := make([]notes.Subject, 0)
	for rows.Next() {
		var subject notes.Subject
		if err := rows.Scan(&subject.ID, &subject.Name); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		subjects = append(subjects, subject)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}
	return subjects, nil
}

type noteScanner interface {
	Scan(...any) error
}

func scanNote(scanner noteScanner) (notes.Note, error) {
	var note notes.Note
	var subjectID sql.NullInt64
	var createdAt, updatedAt string
	if err := scanner.Scan(
		&note.ID,
		&note.Title,
		&note.Body,
		&subjectID,
		&createdAt,
		&updatedAt,
	); err != nil {
		return notes.Note{}, err
	}
	if subjectID.Valid {
		note.SubjectID = &subjectID.Int64
	}
	var err error
	note.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return notes.Note{}, fmt.Errorf("parse created_at: %w", err)
	}
	note.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return notes.Note{}, fmt.Errorf("parse updated_at: %w", err)
	}
	return note, nil
}

func requireAffectedRow(result sql.Result, id int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check affected rows: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("note %d: %w", id, notes.ErrNotFound)
	}
	return nil
}
