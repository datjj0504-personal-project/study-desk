# internal/storage

Reserved for SQLite access: opening the database, running [migrations](../../migrations/README.md), and repository implementations of interfaces defined by domain packages. `OpenAppDatabase` stores the database in the user's application config directory. The SQLite-backed note repository implements note and subject operations from `internal/notes`; the settings repository persists timer preferences and the timer-state repository stores Pomodoro and stopwatch snapshots in the existing `settings` table. The only package that talks to SQLite.
