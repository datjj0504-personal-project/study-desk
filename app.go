package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"studydesk/internal/notes"
	"studydesk/internal/storage"
	"studydesk/internal/timer"
)

const appVersion = "0.1.0"

// AppInfo describes the running application.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// PomodoroState is the UI-facing timer snapshot. Durations are in milliseconds.
type PomodoroState struct {
	Phase            timer.Phase  `json:"phase"`
	Status           timer.Status `json:"status"`
	DurationMillis   int64        `json:"durationMillis"`
	ElapsedMillis    int64        `json:"elapsedMillis"`
	RemainingMillis  int64        `json:"remainingMillis"`
	CompletedFocuses int          `json:"completedFocuses"`
}

// App is the Wails-bound adapter. Methods are added here as features land.
type App struct {
	ctx        context.Context
	db         *sql.DB
	noteRepo   notes.Repository
	pomodoro   *timer.Timer
	timerErr   error
	startupErr error
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.pomodoro, a.timerErr = timer.NewPomodoro(wallClock{}, timer.DefaultDurations())
	if a.timerErr != nil {
		log.Printf("initialize Pomodoro timer: %v", a.timerErr)
	}
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

// GetPomodoroState returns the current countdown and phase state.
func (a *App) GetPomodoroState() (PomodoroState, error) {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return PomodoroState{}, err
	}
	state := pomodoro.State()
	return PomodoroState{
		Phase:            state.Phase,
		Status:           state.Status,
		DurationMillis:   state.Duration.Milliseconds(),
		ElapsedMillis:    state.Elapsed.Milliseconds(),
		RemainingMillis:  state.Remaining.Milliseconds(),
		CompletedFocuses: state.CompletedFocuses,
	}, nil
}

// StartPomodoro starts an idle Pomodoro phase.
func (a *App) StartPomodoro() error {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	if err := pomodoro.Start(); err != nil {
		return fmt.Errorf("start Pomodoro: %w", err)
	}
	return nil
}

// PausePomodoro pauses a running Pomodoro phase.
func (a *App) PausePomodoro() error {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	if err := pomodoro.Pause(); err != nil {
		return fmt.Errorf("pause Pomodoro: %w", err)
	}
	return nil
}

// ResumePomodoro resumes a paused Pomodoro phase.
func (a *App) ResumePomodoro() error {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	if err := pomodoro.Resume(); err != nil {
		return fmt.Errorf("resume Pomodoro: %w", err)
	}
	return nil
}

// ResetPomodoro resets the complete Pomodoro cycle to its initial focus phase.
func (a *App) ResetPomodoro() error {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	pomodoro.Reset()
	return nil
}

// AdvancePomodoro moves a completed phase to the next phase.
func (a *App) AdvancePomodoro() error {
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	if err := pomodoro.Advance(); err != nil {
		return fmt.Errorf("advance Pomodoro: %w", err)
	}
	return nil
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

func (a *App) pomodoroTimer() (*timer.Timer, error) {
	if a.timerErr != nil {
		return nil, fmt.Errorf("initialize Pomodoro timer: %w", a.timerErr)
	}
	if a.pomodoro == nil || a.ctx == nil {
		return nil, errors.New("Pomodoro timer is unavailable before application startup")
	}
	return a.pomodoro, nil
}

type wallClock struct{}

func (wallClock) Now() time.Time {
	return time.Now()
}
