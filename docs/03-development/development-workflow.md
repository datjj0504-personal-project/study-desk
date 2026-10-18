# Development Workflow

## Cycle

1. Pick or write a task using [ai/task-template.md](../../ai/task-template.md).
2. Create a branch: `feature/<short-name>`, `fix/<short-name>`, or `docs/<short-name>`.
3. Implement the smallest scope that satisfies the task.
4. Run checks: `gofmt -l .`, `go vet ./...`, `go test ./...`, and in `frontend/`: type check, tests, build.
5. Update docs if behavior or architecture changed; add an [ADR](../../ai/decisions/README.md) for significant decisions.
6. Commit with clear imperative messages (e.g. `Add notes repository`), then open a pull request.

## Rules

- Keep `main` buildable.
- Small, focused commits and PRs.
- Never commit secrets, `.env`, database files, or generated artifacts (`build/bin`, `frontend/wailsjs`, `node_modules`).

AI agents follow [ai/workflow.md](../../ai/workflow.md). Conventions: [coding conventions](coding-conventions.md).
