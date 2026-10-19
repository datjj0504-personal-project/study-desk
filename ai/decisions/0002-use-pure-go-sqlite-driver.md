# 0002. Use a Pure-Go SQLite Driver

- Status: Accepted
- Date: 2026-10-19

## Context

StudyDesk is a Windows Wails desktop application that embeds a Go backend. The storage layer needs an SQLite driver usable through Go's `database/sql`. Production builds should not require developers or CI to install a separate C compiler and configure CGO just to build the application.

## Decision

Use [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) v1.60.1, a pure-Go SQLite driver registered with `database/sql`. The storage package exposes `OpenSQLite` and owns the driver import.

## Alternatives Considered

- `github.com/mattn/go-sqlite3`: mature CGO-based driver, but Windows builds require CGO and a compatible C toolchain. This adds build and packaging requirements that are unnecessary for the current single-user desktop application.
- Keep the driver choice open until database initialization: this would defer the Windows build decision and leave storage unable to use SQLite through `database/sql`.

## Consequences

- Windows builds do not require a separately installed C compiler or CGO toolchain for SQLite.
- The driver and its pure-Go SQLite implementation add Go dependencies and may increase compile time and binary size compared with a smaller CGO binding.
- Database location, connection configuration, migrations, and storage behavior remain to be implemented in P2-002 and later tasks.
