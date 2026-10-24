package notes

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidNote = errors.New("invalid note")
	ErrNotFound    = errors.New("note not found")
)

// Note is a persisted study note.
type Note struct {
	ID        int64
	Title     string
	Body      string
	SubjectID *int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks the fields required for a note.
func (n Note) Validate() error {
	if strings.TrimSpace(n.Title) == "" {
		return ErrInvalidNote
	}
	return nil
}
