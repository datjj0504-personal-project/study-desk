package timer

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidConfiguration = errors.New("invalid timer configuration")
	ErrInvalidTransition    = errors.New("invalid timer state transition")
)

type Mode string

const (
	ModePomodoro  Mode = "pomodoro"
	ModeStopwatch Mode = "stopwatch"
)

type Phase string

const (
	PhaseFocus      Phase = "focus"
	PhaseShortBreak Phase = "short_break"
	PhaseLongBreak  Phase = "long_break"
	PhaseStopwatch  Phase = "stopwatch"
)

type Status string

const (
	StatusIdle      Status = "idle"
	StatusRunning   Status = "running"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
)

// Clock supplies the current time to timer state transitions.
type Clock interface {
	Now() time.Time
}

type Durations struct {
	Focus          time.Duration
	ShortBreak     time.Duration
	LongBreak      time.Duration
	LongBreakEvery int
}

func DefaultDurations() Durations {
	return Durations{
		Focus:          25 * time.Minute,
		ShortBreak:     5 * time.Minute,
		LongBreak:      15 * time.Minute,
		LongBreakEvery: 4,
	}
}

type State struct {
	Mode             Mode          `json:"mode"`
	Phase            Phase         `json:"phase"`
	Status           Status        `json:"status"`
	Duration         time.Duration `json:"duration"`
	Elapsed          time.Duration `json:"elapsed"`
	Remaining        time.Duration `json:"remaining"`
	CompletedFocuses int           `json:"completedFocuses"`
}

// Timer owns one Pomodoro or stopwatch timeline.
type Timer struct {
	clock     Clock
	mode      Mode
	durations Durations
	phase     Phase
	status    Status
	duration  time.Duration
	elapsed   time.Duration
	startedAt time.Time
	completed int
}

func NewPomodoro(clock Clock, durations Durations) (*Timer, error) {
	if clock == nil {
		return nil, fmt.Errorf("create Pomodoro: %w: clock is required", ErrInvalidConfiguration)
	}
	if durations.Focus <= 0 || durations.ShortBreak <= 0 || durations.LongBreak <= 0 || durations.LongBreakEvery <= 0 {
		return nil, fmt.Errorf("create Pomodoro: %w: durations and long-break interval must be positive", ErrInvalidConfiguration)
	}
	return &Timer{
		clock:     clock,
		mode:      ModePomodoro,
		durations: durations,
		phase:     PhaseFocus,
		status:    StatusIdle,
		duration:  durations.Focus,
	}, nil
}

func NewStopwatch(clock Clock) (*Timer, error) {
	if clock == nil {
		return nil, fmt.Errorf("create stopwatch: %w: clock is required", ErrInvalidConfiguration)
	}
	return &Timer{
		clock:  clock,
		mode:   ModeStopwatch,
		phase:  PhaseStopwatch,
		status: StatusIdle,
	}, nil
}

func (t *Timer) Start() error {
	t.refresh()
	if t.status != StatusIdle {
		return fmt.Errorf("start timer: %w: current status is %s", ErrInvalidTransition, t.status)
	}
	t.startedAt = t.clock.Now()
	t.status = StatusRunning
	return nil
}

func (t *Timer) Pause() error {
	t.refresh()
	if t.status != StatusRunning {
		return fmt.Errorf("pause timer: %w: current status is %s", ErrInvalidTransition, t.status)
	}
	t.elapsed += nonNegative(t.clock.Now().Sub(t.startedAt))
	t.startedAt = time.Time{}
	t.status = StatusPaused
	return nil
}

func (t *Timer) Resume() error {
	t.refresh()
	if t.status != StatusPaused {
		return fmt.Errorf("resume timer: %w: current status is %s", ErrInvalidTransition, t.status)
	}
	t.startedAt = t.clock.Now()
	t.status = StatusRunning
	return nil
}

func (t *Timer) Reset() {
	t.phase = PhaseStopwatch
	t.status = StatusIdle
	t.duration = 0
	t.elapsed = 0
	t.startedAt = time.Time{}
	t.completed = 0
	if t.mode == ModePomodoro {
		t.phase = PhaseFocus
		t.duration = t.durations.Focus
	}
}

// Advance moves a completed Pomodoro to its next phase and leaves it idle.
func (t *Timer) Advance() error {
	t.refresh()
	if t.mode != ModePomodoro || t.status != StatusCompleted {
		return fmt.Errorf("advance timer: %w: a completed Pomodoro phase is required", ErrInvalidTransition)
	}

	switch t.phase {
	case PhaseFocus:
		t.completed++
		if t.completed%t.durations.LongBreakEvery == 0 {
			t.phase = PhaseLongBreak
			t.duration = t.durations.LongBreak
		} else {
			t.phase = PhaseShortBreak
			t.duration = t.durations.ShortBreak
		}
	case PhaseShortBreak, PhaseLongBreak:
		t.phase = PhaseFocus
		t.duration = t.durations.Focus
	default:
		return fmt.Errorf("advance timer: %w: unsupported phase %s", ErrInvalidTransition, t.phase)
	}

	t.elapsed = 0
	t.startedAt = time.Time{}
	t.status = StatusIdle
	return nil
}

func (t *Timer) State() State {
	t.refresh()
	elapsed := t.elapsed
	if t.status == StatusRunning {
		elapsed += nonNegative(t.clock.Now().Sub(t.startedAt))
	}
	if t.mode == ModePomodoro && elapsed > t.duration {
		elapsed = t.duration
	}
	remaining := time.Duration(0)
	if t.mode == ModePomodoro && elapsed < t.duration {
		remaining = t.duration - elapsed
	}
	return State{
		Mode:             t.mode,
		Phase:            t.phase,
		Status:           t.status,
		Duration:         t.duration,
		Elapsed:          elapsed,
		Remaining:        remaining,
		CompletedFocuses: t.completed,
	}
}

func (t *Timer) refresh() {
	if t.status != StatusRunning || t.mode != ModePomodoro {
		return
	}
	if nonNegative(t.clock.Now().Sub(t.startedAt))+t.elapsed >= t.duration {
		t.elapsed = t.duration
		t.startedAt = time.Time{}
		t.status = StatusCompleted
	}
}

func nonNegative(duration time.Duration) time.Duration {
	if duration < 0 {
		return 0
	}
	return duration
}
