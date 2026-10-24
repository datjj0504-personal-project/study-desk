# internal/storage

Reserved for SQLite access: opening the database, running [migrations](../../migrations/README.md), and repository implementations of interfaces defined by domain packages. `OpenAppDatabase` stores the database in the user's application config directory. The SQLite-backed note repository implements the contract in `internal/notes`. The only package that talks to SQLite.
