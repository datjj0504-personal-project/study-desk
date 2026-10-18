# Coding Conventions

## Go

- Idiomatic Go; format with `gofmt`; run `go vet ./...`.
- Handle every error explicitly; wrap with `fmt.Errorf("doing x: %w", err)`.
- Pass `context.Context` as the first parameter to operations that do I/O.
- Define interfaces in the consumer package and keep them small.
- Inject time (a clock) rather than calling `time.Now()` in logic under test.
- Table-driven tests in `*_test.go` next to the code.
- Domain packages under `internal/` must not import Wails.

## TypeScript / React

- `strict` mode on; avoid `any`.
- Function components and hooks; keep components small and focused.
- Call Go only through `src/services/`.
- Tailwind utility classes for styling; extract a component rather than copying class lists.
- Files: `PascalCase.tsx` for components, `camelCase.ts` for hooks and utils (`useThing.ts`).

## General

- No secrets or real user data in the repo.
- Avoid new dependencies unless clearly justified.

Agent-facing summary: [ai/coding-rules.md](../../ai/coding-rules.md).
