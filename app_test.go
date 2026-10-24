package main

import (
	"context"
	"errors"
	"testing"

	"studydesk/internal/notes"
	"studydesk/internal/storage"
)

func TestGetAppInfo(t *testing.T) {
	app := NewApp()

	got, err := app.GetAppInfo()
	if err != nil {
		t.Fatalf("GetAppInfo() returned an error: %v", err)
	}

	want := AppInfo{Name: "StudyDesk", Version: "0.1.0"}
	if got != want {
		t.Errorf("GetAppInfo() = %+v, want %+v", got, want)
	}
}

func TestAppNoteCRUD(t *testing.T) {
	ctx := context.Background()
	db, err := storage.OpenDatabase(ctx, ":memory:")
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})
	app := &App{
		ctx:      ctx,
		db:       db,
		noteRepo: storage.NewNoteRepository(db),
	}

	created, err := app.CreateNote(notes.Note{Title: "App note", Body: "Content"})
	if err != nil {
		t.Fatalf("CreateNote() returned an error: %v", err)
	}
	got, err := app.GetNote(created.ID)
	if err != nil {
		t.Fatalf("GetNote() returned an error: %v", err)
	}
	if got.Title != "App note" || got.Body != "Content" {
		t.Errorf("GetNote() = %+v, want created note", got)
	}

	got.Body = "Edited content"
	if err := app.UpdateNote(got); err != nil {
		t.Fatalf("UpdateNote() returned an error: %v", err)
	}
	listed, err := app.ListNotes()
	if err != nil {
		t.Fatalf("ListNotes() returned an error: %v", err)
	}
	if len(listed) != 1 || listed[0].Body != "Edited content" {
		t.Errorf("ListNotes() = %+v, want one updated note", listed)
	}
	subject, err := app.CreateSubject(notes.Subject{Name: "Physics"})
	if err != nil {
		t.Fatalf("CreateSubject() returned an error: %v", err)
	}
	got.SubjectID = &subject.ID
	if err := app.UpdateNote(got); err != nil {
		t.Fatalf("UpdateNote() assigning subject returned an error: %v", err)
	}
	subjects, err := app.ListSubjects()
	if err != nil {
		t.Fatalf("ListSubjects() returned an error: %v", err)
	}
	if len(subjects) != 1 || subjects[0].ID != subject.ID {
		t.Errorf("ListSubjects() = %+v, want created subject %+v", subjects, subject)
	}

	if err := app.DeleteNote(created.ID); err != nil {
		t.Fatalf("DeleteNote() returned an error: %v", err)
	}
	if _, err := app.GetNote(created.ID); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("GetNote() after deletion error = %v, want %v", err, notes.ErrNotFound)
	}
}

func TestAppNoteMethodsRequireStartup(t *testing.T) {
	app := NewApp()
	if _, err := app.ListNotes(); err == nil {
		t.Fatal("ListNotes() succeeded before startup")
	}
}
