package timer

import (
	"errors"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.now
}

func (c *fakeClock) Advance(duration time.Duration) {
	c.now = c.now.Add(duration)
}

func TestNewPomodoroValidatesConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		clock     Clock
		durations Durations
	}{
		{name: "missing clock", durations: DefaultDurations()},
		{name: "zero focus duration", clock: &fakeClock{}, durations: Durations{ShortBreak: time.Minute, LongBreak: time.Minute, LongBreakEvery: 4}},
		{name: "zero break duration", clock: &fakeClock{}, durations: Durations{Focus: time.Minute, LongBreak: time.Minute, LongBreakEvery: 4}},
		{name: "zero long break interval", clock: &fakeClock{}, durations: Durations{Focus: time.Minute, ShortBreak: time.Minute, LongBreak: time.Minute}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewPomodoro(test.clock, test.durations); !errors.Is(err, ErrInvalidConfiguration) {
				t.Errorf("NewPomodoro() error = %v, want %v", err, ErrInvalidConfiguration)
			}
		})
	}
}

func TestPomodoroPauseResumeAndCompletion(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	durations := Durations{Focus: 10 * time.Minute, ShortBreak: 3 * time.Minute, LongBreak: 7 * time.Minute, LongBreakEvery: 4}
	timer, err := NewPomodoro(clock, durations)
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}

	initial := timer.State()
	if initial.Status != StatusIdle || initial.Phase != PhaseFocus || initial.Remaining != durations.Focus {
		t.Fatalf("initial State() = %+v", initial)
	}
	if err := timer.Start(); err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	clock.Advance(4 * time.Minute)
	if err := timer.Pause(); err != nil {
		t.Fatalf("Pause() returned an error: %v", err)
	}
	paused := timer.State()
	if paused.Status != StatusPaused || paused.Elapsed != 4*time.Minute || paused.Remaining != 6*time.Minute {
		t.Errorf("paused State() = %+v", paused)
	}

	clock.Advance(30 * time.Minute)
	if got := timer.State().Remaining; got != 6*time.Minute {
		t.Errorf("paused remaining after clock advance = %v, want 6m", got)
	}
	if err := timer.Resume(); err != nil {
		t.Fatalf("Resume() returned an error: %v", err)
	}
	clock.Advance(6 * time.Minute)
	completed := timer.State()
	if completed.Status != StatusCompleted || completed.Elapsed != durations.Focus || completed.Remaining != 0 {
		t.Errorf("completed State() = %+v", completed)
	}
}

func TestPomodoroUsesDefaultDurations(t *testing.T) {
	timer, err := NewPomodoro(&fakeClock{}, DefaultDurations())
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}
	state := timer.State()
	if state.Duration != 25*time.Minute || state.Remaining != 25*time.Minute {
		t.Errorf("default focus state = %+v, want 25 minutes", state)
	}
}

func TestPomodoroPhaseCycle(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	durations := Durations{Focus: time.Minute, ShortBreak: 10 * time.Second, LongBreak: 30 * time.Second, LongBreakEvery: 2}
	timer, err := NewPomodoro(clock, durations)
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}

	completePhase := func(wantPhase Phase, wantCompleted int) {
		t.Helper()
		if err := timer.Start(); err != nil {
			t.Fatalf("Start() returned an error: %v", err)
		}
		clock.Advance(timer.State().Remaining)
		state := timer.State()
		if state.Status != StatusCompleted {
			t.Fatalf("phase did not complete: %+v", state)
		}
		if err := timer.Advance(); err != nil {
			t.Fatalf("Advance() returned an error: %v", err)
		}
		state = timer.State()
		if state.Phase != wantPhase || state.CompletedFocuses != wantCompleted || state.Status != StatusIdle {
			t.Errorf("state after advance = %+v, want phase %s and %d completed focuses", state, wantPhase, wantCompleted)
		}
	}

	completePhase(PhaseShortBreak, 1)
	completePhase(PhaseFocus, 1)
	completePhase(PhaseLongBreak, 2)
}

func TestStopwatchPauseResumeAndReset(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	timer, err := NewStopwatch(clock)
	if err != nil {
		t.Fatalf("NewStopwatch() returned an error: %v", err)
	}
	if err := timer.Start(); err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	clock.Advance(5 * time.Second)
	if err := timer.Pause(); err != nil {
		t.Fatalf("Pause() returned an error: %v", err)
	}
	if state := timer.State(); state.Elapsed != 5*time.Second || state.Status != StatusPaused {
		t.Errorf("paused stopwatch State() = %+v", state)
	}

	clock.Advance(time.Hour)
	if state := timer.State(); state.Elapsed != 5*time.Second {
		t.Errorf("paused stopwatch elapsed after clock advance = %v, want 5s", state.Elapsed)
	}
	if err := timer.Resume(); err != nil {
		t.Fatalf("Resume() returned an error: %v", err)
	}
	clock.Advance(2 * time.Second)
	if state := timer.State(); state.Elapsed != 7*time.Second || state.Status != StatusRunning {
		t.Errorf("resumed stopwatch State() = %+v", state)
	}

	timer.Reset()
	if state := timer.State(); state.Status != StatusIdle || state.Elapsed != 0 || state.Phase != PhaseStopwatch {
		t.Errorf("reset stopwatch State() = %+v", state)
	}
}

func TestInvalidTransitionsAndReset(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	timer, err := NewPomodoro(clock, Durations{Focus: time.Minute, ShortBreak: time.Minute, LongBreak: time.Minute, LongBreakEvery: 4})
	if err != nil {
		t.Fatalf("NewPomodoro() returned an error: %v", err)
	}

	if err := timer.Pause(); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("Pause() while idle error = %v, want %v", err, ErrInvalidTransition)
	}
	if err := timer.Resume(); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("Resume() while idle error = %v, want %v", err, ErrInvalidTransition)
	}
	if err := timer.Advance(); !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("Advance() while idle error = %v, want %v", err, ErrInvalidTransition)
	}

	if err := timer.Start(); err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	clock.Advance(10 * time.Second)
	timer.Reset()
	state := timer.State()
	if state.Status != StatusIdle || state.Phase != PhaseFocus || state.Elapsed != 0 || state.Remaining != time.Minute {
		t.Errorf("reset Pomodoro State() = %+v", state)
	}
}

func TestClockMovingBackwardDoesNotReduceElapsedBelowZero(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	timer, err := NewStopwatch(clock)
	if err != nil {
		t.Fatalf("NewStopwatch() returned an error: %v", err)
	}
	if err := timer.Start(); err != nil {
		t.Fatalf("Start() returned an error: %v", err)
	}
	clock.Advance(-time.Second)
	if state := timer.State(); state.Elapsed != 0 {
		t.Errorf("elapsed after clock moved backward = %v, want 0", state.Elapsed)
	}
}
