Write tests for: <package, file, or feature>.

First read `ai/project-context.md`, `ai/architecture-rules.md`, `ai/coding-rules.md`, and `docs/04-testing/testing-strategy.md`. Inspect the code under test.

Instructions:
- Go: table-driven tests next to the code; use fakes for interfaces and an injected clock; use a temp SQLite file only for storage integration tests.
- Frontend: Vitest and React Testing Library; mock `src/services/`.
- Cover meaningful behavior, edge cases, and error paths; avoid testing trivia.
- Do not change production code except to make it testable, and say so if you do.
- Run the tests and report the results.
