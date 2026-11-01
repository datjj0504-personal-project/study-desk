package notes

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidNote    = errors.New("invalid note")
	ErrNotFound       = errors.New("note not found")
	ErrInvalidSubject = errors.New("invalid subject")
)

// Subject groups notes for organization.
type Subject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Validate checks the fields required for a subject.
func (s Subject) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return ErrInvalidSubject
	}
	return nil
}

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
