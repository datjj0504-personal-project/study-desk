# StudyDesk

A local-first Windows desktop study companion built with Go, Wails, React, TypeScript, Tailwind CSS and SQLite.

> Status: desktop shell only (sidebar and placeholder pages). No features are implemented yet. Progress: [ai/PROJECT_STATUS.md](ai/PROJECT_STATUS.md).

## Features (planned)

- Notes management
- Focus timer (Pomodoro)
- Stopwatch
- Study session tracking
- Basic study statistics
- Local data persistence (no server, no Docker)

Details: [functional requirements](docs/01-requirements/functional-requirements.md).

## Technology Stack

| Layer | Technology |
| --- | --- |
| Backend | Go |
| Desktop shell | Wails |
| Frontend | React + TypeScript |
| Styling | Tailwind CSS |
| Database | SQLite (local file) |

## Prerequisites

- Windows 10/11 with WebView2 runtime
- Go (1.21+)
- Node.js (LTS) and npm
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

Verify with `wails doctor`. Full instructions: [setup](docs/03-development/setup.md).

## Setup

```powershell
wails doctor
cd frontend; npm install; cd ..
```

The `frontend/dist` folder must exist for a plain `go build`; `wails dev` and `wails build` create it automatically.

## Development

```powershell
wails dev
```

Workflow and conventions: [development workflow](docs/03-development/development-workflow.md), [coding conventions](docs/03-development/coding-conventions.md).

## Testing

```powershell
go test ./...
cd frontend; npm run test; npm run build
```

Strategy: [testing strategy](docs/04-testing/testing-strategy.md).

## Build

```powershell
wails build
```

Output goes to `build/bin/`.

## Documentation

- [Project overview](docs/00-project-context/project-overview.md)
- [Architecture](docs/02-design/architecture.md)
- [Folder structure](docs/02-design/folder-structure.md)
- [Database design (proposal)](docs/02-design/database-design.md)
- [AI agent context](ai/README.md)
- [Project status](ai/PROJECT_STATUS.md)
