package storage

import (
	"context"
	"errors"
	"testing"

	"studydesk/internal/notes"
)

func TestNoteRepositoryCRUD(t *testing.T) {
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

	repo := NewNoteRepository(db)
	created, err := repo.Create(ctx, notes.Note{Title: "First note", Body: "Initial body"})
	if err != nil {
		t.Fatalf("Create() returned an error: %v", err)
	}
	if created.ID < 1 || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("Create() returned incomplete persisted fields: %+v", created)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() returned an error: %v", err)
	}
	if got.Title != created.Title || got.Body != created.Body {
		t.Errorf("GetByID() = %+v, want title/body %q/%q", got, created.Title, created.Body)
	}

	second, err := repo.Create(ctx, notes.Note{Title: "Second note"})
	if err != nil {
		t.Fatalf("Create() second note returned an error: %v", err)
	}
	if _, err := db.ExecContext(ctx,
		"UPDATE notes SET updated_at = ? WHERE id = ?",
		"2000-01-01T00:00:00.000000000Z", second.ID,
	); err != nil {
		t.Fatalf("make second note older for ordering check: %v", err)
	}
	update := got
	update.Title = "Updated first note"
	update.Body = "Updated body"
	if err := repo.Update(ctx, update); err != nil {
		t.Fatalf("Update() returned an error: %v", err)
	}

	got, err = repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID() after update returned an error: %v", err)
	}
	if got.Title != update.Title || got.Body != update.Body {
		t.Errorf("updated note = %+v, want title/body %q/%q", got, update.Title, update.Body)
	}
	if got.CreatedAt != created.CreatedAt {
		t.Errorf("updated CreatedAt = %v, want preserved value %v", got.CreatedAt, created.CreatedAt)
	}
	if got.UpdatedAt.Before(created.UpdatedAt) {
		t.Errorf("updated UpdatedAt %v is before original %v", got.UpdatedAt, created.UpdatedAt)
	}

	listed, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() returned an error: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("List() returned %d notes, want 2", len(listed))
	}
	if listed[0].ID != created.ID {
		t.Errorf("first listed note ID = %d, want updated note ID %d", listed[0].ID, created.ID)
	}
	if listed[1].ID != second.ID {
		t.Errorf("second listed note ID = %d, want %d", listed[1].ID, second.ID)
	}

	if err := repo.Delete(ctx, second.ID); err != nil {
		t.Fatalf("Delete() returned an error: %v", err)
	}
	if _, err := repo.GetByID(ctx, second.ID); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("GetByID() after delete error = %v, want %v", err, notes.ErrNotFound)
	}
	if err := repo.Delete(ctx, second.ID); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("Delete() for missing note error = %v, want %v", err, notes.ErrNotFound)
	}
}

func TestNoteRepositoryRejectsInvalidNotesAndMissingUpdates(t *testing.T) {
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
	repo := NewNoteRepository(db)

	if _, err := repo.Create(ctx, notes.Note{Title: " \t"}); !errors.Is(err, notes.ErrInvalidNote) {
		t.Errorf("Create() error = %v, want %v", err, notes.ErrInvalidNote)
	}
	if err := repo.Update(ctx, notes.Note{ID: 1, Title: "Missing note"}); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("Update() for missing note error = %v, want %v", err, notes.ErrNotFound)
	}
	if _, err := repo.GetByID(ctx, 1); !errors.Is(err, notes.ErrNotFound) {
		t.Errorf("GetByID() for missing note error = %v, want %v", err, notes.ErrNotFound)
	}
}

func TestNoteRepositorySubjects(t *testing.T) {
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
	repo := NewNoteRepository(db)

	created, err := repo.CreateSubject(ctx, notes.Subject{Name: "  Physics  "})
	if err != nil {
		t.Fatalf("CreateSubject() returned an error: %v", err)
	}
	if created.ID == 0 || created.Name != "Physics" {
		t.Errorf("CreateSubject() = %+v, want a generated ID and trimmed name", created)
	}
	if _, err := repo.CreateSubject(ctx, notes.Subject{Name: "   "}); !errors.Is(err, notes.ErrInvalidSubject) {
		t.Errorf("CreateSubject() for blank name error = %v, want %v", err, notes.ErrInvalidSubject)
	}
	if _, err := repo.CreateSubject(ctx, notes.Subject{Name: "Physics"}); err == nil {
		t.Fatal("CreateSubject() accepted a duplicate subject name")
	}
	if _, err := repo.CreateSubject(ctx, notes.Subject{Name: "Biology"}); err != nil {
		t.Fatalf("CreateSubject() second subject returned an error: %v", err)
	}
	createdNote, err := repo.Create(ctx, notes.Note{
		Title:     "Motion",
		SubjectID: &created.ID,
	})
	if err != nil {
		t.Fatalf("Create() with subject returned an error: %v", err)
	}
	gotNote, err := repo.GetByID(ctx, createdNote.ID)
	if err != nil {
		t.Fatalf("GetByID() for subject note returned an error: %v", err)
	}
	if gotNote.SubjectID == nil || *gotNote.SubjectID != created.ID {
		t.Errorf("persisted note subject ID = %v, want %d", gotNote.SubjectID, created.ID)
	}

	subjects, err := repo.ListSubjects(ctx)
	if err != nil {
		t.Fatalf("ListSubjects() returned an error: %v", err)
	}
	if len(subjects) != 2 || subjects[0].Name != "Biology" || subjects[1].Name != "Physics" {
		t.Errorf("ListSubjects() = %+v, want alphabetical subjects", subjects)
	}
}
