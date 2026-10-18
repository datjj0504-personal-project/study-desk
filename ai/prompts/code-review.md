Review the following change in StudyDesk: <files, diff, or branch>.

First read `ai/project-context.md`, `ai/architecture-rules.md`, and `ai/coding-rules.md`.

Check for:
- Architecture violations (frontend DB access, Wails imports in domain code, circular dependencies).
- Correctness, unhandled errors, and edge cases.
- Missing or weak tests.
- Unnecessary dependencies or abstractions, and duplicated logic.
- Secrets or real user data.
- Docs that need updating.

Report findings by severity (blocking, should fix, nit) with file and line references. Do not modify code unless asked.
