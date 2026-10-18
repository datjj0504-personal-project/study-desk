# 0001. Use Wails, Go, React and SQLite

- Status: Accepted
- Date: 2026-10-18

## Context
StudyDesk must be a Windows desktop app that works fully offline, without a separate server or Docker.

## Decision
Use Wails (Go backend + WebView2 frontend), React with TypeScript and Tailwind CSS for the UI, and a local SQLite file for persistence. The frontend talks to Go only through Wails bindings.

## Alternatives Considered
- Electron: heavier runtime, and the backend would be Node rather than Go.
- Native Go UI toolkits (Fyne, Walk): less flexible UI and smaller ecosystem than React.
- Local HTTP server plus browser UI: adds an unnecessary server and port management.

## Consequences
- Small native binary and Go-based testable logic.
- Requires the WebView2 runtime and Wails tooling on developer machines.
- SQLite driver choice (CGO vs pure Go) is still open; see [database design](../../docs/02-design/database-design.md).
