package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"studydesk/internal/notes"
	"studydesk/internal/settings"
	"studydesk/internal/storage"
	"studydesk/internal/timer"
)

type appTestClock struct {
	now time.Time
}

type failingTimerStateRepository struct {
	err error
}

func (r failingTimerStateRepository) LoadTimerState(context.Context, timer.Mode) (timer.Snapshot, bool, error) {
	return timer.Snapshot{}, false, r.err
}

func (r failingTimerStateRepository) SaveTimerState(context.Context, timer.Snapshot) error {
	return r.err
}

func (c *appTestClock) Now() time.Time {
	return c.now
}

func newAppTimerStateRepository(t *testing.T) timer.StateRepository {
	t.Helper()
	db, err := storage.OpenDatabase(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("OpenDatabase() returned an error: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close SQLite database: %v", err)
		}
	})
	return storage.NewTimerStateRepository(db)
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
	app := &App{ctx: ctx, pomodoro: pomodoro, timerStates: newAppTimerStateRepository(t)}

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
	app.timerStates = newAppTimerStateRepository(t)
	if err := app.PausePomodoro(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("PausePomodoro() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
	if err := app.AdvancePomodoro(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("AdvancePomodoro() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
}

func TestAppTimerSettingsPersistAndRequireIdleForDurationChanges(t *testing.T) {
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
	pomodoro, err := timer.NewPomodoro(&appTestClock{now: time.Now()}, timer.DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	app := &App{
		ctx:          ctx,
		settingsRepo: storage.NewSettingsRepository(db),
		pomodoro:     pomodoro,
		timerStates:  storage.NewTimerStateRepository(db),
	}

	updated := settings.TimerSettings{
		FocusMinutes: 40, ShortBreakMinutes: 7, LongBreakMinutes: 18,
		LongBreakEvery: 3, NotificationsEnabled: true,
	}
	if err := app.UpdateTimerSettings(updated); err != nil {
		t.Fatalf("UpdateTimerSettings() returned an error: %v", err)
	}
	got, err := app.GetTimerSettings()
	if err != nil {
		t.Fatalf("GetTimerSettings() returned an error: %v", err)
	}
	if got != updated {
		t.Errorf("GetTimerSettings() = %+v, want %+v", got, updated)
	}
	state, err := app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() after update returned an error: %v", err)
	}
	if state.DurationMillis != 40*60*1000 {
		t.Errorf("updated focus duration = %dms, want 2400000ms", state.DurationMillis)
	}

	if err := app.StartPomodoro(); err != nil {
		t.Fatalf("StartPomodoro() returned an error: %v", err)
	}
	notificationOnly := updated
	notificationOnly.NotificationsEnabled = false
	if err := app.UpdateTimerSettings(notificationOnly); err != nil {
		t.Fatalf("UpdateTimerSettings() changing only notifications while running returned an error: %v", err)
	}
	changedDurations := notificationOnly
	changedDurations.FocusMinutes = 45
	if err := app.UpdateTimerSettings(changedDurations); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("UpdateTimerSettings() changing durations while running error = %v, want %v", err, timer.ErrInvalidTransition)
	}
	got, err = app.GetTimerSettings()
	if err != nil {
		t.Fatalf("GetTimerSettings() after rejected duration change returned an error: %v", err)
	}
	if got != notificationOnly {
		t.Errorf("settings after rejected duration change = %+v, want %+v", got, notificationOnly)
	}
}

func TestAppRestoresTimersAfterRestartAndSleep(t *testing.T) {
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
	clock := &appTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	settingsRepo := storage.NewSettingsRepository(db)
	stateRepo := storage.NewTimerStateRepository(db)
	pomodoro, err := timer.NewPomodoro(clock, timer.DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	stopwatch, err := timer.NewStopwatch(clock)
	if err != nil {
		t.Fatalf("NewStopwatch() returned an error: %v", err)
	}
	firstRun := &App{
		ctx: ctx, settingsRepo: settingsRepo, timerStates: stateRepo,
		pomodoro: pomodoro, stopwatch: stopwatch,
	}
	if err := firstRun.StartPomodoro(); err != nil {
		t.Fatalf("StartPomodoro() returned an error: %v", err)
	}
	if err := firstRun.StartStopwatch(); err != nil {
		t.Fatalf("StartStopwatch() returned an error: %v", err)
	}
	clock.now = clock.now.Add(2 * time.Minute)

	restartedPomodoro, err := timer.NewPomodoro(clock, timer.DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() after restart returned an error: %v", err)
	}
	restartedStopwatch, err := timer.NewStopwatch(clock)
	if err != nil {
		t.Fatalf("NewStopwatch() after restart returned an error: %v", err)
	}
	restored := &App{
		ctx: ctx, settingsRepo: settingsRepo, timerStates: stateRepo,
		pomodoro: restartedPomodoro, stopwatch: restartedStopwatch,
	}
	loadedSettings, err := settingsRepo.LoadTimerSettings(ctx)
	if err != nil {
		t.Fatalf("LoadTimerSettings() returned an error: %v", err)
	}
	if err := restartedPomodoro.Configure(loadedSettings.Durations()); err != nil {
		t.Fatalf("Configure() returned an error: %v", err)
	}
	if err := restored.restoreTimerState(restartedPomodoro); err != nil {
		t.Fatalf("restore Pomodoro state returned an error: %v", err)
	}
	if err := restored.restoreTimerState(restartedStopwatch); err != nil {
		t.Fatalf("restore stopwatch state returned an error: %v", err)
	}

	clock.now = clock.now.Add(24 * time.Minute)
	pomodoroState, err := restored.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() after restart returned an error: %v", err)
	}
	if pomodoroState.Status != timer.StatusCompleted || pomodoroState.ElapsedMillis != 25*60*1000 {
		t.Errorf("Pomodoro state after restart and sleep = %+v", pomodoroState)
	}
	stopwatchState, err := restored.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() after restart returned an error: %v", err)
	}
	if stopwatchState.Status != timer.StatusRunning || stopwatchState.ElapsedMillis != 26*60*1000 {
		t.Errorf("stopwatch state after restart and sleep = %+v", stopwatchState)
	}
}

func TestAppTimerControlRollsBackWhenPersistenceFails(t *testing.T) {
	ctx := context.Background()
	pomodoro, err := timer.NewPomodoro(&appTestClock{now: time.Now()}, timer.DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	app := &App{
		ctx:         ctx,
		pomodoro:    pomodoro,
		timerStates: failingTimerStateRepository{err: errors.New("disk unavailable")},
	}

	if err := app.StartPomodoro(); err == nil {
		t.Fatal("StartPomodoro() succeeded despite persistence failure")
	}
	state, err := app.GetPomodoroState()
	if err != nil {
		t.Fatalf("GetPomodoroState() returned an error: %v", err)
	}
	if state.Status != timer.StatusIdle {
		t.Errorf("timer status after failed persistence = %s, want idle", state.Status)
	}
}

func TestAppStopwatchControls(t *testing.T) {
	ctx := context.Background()
	clock := &appTestClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	stopwatch, err := timer.NewStopwatch(clock)
	if err != nil {
		t.Fatalf("NewStopwatch() returned an error: %v", err)
	}
	app := &App{ctx: ctx, stopwatch: stopwatch, timerStates: newAppTimerStateRepository(t)}

	state, err := app.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() returned an error: %v", err)
	}
	if state.Status != timer.StatusIdle || state.ElapsedMillis != 0 {
		t.Errorf("initial state = %+v, want idle with no elapsed time", state)
	}
	if err := app.StartStopwatch(); err != nil {
		t.Fatalf("StartStopwatch() returned an error: %v", err)
	}
	clock.now = clock.now.Add(4*time.Second + 500*time.Millisecond)
	state, err = app.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() while running returned an error: %v", err)
	}
	if state.Status != timer.StatusRunning || state.ElapsedMillis != 4_500 {
		t.Errorf("running state = %+v, want running with 4500ms elapsed", state)
	}

	if err := app.PauseStopwatch(); err != nil {
		t.Fatalf("PauseStopwatch() returned an error: %v", err)
	}
	clock.now = clock.now.Add(3 * time.Second)
	state, err = app.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() while paused returned an error: %v", err)
	}
	if state.Status != timer.StatusPaused || state.ElapsedMillis != 4_500 {
		t.Errorf("paused state = %+v, want paused with 4500ms elapsed", state)
	}

	if err := app.ResumeStopwatch(); err != nil {
		t.Fatalf("ResumeStopwatch() returned an error: %v", err)
	}
	clock.now = clock.now.Add(2 * time.Second)
	state, err = app.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() after resume returned an error: %v", err)
	}
	if state.Status != timer.StatusRunning || state.ElapsedMillis != 6_500 {
		t.Errorf("resumed state = %+v, want running with 6500ms elapsed", state)
	}

	if err := app.ResetStopwatch(); err != nil {
		t.Fatalf("ResetStopwatch() returned an error: %v", err)
	}
	state, err = app.GetStopwatchState()
	if err != nil {
		t.Fatalf("GetStopwatchState() after reset returned an error: %v", err)
	}
	if state.Status != timer.StatusIdle || state.ElapsedMillis != 0 {
		t.Errorf("reset state = %+v, want idle with no elapsed time", state)
	}
}

func TestAppStopwatchRejectsInvalidTransitions(t *testing.T) {
	app := NewApp()
	if _, err := app.GetStopwatchState(); err == nil {
		t.Fatal("GetStopwatchState() succeeded before startup")
	}

	stopwatch, err := timer.NewStopwatch(&appTestClock{now: time.Now()})
	if err != nil {
		t.Fatalf("NewStopwatch() returned an error: %v", err)
	}
	app.ctx = context.Background()
	app.stopwatch = stopwatch
	app.timerStates = newAppTimerStateRepository(t)
	if err := app.PauseStopwatch(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("PauseStopwatch() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
	if err := app.ResumeStopwatch(); !errors.Is(err, timer.ErrInvalidTransition) {
		t.Errorf("ResumeStopwatch() while idle error = %v, want %v", err, timer.ErrInvalidTransition)
	}
}
