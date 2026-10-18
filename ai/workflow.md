# Agent Workflow

1. Read [project-context.md](project-context.md) and the relevant docs under `docs/` before modifying code.
2. Inspect existing code before proposing changes.
3. Understand the task and identify affected components.
4. Present a short implementation plan for non-trivial tasks.
5. Implement only the requested scope.
6. Add or update tests when appropriate.
7. Run relevant formatting, tests, and build checks when available.
8. Report changed files, important decisions, test results, and remaining issues.
9. Update documentation when behavior or architecture changes.
10. Ask for clarification when a decision could materially affect architecture or existing user work.

Follow [architecture-rules.md](architecture-rules.md) and [coding-rules.md](coding-rules.md). Never claim a command succeeded unless you ran it. Do not commit or push unless asked.

## Mandatory Project Progress Tracking

Before starting any task, read `ai/PROJECT_STATUS.md` if it exists. Verify the repository state before relying on its contents.

After every task, update `ai/PROJECT_STATUS.md` to reflect:
- Actual implementation progress
- Completed and remaining work
- Validation results
- Known issues and blockers
- The next recommended task

Never mark a task as completed without satisfying its acceptance criteria. If a task is partially implemented, blocked, or untested, record that explicitly.

Keep task IDs stable and preserve the completed task history. Do not replace actual implementation evidence with assumptions based on plans, directory names, or previous AI responses.

Every final task report must state the next task ID and explain what should be done next.