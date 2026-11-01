# Coding Rules

- Write idiomatic Go; format all Go files with `gofmt`.
- Use TypeScript with `strict` type checking.
- Keep functions and components focused and small.
- Handle errors explicitly; never silently ignore them.
- Do not duplicate business logic; keep it in Go domain packages.
- Add tests for meaningful business logic.
- Avoid unnecessary dependencies and speculative abstractions.
- Do not add secrets, credentials, or real user data to the repository.
- Do not edit generated files (`frontend/wailsjs`, `build/bin`).

Details: [coding conventions](../docs/03-development/coding-conventions.md).
