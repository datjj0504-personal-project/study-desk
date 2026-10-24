package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"studydesk/internal/notes"
	"studydesk/internal/settings"
	"studydesk/internal/storage"
	"studydesk/internal/study"
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

// StopwatchState is the UI-facing stopwatch snapshot. Elapsed time is in milliseconds.
type StopwatchState struct {
	Status        timer.Status `json:"status"`
	ElapsedMillis int64        `json:"elapsedMillis"`
}

// App is the Wails-bound adapter. Methods are added here as features land.
type App struct {
	ctx          context.Context
	db           *sql.DB
	noteRepo     notes.Repository
	sessions     study.SessionRepository
	settingsRepo settings.Repository
	settingsErr  error
	timerStates  timer.StateRepository
	pomodoro     *timer.Timer
	timerErr     error
	pomodoroMu   sync.Mutex
	stopwatch    *timer.Timer
	stopwatchErr error
	stopwatchMu  sync.Mutex
	startupErr   error
	clock        timer.Clock
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.clock = wallClock{}
	a.pomodoro, a.timerErr = timer.NewPomodoro(a.clock, timer.DefaultDurations())
	if a.timerErr != nil {
		log.Printf("initialize Pomodoro timer: %v", a.timerErr)
	}
	a.stopwatch, a.stopwatchErr = timer.NewStopwatch(a.clock)
	if a.stopwatchErr != nil {
		log.Printf("initialize stopwatch: %v", a.stopwatchErr)
	}
	a.db, a.startupErr = storage.OpenAppDatabase(ctx)
	if a.startupErr != nil {
		log.Printf("initialize local database: %v", a.startupErr)
		return
	}
	a.noteRepo = storage.NewNoteRepository(a.db)
	a.sessions = storage.NewSessionRepository(a.db)
	a.settingsRepo = storage.NewSettingsRepository(a.db)
	a.timerStates = storage.NewTimerStateRepository(a.db)
	timerSettings, err := a.settingsRepo.LoadTimerSettings(ctx)
	if err != nil {
		a.settingsErr = err
		a.timerErr = err
		log.Printf("load timer settings: %v", err)
	} else {
		if err := a.pomodoro.Configure(timerSettings.Durations()); err != nil {
			a.timerErr = err
			log.Printf("configure Pomodoro timer: %v", err)
		} else if err := a.restoreTimerState(a.pomodoro); err != nil {
			a.timerErr = err
			log.Printf("restore Pomodoro timer: %v", err)
		}
	}
	if err := a.restoreTimerState(a.stopwatch); err != nil {
		a.stopwatchErr = err
		log.Printf("restore stopwatch: %v", err)
	}
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

// GetStudyStatistics returns study durations for the local day, local week, all time, and each subject.
func (a *App) GetStudyStatistics() (study.Statistics, error) {
	if a.sessions == nil || a.ctx == nil {
		return study.Statistics{}, errors.New("study session persistence is unavailable before application startup")
	}
	statistics, err := a.sessions.GetStatistics(a.ctx, study.NewStatisticsWindow(a.now()))
	if err != nil {
		return study.Statistics{}, fmt.Errorf("get study statistics: %w", err)
	}
	return statistics, nil
}

// GetPomodoroState returns the current countdown and phase state.
func (a *App) GetPomodoroState() (PomodoroState, error) {
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return PomodoroState{}, err
	}
	state := pomodoro.State()
	if state.Status == timer.StatusCompleted && state.Phase == timer.PhaseFocus {
		if err := a.recordCompletedPomodoro(pomodoro.Snapshot()); err != nil {
			return PomodoroState{}, err
		}
	}
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
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	previous := pomodoro.Snapshot()
	if err := a.recordPomodoroIfCompleted(previous); err != nil {
		return err
	}
	if err := pomodoro.Start(); err != nil {
		return fmt.Errorf("start Pomodoro: %w", err)
	}
	return a.persistTimerChange(pomodoro, previous)
}

// PausePomodoro pauses a running Pomodoro phase.
func (a *App) PausePomodoro() error {
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	previous := pomodoro.Snapshot()
	if err := a.recordPomodoroIfCompleted(previous); err != nil {
		return err
	}
	if err := pomodoro.Pause(); err != nil {
		return fmt.Errorf("pause Pomodoro: %w", err)
	}
	return a.persistTimerChange(pomodoro, previous)
}

// ResumePomodoro resumes a paused Pomodoro phase.
func (a *App) ResumePomodoro() error {
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	previous := pomodoro.Snapshot()
	if err := a.recordPomodoroIfCompleted(previous); err != nil {
		return err
	}
	if err := pomodoro.Resume(); err != nil {
		return fmt.Errorf("resume Pomodoro: %w", err)
	}
	return a.persistTimerChange(pomodoro, previous)
}

// ResetPomodoro resets the complete Pomodoro cycle to its initial focus phase.
func (a *App) ResetPomodoro() error {
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	previous := pomodoro.Snapshot()
	if err := a.recordPomodoroIfCompleted(previous); err != nil {
		return err
	}
	pomodoro.Reset()
	return a.persistTimerChange(pomodoro, previous)
}

// AdvancePomodoro moves a completed phase to the next phase.
func (a *App) AdvancePomodoro() error {
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	previous := pomodoro.Snapshot()
	if err := a.recordPomodoroIfCompleted(previous); err != nil {
		return err
	}
	if err := pomodoro.Advance(); err != nil {
		return fmt.Errorf("advance Pomodoro: %w", err)
	}
	return a.persistTimerChange(pomodoro, previous)
}

// GetTimerSettings returns the persisted Pomodoro configuration.
func (a *App) GetTimerSettings() (settings.TimerSettings, error) {
	repo, err := a.timerSettingsRepository()
	if err != nil {
		return settings.TimerSettings{}, err
	}
	value, err := repo.LoadTimerSettings(a.ctx)
	if err != nil {
		return settings.TimerSettings{}, fmt.Errorf("load timer settings: %w", err)
	}
	return value, nil
}

// UpdateTimerSettings persists the Pomodoro configuration when the timer is idle.
func (a *App) UpdateTimerSettings(value settings.TimerSettings) error {
	if err := value.Validate(); err != nil {
		return err
	}
	repo, err := a.timerSettingsRepository()
	if err != nil {
		return err
	}
	a.pomodoroMu.Lock()
	defer a.pomodoroMu.Unlock()
	pomodoro, err := a.pomodoroTimer()
	if err != nil {
		return err
	}
	current, err := repo.LoadTimerSettings(a.ctx)
	if err != nil {
		return fmt.Errorf("load timer settings before update: %w", err)
	}
	durationsChanged := value.FocusMinutes != current.FocusMinutes ||
		value.ShortBreakMinutes != current.ShortBreakMinutes ||
		value.LongBreakMinutes != current.LongBreakMinutes ||
		value.LongBreakEvery != current.LongBreakEvery
	if durationsChanged && pomodoro.State().Status != timer.StatusIdle {
		return fmt.Errorf("update timer settings: %w: reset the timer before changing durations", timer.ErrInvalidTransition)
	}
	if err := repo.SaveTimerSettings(a.ctx, value); err != nil {
		return fmt.Errorf("save timer settings: %w", err)
	}
	if durationsChanged {
		previous := pomodoro.Snapshot()
		if err := pomodoro.Configure(value.Durations()); err != nil {
			if rollbackErr := repo.SaveTimerSettings(a.ctx, current); rollbackErr != nil {
				return fmt.Errorf("configure timer settings: %w (restore previous settings: %v)", err, rollbackErr)
			}
			return fmt.Errorf("configure timer settings: %w", err)
		}
		if a.timerStates == nil || a.ctx == nil {
			err = errors.New("timer state persistence is unavailable before application startup")
		} else {
			err = a.timerStates.SaveTimerState(a.ctx, pomodoro.Snapshot())
		}
		if err != nil {
			if restoreErr := pomodoro.Configure(current.Durations()); restoreErr != nil {
				return fmt.Errorf("save updated timer state: %w (restore previous timer configuration: %v)", err, restoreErr)
			}
			if restoreErr := pomodoro.Restore(previous); restoreErr != nil {
				return fmt.Errorf("save updated timer state: %w (restore previous timer state: %v)", err, restoreErr)
			}
			if rollbackErr := repo.SaveTimerSettings(a.ctx, current); rollbackErr != nil {
				return fmt.Errorf("save updated timer state: %w (restore previous settings: %v)", err, rollbackErr)
			}
			return fmt.Errorf("save updated timer state: %w", err)
		}
	}
	return nil
}

// GetStopwatchState returns the current stopwatch status and elapsed time.
func (a *App) GetStopwatchState() (StopwatchState, error) {
	a.stopwatchMu.Lock()
	defer a.stopwatchMu.Unlock()
	stopwatch, err := a.stopwatchTimer()
	if err != nil {
		return StopwatchState{}, err
	}
	state := stopwatch.State()
	return StopwatchState{
		Status:        state.Status,
		ElapsedMillis: state.Elapsed.Milliseconds(),
	}, nil
}

// StartStopwatch starts an idle stopwatch.
func (a *App) StartStopwatch() error {
	a.stopwatchMu.Lock()
	defer a.stopwatchMu.Unlock()
	stopwatch, err := a.stopwatchTimer()
	if err != nil {
		return err
	}
	previous := stopwatch.Snapshot()
	if err := stopwatch.Start(); err != nil {
		return fmt.Errorf("start stopwatch: %w", err)
	}
	return a.persistTimerChange(stopwatch, previous)
}

// PauseStopwatch pauses a running stopwatch.
func (a *App) PauseStopwatch() error {
	a.stopwatchMu.Lock()
	defer a.stopwatchMu.Unlock()
	stopwatch, err := a.stopwatchTimer()
	if err != nil {
		return err
	}
	previous := stopwatch.Snapshot()
	if err := stopwatch.Pause(); err != nil {
		return fmt.Errorf("pause stopwatch: %w", err)
	}
	return a.persistTimerChange(stopwatch, previous)
}

// ResumeStopwatch resumes a paused stopwatch.
func (a *App) ResumeStopwatch() error {
	a.stopwatchMu.Lock()
	defer a.stopwatchMu.Unlock()
	stopwatch, err := a.stopwatchTimer()
	if err != nil {
		return err
	}
	previous := stopwatch.Snapshot()
	if err := stopwatch.Resume(); err != nil {
		return fmt.Errorf("resume stopwatch: %w", err)
	}
	return a.persistTimerChange(stopwatch, previous)
}

// ResetStopwatch resets the stopwatch to idle with no elapsed time.
func (a *App) ResetStopwatch() error {
	a.stopwatchMu.Lock()
	defer a.stopwatchMu.Unlock()
	stopwatch, err := a.stopwatchTimer()
	if err != nil {
		return err
	}
	previous := stopwatch.Snapshot()
	if previous.SessionStartedAtUnixNano != 0 && previous.Elapsed > 0 {
		if err := a.recordCompletedStopwatch(previous); err != nil {
			return err
		}
	}
	stopwatch.Reset()
	return a.persistTimerChange(stopwatch, previous)
}

func (a *App) recordCompletedPomodoro(snapshot timer.Snapshot) error {
	if a.sessions == nil || a.timerStates == nil || a.ctx == nil {
		return errors.New("study session persistence is unavailable before application startup")
	}
	if snapshot.SessionStartedAtUnixNano == 0 {
		return errors.New("completed Pomodoro phase has no start time")
	}
	persisted, found, err := a.timerStates.LoadTimerState(a.ctx, timer.ModePomodoro)
	if err != nil {
		return fmt.Errorf("check completed Pomodoro persistence: %w", err)
	}
	if found && persisted.Status == timer.StatusCompleted &&
		persisted.Phase == timer.PhaseFocus &&
		persisted.SessionStartedAtUnixNano == snapshot.SessionStartedAtUnixNano {
		return nil
	}
	startedAt := time.Unix(0, snapshot.SessionStartedAtUnixNano).UTC()
	endedAt := a.now().UTC()
	if endedAt.Before(startedAt) {
		endedAt = startedAt
	}
	session := study.Session{
		Kind:            study.SessionKindPomodoro,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		DurationSeconds: int64(snapshot.Elapsed / time.Second),
	}
	completionKey := fmt.Sprintf("pomodoro:%d", snapshot.SessionStartedAtUnixNano)
	if _, err := a.sessions.CreateCompletedSession(a.ctx, session, completionKey); err != nil {
		return fmt.Errorf("record completed Pomodoro session: %w", err)
	}
	if err := a.timerStates.SaveTimerState(a.ctx, snapshot); err != nil {
		return fmt.Errorf("save completed Pomodoro state: %w", err)
	}
	return nil
}

func (a *App) recordPomodoroIfCompleted(snapshot timer.Snapshot) error {
	if snapshot.Status == timer.StatusCompleted && snapshot.Phase == timer.PhaseFocus {
		return a.recordCompletedPomodoro(snapshot)
	}
	return nil
}

func (a *App) recordCompletedStopwatch(snapshot timer.Snapshot) error {
	if a.sessions == nil || a.ctx == nil {
		return errors.New("study session persistence is unavailable before application startup")
	}
	startedAt := time.Unix(0, snapshot.SessionStartedAtUnixNano).UTC()
	endedAt := a.now().UTC()
	if endedAt.Before(startedAt) {
		endedAt = startedAt
	}
	session := study.Session{
		Kind:            study.SessionKindStopwatch,
		StartedAt:       startedAt,
		EndedAt:         endedAt,
		DurationSeconds: int64(snapshot.Elapsed / time.Second),
	}
	completionKey := fmt.Sprintf("stopwatch:%d", snapshot.SessionStartedAtUnixNano)
	if _, err := a.sessions.CreateCompletedSession(a.ctx, session, completionKey); err != nil {
		return fmt.Errorf("record completed stopwatch session: %w", err)
	}
	return nil
}

func (a *App) now() time.Time {
	if a.clock == nil {
		return time.Now()
	}
	return a.clock.Now()
}

func (a *App) restoreTimerState(value *timer.Timer) error {
	if a.timerStates == nil || a.ctx == nil {
		return errors.New("timer state persistence is unavailable before application startup")
	}
	mode := value.State().Mode
	snapshot, found, err := a.timerStates.LoadTimerState(a.ctx, mode)
	if err != nil {
		return fmt.Errorf("load %s state: %w", mode, err)
	}
	if !found {
		return nil
	}
	if err := value.Restore(snapshot); err != nil {
		return fmt.Errorf("restore %s state: %w", mode, err)
	}
	return nil
}

func (a *App) persistTimerChange(value *timer.Timer, previous timer.Snapshot) error {
	if a.timerStates == nil || a.ctx == nil {
		return errors.New("timer state persistence is unavailable before application startup")
	}
	snapshot := value.Snapshot()
	if err := a.timerStates.SaveTimerState(a.ctx, snapshot); err != nil {
		if restoreErr := value.Restore(previous); restoreErr != nil {
			return fmt.Errorf("save %s state: %w (restore previous state: %v)", snapshot.Mode, err, restoreErr)
		}
		return fmt.Errorf("save %s state: %w", snapshot.Mode, err)
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

func (a *App) timerSettingsRepository() (settings.Repository, error) {
	if a.startupErr != nil {
		return nil, fmt.Errorf("initialize local database: %w", a.startupErr)
	}
	if a.settingsErr != nil {
		return nil, fmt.Errorf("load timer settings: %w", a.settingsErr)
	}
	if a.settingsRepo == nil || a.ctx == nil {
		return nil, errors.New("timer settings are unavailable before application startup")
	}
	return a.settingsRepo, nil
}

func (a *App) stopwatchTimer() (*timer.Timer, error) {
	if a.stopwatchErr != nil {
		return nil, fmt.Errorf("initialize stopwatch: %w", a.stopwatchErr)
	}
	if a.stopwatch == nil || a.ctx == nil {
		return nil, errors.New("stopwatch is unavailable before application startup")
	}
	return a.stopwatch, nil
}

type wallClock struct{}

func (wallClock) Now() time.Time {
	return time.Now()
}
