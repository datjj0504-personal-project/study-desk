# internal/storage

Reserved for SQLite access: opening the database, running [migrations](../../migrations/README.md), and repository implementations of interfaces defined by domain packages. `OpenAppDatabase` stores the database in the user's application config directory. Repositories implement notes and subjects from `internal/notes`, study sessions from `internal/study`, and settings and timer snapshots using the existing `settings` table. The only package that talks to SQLite.
