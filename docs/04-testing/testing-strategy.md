# Testing Strategy

## Go unit tests

- Cover domain logic in `internal/notes`, `timer`, `study`, `settings` with table-driven tests.
- Use fakes of repository interfaces and an injected clock; no Wails or SQLite needed.
- Run: `go test ./...`
- Timer state coverage: `go test ./internal/timer -cover` and `go test -race ./internal/timer ./internal/storage`.

## Integration tests

- Test `internal/storage` against temporary SQLite files created per test, applying real [migrations](../../migrations/README.md). Verify persistence after closing and reopening, schema constraints and foreign-key behavior, migration idempotence, and rollback on migration failure.
- Place cross-package scenarios in [tests/](../../tests/README.md).

## Frontend tests

- Vitest and React Testing Library for components, hooks, and utils.
- Mock `src/services/` rather than Wails bindings.
- Checks: type check (`tsc --noEmit`), tests, and `npm run build`.

## Manual checks

- Smoke test via `wails dev` for timer behavior, window close, restart persistence, and elapsed-time behavior after system sleep, until end-to-end automation is justified. Automated timer tests use an injected clock to simulate sleep and app restarts deterministically.

Requirements being verified: [functional](../01-requirements/functional-requirements.md), [non-functional](../01-requirements/non-functional-requirements.md).
