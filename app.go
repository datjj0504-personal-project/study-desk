package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"studydesk/internal/notes"
	"studydesk/internal/storage"
)

const appVersion = "0.1.0"

// AppInfo describes the running application.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// App is the Wails-bound adapter. Methods are added here as features land.
type App struct {
	ctx        context.Context
	db         *sql.DB
	noteRepo   notes.Repository
	startupErr error
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.db, a.startupErr = storage.OpenAppDatabase(ctx)
	if a.startupErr != nil {
		log.Printf("initialize local database: %v", a.startupErr)
		return
	}
	a.noteRepo = storage.NewNoteRepository(a.db)
}

func (a *App) shutdown(context.Context) {
	if a.db == nil {
		return
	}
	if err := a.db.Close(); err != nil {
		log.Printf("close local database: %v", err)
	}
}

// GetAppInfo returns basic application metadata to the frontend.
func (a *App) GetAppInfo() (AppInfo, error) {
	if a.startupErr != nil {
		return AppInfo{}, fmt.Errorf("initialize local database: %w", a.startupErr)
	}
	return AppInfo{Name: "StudyDesk", Version: appVersion}, nil
}

// CreateNote validates and saves a note.
func (a *App) CreateNote(note notes.Note) (notes.Note, error) {
	repo, err := a.noteRepository()
	if err != nil {
		return notes.Note{}, err
	}
	created, err := repo.Create(a.ctx, note)
	if err != nil {
		return notes.Note{}, fmt.Errorf("create note: %w", err)
	}
	return created, nil
}

// GetNote retrieves a note by its ID.
func (a *App) GetNote(id int64) (notes.Note, error) {
	repo, err := a.noteRepository()
	if err != nil {
		return notes.Note{}, err
	}
	note, err := repo.GetByID(a.ctx, id)
	if err != nil {
		return notes.Note{}, fmt.Errorf("get note: %w", err)
	}
	return note, nil
}

// UpdateNote updates a note's title, body, and optional subject.
func (a *App) UpdateNote(note notes.Note) error {
	repo, err := a.noteRepository()
	if err != nil {
		return err
	}
	if err := repo.Update(a.ctx, note); err != nil {
		return fmt.Errorf("update note: %w", err)
	}
	return nil
}

// DeleteNote deletes a note by its ID.
func (a *App) DeleteNote(id int64) error {
	repo, err := a.noteRepository()
	if err != nil {
		return err
	}
	if err := repo.Delete(a.ctx, id); err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return nil
}

// ListNotes returns notes ordered by most recently updated first.
func (a *App) ListNotes() ([]notes.Note, error) {
	repo, err := a.noteRepository()
	if err != nil {
		return nil, err
	}
	result, err := repo.List(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	return result, nil
}

// CreateSubject creates a subject for organizing notes.
func (a *App) CreateSubject(subject notes.Subject) (notes.Subject, error) {
	repo, err := a.noteRepository()
	if err != nil {
		return notes.Subject{}, err
	}
	created, err := repo.CreateSubject(a.ctx, subject)
	if err != nil {
		return notes.Subject{}, fmt.Errorf("create subject: %w", err)
	}
	return created, nil
}

// ListSubjects returns available note subjects.
func (a *App) ListSubjects() ([]notes.Subject, error) {
	repo, err := a.noteRepository()
	if err != nil {
		return nil, err
	}
	result, err := repo.ListSubjects(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	return result, nil
}

func (a *App) noteRepository() (notes.Repository, error) {
	if a.startupErr != nil {
		return nil, fmt.Errorf("initialize local database: %w", a.startupErr)
	}
	if a.noteRepo == nil || a.ctx == nil {
		return nil, errors.New("notes are unavailable before application startup")
	}
	return a.noteRepo, nil
}
