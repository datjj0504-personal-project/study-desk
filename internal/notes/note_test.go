package notes

import (
	"errors"
	"testing"
)

func TestNoteValidate(t *testing.T) {
	tests := []struct {
		name    string
		note    Note
		wantErr bool
	}{
		{
			name: "title and empty body are valid",
			note: Note{Title: "Study plan"},
		},
		{
			name:    "empty title is invalid",
			note:    Note{Body: "Review later"},
			wantErr: true,
		},
		{
			name:    "whitespace title is invalid",
			note:    Note{Title: " \t\n"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.note.Validate()
			if test.wantErr {
				if !errors.Is(err, ErrInvalidNote) {
					t.Errorf("Validate() error = %v, want %v", err, ErrInvalidNote)
				}
				return
			}
			if err != nil {
				t.Errorf("Validate() returned an error: %v", err)
			}
		})
	}
}
