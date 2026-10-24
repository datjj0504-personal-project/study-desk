# internal/timer

Wails-independent Pomodoro and stopwatch state logic. Time is read through a
`Clock` so transitions can be tested without sleeping. Pomodoro completion is
explicit in the state; callers advance to the next focus/break phase after
completion. Stopwatch elapsed time continues until paused or reset. Clock
deltas below zero are treated as zero.

Default Pomodoro durations are 25 minutes of focus, 5 minutes of short break,
and 15 minutes of long break. A long break follows every fourth completed focus
phase. `Configure` can replace these durations while the timer is idle and
resets the Pomodoro cycle; active or completed phases must be reset first.

The app adapter exposes Pomodoro and stopwatch snapshots and controls through
Wails. The UI can poll Pomodoro state to update the countdown; a snapshot
reports completion after the phase deadline has passed. The app monitors
snapshots independently of the selected page so phase completion alerts are
not limited to the timer screen. Stopwatch snapshots report elapsed
milliseconds while running and preserve elapsed time while paused.
