# Architecture Rules

- React frontend handles UI and user interaction only.
- Go handles application logic and desktop/backend operations.
- SQLite handles local persistent data.
- The frontend must never access SQLite directly.
- Frontend-to-Go communication uses Wails bindings (bound methods on the App struct and runtime events), wrapped in `frontend/src/services/`.
- `app.go` is a thin adapter: validate/convert input, call domain packages, return results.
- Business logic lives in `internal/*` and must not import Wails.
- Domain packages declare small interfaces for what they need (e.g. a repository, a clock); `internal/storage` implements them. Add interfaces only where they give real testability or decoupling.
- Dependencies flow one way: UI -> `app.go` -> domain -> interfaces <- storage. No circular imports.
- Schema changes go through versioned files in `migrations/`.
- Significant changes to these rules require an ADR in [decisions/](decisions/README.md).

See [architecture](../docs/02-design/architecture.md).
