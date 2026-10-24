# internal/timer

Wails-independent Pomodoro and stopwatch state logic. Time is read through a
`Clock` so transitions can be tested without sleeping. Pomodoro completion is
explicit in the state; callers advance to the next focus/break phase after
completion. Stopwatch elapsed time continues until paused or reset. Clock
deltas below zero are treated as zero.

Default Pomodoro durations are 25 minutes of focus, 5 minutes of short break,
and 15 minutes of long break. A long break follows every fourth completed focus
phase.
