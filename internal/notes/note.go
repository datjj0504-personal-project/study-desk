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
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	SubjectID *int64    `json:"subjectId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Validate checks the fields required for a note.
func (n Note) Validate() error {
	if strings.TrimSpace(n.Title) == "" {
		return ErrInvalidNote
	}
	return nil
}
