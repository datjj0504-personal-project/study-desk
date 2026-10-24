package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"studydesk/internal/notes"
	"studydesk/internal/storage"
	"studydesk/internal/timer"
)

type appTestClock struct {
	now time.Time
}

func (c *appTestClock) Now() time.Time {
	return c.now
}

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

func TestAppPomodoroControls(t *testing.T) {
	ctx := context.Background()
	clock := &appTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	durations := timer.Durations{
		Focus:          10 * time.Second,
		ShortBreak:     3 * time.Second,
		LongBreak:      8 * time.Second,
		LongBreakEvery: 2,
	}
	pomodoro, err := timer.NewPomodoro(clock, durations)
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	app := &App{ctx: ctx, pomodoro: pomodoro}

	state, err := app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() returned an error: %v", err)
	}
	if state.Phase != timer.PhaseFocus || state.Status != timer.StatusIdle || state.RemainingMillis != 10_000 {
		t.Errorf("initial state = %+v", state)
	}
	if err := app.StartPomodoro(); err != nil {
		t.Fatalf("StartPomodoro() returned an error: %v", err)
	}
	clock.now = clock.now.Add(4 * time.Second)
	if err := app.PausePomodoro(); err != nil {
		t.Fatalf("PausePomodoro() returned an error: %v", err)
	}
	state, err = app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() after pause returned an error: %v", err)
	}
	if state.Status != timer.StatusPaused || state.ElapsedMillis != 4_000 || state.RemainingMillis != 6_000 {
		t.Errorf("paused state = %+v", state)
	}
	if err := app.ResumePomodoro(); err != nil {
		t.Fatalf("ResumePomodoro() returned an error: %v", err)
	}
	clock.now = clock.now.Add(6 * time.Second)
	state, err = app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() at completion returned an error: %v", err)
	}
	if state.Status != timer.StatusCompleted || state.RemainingMillis != 0 {
		t.Errorf("completed state = %+v", state)
	}
	if err := app.AdvancePomodoro(); err != nil {
		t.Fatalf("AdvancePomodoro() returned an error: %v", err)
	}
	state, err = app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() after advancing returned an error: %v", err)
	}
	if state.Phase != timer.PhaseShortBreak || state.Status != timer.StatusIdle || state.DurationMillis != 3_000 || state.CompletedFocuses != 1 {
		t.Errorf("state after advance = %+v", state)
	}

	if err := app.ResetPomodoro(); err != nil {
		t.Fatalf("ResetPomodoro() returned an error: %v", err)
	}
	state, err = app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() after reset returned an error: %v", err)
	}
	if state.Phase != timer.PhaseFocus || state.Status != timer.StatusIdle || state.ElapsedMillis != 0 || state.DurationMillis != 10_000 || state.CompletedFocuses != 0 {
		t.Errorf("reset state = %+v", state)
	}
}

func TestAppPomodoroRejectsInvalidTransitions(t *testing.T) {
	app := NewApp()
	if _, err := app.GetPomodoroState(); err == nil {
		t.Fatal("GetPomodoroState() succeeded before startup")
	}

	pomodoro, err := timer.NewPomodoro(&appTestClock{now: time.Now()}, timer.DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	app.ctx = context.Background()
	app.pomodoro = pomodoro
	if err := app.PausePomodoro(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("PausePomodoro() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
	if err := app.AdvancePomodoro(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("AdvancePomodoro() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
}
