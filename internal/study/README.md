# internal/study

Study-session domain types, validation, and repository contracts live here.
This package must not import Wails or SQLite.

Sessions store their kind (`pomodoro`, `stopwatch`, or `manual`), UTC start and
end instants, whole-second duration, optional subject, and optional note.
Durations are kept separately from the time interval because a session can
include timer pauses. New sessions require a valid kind, non-zero times with
the end at or after the start, a non-negative duration, and a positive subject
ID when one is provided.

`SessionRepository` creates sessions, retrieves them by ID, and lists them
newest-started first. SQLite implementations belong in `internal/storage`.
